// WebviewViewProvider that renders Flint's sidebar panel.
import type { WebviewView, WebviewViewProvider, ExtensionContext } from 'vscode';
import type { Observation, AutoFix } from '../ipc/messages';
import type { DaemonConnection } from '../ipc/connection';
import { asAutoFix, asDaemonStatus, asSessionState } from '../ipc/connection';

export class FlintSidebarProvider implements WebviewViewProvider {
  private view: WebviewView | null = null;
  private pendingObs: Observation[] = [];
  private conn: DaemonConnection;
  private context: ExtensionContext;

  // Callbacks wired in by extension.ts
  onDismiss?: (obsId: string) => void;
  onGotoFile?: (filePath: string, lineNumber: number) => void;
  onTellMeMore?: (obs: Observation) => void;
  onRunTool?: (command: string) => void;

  constructor(conn: DaemonConnection, context: ExtensionContext) {
    this.conn = conn;
    this.context = context;

    // Reflect connection state changes into the webview immediately
    conn.onStateChange = state => {
      this.postToView({ type: 'connectionState', payload: { connected: state === 'connected' } });
    };

    conn.onMessage(msg => {
      const fix = asAutoFix(msg);
      if (fix) { this.pushAutoFix(fix); return; }

      const status = asDaemonStatus(msg);
      if (status) { this.postToView({ type: 'daemonStatus', payload: status }); return; }

      const state = asSessionState(msg);
      if (state) { this.postToView({ type: 'sessionState', payload: state }); return; }
    });
  }

  resolveWebviewView(webviewView: WebviewView): void {
    this.view = webviewView;
    webviewView.webview.options = { enableScripts: true };
    webviewView.webview.html = this.buildHTML();

    // Seed the current connection state so the webview starts in the right mode
    const connected = this.conn.connectionState === 'connected';
    this.postToView({ type: 'connectionState', payload: { connected } });

    // Flush observations that arrived before the view was ready
    for (const obs of this.pendingObs) {
      this.postToView({ type: 'observation', payload: obs });
    }
    this.pendingObs = [];

    webviewView.webview.onDidReceiveMessage((msg: {
      type: string;
      observationId?: string;
      obs?: Observation;
      filePath?: string;
      lineNumber?: number;
      command?: string;
    }) => {
      switch (msg.type) {
        case 'dismiss':
          if (msg.observationId) {
            this.conn.send('dismissal', {
              observationId: msg.observationId,
              timestamp: new Date().toISOString(),
            });
            this.onDismiss?.(msg.observationId);
          }
          break;

        case 'gotoFile':
          if (msg.filePath) {
            this.onGotoFile?.(msg.filePath, msg.lineNumber ?? 1);
          }
          break;

        case 'tellMeMore':
          if (msg.obs) {
            this.onTellMeMore?.(msg.obs);
          }
          break;

        case 'runTool':
          if (msg.command) {
            this.onRunTool?.(msg.command);
          }
          break;
      }
    });
  }

  pushObservation(obs: Observation): void {
    if (this.view) {
      this.postToView({ type: 'observation', payload: obs });
    } else {
      this.pendingObs.push(obs);
    }
  }

  highlightCard(obsId: string): void {
    this.postToView({ type: 'highlightCard', obsId });
  }

  private pushAutoFix(fix: AutoFix): void {
    this.postToView({ type: 'autoFix', payload: fix });
  }

  private postToView(msg: unknown): void {
    this.view?.webview.postMessage(msg);
  }

  // ---------------------------------------------------------------------------
  // HTML shell
  // ---------------------------------------------------------------------------
  private buildHTML(): string {
    return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline';">
<style>
  * { box-sizing:border-box; margin:0; padding:0; }
  body { font-family:var(--vscode-font-family); font-size:13px; color:var(--vscode-foreground); background:var(--vscode-sideBar-background); padding:8px; }

  /* ── Cards mode ── */
  #empty { padding:20px 8px; color:var(--vscode-descriptionForeground); line-height:1.5; }
  .card { border:1px solid var(--vscode-panel-border); border-radius:4px; padding:10px 12px; margin-bottom:8px; transition:opacity .6s, background .3s; }
  .card.highlight { background:var(--vscode-editor-findMatchHighlightBackground); }
  .card-meta { font-size:10px; font-weight:600; letter-spacing:.05em; text-transform:uppercase; opacity:.6; margin-bottom:4px; display:flex; justify-content:space-between; align-items:center; }
  .card-loc { font-size:10px; opacity:.5; font-weight:normal; text-transform:none; letter-spacing:0; }
  .card-text { line-height:1.5; }
  .card-why { font-size:11px; opacity:.7; margin-top:6px; font-style:italic; }
  .card-actions { margin-top:8px; display:flex; gap:6px; flex-wrap:wrap; }
  button { background:var(--vscode-button-secondaryBackground); color:var(--vscode-button-secondaryForeground); border:none; border-radius:3px; padding:3px 8px; cursor:pointer; font-size:12px; }
  button:hover { background:var(--vscode-button-secondaryHoverBackground); }
  button.primary { background:var(--vscode-button-background); color:var(--vscode-button-foreground); }
  button.primary:hover { background:var(--vscode-button-hoverBackground); }
  .fix-card { border-color:var(--vscode-gitDecoration-addedResourceForeground); }
  .fix-label { color:var(--vscode-gitDecoration-addedResourceForeground); font-weight:600; font-size:11px; margin-bottom:4px; }
  .kind-human     { border-left:3px solid var(--vscode-editorWarning-foreground); }
  .kind-win       { border-left:3px solid var(--vscode-terminal-ansiGreen,#4ec9b0); }
  .kind-technical { border-left:3px solid var(--vscode-editorInfo-foreground); }
  .kind-question  { border-left:3px solid var(--vscode-editorInfo-foreground); }

  /* ── Standalone mode ── */
  #standalone { display:none; padding:12px 8px; }
  .standalone-header { font-size:11px; font-weight:700; text-transform:uppercase; letter-spacing:.06em; opacity:.5; margin-bottom:4px; }
  .standalone-sub { font-size:12px; opacity:.6; margin-bottom:16px; line-height:1.5; }
  .tool-grid { display:grid; grid-template-columns:1fr 1fr; gap:6px; margin-bottom:16px; }
  .tool-btn { padding:8px 6px; font-size:12px; text-align:center; cursor:pointer; border:1px solid var(--vscode-panel-border); border-radius:4px; background:transparent; color:var(--vscode-foreground); transition:background .15s; }
  .tool-btn:hover { background:var(--vscode-list-hoverBackground); }
  .tool-btn.start { grid-column:1/-1; background:var(--vscode-button-background); color:var(--vscode-button-foreground); border:none; }
  .tool-btn.start:hover { background:var(--vscode-button-hoverBackground); }
</style>
</head>
<body>

<!-- ── Cards mode (daemon connected) ── -->
<div id="cards-mode">
  <div id="empty">Flint is watching.<br>It will speak up when it has something worth saying.</div>
  <div id="cards"></div>
</div>

<!-- ── Standalone mode (daemon not running) ── -->
<div id="standalone">
  <div class="standalone-header">Daemon not running</div>
  <div class="standalone-sub">Start the daemon for passive observations, or run tools manually below.</div>
  <div class="tool-grid">
    <button class="tool-btn start" onclick="startDaemon()">Start daemon</button>
    <button class="tool-btn" onclick="runTool('scan')">Scan codebase</button>
    <button class="tool-btn" onclick="runTool('status')">Status</button>
    <button class="tool-btn" onclick="runTool('diagnose')">Diagnose</button>
    <button class="tool-btn" onclick="runTool('memory obs')">Observations</button>
    <button class="tool-btn" onclick="runTool('errors')">Errors</button>
  </div>
</div>

<script>
const vscode    = acquireVsCodeApi();
const cardsMode = document.getElementById('cards-mode');
const standalone= document.getElementById('standalone');
const cards     = document.getElementById('cards');
const empty     = document.getElementById('empty');

// Keyed obs store for "Tell me more"
const obsMap = new Map();

// ── Mode switching ──────────────────────────────────────────────────────────

function setConnected(connected) {
  cardsMode.style.display  = connected ? '' : 'none';
  standalone.style.display = connected ? 'none' : '';
}

// ── Message router ──────────────────────────────────────────────────────────

window.addEventListener('message', e => {
  const { type, payload, obsId } = e.data;
  switch (type) {
    case 'connectionState': setConnected(payload.connected); break;
    case 'observation':     renderObs(payload); break;
    case 'autoFix':         renderFix(payload); break;
    case 'highlightCard':   highlightCard(obsId); break;
    case 'daemonStatus':    setConnected(true); break;
  }
});

// ── Observation cards ───────────────────────────────────────────────────────

function renderObs(obs) {
  obsMap.set(obs.id, obs);
  empty.style.display = 'none';

  let locBadge = '';
  if (obs.filePath) {
    const parts = obs.filePath.replace(/\\\\/g, '/').split('/');
    const name  = parts[parts.length - 1];
    const loc   = obs.lineNumber ? name + ':' + obs.lineNumber : name;
    locBadge = '<span class="card-loc">' + esc(loc) + '</span>';
  }

  const why = obs.explainWhy
    ? '<div class="card-why">' + esc(obs.explainWhy) + '</div>'
    : '';

  const gotoBtn = obs.filePath
    ? '<button class="primary" onclick="gotoFile(' +
        JSON.stringify(obs.filePath) + ',' + (obs.lineNumber || 1) +
      ')">Go to file</button>'
    : '';

  const div = document.createElement('div');
  div.className = 'card kind-' + (obs.kind || 'technical');
  div.dataset.id = obs.id;
  div.innerHTML =
    '<div class="card-meta">' +
      '<span>' + esc(obs.kind) + ' · ' + esc(obs.category) + '</span>' +
      locBadge +
    '</div>' +
    '<div class="card-text">' + esc(obs.text) + '</div>' +
    why +
    '<div class="card-actions">' +
      gotoBtn +
      '<button onclick="tellMeMore(\\''+obs.id+'\\')">Tell me more</button>' +
      '<button onclick="dismiss(\\''+obs.id+'\\')">Dismiss</button>' +
    '</div>';

  cards.prepend(div);

  // Auto-fade win and human cards after 8 seconds
  if (obs.kind === 'win' || obs.kind === 'human') {
    setTimeout(() => fadeOut(div), 8000);
  }
}

function renderFix(fix) {
  empty.style.display = 'none';
  const div = document.createElement('div');
  div.className = 'card fix-card';
  div.innerHTML =
    '<div class="fix-label">Auto-fix staged</div>' +
    '<div class="card-text">' + esc(fix.description) + '</div>';
  cards.prepend(div);
}

// ── Actions ─────────────────────────────────────────────────────────────────

function dismiss(id) {
  vscode.postMessage({ type: 'dismiss', observationId: id });
  const card = document.querySelector('[data-id="' + id + '"]');
  if (card) card.remove();
  if (!cards.firstChild) empty.style.display = '';
}

function tellMeMore(id) {
  const obs = obsMap.get(id);
  if (obs) vscode.postMessage({ type: 'tellMeMore', obs });
}

function gotoFile(filePath, lineNumber) {
  vscode.postMessage({ type: 'gotoFile', filePath, lineNumber });
}

function startDaemon() {
  vscode.postMessage({ type: 'runTool', command: 'start' });
}

function runTool(command) {
  vscode.postMessage({ type: 'runTool', command });
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function fadeOut(el) {
  el.style.opacity = '0';
  setTimeout(() => {
    el.remove();
    if (!cards.firstChild) empty.style.display = '';
  }, 600);
}

function highlightCard(obsId) {
  const card = document.querySelector('[data-id="' + obsId + '"]');
  if (!card) return;
  card.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
  card.classList.add('highlight');
  setTimeout(() => card.classList.remove('highlight'), 1500);
}

function esc(s) {
  return String(s ?? '')
    .replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
</script>
</body>
</html>`;
  }
}
