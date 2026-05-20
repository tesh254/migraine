package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tesh254/migraine/internal/ui"
)

type AgentConfig struct {
	Name        string
	DisplayName string
	ConfigFile  string
	Generate    func(dir string, skills []Skill) error
}

var supportedAgents = []AgentConfig{
	{
		Name:        "opencode",
		DisplayName: "opencode",
		ConfigFile:  "AGENTS.md",
		Generate:    generateOpenCode,
	},
	{
		Name:        "claude",
		DisplayName: "Claude Code",
		ConfigFile:  "CLAUDE.md",
		Generate:    generateClaude,
	},
	{
		Name:        "codex",
		DisplayName: "OpenAI Codex",
		ConfigFile:  "AGENTS.md",
		Generate:    generateCodex,
	},
	{
		Name:        "cursor",
		DisplayName: "Cursor",
		ConfigFile:  ".cursorrules",
		Generate:    generateCursor,
	},
	{
		Name:        "windsurf",
		DisplayName: "Windsurf",
		ConfigFile:  ".windsurfrules",
		Generate:    generateWindsurf,
	},
	{
		Name:        "cline",
		DisplayName: "Cline",
		ConfigFile:  ".clinerules",
		Generate:    generateCline,
	},
}

func SupportedAgents() []AgentConfig {
	return supportedAgents
}

func FindAgent(name string) (*AgentConfig, bool) {
	for _, a := range supportedAgents {
		if a.Name == name {
			return &a, true
		}
	}
	return nil, false
}

func SetupAgent(agentName string, skills []Skill) error {
	agent, ok := FindAgent(agentName)
	if !ok {
		return fmt.Errorf("unsupported agent '%s'. Supported: %s", agentName, AgentList())
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	if err := agent.Generate(cwd, skills); err != nil {
		return err
	}

	ui.LogSuccessBordered(fmt.Sprintf("Agent config generated for %s", agent.DisplayName))
	fmt.Printf("  Config file: %s\n", filepath.Join(cwd, agent.ConfigFile))
	return nil
}

func AgentList() string {
	names := make([]string, len(supportedAgents))
	for i, a := range supportedAgents {
		names[i] = a.Name
	}
	return strings.Join(names, ", ")
}

func skillReference(skills []Skill) string {
	var b strings.Builder
	for _, s := range skills {
		b.WriteString(fmt.Sprintf("- **%s** — %s\n", s.Name, s.Description))
		if len(s.Variables) > 0 {
			b.WriteString(fmt.Sprintf("  Variables: %s\n", strings.Join(mapKeys(s.Variables), ", ")))
		}
		b.WriteString(fmt.Sprintf("  Usage: `migraine run %s`\n", s.Name))
	}
	return b.String()
}

func generateOpenCode(dir string, skills []Skill) error {
	path := filepath.Join(dir, "AGENTS.md")
	content := fmt.Sprintf(`# Migraine CLI — Agent Instructions

## Overview
Migraine is a workflow automation CLI. It defines, stores, and runs sequences of shell commands with variable substitution, pre-flight checks, and discrete actions.

## Key Commands
- `+"`migraine run <workflow>`"+` — Run a workflow by name
- `+"`migraine run`"+` — Run the project workflow (migraine.yml in CWD)
- `+"`migraine workflow list`"+` — List all workflows
- `+"`migraine skill add <name> --project`"+` — Install a skill per-project
- `+"`migraine skill add <name> --global`"+` — Install a skill globally
- `+"`migraine skill list`"+` — List available skills
- `+"`migraine init --editor <vscode|neovim|vim|helix>`"+` — Configure editor LSP

## Workflow Format (.mg)
Workflows can be defined in .mg files:
`+"```"+`mg
metadata {
    name = "my-workflow"
    desc = "Description"
}

workflow {
    steps [
        { cmd = `+"`echo hello`"+`, desc = "Say hello" }
    ]
}
`+"```"+`

Variables use `+"`args:VAR`"+`, `+"`env:VAR`"+`, or `+"`vault:VAR`"+` prefixes.

## Available Skills
%s

## Working with Skills
When the user asks to set up a workflow:
1. Check if a built-in skill matches: `+"`migraine skill list`"+`
2. Install it: `+"`migraine skill add <name> --project`"+`
3. Run it: `+"`migraine run <name>`"+`

For custom workflows, create a .mg file or migraine.yml in the project root.
`, skillReference(skills))
	return writeFile(path, content)
}

func generateClaude(dir string, skills []Skill) error {
	path := filepath.Join(dir, "CLAUDE.md")
	content := fmt.Sprintf(`# Migraine CLI

Migraine is a workflow automation CLI for defining, storing, and running sequences of shell commands.

## Commands
- `+"`migraine run <name>`"+` — Run a workflow
- `+"`migraine skill add <name> --project`"+` — Install a skill
- `+"`migraine skill list`"+` — List available skills
- `+"`migraine init --editor <editor>`"+` — Set up editor LSP

## Available Skills
%s

When building workflows, prefer installing a skill over writing from scratch:
`+"`migraine skill add <name> --project`"+`

For custom workflows use .mg format:
`+"```mg"+`
metadata { name = "my-workflow" }
workflow {
    steps [{ cmd = `+"`command`"+`, desc = "description" }]
}
`+"```"+`
`, skillReference(skills))
	return writeFile(path, content)
}

func generateCodex(dir string, skills []Skill) error {
	path := filepath.Join(dir, "AGENTS.md")
	content := fmt.Sprintf(`# Migraine CLI

Migraine automates workflows via CLI with variable substitution and hook support.

## Commands
- `+"`migraine run <workflow>`"+` — Run a workflow by name
- `+"`migraine skill add <name> --project`"+` — Install a skill for this project
- `+"`migraine skill add <name> --global`"+` — Install a skill globally
- `+"`migraine skill list`"+` — List available skills

## Skills
%s

Use `+"`migraine skill add <name> --project`"+` to install a skill, then `+"`migraine run <name>`"+` to run it.
`, skillReference(skills))
	return writeFile(path, content)
}

func generateCursor(dir string, skills []Skill) error {
	path := filepath.Join(dir, ".cursorrules")
	content := fmt.Sprintf(`# Migraine CLI Rules

Migraine is a workflow automation CLI.

## Available Skills
%s

## Usage
- Install: `+"`migraine skill add <name> --project`"+`
- Run: `+"`migraine run <name>`"+`
- List: `+"`migraine skill list`"+`

When asked to set up CI, deployment, or Git workflows, check `+"`migraine skill list`"+` first.
`, skillReference(skills))
	return writeFile(path, content)
}

func generateWindsurf(dir string, skills []Skill) error {
	path := filepath.Join(dir, ".windsurfrules")
	content := fmt.Sprintf(`# Migraine CLI Rules

Migraine is a workflow automation CLI.

## Available Skills
%s

## Usage
- Install: `+"`migraine skill add <name> --project`"+`
- Run: `+"`migraine run <name>`"+`
- List: `+"`migraine skill list`"+`

When asked to set up CI, deployment, or Git workflows, check `+"`migraine skill list`"+` first.
`, skillReference(skills))
	return writeFile(path, content)
}

func generateCline(dir string, skills []Skill) error {
	path := filepath.Join(dir, ".clinerules")
	content := fmt.Sprintf(`# Migraine CLI

Migraine is a workflow automation CLI.

## Available Skills
%s

Use `+"`migraine skill add <name> --project`"+` to install, `+"`migraine run <name>`"+` to run.
`, skillReference(skills))
	return writeFile(path, content)
}

func writeFile(path, content string) error {
	existing, err := os.ReadFile(path)
	if err == nil && len(existing) > 0 {
		marker := "<!-- migraine-skills -->"
		if strings.Contains(string(existing), marker) {
			return os.WriteFile(path, []byte(content), 0644)
		}
		content = string(existing) + "\n\n" + content
	}
	return os.WriteFile(path, []byte(content), 0644)
}