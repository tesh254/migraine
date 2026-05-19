# Migraine Workflow Language — VS Code Extension

Syntax highlighting, diagnostics, autocompletion, hover docs, and document outline for `.mg` (Migraine) workflow files.

## Features

- **Syntax highlighting** — Full TextMate grammar for the Migraine DSL: blocks (`metadata`, `variables`, `workflow`, `config`), sections (`steps`, `pre_checks`, `actions`), properties, template variables (`{{var}}`), value prefixes (`args:`, `env:`, `vault:`, `action:`, `run:`), strings, booleans, numbers, and comments.
- **Real-time diagnostics** — Parse errors are highlighted as you type, powered by the Migraine LSP server.
- **Autocompletion** — Keyword, block, property, and value-prefix completions.
- **Hover documentation** — Hover over any keyword to see markdown docs.
- **Document outline** — `metadata`, `workflow`, `config` blocks appear in the editor outline.
- **Semantic tokens** — Rich tokenization for enhanced coloring.

## Requirements

- [Migraine CLI](https://github.com/tesh254/migraine) must be installed and on your `PATH` (the `migraine lsp` command is used to start the language server).
- VS Code 1.85.0 or newer.

## Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `migraine.lsp.path` | `migraine` | Path to the migraine binary |
| `migraine.lsp.enabled` | `true` | Enable/disable the LSP server |

## Example `.mg` file

```mg
metadata {
    name = "deploy-app"
    desc = "Build and deploy to staging"
}

variables {
    app_name = "args:APP_NAME"
    env = "args:ENV"
    deploy_host = "env:DEPLOY_HOST"
}

workflow {
    pre_checks [
        {
            cmd = `docker info`
            desc = "Verify Docker daemon is running"
            on_fail = "action:notify_failure"
        }
    ]

    steps [
        {
            cmd = `docker build -t {{app_name}}:{{env}} .`
            desc = "Build the Docker image"
            on_fail = "action:notify_failure"
        }
    ]

    actions {
        notify_failure {
            cmd = `echo "deploy failed"`
            desc = "Handle failure"
        }
    }
}

config {
    store_variables = true
    store_logs = true
    background = false
    global = false
}
```

## License

MIT