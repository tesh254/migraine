# Migraine CLI — Agent Instructions

## Overview
Migraine is a workflow automation CLI. It defines, stores, and runs sequences of shell commands with variable substitution, pre-flight checks, and discrete actions.

## Key Commands
- `migraine run <workflow>` — Run a workflow by name
- `migraine run` — Run the project workflow (migraine.yml in CWD)
- `migraine workflow list` — List all workflows
- `migraine skill add <name> --project` — Install a skill per-project
- `migraine skill add <name> --global` — Install a skill globally
- `migraine skill list` — List available skills
- `migraine init --editor <vscode|neovim|vim|helix>` — Configure editor LSP

## Workflow Format (.mg)
Workflows can be defined in .mg files:
```mg
metadata {
    name = "my-workflow"
    desc = "Description"
}

workflow {
    steps [
        { cmd = `echo hello`, desc = "Say hello" }
    ]
}
```

Variables use `args:VAR`, `env:VAR`, or `vault:VAR` prefixes.

## Available Skills
- **migraine-workflow** — Scaffold a new migraine workflow from scratch
  Usage: `migraine run migraine-workflow`


## Working with Skills
When the user asks to set up a workflow:
1. Check if a built-in skill matches: `migraine skill list`
2. Install it: `migraine skill add <name> --project`
3. Run it: `migraine run <name>`

For custom workflows, create a .mg file or migraine.yml in the project root.


# Migraine CLI

Migraine automates workflows via CLI with variable substitution and hook support.

## Commands
- `migraine run <workflow>` — Run a workflow by name
- `migraine skill add <name> --project` — Install a skill for this project
- `migraine skill add <name> --global` — Install a skill globally
- `migraine skill list` — List available skills

## Skills
- **migraine-workflow** — Scaffold a new migraine workflow from scratch
  Usage: `migraine run migraine-workflow`


Use `migraine skill add <name> --project` to install a skill, then `migraine run <name>` to run it.
