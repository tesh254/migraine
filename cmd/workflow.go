package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tesh254/migraine/kv"
	"github.com/tesh254/migraine/utils"
)

var workflowCmd = &cobra.Command{
	Use:     "workflow",
	Aliases: []string{"wk"},
	Short:   "Manage workflow templates",
}

var workflowNewCmd = &cobra.Command{
	Use:   "new [template_file]",
	Short: "Create a new workflow template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templatePath := args[0]
		return handleNewWorkflow(templatePath)
	},
}

var workflowTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "List all templates",
	Run: func(cmd *cobra.Command, args []string) {
		handleListTemplates()
	},
}

var workflowTemplatesDeleteCmd = &cobra.Command{
	Use:   "delete [template_name]",
	Short: "Delete a template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleDeleteTemplate(args[0])
	},
}

func init() {
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.AddCommand(workflowNewCmd)
	workflowCmd.AddCommand(workflowTemplatesCmd)
	workflowTemplatesCmd.AddCommand(workflowTemplatesDeleteCmd)
}

func handleNewWorkflow(templatePath string) error {
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
