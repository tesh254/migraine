package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tesh254/migraine/kv"
	"github.com/tesh254/migraine/utils"
	"github.com/tesh254/migraine/workflow"
)

func handleNewTemplate(templatePath string) error {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %v", err)
	}

	// Extract slug from filename
	slug := filepath.Base(templatePath)
	slug = slug[:len(slug)-len(filepath.Ext(slug))]
	slug = utils.FormatString(slug)

	// Initialize KV store
	kvDB, err := kv.InitDB("migraine")
	if err != nil {
		return fmt.Errorf("failed to initialize kv store: %v", err)
	}
	defer kvDB.Close()

	store := kv.New(kvDB)
	templateStore := kv.NewTemplateStoreManager(store)

	// Check if template exists
	existing, err := templateStore.GetTemplate(slug)
	if err == nil && existing != nil {
		return fmt.Errorf("template with name '%s' already exists", slug)
	}

	// Extract variables
	variables := utils.ExtractTemplateVars(string(content))
	if len(variables) > 0 {
		utils.LogInfo("Template variables detected:")
		for _, v := range variables {
			fmt.Printf("  • %s\n", v)
		}
		fmt.Println()
	}

	// Create template
	template := kv.TemplateItem{
		Slug:     slug,
		Workflow: string(content),
	}

	if err := templateStore.CreateTemplate(template); err != nil {
		return fmt.Errorf("failed to create template: %v", err)
	}

	utils.LogSuccess(fmt.Sprintf("Template '%s' created successfully", slug))
	return nil
}

func handleListTemplates() {
	kvDB, err := kv.InitDB("migraine")
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to initialize kv store: %v", err))
		return
	}
	defer kvDB.Close()

	store := kv.New(kvDB)
	templateStore := kv.NewTemplateStoreManager(store)

	templates, err := templateStore.ListTemplates()
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to list templates: %v", err))
		return
	}

	if len(templates) == 0 {
		fmt.Println("No templates found")
		return
	}

	fmt.Println("\nAvailable templates:")
	for _, t := range templates {
		fmt.Printf("  • %s\n", t.Slug)
	}
}

func handleDeleteTemplate(templateName string) error {
	kvDB, err := kv.InitDB("migraine")
	if err != nil {
		return fmt.Errorf("failed to initialize kv store: %v", err)
	}
	defer kvDB.Close()

	store := kv.New(kvDB)
	templateStore := kv.NewTemplateStoreManager(store)

	if err := templateStore.DeleteTemplate(templateName); err != nil {
		return fmt.Errorf("failed to delete template: %v", err)
	}

	utils.LogSuccess(fmt.Sprintf("Template '%s' deleted successfully", templateName))
	return nil
}

func handleAddWorkflow() {
	kvDB, err := kv.InitDB("migraine")
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to initialize kv store: %v", err))
		return
	}
	defer kvDB.Close()

	store := kv.New(kvDB)
	templateStore := kv.NewTemplateStoreManager(store)
	workflowStore := kv.NewWorkflowStore(store)

	templates, err := templateStore.ListTemplates()
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to list templates: %v", err))
		os.Exit(1)
	}

	if len(templates) == 0 {
		utils.LogError(fmt.Sprintf("No template found. Fork or Create a template."))
		os.Exit(1)
	}

	fmt.Println("\nAvailable templates:")
	for i, t := range templates {
		fmt.Printf("[%d] %s\n", i+1, t.Slug)
	}

	// Template selection
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nEnter template number: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to read input: %v", err))
		os.Exit(1)
	}

	// Parse template selection
	selection, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || selection < 1 || selection > len(templates) {
		utils.LogError("Invalid template selection")
		os.Exit(1)
	}

	selectedTemplate := templates[selection-1]

	// Parse template into workflow to access config
	parser := workflow.NewTemplateParser(selectedTemplate.Workflow)
	wk, err := parser.ParseToWorkflow()
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to parse workflow template: %v", err))
		os.Exit(1)
	}

	// Get workflow name first
	fmt.Print("\nEnter workflow name: ")
	workflowName, err := reader.ReadString('\n')
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to read input: %v", err))
		os.Exit(1)
	}

	workflowName = strings.TrimSpace(workflowName)
	if workflowName == "" {
		utils.LogError("Workflow name cannot be empty")
		os.Exit(1)
	}

	slugifiedName := utils.FormatString(workflowName)

	// Initialize the workflow with basic information
	kvWorkflow := kv.Workflow{
		ID:          slugifiedName, // Add the ID field
		Name:        wk.Name,
		Steps:       make([]kv.Atom, len(wk.Steps)),
		Description: wk.Description,
		Actions:     make(map[string]kv.Atom),
		Config: kv.Config{
			Variables:      wk.Config.Variables,
			StoreVariables: wk.Config.StoreVariables,
		},
		UsesSudo: false,
	}

	// Handle variables based on StoreVariables config
	if wk.Config.StoreVariables {
		// Extract variables and get their values
		variables := utils.ExtractTemplateVars(selectedTemplate.Workflow)
		if len(variables) > 0 {
			fmt.Printf("\nEnter variables:\n")
			variableValues := make(map[string]string)

			for _, v := range variables {
				fmt.Printf("%s: ", v)
				value, err := reader.ReadString('\n')
				if err != nil {
					utils.LogError(fmt.Sprintf("Failed to read input: %v", err))
					os.Exit(1)
				}

				value = strings.TrimSpace(value)
				if value == "" {
					utils.LogError(fmt.Sprintf("Variable %s cannot be empty", v))
					os.Exit(1)
				}

				// Process variable based on config rules
				if rules, exists := wk.Config.Variables[v]; exists {
					if rulesArray, ok := rules.([]interface{}); ok {
						for _, rule := range rulesArray {
							if ruleStr, ok := rule.(string); ok {
								switch ruleStr {
								case "slugify":
									value = utils.FormatString(value)
								}
							}
						}
					}
				}

				variableValues[v] = value
			}

			// Process Steps with variables
			for i, step := range wk.Steps {
				command, err := utils.ApplyVariablesToCommand(step.Command, variableValues)
				if err != nil {
					utils.LogError(fmt.Sprintf("Failed to process step command: %v", err))
					os.Exit(1)
				}

				kvWorkflow.Steps[i] = kv.Atom{
					Command:     command,
					Description: step.Description,
				}
			}

			// Process Actions with variables
			for key, action := range wk.Actions {
				command, err := utils.ApplyVariablesToCommand(action.Command, variableValues)
				if err != nil {
					utils.LogError(fmt.Sprintf("Failed to process action command: %v", err))
					os.Exit(1)
				}

				kvWorkflow.Actions[key] = kv.Atom{
					Command:     command,
					Description: action.Description,
				}
			}
		}
	} else {
		// If StoreVariables is false, copy steps and actions as is
		for i, step := range wk.Steps {
			kvWorkflow.Steps[i] = kv.Atom{
				Command:     step.Command,
				Description: step.Description,
			}
		}

		for key, action := range wk.Actions {
			kvWorkflow.Actions[key] = kv.Atom{
				Command:     action.Command,
				Description: action.Description,
			}
		}
	}

	// Create the workflow
	err = workflowStore.CreateWorkflow(slugifiedName, kvWorkflow)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to create workflow: %v", err))
		os.Exit(1)
	}

	utils.LogSuccess(fmt.Sprintf("Workflow '%s' created successfully", slugifiedName))
}

func handleListWorkflows() {
	kvDB, err := kv.InitDB("migraine")
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to initialize kv store: %v", err))
		return
	}
	defer kvDB.Close()

	store := kv.New(kvDB)
	workflowStore := kv.NewWorkflowStore(store)

	workflows, err := workflowStore.ListWorkflows()
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to list workflows: %v", err))
		return
	}

	if len(workflows) == 0 {
		fmt.Println("\nNo workflows found")
		return
	}

	fmt.Printf("\n%sAvailable Workflows:%s\n\n", utils.BOLD, utils.RESET)

	for i, workflow := range workflows {
		// Print workflow name, ID (slug), and description
		fmt.Printf("%d. (%s%s%s) %s\n", i+1, utils.BOLD, workflow.ID, utils.RESET, workflow.Name)
		if workflow.Description != nil && *workflow.Description != "" {
			fmt.Printf("   Description: %s\n", *workflow.Description)
		}

		// Extract and display unique variables from steps and actions
		variables := make(map[string]bool)

		// Extract from steps
		for _, step := range workflow.Steps {
			vars := utils.ExtractTemplateVars(step.Command)
			for _, v := range vars {
				variables[v] = true
			}
		}

		// Extract from actions
		for _, action := range workflow.Actions {
			vars := utils.ExtractTemplateVars(action.Command)
			for _, v := range vars {
				variables[v] = true
			}
		}

		// Display variables if any exist
		if len(variables) > 0 {
			fmt.Printf("   Required Variables:\n")
			for v := range variables {
				fmt.Printf("   • %s\n", v)
			}
		}

		// Add a newline between workflows for better readability
		fmt.Println()
	}
}

func handleDeleteWorkflow(workflowId string) {
	kvDB, err := kv.InitDB("migraine")
	if err != nil {
		utils.LogError(fmt.Sprintf("failed to initialize kv store: %v", err))
	}
	defer kvDB.Close()

	store := kv.New(kvDB)
	workflowStore := kv.NewWorkflowStore(store)

	if err := workflowStore.DeleteWorkflow(workflowId); err != nil {
		utils.LogError(fmt.Sprintf("failed to delete workflow: %v", err))
	}

	utils.LogSuccess(fmt.Sprintf("Workflow '%s' deleted successfully", workflowId))
}
