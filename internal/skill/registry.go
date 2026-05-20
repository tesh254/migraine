package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tesh254/migraine/internal/ui"
	"gopkg.in/yaml.v3"
)

type Skill struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Category    string            `yaml:"category"`
	Tags        []string          `yaml:"tags,omitempty"`
	Variables   map[string]string `yaml:"variables,omitempty"`
	Workflow    SkillWorkflow     `yaml:"workflow"`
}

type SkillWorkflow struct {
	PreChecks []SkillStep          `yaml:"pre_checks,omitempty"`
	Steps     []SkillStep          `yaml:"steps"`
	Actions   map[string]SkillStep `yaml:"actions,omitempty"`
}

type SkillStep struct {
	Command     string  `yaml:"command"`
	Description *string `yaml:"description,omitempty"`
	OnFail      string  `yaml:"on_fail,omitempty"`
	OnSuccess   string  `yaml:"on_success,omitempty"`
}

var builtinSkills = []Skill{
	{
		Name:        "migraine-workflow",
		Description:  "Scaffold a new migraine workflow from scratch",
		Category:    "migraine",
		Tags:        []string{"migraine", "workflow", "scaffold", "init", "setup"},
		Workflow: SkillWorkflow{
			Steps: []SkillStep{
				{Command: "migraine workflow init", Description: strPtr("Create migraine.yml project config")},
				{Command: "migraine workflow validate migraine.yml", Description: strPtr("Validate the generated workflow")},
			},
		},
	},
}

func strPtr(s string) *string { return &s }

var globalSkillDirFn = defaultGlobalSkillDir
var projectSkillDirFn = defaultProjectSkillDir

func defaultGlobalSkillDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".migraine", "skills"), nil
}

func defaultProjectSkillDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, ".migraine", "skills"), nil
}

func List() []Skill {
	return builtinSkills
}

func Find(name string) (*Skill, bool) {
	for _, s := range builtinSkills {
		if s.Name == name {
			return &s, true
		}
	}
	return nil, false
}

func Search(query string) []Skill {
	q := strings.ToLower(query)
	var results []Skill
	for _, s := range builtinSkills {
		if strings.Contains(strings.ToLower(s.Name), q) ||
			strings.Contains(strings.ToLower(s.Description), q) ||
			strings.Contains(strings.ToLower(s.Category), q) {
			results = append(results, s)
			continue
		}
		for _, t := range s.Tags {
			if strings.Contains(strings.ToLower(t), q) {
				results = append(results, s)
				break
			}
		}
	}
	return results
}

func Install(name string, scope string) error {
	s, ok := Find(name)
	if !ok {
		return fmt.Errorf("skill '%s' not found. Use 'migraine skill list' to see available skills", name)
	}

	var dir string
	var err error
	switch scope {
	case "global":
		dir, err = globalSkillDirFn()
	case "project":
		dir, err = projectSkillDirFn()
	default:
		return fmt.Errorf("invalid scope '%s': must be 'global' or 'project'", scope)
	}
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %v", err)
	}

	filename := filepath.Join(dir, s.Name+".yaml")

	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal skill: %v", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write skill file: %v", err)
	}

	ui.LogSuccessBordered(fmt.Sprintf("Skill '%s' installed (%s)", s.Name, scope))
	ui.SectionHeader("Details")
	fmt.Printf("  Name:        %s\n", s.Name)
	fmt.Printf("  Description: %s\n", s.Description)
	fmt.Printf("  Category:    %s\n", s.Category)
	fmt.Printf("  Location:    %s\n", filename)
	if len(s.Variables) > 0 {
		fmt.Printf("  Variables:   %s\n", strings.Join(mapKeys(s.Variables), ", "))
	}
	fmt.Printf("\n  Run with: migraine run %s\n", s.Name)
	return nil
}

func Remove(name string) error {
	dirs := []string{}

	globalDir, err := globalSkillDirFn()
	if err == nil {
		dirs = append(dirs, globalDir)
	}
	projectDir, err := projectSkillDirFn()
	if err == nil {
		dirs = append(dirs, projectDir)
	}

	for _, dir := range dirs {
		filename := filepath.Join(dir, name+".yaml")
		if _, err := os.Stat(filename); err == nil {
			if err := os.Remove(filename); err != nil {
				return fmt.Errorf("failed to remove skill file: %v", err)
			}
			ui.LogSuccessBordered(fmt.Sprintf("Skill '%s' removed from %s", name, dir))
			return nil
		}
	}

	return fmt.Errorf("skill '%s' is not installed", name)
}

func ListInstalled() ([]string, error) {
	var installed []string

	globalDir, err := globalSkillDirFn()
	if err == nil {
		files, _ := filepath.Glob(filepath.Join(globalDir, "*.yaml"))
		for _, f := range files {
			name := strings.TrimSuffix(filepath.Base(f), ".yaml")
			installed = append(installed, name+" (global)")
		}
	}

	projectDir, err := projectSkillDirFn()
	if err == nil {
		files, _ := filepath.Glob(filepath.Join(projectDir, "*.yaml"))
		for _, f := range files {
			name := strings.TrimSuffix(filepath.Base(f), ".yaml")
			installed = append(installed, name+" (project)")
		}
	}

	return installed, nil
}

func RenderMG(s *Skill) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("metadata {\n"))
	b.WriteString(fmt.Sprintf("    name = \"%s\"\n", s.Name))
	b.WriteString(fmt.Sprintf("    desc = \"%s\"\n", s.Description))
	b.WriteString("}\n\n")

	if len(s.Variables) > 0 {
		b.WriteString("variables {\n")
		for k, v := range s.Variables {
			b.WriteString(fmt.Sprintf("    %s = \"%s\"\n", k, v))
		}
		b.WriteString("}\n\n")
	}

	b.WriteString("workflow {\n")

	if len(s.Workflow.PreChecks) > 0 {
		b.WriteString("    pre_checks [\n")
		for _, pc := range s.Workflow.PreChecks {
			b.WriteString("        {\n")
			b.WriteString(fmt.Sprintf("            cmd = `%s`\n", pc.Command))
			if pc.Description != nil {
				b.WriteString(fmt.Sprintf("            desc = \"%s\"\n", *pc.Description))
			}
			if pc.OnFail != "" {
				b.WriteString(fmt.Sprintf("            on_fail = \"%s\"\n", pc.OnFail))
			}
			if pc.OnSuccess != "" {
				b.WriteString(fmt.Sprintf("            on_success = \"%s\"\n", pc.OnSuccess))
			}
			b.WriteString("        },\n")
		}
		b.WriteString("    ]\n\n")
	}

	b.WriteString("    steps [\n")
	for _, step := range s.Workflow.Steps {
		b.WriteString("        {\n")
		b.WriteString(fmt.Sprintf("            cmd = `%s`\n", step.Command))
		if step.Description != nil {
			b.WriteString(fmt.Sprintf("            desc = \"%s\"\n", *step.Description))
		}
		if step.OnFail != "" {
			b.WriteString(fmt.Sprintf("            on_fail = \"%s\"\n", step.OnFail))
		}
		if step.OnSuccess != "" {
			b.WriteString(fmt.Sprintf("            on_success = \"%s\"\n", step.OnSuccess))
		}
		b.WriteString("        },\n")
	}
	b.WriteString("    ]\n")

	if len(s.Workflow.Actions) > 0 {
		b.WriteString("\n    actions {\n")
		for name, action := range s.Workflow.Actions {
			b.WriteString(fmt.Sprintf("        %s {\n", name))
			b.WriteString(fmt.Sprintf("            cmd = `%s`\n", action.Command))
			if action.Description != nil {
				b.WriteString(fmt.Sprintf("            desc = \"%s\"\n", *action.Description))
			}
			if action.OnFail != "" {
				b.WriteString(fmt.Sprintf("            on_fail = \"%s\"\n", action.OnFail))
			}
			b.WriteString("        },\n")
		}
		b.WriteString("    }\n")
	}

	b.WriteString("}\n")
	return b.String()
}

func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}