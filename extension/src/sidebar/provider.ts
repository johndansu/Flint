// WebviewViewProvider that renders Flint's sidebar panel.
import type { WebviewView, WebviewViewProvider, ExtensionContext } from 'vscode';
import type { Observation, AutoFix } from '../ipc/messages';
import type { DaemonConnection } from '../ipc/connection';
import { asObservation, asAutoFix, asDaemonStatus, asSessionState } from '../ipc/connection';

export class FlintSidebarProvider implements WebviewViewProvider {
  private view: WebviewView | null = null;
  private pendingObs: Observation[] = [];
  private conn: DaemonConnection;
  private context: ExtensionContext;

  constructor(conn: DaemonConnection, context: ExtensionContext) {
    this.conn = conn;
    this.context = context;

    conn.onMessage(msg => {
      const obs = asObservation(msg);
      if (obs) { this.pushObservation(obs); return; }

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

    // Flush any observations that arrived before the view was ready
    for (const obs of this.pendingObs) {
      this.postToView({ type: 'observation', payload: obs });
    }
    this.pendingObs = [];

    // Handle messages from the webview (dismiss button, follow-up)
    webviewView.webview.onDidReceiveMessage((msg: { type: string; observationId?: string }) => {
      if (msg.type === 'dismiss') {
        this.conn.send('dismissal', {
          observationId: msg.observationId,
          timestamp: new Date().toISOString(),
        });
      }
    });
  }

  private pushObservation(obs: Observation): void {
    if (this.view) {
      this.postToView({ type: 'observation', payload: obs });
    } else {
      this.pendingObs.push(obs);
    }
  }

  private pushAutoFix(fix: AutoFix): void {
    this.postToView({ type: 'autoFix', payload: fix });
  }

  private postToView(msg: unknown): void {
    this.view?.webview.postMessage(msg);
  }

  // ---------------------------------------------------------------------------
  // HTML shell — the actual UI is in the webview script below.
  // In production this would be a compiled React/Svelte bundle.
  // ---------------------------------------------------------------------------
  private buildHTML(): string {
    return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline';">
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: var(--vscode-font-family); font-size: 13px; color: var(--vscode-foreground); background: var(--vscode-sideBar-background); padding: 8px; }
  #empty { padding: 20px 8px; color: var(--vscode-descriptionForeground); line-height: 1.5; }
  .card { border: 1px solid var(--vscode-panel-border); border-radius: 4px; padding: 10px 12px; margin-bottom: 8px; }
  .card-kind { font-size: 10px; font-weight: 600; letter-spacing: 0.05em; text-transform: uppercase; opacity: 0.6; margin-bottom: 4px; }
  .card-text { line-height: 1.5; }
  .card-why { font-size: 11px; opacity: 0.7; margin-top: 6px; }
  .card-actions { margin-top: 8px; display: flex; gap: 6px; }
  button { background: var(--vscode-button-secondaryBackground); color: var(--vscode-button-secondaryForeground); border: none; border-radius: 3px; padding: 3px 8px; cursor: pointer; font-size: 12px; }
  button:hover { background: var(--vscode-button-secondaryHoverBackground); }
  .fix-card { border-color: var(--vscode-gitDecoration-addedResourceForeground); }
  .fix-label { color: var(--vscode-gitDecoration-addedResourceForeground); font-weight: 600; font-size: 11px; margin-bottom: 4px; }
</style>
</head>
<body>
<div id="empty">Flint is watching.<br>It will speak up when it has something worth saying.</div>
<div id="cards"></div>
<script>
const vscode = acquireVsCodeApi();
const cards  = document.getElementById('cards');
const empty  = document.getElementById('empty');

function showCards() { empty.style.display = 'none'; }

window.addEventListener('message', e => {
  const { type, payload } = e.data;
  if (type === 'observation') renderObs(payload);
  if (type === 'autoFix')     renderFix(payload);
});

function renderObs(obs) {
  showCards();
  const div = document.createElement('div');
  div.className = 'card';
  div.dataset.id = obs.id;
  const why = obs.explainWhy ? '<div class="card-why">' + esc(obs.explainWhy) + '</div>' : '';
  div.innerHTML =
    '<div class="card-kind">' + esc(obs.kind) + ' · ' + esc(obs.category) + '</div>' +
    '<div class="card-text">' + esc(obs.text) + '</div>' +
    why +
    '<div class="card-actions">' +
    '<button onclick="dismiss(\'' + obs.id + '\')">Dismiss</button>' +
    '</div>';
  cards.prepend(div);
}

function renderFix(fix) {
  showCards();
  const div = document.createElement('div');
  div.className = 'card fix-card';
  div.innerHTML =
    '<div class="fix-label">Auto-fix staged</div>' +
    '<div class="card-text">' + esc(fix.description) + '</div>';
  cards.prepend(div);
}

function dismiss(id) {
  vscode.postMessage({ type: 'dismiss', observationId: id });
  const card = document.querySelector('[data-id="' + id + '"]');
  if (card) card.remove();
  if (!cards.firstChild) empty.style.display = '';
}

function esc(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}
</script>
</body>
</html>`;
  }
}
