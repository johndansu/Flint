// CodeLens provider for Flint observations.
// Shows a clickable lens above each line that has an observation, allowing
// the developer to jump to the sidebar card without leaving the editor.
import type {
  CodeLensProvider, CodeLens, TextDocument, CancellationToken, Event, EventEmitter,
} from 'vscode';
import type { Observation } from '../ipc/messages';

const vscodeApi = (() => {
  try { return require('vscode'); } catch { return null; }
})();

const LENS_MAX_CHARS = 80;

export class FlintCodeLensProvider implements CodeLensProvider {
  // filePath → observations on that file
  private byFile = new Map<string, Map<string, Observation>>(); // filePath → obsId → obs

  private _emitter: EventEmitter<void> | null = null;
  readonly onDidChangeCodeLenses: Event<void>;

  constructor() {
    if (vscodeApi) {
      // Cast needed because vscodeApi is untyped at runtime (require pattern)
      this._emitter = new (vscodeApi.EventEmitter as new() => EventEmitter<void>)();
    }
    // EventEmitter.event is the Event<void> — fall back to a no-op if vscode unavailable
    this.onDidChangeCodeLenses = this._emitter?.event ?? (() => ({ dispose: () => {} }));
  }

  addObservation(obs: Observation): void {
    if (!obs.filePath || obs.lineNumber == null) return;

    let fileMap = this.byFile.get(obs.filePath);
    if (!fileMap) {
      fileMap = new Map();
      this.byFile.set(obs.filePath, fileMap);
    }
    fileMap.set(obs.id, obs);
    this._emitter?.fire();
  }

  removeObservation(obsId: string): void {
    for (const [filePath, fileMap] of this.byFile) {
      if (fileMap.delete(obsId)) {
        if (fileMap.size === 0) this.byFile.delete(filePath);
        this._emitter?.fire();
        return;
      }
    }
  }

  provideCodeLenses(document: TextDocument, _token: CancellationToken): CodeLens[] {
    if (!vscodeApi) return [];

    const fileMap = this.byFile.get(document.uri.fsPath);
    if (!fileMap || fileMap.size === 0) return [];

    const lenses: CodeLens[] = [];

    for (const obs of fileMap.values()) {
      const line = (obs.lineNumber ?? 1) - 1; // 1-based → 0-based
      if (line < 0 || line >= document.lineCount) continue;

      const range = document.lineAt(line).range;
      const label = formatLensLabel(obs);

      // The command opens the file at the right line AND reveals the sidebar
      lenses.push(new vscodeApi.CodeLens(range, {
        title: label,
        command: 'flint.revealObservation',
        arguments: [obs.id, obs.filePath, obs.lineNumber],
      }));
    }

    return lenses;
  }

  // resolveCodeLens is not needed — command is already filled in provideCodeLenses
}

function formatLensLabel(obs: Observation): string {
  const kindIcon: Record<string, string> = {
    technical: '⊙',
    human:     '◈',
    win:       '✓',
    question:  '?',
    taste:     '~',
  };
  const icon = kindIcon[obs.kind ?? 'technical'] ?? '⊙';
  const cat  = obs.category ? ` [${obs.category}]` : '';
  const text = truncate(obs.text, LENS_MAX_CHARS);
  return `${icon} Flint${cat}: ${text}`;
}

function truncate(text: string, max: number): string {
  if (text.length <= max) return text;
  return text.slice(0, max - 1) + '…';
}
