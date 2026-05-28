// Manages the Unix domain socket connection to flintd.
import * as net from 'net';
import * as path from 'path';
import * as os from 'os';
import * as crypto from 'crypto';
import { Observation, SessionState, DaemonStatus, AutoFix, TripwireFired, Ack } from './messages';

export type ConnectionState = 'connected' | 'reconnecting' | 'disconnected';

export interface IncomingMessage {
  type: string;
  payload: unknown;
}

export type MessageHandler = (msg: IncomingMessage) => void;

export class DaemonConnection {
  private socket: net.Socket | null = null;
  private state: ConnectionState = 'disconnected';
  private handlers: MessageHandler[] = [];
  private buffer = '';
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private readonly socketPath: string;

  constructor(workspaceRoot: string) {
    this.socketPath = resolveSocketPath(workspaceRoot);
  }

  get connectionState(): ConnectionState { return this.state; }

  connect(): void {
    if (this.socket) return;
    this.tryConnect();
  }

  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.socket?.destroy();
    this.socket = null;
    this.setState('disconnected');
  }

  onMessage(handler: MessageHandler): () => void {
    this.handlers.push(handler);
    return () => { this.handlers = this.handlers.filter(h => h !== handler); };
  }

  send(type: string, payload: unknown): void {
    if (!this.socket || this.state !== 'connected') return;
    const line = JSON.stringify({ type, payload }) + '\n';
    this.socket.write(line);
  }

  private tryConnect(): void {
    const sock = net.createConnection(this.socketPath);
    this.socket = sock;

    sock.on('connect', () => {
      this.setState('connected');
      this.buffer = '';
      // Subscribe immediately on connect
      this.send('subscribe', { clientType: 'extension' });
    });

    sock.on('data', (chunk: Buffer) => {
      this.buffer += chunk.toString('utf8');
      const lines = this.buffer.split('\n');
      this.buffer = lines.pop() ?? '';
      for (const line of lines) {
        if (!line.trim()) continue;
        try {
          const msg = JSON.parse(line) as IncomingMessage;
          this.dispatch(msg);
        } catch { /* ignore malformed lines */ }
      }
    });

    sock.on('close', () => {
      this.socket = null;
      this.setState('reconnecting');
      this.scheduleReconnect();
    });

    sock.on('error', () => {
      // close event follows; setState handled there
    });
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return;
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      if (this.state !== 'disconnected') {
        this.tryConnect();
      }
    }, 3000);
  }

  private dispatch(msg: IncomingMessage): void {
    for (const h of this.handlers) {
      try { h(msg); } catch { /* isolate handler errors */ }
    }
  }

  private setState(s: ConnectionState): void {
    this.state = s;
  }
}

// ---------------------------------------------------------------------------
// Typed helpers — narrow IncomingMessage.payload by type discriminator
// ---------------------------------------------------------------------------

export function asObservation(msg: IncomingMessage): Observation | null {
  return msg.type === 'observation' ? msg.payload as Observation : null;
}

export function asSessionState(msg: IncomingMessage): SessionState | null {
  return msg.type === 'sessionState' ? msg.payload as SessionState : null;
}

export function asDaemonStatus(msg: IncomingMessage): DaemonStatus | null {
  return msg.type === 'daemonStatus' ? msg.payload as DaemonStatus : null;
}

export function asAutoFix(msg: IncomingMessage): AutoFix | null {
  return msg.type === 'autoFix' ? msg.payload as AutoFix : null;
}

export function asTripwireFired(msg: IncomingMessage): TripwireFired | null {
  return msg.type === 'tripwireFired' ? msg.payload as TripwireFired : null;
}

export function asAck(msg: IncomingMessage): Ack | null {
  return msg.type === 'ack' ? msg.payload as Ack : null;
}

// ---------------------------------------------------------------------------
// Socket path resolution — mirrors Go's ipc.SocketPath
// ---------------------------------------------------------------------------

function resolveSocketPath(workspaceRoot: string): string {
  const hash = crypto.createHash('md5').update(workspaceRoot).digest('hex').slice(0, 8);
  const flintDir = path.join(os.homedir(), '.flint');
  return path.join(flintDir, `daemon-${hash}.sock`);
}
