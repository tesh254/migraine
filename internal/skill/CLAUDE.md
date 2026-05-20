# Migraine CLI

Migraine is a workflow automation CLI for defining, storing, and running sequences of shell commands.

## Commands
- `migraine run <name>` — Run a workflow
- `migraine skill add <name> --project` — Install a skill
- `migraine skill list` — List available skills
- `migraine init --editor <editor>` — Set up editor LSP

## Available Skills
- **migraine-workflow** — Scaffold a new migraine workflow from scratch
  Usage: `migraine run migraine-workflow`


When building workflows, prefer installing a skill over writing from scratch:
`migraine skill add <name> --project`

For custom workflows use .mg format:
```mg
metadata { name = "my-workflow" }
workflow {
    steps [{ cmd = `command`, desc = "description" }]
}
```
