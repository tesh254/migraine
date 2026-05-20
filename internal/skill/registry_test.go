package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	skills := List()
	if len(skills) == 0 {
		t.Fatal("expected at least one built-in skill, got none")
	}

	found := false
	for _, s := range skills {
		if s.Name == "migraine-workflow" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected migraine-workflow skill to exist in built-in skills")
	}
}

func TestFind(t *testing.T) {
	s, ok := Find("migraine-workflow")
	if !ok {
		t.Fatal("expected to find migraine-workflow skill")
	}
	if s.Name != "migraine-workflow" {
		t.Errorf("expected name=migraine-workflow, got %s", s.Name)
	}
	if s.Category != "migraine" {
		t.Errorf("expected category=migraine, got %s", s.Category)
	}
}

func TestFindNotFound(t *testing.T) {
	_, ok := Find("nonexistent-skill")
	if ok {
		t.Error("expected Find to return false for nonexistent skill")
	}
}

func TestSearch(t *testing.T) {
	results := Search("migraine")
	if len(results) == 0 {
		t.Fatal("expected at least one result for 'migraine' search")
	}

	found := false
	for _, s := range results {
		if s.Name == "migraine-workflow" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected migraine-workflow in search results for 'migraine'")
	}
}

func TestSearchByTag(t *testing.T) {
	results := Search("scaffold")
	if len(results) == 0 {
		t.Fatal("expected at least one result for 'scaffold' search (tag)")
	}
}

func TestSearchNoResults(t *testing.T) {
	results := Search("zzznonexistent")
	if len(results) != 0 {
		t.Error("expected no results for nonsensical search")
	}
}

func TestInstallAndRemove(t *testing.T) {
	tmpDir := t.TempDir()

	origGlobal := globalSkillDirFn
	origProject := projectSkillDirFn
	defer func() {
		globalSkillDirFn = origGlobal
		projectSkillDirFn = origProject
	}()

	globalSkillDirFn = func() (string, error) { return filepath.Join(tmpDir, "global", ".migraine", "skills"), nil }
	projectSkillDirFn = func() (string, error) { return filepath.Join(tmpDir, "project", ".migraine", "skills"), nil }

	err := Install("migraine-workflow", "global")
	if err != nil {
		t.Fatalf("Install global failed: %v", err)
	}

	globalPath := filepath.Join(tmpDir, "global", ".migraine", "skills", "migraine-workflow.yaml")
	if _, err := os.Stat(globalPath); os.IsNotExist(err) {
		t.Error("expected skill file to exist at global path")
	}

	err = Install("migraine-workflow", "project")
	if err != nil {
		t.Fatalf("Install project failed: %v", err)
	}

	projectPath := filepath.Join(tmpDir, "project", ".migraine", "skills", "migraine-workflow.yaml")
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Error("expected skill file to exist at project path")
	}

	installed, err := ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled failed: %v", err)
	}
	if len(installed) == 0 {
		t.Error("expected at least one installed skill")
	}

	foundGlobal := false
	foundProject := false
	for _, i := range installed {
		if strings.Contains(i, "migraine-workflow") && strings.Contains(i, "global") {
			foundGlobal = true
		}
		if strings.Contains(i, "migraine-workflow") && strings.Contains(i, "project") {
			foundProject = true
		}
	}
	if !foundGlobal {
		t.Error("expected migraine-workflow (global) in installed list")
	}
	if !foundProject {
		t.Error("expected migraine-workflow (project) in installed list")
	}

	err = Remove("migraine-workflow")
	if err != nil {
		t.Fatalf("first Remove failed: %v", err)
	}

	// Second call removes the project copy
	err = Remove("migraine-workflow")
	if err != nil {
		t.Fatalf("second Remove failed: %v", err)
	}

	// Third call should fail since neither copy exists
	err = Remove("migraine-workflow")
	if err == nil {
		t.Error("expected error when removing skill with no remaining copies")
	}
}

func TestInstallInvalidScope(t *testing.T) {
	err := Install("migraine-workflow", "invalid")
	if err == nil {
		t.Error("expected error for invalid scope")
	}
}

func TestInstallNonexistentSkill(t *testing.T) {
	err := Install("nonexistent-skill", "project")
	if err == nil {
		t.Error("expected error for nonexistent skill")
	}
}

func TestRemoveNonexistentSkill(t *testing.T) {
	err := Remove("totally-fake-skill")
	if err == nil {
		t.Error("expected error when removing nonexistent skill")
	}
}

func TestRenderMG(t *testing.T) {
	s, ok := Find("migraine-workflow")
	if !ok {
		t.Fatal("migraine-workflow skill not found")
	}

	output := RenderMG(s)

	if !strings.Contains(output, "metadata {") {
		t.Error("expected metadata block in .mg output")
	}
	if !strings.Contains(output, `name = "migraine-workflow"`) {
		t.Error("expected name in .mg output")
	}
	if !strings.Contains(output, "workflow {") {
		t.Error("expected workflow block in .mg output")
	}
	if !strings.Contains(output, "steps [") {
		t.Error("expected steps block in .mg output")
	}
	if !strings.Contains(output, "cmd =") {
		t.Error("expected cmd fields in .mg output")
	}
}

func TestRenderMGWithAllFields(t *testing.T) {
	desc := "test step"
	s := &Skill{
		Name:        "test-skill",
		Description:  "A test",
		Category:    "test",
		Tags:        []string{"test"},
		Variables:   map[string]string{"VAR1": "args:VAR1"},
		Workflow: SkillWorkflow{
			PreChecks: []SkillStep{
				{Command: "echo precheck", Description: &desc, OnFail: "action:fail"},
			},
			Steps: []SkillStep{
				{Command: "echo step1", Description: &desc, OnFail: "action:fail", OnSuccess: "action:ok"},
			},
			Actions: map[string]SkillStep{
				"fail": {Command: "echo failed", Description: &desc},
			},
		},
	}

	output := RenderMG(s)

	if !strings.Contains(output, "pre_checks [") {
		t.Error("expected pre_checks in output")
	}
	if !strings.Contains(output, `on_fail = "action:fail"`) {
		t.Error("expected on_fail in output")
	}
	if !strings.Contains(output, `on_success = "action:ok"`) {
		t.Error("expected on_success in output")
	}
	if !strings.Contains(output, "actions {") {
		t.Error("expected actions block in output")
	}
	if !strings.Contains(output, "variables {") {
		t.Error("expected variables block in output")
	}
	if !strings.Contains(output, `VAR1 = "args:VAR1"`) {
		t.Error("expected VAR1 in output")
	}
}