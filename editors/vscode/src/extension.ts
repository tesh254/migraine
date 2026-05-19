import * as vscode from "vscode";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  Trace,
} from "vscode-languageclient/node";

let client: LanguageClient | undefined;

export function activate(context: vscode.ExtensionContext) {
  const config = vscode.workspace.getConfiguration("migraine.lsp");
  const enabled = config.get<boolean>("enabled", true);
  if (!enabled) {
    return;
  }

  const migrainePath = config.get<string>("path", "migraine");

  const serverOptions: ServerOptions = {
    run: {
      command: migrainePath,
      args: ["lsp"],
    },
    debug: {
      command: migrainePath,
      args: ["lsp"],
    },
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [{ scheme: "file", language: "migraine" }],
    traceOutputChannel: vscode.window.createOutputChannel(
      "Migraine LSP Trace"
    ),
  };

  client = new LanguageClient(
    "migraine-lsp",
    "Migraine Language Server",
    serverOptions,
    clientOptions
  );

  client.start();

  context.subscriptions.push(
    vscode.commands.registerCommand("migraine.restartLsp", async () => {
      if (client) {
        await client.stop();
        client = new LanguageClient(
          "migraine-lsp",
          "Migraine Language Server",
          serverOptions,
          clientOptions
        );
        client.start();
        vscode.window.showInformationMessage("Migraine LSP restarted.");
      }
    })
  );
}

export function deactivate(): Thenable<void> | undefined {
  return client?.stop();
}