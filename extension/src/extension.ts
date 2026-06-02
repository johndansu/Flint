// Flint VS Code extension entry point.
import type { ExtensionContext } from 'vscode';
import { DaemonConnection } from './ipc/connection';
import { FlintStatusBar } from './status/statusbar';
import { FlintSidebarProvider } from './sidebar/provider';
import { FlintDecorationManager } from './editor/decorations';
import { FlintCodeLensProvider } from './editor/codelens';
import { registerCommands } from './commands/index';
import { asObservation, asDaemonStatus } from './ipc/connection';
import { openFollowupPanel } from './panels/followup';
import { openToolOutputPanel } from './panels/tooloutput';

const vscodeApi = (() => {
  try { return require('vscode'); } catch { return null; }
})();

export function activate(context: ExtensionContext): void {
  const { workspace, window, languages, Uri } = vscodeApi ?? {};
  const workspaceRoot = workspace?.workspaceFolders?.[0]?.uri.fsPath ?? process.cwd();

  const conn        = new DaemonConnection(workspaceRoot);
  const statusBar   = new FlintStatusBar();
  const decorations = new FlintDecorationManager(context);
  const codeLens    = new FlintCodeLensProvider();

  // Register CodeLens provider for all file types
  if (languages) {
    context.subscriptions.push(
      languages.registerCodeLensProvider({ scheme: 'file' }, codeLens),
    );
  }

  // Sidebar webview
  const sidebarProvider = new FlintSidebarProvider(conn, context);
  if (vscodeApi) {
    context.subscriptions.push(
      window.registerWebviewViewProvider('flint.sidebar', sidebarProvider),
    );
  }

  // Route incoming IPC messages
  conn.onMessage(msg => {
    // Status bar: update on daemonStatus
    const status = asDaemonStatus(msg);
    if (status) {
      statusBar.setConnected(status.awareness);
      return;
    }

    // Observation: push to sidebar, decorations, and CodeLens
    const obs = asObservation(msg);
    if (obs) {
      sidebarProvider.pushObservation(obs);
      statusBar.incrementObservations();

      if (obs.filePath && obs.lineNumber != null) {
        decorations.addObservation(obs);
        codeLens.addObservation(obs);
      }
    }
  });

  // Wire sidebar callbacks
  sidebarProvider.onTellMeMore = obs => openFollowupPanel(context, obs, vscodeApi);
  sidebarProvider.onRunTool    = cmd => openToolOutputPanel(context, cmd, vscodeApi);

  // Handle the "reveal observation" command fired by CodeLens clicks
  // and the "go to file" action from sidebar cards.
  registerCommands(context, conn, {
    onRevealObservation: (obsId: string, filePath: string, lineNumber: number) => {
      // Focus sidebar (the view container)
      vscodeApi?.commands.executeCommand('flint.sidebar.focus');
      sidebarProvider.highlightCard(obsId);

      // Also navigate the editor to the relevant line
      if (filePath && lineNumber && Uri) {
        const uri = Uri.file(filePath);
        const pos = new vscodeApi.Position(lineNumber - 1, 0);
        const range = new vscodeApi.Range(pos, pos);
        window?.showTextDocument(uri, { selection: range, preserveFocus: false });
      }
    },

    onDismissObservation: (obsId: string) => {
      decorations.removeObservation(obsId);
      codeLens.removeObservation(obsId);
      statusBar.decrementObservations();
    },

    onGotoFile: (filePath: string, lineNumber: number) => {
      if (!filePath || !Uri) return;
      const uri = Uri.file(filePath);
      const pos = new vscodeApi.Position(Math.max(0, lineNumber - 1), 0);
      window?.showTextDocument(uri, {
        selection: new vscodeApi.Range(pos, pos),
        preserveFocus: false,
      });
    },
  });

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

export function deactivate(): void {}
