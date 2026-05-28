// Flint VS Code extension entry point.
import type { ExtensionContext } from 'vscode';
import { DaemonConnection } from './ipc/connection';
import { FlintStatusBar } from './status/statusbar';
import { FlintSidebarProvider } from './sidebar/provider';
import { registerCommands } from './commands/index';
import { asDaemonStatus } from './ipc/connection';

const vscodeApi = (() => {
  try { return require('vscode'); } catch { return null; }
})();

export function activate(context: ExtensionContext): void {
  const { workspace, window } = vscodeApi ?? {};
  const workspaceRoot = workspace?.workspaceFolders?.[0]?.uri.fsPath ?? process.cwd();

  const conn = new DaemonConnection(workspaceRoot);
  const statusBar = new FlintStatusBar();

  // Track connection state in the status bar
  conn.onMessage(msg => {
    const status = asDaemonStatus(msg);
    if (status) {
      statusBar.setConnected(status.awareness);
    }
  });

  // Sidebar webview
  const sidebarProvider = new FlintSidebarProvider(conn, context);
  if (vscodeApi) {
    context.subscriptions.push(
      vscodeApi.window.registerWebviewViewProvider('flint.sidebar', sidebarProvider),
    );
  }

  // Commands
  registerCommands(context, conn);

  // Auto-connect
  conn.connect();

  // Cleanup on deactivate
  context.subscriptions.push({
    dispose() {
      conn.disconnect();
      statusBar.dispose();
    },
  });
}

export function deactivate(): void {
  // VS Code calls dispose() on all subscriptions; nothing extra needed here.
}
