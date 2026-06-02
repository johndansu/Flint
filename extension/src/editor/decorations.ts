// Manages inline editor decorations for Flint observations.
// When an observation has a filePath + lineNumber, a subtle after-line
// annotation is painted in the open editor so the developer sees it in context.
import type { ExtensionContext, TextEditor, TextEditorDecorationType, DecorationOptions } from 'vscode';
import type { Observation } from '../ipc/messages';

const vscodeApi = (() => {
  try { return require('vscode'); } catch { return null; }
})();

// Max characters of observation text shown inline — keeps it unobtrusive.
const INLINE_MAX_CHARS = 72;

// One decoration type per observation kind, created once and reused.
interface KindStyle {
  color: string;
  prefix: string;
}

const KIND_STYLES: Record<string, KindStyle> = {
  technical: { color: 'editorCodeLens.foreground', prefix: '⊙' },
  human:     { color: 'editorWarning.foreground',   prefix: '◈' },
  win:       { color: 'terminal.ansiGreen',          prefix: '✓' },
  question:  { color: 'editorInfo.foreground',       prefix: '?' },
  taste:     { color: 'editorCodeLens.foreground',   prefix: '~' },
};

interface TrackedObs {
  obs: Observation;
  filePath: string;
  lineNumber: number; // 1-based from daemon; 0-based in VS Code range
}

export class FlintDecorationManager {
  // obs id → tracked record
  private tracked = new Map<string, TrackedObs>();

  // kind → decoration type (created lazily)
  private types = new Map<string, TextEditorDecorationType>();

  private disposables: { dispose(): void }[] = [];

  constructor(context: ExtensionContext) {
    if (!vscodeApi) return;

    // Re-apply decorations whenever any editor becomes visible
    this.disposables.push(
      vscodeApi.window.onDidChangeVisibleTextEditors((editors: TextEditor[]) => {
        for (const editor of editors) {
          this.applyToEditor(editor);
        }
      }),
    );

    context.subscriptions.push({ dispose: () => this.dispose() });
  }

  addObservation(obs: Observation): void {
    if (!vscodeApi || !obs.filePath || obs.lineNumber == null) return;

    this.tracked.set(obs.id, {
      obs,
      filePath: obs.filePath,
      lineNumber: obs.lineNumber,
    });

    // Apply immediately to any editor that already has this file open
    for (const editor of vscodeApi.window.visibleTextEditors as TextEditor[]) {
      if (editor.document.uri.fsPath === obs.filePath) {
        this.applyToEditor(editor);
      }
    }
  }

  removeObservation(obsId: string): void {
    if (!this.tracked.has(obsId)) return;
    this.tracked.delete(obsId);
    this.reapplyAll();
  }

  private applyToEditor(editor: TextEditor): void {
    if (!vscodeApi) return;

    // Group tracked observations by kind for this file
    const byKind = new Map<string, DecorationOptions[]>();

    for (const tracked of this.tracked.values()) {
      if (tracked.filePath !== editor.document.uri.fsPath) continue;

      const line = tracked.lineNumber - 1; // daemon is 1-based; VS Code is 0-based
      if (line < 0 || line >= editor.document.lineCount) continue;

      const lineRange = editor.document.lineAt(line).range;
      const kind = tracked.obs.kind ?? 'technical';
      const style = KIND_STYLES[kind] ?? KIND_STYLES.technical;
      const truncated = truncate(tracked.obs.text, INLINE_MAX_CHARS);

      const decoration: DecorationOptions = {
        range: lineRange,
        renderOptions: {
          after: {
            contentText: `  ${style.prefix} ${truncated}`,
            color: new vscodeApi.ThemeColor(style.color),
            fontStyle: 'italic',
            margin: '0 0 0 16px',
          },
        },
      };

      if (!byKind.has(kind)) byKind.set(kind, []);
      byKind.get(kind)!.push(decoration);
    }

    // Apply each kind's decorations using its shared decoration type
    for (const [kind, decorations] of byKind) {
      editor.setDecorations(this.getType(kind), decorations);
    }

    // Clear any kinds that have no decorations for this editor
    for (const [kind, type] of this.types) {
      if (!byKind.has(kind)) {
        editor.setDecorations(type, []);
      }
    }
  }

  private reapplyAll(): void {
    if (!vscodeApi) return;
    for (const editor of vscodeApi.window.visibleTextEditors as TextEditor[]) {
      this.applyToEditor(editor);
    }
  }

  private getType(kind: string): TextEditorDecorationType {
    if (!this.types.has(kind)) {
      const type = vscodeApi.window.createTextEditorDecorationType({});
      this.types.set(kind, type);
      this.disposables.push(type);
    }
    return this.types.get(kind)!;
  }

  dispose(): void {
    for (const d of this.disposables) d.dispose();
    this.disposables = [];
    this.types.clear();
    this.tracked.clear();
  }
}

function truncate(text: string, max: number): string {
  if (text.length <= max) return text;
  return text.slice(0, max - 1) + '…';
}
