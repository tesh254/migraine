package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testSkills() []Skill {
	return []Skill{
		{
			Name:        "migraine-workflow",
			Description:  "Scaffold a new migraine workflow from scratch",
			Category:    "migraine",
			Tags:        []string{"migraine", "workflow", "scaffold"},
			Workflow: SkillWorkflow{
				Steps: []SkillStep{
					{Command: "migraine workflow init", Description: strPtr("Create migraine.yml project config")},
				},
			},
		},
	}
}

func TestSupportedAgents(t *testing.T) {
	agents := SupportedAgents()
	if len(agents) == 0 {
		t.Fatal("expected at least one supported agent")
	}

	expectedAgents := []string{"opencode", "claude", "codex", "cursor", "windsurf", "cline"}
	for _, name := range expectedAgents {
		found := false
		for _, a := range agents {
			if a.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected agent '%s' to be supported", name)
		}
	}
}

func TestFindAgent(t *testing.T) {
	agent, ok := FindAgent("opencode")
	if !ok {
		t.Fatal("expected to find opencode agent")
	}
	if agent.ConfigFile != "AGENTS.md" {
		t.Errorf("expected AGENTS.md config file, got %s", agent.ConfigFile)
	}

	agent, ok = FindAgent("claude")
	if !ok {
		t.Fatal("expected to find claude agent")
	}
	if agent.ConfigFile != "CLAUDE.md" {
		t.Errorf("expected CLAUDE.md config file, got %s", agent.ConfigFile)
	}

	agent, ok = FindAgent("codex")
	if !ok {
		t.Fatal("expected to find codex agent")
	}

	_, ok = FindAgent("nonexistent")
	if ok {
		t.Error("expected FindAgent to return false for nonexistent agent")
	}
}

func TestAgentList(t *testing.T) {
	list := AgentList()
	if list == "" {
		t.Error("expected non-empty agent list")
	}
	for _, name := range []string{"opencode", "claude", "codex", "cursor", "windsurf", "cline"} {
		if !strings.Contains(list, name) {
			t.Errorf("expected agent list to contain '%s'", name)
		}
	}
}

func TestSetupAgent(t *testing.T) {
	skills := testSkills()

	for _, agentName := range []string{"opencode", "claude", "codex", "cursor", "windsurf", "cline"} {
		t.Run(agentName, func(t *testing.T) {
			tmpDir := t.TempDir()
			origDir, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(origDir)

			err := SetupAgent(agentName, skills)
			if err != nil {
				t.Fatalf("SetupAgent(%s) failed: %v", agentName, err)
			}

			agent, _ := FindAgent(agentName)
			configPath := filepath.Join(tmpDir, agent.ConfigFile)

			data, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("failed to read %s config: %v", agentName, err)
			}

			content := string(data)

			if !strings.Contains(content, "migraine") {
				t.Errorf("expected %s config to mention migraine", agentName)
			}
			if !strings.Contains(content, "migraine-workflow") {
				t.Errorf("expected %s config to mention migraine-workflow skill", agentName)
			}
			if !strings.Contains(content, "migraine run") {
				t.Errorf("expected %s config to mention 'migraine run'", agentName)
			}
			if !strings.Contains(content, "migraine skill add") {
				t.Errorf("expected %s config to mention 'migraine skill add'", agentName)
			}
		})
	}
}

func TestSetupAgentInvalid(t *testing.T) {
	err := SetupAgent("nonexistent", testSkills())
	if err == nil {
		t.Error("expected error for invalid agent name")
	}
}

func TestOpenCodeConfigContent(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	skills := testSkills()
	err := SetupAgent("opencode", skills)
	if err != nil {
		t.Fatalf("SetupAgent(opencode) failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}

	content := string(data)
	for _, expected := range []string{"migraine run", "migraine skill add", "migraine workflow list", ".mg", "args:VAR", "env:VAR", "vault:VAR"} {
		if !strings.Contains(content, expected) {
			t.Errorf("expected opencode config to contain '%s'", expected)
		}
	}
}

func TestClaudeConfigContent(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	skills := testSkills()
	err := SetupAgent("claude", skills)
	if err != nil {
		t.Fatalf("SetupAgent(claude) failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "migraine skill add") {
		t.Error("expected claude config to mention 'migraine skill add'")
	}
	if !strings.Contains(content, ".mg") {
		t.Error("expected claude config to mention .mg format")
	}
}

func TestCursorConfigContent(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	skills := testSkills()
	err := SetupAgent("cursor", skills)
	if err != nil {
		t.Fatalf("SetupAgent(cursor) failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, ".cursorrules"))
	if err != nil {
		t.Fatalf("failed to read .cursorrules: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "migraine skill list") {
		t.Error("expected cursor config to mention 'migraine skill list'")
	}
}