// Follow-up streaming panel — opened when the developer taps "Tell me more".
import * as https from 'https';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import type { ExtensionContext } from 'vscode';
import type { Observation } from '../ipc/messages';

function readFlintConfig(): { apiKey: string; model: string } {
  try {
    const cfgPath = path.join(os.homedir(), '.flint', 'config.json');
    const cfg = JSON.parse(fs.readFileSync(cfgPath, 'utf8'));
    return {
      apiKey: cfg.api_key ?? '',
      model: cfg.model || 'claude-sonnet-4-6',
    };
  } catch {
    return { apiKey: '', model: 'claude-sonnet-4-6' };
  }
}

export function openFollowupPanel(
  context: ExtensionContext,
  obs: Observation,
  vscodeApi: typeof import('vscode'),
): void {
  const panel = vscodeApi.window.createWebviewPanel(
    'flint.followup',
    `Flint: ${obs.category}`,
    vscodeApi.ViewColumn.Beside,
    { enableScripts: true },
  );
  panel.webview.html = buildHTML(obs);
  context.subscriptions.push(panel);

  const { apiKey, model } = readFlintConfig();
  if (!apiKey) {
    panel.webview.postMessage({ type: 'error', text: 'No API key — run flint init to configure.' });
    return;
  }

  stream(panel, obs, apiKey, model);
}

function stream(
  panel: import('vscode').WebviewPanel,
  obs: Observation,
  apiKey: string,
  model: string,
): void {
  const system =
    `You are Flint, a senior developer. The developer tapped "Tell me more" on one of your observations. ` +
    `Expand on it in 4–6 sentences. Be specific and actionable. Reference the file or line if provided. ` +
    `No preamble. No "Certainly" or "Of course". Start directly with the insight.`;

  let userMsg = `Observation (${obs.kind}/${obs.category}): ${obs.text}`;
  if (obs.filePath) userMsg += `\nFile: ${obs.filePath}${obs.lineNumber ? ':' + obs.lineNumber : ''}`;
  if (obs.explainWhy) userMsg += `\nContext: ${obs.explainWhy}`;

  const body = JSON.stringify({
    model,
    max_tokens: 512,
    stream: true,
    system,
    messages: [{ role: 'user', content: userMsg }],
  });

  const req = https.request(
    {
      hostname: 'api.anthropic.com',
      path: '/v1/messages',
      method: 'POST',
      headers: {
        'content-type': 'application/json',
        'x-api-key': apiKey,
        'anthropic-version': '2023-06-01',
      },
    },
    res => {
      let buf = '';
      res.on('data', (chunk: Buffer) => {
        buf += chunk.toString();
        const lines = buf.split('\n');
        buf = lines.pop() ?? '';
        for (const line of lines) {
          if (!line.startsWith('data: ')) continue;
          const data = line.slice(6).trim();
          if (data === '[DONE]') { panel.webview.postMessage({ type: 'done' }); return; }
          try {
            const evt = JSON.parse(data);
            if (evt.type === 'content_block_delta' && evt.delta?.type === 'text_delta') {
              panel.webview.postMessage({ type: 'delta', text: evt.delta.text });
            }
          } catch { /* skip malformed SSE lines */ }
        }
      });
      res.on('end', () => panel.webview.postMessage({ type: 'done' }));
      res.on('error', (err: Error) => panel.webview.postMessage({ type: 'error', text: err.message }));
    },
  );

  req.on('error', (err: Error) => panel.webview.postMessage({ type: 'error', text: err.message }));
  req.write(body);
  req.end();
}

function buildHTML(obs: Observation): string {
  const locLine = obs.filePath
    ? `<div class="loc">${esc(obs.filePath)}${obs.lineNumber ? ':' + obs.lineNumber : ''}</div>`
    : '';

  return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline';">
<style>
  body { font-family:var(--vscode-font-family); font-size:14px; color:var(--vscode-foreground); background:var(--vscode-editor-background); padding:24px 32px; max-width:680px; margin:0 auto; line-height:1.6; }
  h2 { font-size:11px; font-weight:700; text-transform:uppercase; letter-spacing:.08em; opacity:.5; margin-bottom:4px; }
  .obs { font-size:15px; margin-bottom:4px; }
  .loc { font-size:11px; opacity:.4; font-family:var(--vscode-editor-font-family); margin-bottom:20px; }
  hr { border:none; border-top:1px solid var(--vscode-panel-border); margin:16px 0; }
  #response { white-space:pre-wrap; line-height:1.7; }
  .cursor { display:inline-block; width:2px; height:1em; background:var(--vscode-foreground); animation:blink .8s step-end infinite; vertical-align:text-bottom; margin-left:1px; }
  @keyframes blink { 0%,100%{opacity:1} 50%{opacity:0} }
  .error { color:var(--vscode-editorError-foreground); }
</style>
</head>
<body>
<h2>${esc(obs.kind)} · ${esc(obs.category)}</h2>
<div class="obs">${esc(obs.text)}</div>
${locLine}
<hr>
<div id="response"><span id="cur" class="cursor"></span></div>
<script>
const el  = document.getElementById('response');
let   cur = document.getElementById('cur');
window.addEventListener('message', e => {
  const { type, text } = e.data;
  if (type === 'delta') {
    cur?.remove(); cur = null;
    el.appendChild(document.createTextNode(text));
  } else if (type === 'done') {
    cur?.remove(); cur = null;
  } else if (type === 'error') {
    cur?.remove(); cur = null;
    const s = document.createElement('span');
    s.className = 'error';
    s.textContent = 'Error: ' + text;
    el.appendChild(s);
  }
});
</script>
</body>
</html>`;
}

function esc(s: string | undefined): string {
  return String(s ?? '')
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}
