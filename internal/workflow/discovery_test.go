package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverWorkflowsFromCWD_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	yamlContent := []byte("name: test-yaml\nsteps:\n  - command: echo hello\n")
	if err := os.WriteFile(filepath.Join(tmpDir, "test-yaml.yaml"), yamlContent, 0644); err != nil {
		t.Fatal(err)
	}

	workflows, err := DiscoverWorkflowsFromCWD()
	if err != nil {
		t.Fatalf("DiscoverWorkflowsFromCWD() error: %v", err)
	}

	found := false
	for _, wf := range workflows {
		if wf.Name == "test-yaml" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find test-yaml workflow from YAML file")
	}
}

func TestDiscoverWorkflowsFromCWD_MG(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	mgContent := []byte(`metadata {
    name = "test-mg"
    desc = "A test .mg workflow"
}

workflow {
    steps [
        {
            cmd = ` + "`echo hello`" + `
            desc = "Say hello"
        }
    ]
}
`)
	if err := os.WriteFile(filepath.Join(tmpDir, "test-mg.mg"), mgContent, 0644); err != nil {
		t.Fatal(err)
	}

	workflows, err := DiscoverWorkflowsFromCWD()
	if err != nil {
		t.Fatalf("DiscoverWorkflowsFromCWD() error: %v", err)
	}

	found := false
	for _, wf := range workflows {
		if wf.Name == "test-mg" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find test-mg workflow from .mg file")
	}
}

func TestDiscoverWorkflowsFromCWD_SkillDir(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	skillDir := filepath.Join(tmpDir, ".migraine", "skills")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	yamlContent := []byte("name: skill-workflow\nsteps:\n  - command: echo skill\n")
	if err := os.WriteFile(filepath.Join(skillDir, "skill-workflow.yaml"), yamlContent, 0644); err != nil {
		t.Fatal(err)
	}

	workflows, err := DiscoverWorkflowsFromCWD()
	if err != nil {
		t.Fatalf("DiscoverWorkflowsFromCWD() error: %v", err)
	}

	found := false
	for _, wf := range workflows {
		if wf.Name == "skill-workflow" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find skill-workflow from .migraine/skills directory")
	}
}

func TestDiscoverWorkflowsFromCWD_GlobalSkillDir(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	homeSkillDir := filepath.Join(tmpDir, "home", ".migraine", "skills")
	if err := os.MkdirAll(homeSkillDir, 0755); err != nil {
		t.Fatal(err)
	}

	os.Setenv("HOME", filepath.Join(tmpDir, "home"))
	defer os.Unsetenv("HOME")

	yamlContent := []byte("name: global-skill\nsteps:\n  - command: echo global\n")
	if err := os.WriteFile(filepath.Join(homeSkillDir, "global-skill.yaml"), yamlContent, 0644); err != nil {
		t.Fatal(err)
	}

	workflows, err := DiscoverWorkflowsFromCWD()
	if err != nil {
		t.Fatalf("DiscoverWorkflowsFromCWD() error: %v", err)
	}

	found := false
	for _, wf := range workflows {
		if wf.Name == "global-skill" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find global-skill from global skills directory")
	}
}

func TestDiscoverWorkflowsFromCWD_WorkflowsDir(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	wfDir := filepath.Join(tmpDir, "workflows")
	if err := os.MkdirAll(wfDir, 0755); err != nil {
		t.Fatal(err)
	}

	yamlContent := []byte("name: sub-dir-wf\nsteps:\n  - command: echo sub\n")
	if err := os.WriteFile(filepath.Join(wfDir, "sub-dir-wf.yaml"), yamlContent, 0644); err != nil {
		t.Fatal(err)
	}

	workflows, err := DiscoverWorkflowsFromCWD()
	if err != nil {
		t.Fatalf("DiscoverWorkflowsFromCWD() error: %v", err)
	}

	found := false
	for _, wf := range workflows {
		if wf.Name == "sub-dir-wf" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find sub-dir-wf from workflows/ directory")
	}
}

func TestFindWorkflowByName(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	yamlContent := []byte("name: my-workflow\nsteps:\n  - command: echo test\n")
	if err := os.WriteFile(filepath.Join(tmpDir, "my-workflow.yaml"), yamlContent, 0644); err != nil {
		t.Fatal(err)
	}

	wf, err := FindWorkflowByName("my-workflow")
	if err != nil {
		t.Fatalf("FindWorkflowByName() error: %v", err)
	}
	if wf.Name != "my-workflow" {
		t.Errorf("expected workflow name my-workflow, got %s", wf.Name)
	}

	_, err = FindWorkflowByName("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent workflow name")
	}
}

func TestFindWorkflowByName_MG(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	mgContent := []byte(`metadata {
    name = "my-mg-wf"
}

workflow {
    steps [
        {
            cmd = ` + "`echo mg`" + `
        }
    ]
}
`)
	if err := os.WriteFile(filepath.Join(tmpDir, "my-mg-wf.mg"), mgContent, 0644); err != nil {
		t.Fatal(err)
	}

	wf, err := FindWorkflowByName("my-mg-wf")
	if err != nil {
		t.Fatalf("FindWorkflowByName() error: %v", err)
	}
	if wf.Name != "my-mg-wf" {
		t.Errorf("expected workflow name my-mg-wf, got %s", wf.Name)
	}
}

func TestGetWorkflowFilePath_MG(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	mgContent := []byte(`metadata { name = "findme" } workflow { steps [{ cmd = ` + "`echo hi`" + ` }] }`)
	if err := os.WriteFile(filepath.Join(tmpDir, "findme.mg"), mgContent, 0644); err != nil {
		t.Fatal(err)
	}

	path, err := GetWorkflowFilePath("findme")
	if err != nil {
		t.Fatalf("GetWorkflowFilePath() error: %v", err)
	}
	if filepath.Base(path) != "findme.mg" {
		t.Errorf("expected findme.mg, got %s", filepath.Base(path))
	}
}

func TestGetWorkflowFilePath_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	_, err := GetWorkflowFilePath("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent workflow file")
	}
}