// Tool output panel — runs a flint CLI command and streams stdout/stderr into a webview.
import * as child_process from 'child_process';
import type { ExecException } from 'child_process';
import type { ExtensionContext } from 'vscode';

export function openToolOutputPanel(
  context: ExtensionContext,
  command: string,
  vscodeApi: typeof import('vscode'),
): void {
  const panel = vscodeApi.window.createWebviewPanel(
    'flint.toolOutput',
    `flint ${command}`,
    vscodeApi.ViewColumn.Beside,
    { enableScripts: true },
  );
  panel.webview.html = buildHTML(command);
  context.subscriptions.push(panel);

  child_process.exec(
    `flint ${command}`,
    { timeout: 30_000 },
    (err: ExecException | null, stdout: string, stderr: string) => {
      const raw = (stdout + stderr).trim();
      panel.webview.postMessage({
        type: 'output',
        text: raw || '(no output)',
        ok: !err,
      });
    },
  );
}

function buildHTML(command: string): string {
  return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline';">
<style>
  body { font-family:var(--vscode-font-family); font-size:13px; color:var(--vscode-foreground); background:var(--vscode-editor-background); padding:20px; }
  h2 { font-size:11px; font-weight:700; text-transform:uppercase; letter-spacing:.08em; opacity:.5; margin-bottom:12px; }
  pre { white-space:pre-wrap; font-family:var(--vscode-editor-font-family); font-size:12px; line-height:1.5; background:var(--vscode-textCodeBlock-background,rgba(0,0,0,.1)); padding:12px; border-radius:4px; }
  .loading { opacity:.5; }
  .err { color:var(--vscode-editorError-foreground); }
</style>
</head>
<body>
<h2>flint ${esc(command)}</h2>
<pre id="out" class="loading">Running…</pre>
<script>
const out = document.getElementById('out');
window.addEventListener('message', e => {
  const { type, text, ok } = e.data;
  if (type === 'output') {
    out.className = ok ? '' : 'err';
    out.textContent = text;
  }
});
</script>
</body>
</html>`;
}

function esc(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}
