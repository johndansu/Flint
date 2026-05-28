// Manages the Flint status bar item in VS Code.
import type { StatusBarItem, StatusBarAlignment } from 'vscode';
import type { ConnectionState } from '../ipc/connection';

// We import vscode types but reference the actual module at runtime
// so the extension can load without crashing in test environments.
const vscodeApi = (() => {
  try { return require('vscode'); } catch { return null; }
})();

export class FlintStatusBar {
  private item: StatusBarItem | null = null;

  constructor() {
    if (!vscodeApi) return;
    this.item = vscodeApi.window.createStatusBarItem(
      vscodeApi.StatusBarAlignment.Right as StatusBarAlignment,
      100,
    );
    if (this.item) {
      this.item.command = 'flint.showStatus';
      this.item.show();
    }
    this.setIdle();
  }

  setIdle(): void {
    this.update('$(flame) Flint', 'Flint — idle');
  }

  setConnected(awareness: string): void {
    const icon = awarenessIcon(awareness);
    this.update(`${icon} Flint`, `Flint — ${awareness} mode`);
  }

  setReconnecting(): void {
    this.update('$(sync~spin) Flint', 'Flint — reconnecting to daemon...');
  }

  setDisconnected(): void {
    this.update('$(circle-slash) Flint', 'Flint daemon not running — click to start');
  }

  setObservation(count: number): void {
    const label = count === 1 ? '1 observation' : `${count} observations`;
    this.update(`$(flame) Flint  ${count}`, `Flint — ${label} pending`);
  }

  onConnectionChange(state: ConnectionState, awareness?: string): void {
    switch (state) {
      case 'connected':    this.setConnected(awareness ?? 'flame'); break;
      case 'reconnecting': this.setReconnecting(); break;
      case 'disconnected': this.setDisconnected(); break;
    }
  }

  dispose(): void {
    this.item?.dispose();
  }

  private update(text: string, tooltip: string): void {
    if (!this.item) return;
    this.item.text = text;
    this.item.tooltip = tooltip;
  }
}

function awarenessIcon(awareness: string): string {
  switch (awareness) {
    case 'forge': return '$(flame)';
    case 'flame': return '$(flame)';
    default:      return '$(circle-outline)';
  }
}
