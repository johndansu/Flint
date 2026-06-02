// VS Code command registrations for Flint.
import type { ExtensionContext } from 'vscode';
import type { DaemonConnection } from '../ipc/connection';

const vscodeApi = (() => {
  try { return require('vscode'); } catch { return null; }
})();

export interface CommandCallbacks {
  onRevealObservation: (obsId: string, filePath: string, lineNumber: number) => void;
  onDismissObservation: (obsId: string) => void;
  onGotoFile: (filePath: string, lineNumber: number) => void;
}

export function registerCommands(
  context: ExtensionContext,
  conn: DaemonConnection,
  callbacks: CommandCallbacks,
): void {
  if (!vscodeApi) return;
  const { commands, window } = vscodeApi;

  context.subscriptions.push(
    commands.registerCommand('flint.showStatus', () => {
      window.showInformationMessage(
        `Flint — ${conn.connectionState}. Daemon socket: active.`,
      );
    }),

    commands.registerCommand('flint.startDaemon', () => {
      const { workspace } = vscodeApi;
      const root = workspace.workspaceFolders?.[0]?.uri.fsPath ?? '';
      if (!root) {
        window.showErrorMessage('Flint: no workspace folder open.');
        return;
      }
      conn.connect();
      window.showInformationMessage('Flint: connecting to daemon...');
    }),

    commands.registerCommand('flint.stopDaemon', () => {
      conn.disconnect();
      window.showInformationMessage('Flint: disconnected from daemon.');
    }),

    // Fired by CodeLens clicks — reveal the observation in editor + sidebar
    commands.registerCommand(
      'flint.revealObservation',
      (obsId: string, filePath: string, lineNumber: number) => {
        callbacks.onRevealObservation(obsId, filePath, lineNumber);
      },
    ),

    // Fired by sidebar "Go to file" button
    commands.registerCommand(
      'flint.gotoFile',
      (filePath: string, lineNumber: number) => {
        callbacks.onGotoFile(filePath, lineNumber);
      },
    ),
  );
}
