// VS Code command registrations for Flint.
import type { ExtensionContext } from 'vscode';
import type { DaemonConnection } from '../ipc/connection';

const vscodeApi = (() => {
  try { return require('vscode'); } catch { return null; }
})();

export function registerCommands(
  context: ExtensionContext,
  conn: DaemonConnection,
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
  );
}
