package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func skillDirs() []string {
	var dirs []string
	cwd, err := os.Getwd()
	if err == nil {
		dirs = append(dirs, filepath.Join(cwd, ".migraine", "skills"))
	}
	home, err := os.UserHomeDir()
	if err == nil {
		dirs = append(dirs, filepath.Join(home, ".migraine", "skills"))
	}
	return dirs
}

// DiscoverWorkflowsFromCWD discovers all workflows (YAML, .mg, and installed skills) in the current working directory
func DiscoverWorkflowsFromCWD() ([]YAMLWorkflow, error) {
	workflowDirs := []string{"./workflows", "."}

	var allWorkflows []YAMLWorkflow

	for _, dir := range workflowDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		yamlFiles, err := filepath.Glob(filepath.Join(dir, "*.y*ml"))
		if err != nil {
			return nil, err
		}

		for _, file := range yamlFiles {
			workflow, err := LoadYAMLWorkflow(file)
			if err != nil {
				fmt.Printf("Warning: failed to load workflow from %s: %v\n", file, err)
				continue
			}
			allWorkflows = append(allWorkflows, *workflow)
		}

		mgFiles, err := filepath.Glob(filepath.Join(dir, "*.mg"))
		if err != nil {
			return nil, err
		}

		for _, file := range mgFiles {
			workflow, err := LoadMigraineWorkflow(file)
			if err != nil {
				fmt.Printf("Warning: failed to load workflow from %s: %v\n", file, err)
				continue
			}
			allWorkflows = append(allWorkflows, *workflow)
		}

		migraineFile := filepath.Join(dir, "Migraine")
		if _, err := os.Stat(migraineFile); err == nil {
			workflow, err := loadProjectWorkflowFromMigraine(migraineFile)
			if err != nil {
				fmt.Printf("Warning: failed to load workflow from %s: %v\n", migraineFile, err)
			} else {
				allWorkflows = append(allWorkflows, *workflow)
			}
		}
	}

	for _, dir := range skillDirs() {
		yamlFiles, err := filepath.Glob(filepath.Join(dir, "*.y*ml"))
		if err != nil {
			continue
		}
		for _, file := range yamlFiles {
			workflow, err := LoadYAMLWorkflow(file)
			if err != nil {
				fmt.Printf("Warning: failed to load skill from %s: %v\n", file, err)
				continue
			}
			allWorkflows = append(allWorkflows, *workflow)
		}

		mgFiles, err := filepath.Glob(filepath.Join(dir, "*.mg"))
		if err != nil {
			continue
		}
		for _, file := range mgFiles {
			workflow, err := LoadMigraineWorkflow(file)
			if err != nil {
				fmt.Printf("Warning: failed to load skill from %s: %v\n", file, err)
				continue
			}
			allWorkflows = append(allWorkflows, *workflow)
		}
	}

	return allWorkflows, nil
}

// FindWorkflowByName looks for a workflow with the given name in the current working directory
func FindWorkflowByName(name string) (*YAMLWorkflow, error) {
	workflows, err := DiscoverWorkflowsFromCWD()
	if err != nil {
		return nil, err
	}

	for _, workflow := range workflows {
		// Compare names, but also consider the filename without extension as an alternative
		workflowName := workflow.Name
		if workflowName == "" {
			// If name is not set in the file, use filename
			workflowName = strings.TrimSuffix(filepath.Base(workflow.Path), filepath.Ext(workflow.Path))
		}

		if workflowName == name || strings.TrimSuffix(filepath.Base(workflow.Path), filepath.Ext(workflow.Path)) == name {
			return &workflow, nil
		}
	}

	return nil, fmt.Errorf("workflow '%s' not found in current directory", name)
}

// GetWorkflowFilePath returns the file path for a workflow with the given name
func GetWorkflowFilePath(name string) (string, error) {
	workflowDirs := []string{"./workflows", "."}

	for _, dir := range workflowDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		patterns := []string{"*.yaml", "*.yml", "*.mg"}

		for _, pattern := range patterns {
			files, err := filepath.Glob(filepath.Join(dir, pattern))
			if err != nil {
				continue
			}

			for _, file := range files {
				filename := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

				if filename == name {
					return file, nil
				}
			}
		}
	}

	for _, dir := range skillDirs() {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		patterns := []string{"*.yaml", "*.yml", "*.mg"}

		for _, pattern := range patterns {
			files, err := filepath.Glob(filepath.Join(dir, pattern))
			if err != nil {
				continue
			}

			for _, file := range files {
				filename := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
				if filename == name {
					return file, nil
				}
			}
		}
	}

	return "", fmt.Errorf("workflow file for '%s' not found", name)
}
