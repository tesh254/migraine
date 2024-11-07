package core

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tesh254/migraine/constants"
	"github.com/tesh254/migraine/kv"
	"github.com/tesh254/migraine/utils"
)

type CLI struct{}

func (cli *CLI) RunCLI() {
	var (
		help    = flag.Bool("help", false, "Show flag options for migraine")
		version = flag.Bool("version", false, "Show migraine current installed version")
	)

	flag.Usage = func() {
		fmt.Print(constants.MIGRAINE_ASCII_V2)
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Println(constants.CurrentOSWithVersion())
		fmt.Print(constants.MIGRAINE_USAGE)
	}

	// Custom command handling for workflow commands
	if len(os.Args) > 1 && (os.Args[1] == "workflow" || os.Args[1] == "wk") {
		handleWorkflowCommands()
		return
	}

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *version {
		fmt.Println(constants.CurrentOSWithVersion())
		return
	}

	flag.Usage()
}

func displayRecentLogs(numLines int) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	logPath := filepath.Join(homeDir, ".migraine_db", "logs", "badger.log")
	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer file.Close()

	// Create a scanner to read the file
	scanner := bufio.NewScanner(file)

	// Store the lines in a circular buffer
	lines := make([]string, numLines)
	currentIndex := 0
	totalLines := 0

	// Read all lines
	for scanner.Scan() {
		lines[currentIndex] = scanner.Text()
		currentIndex = (currentIndex + 1) % numLines
		totalLines++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log file: %v", err)
	}

	// Print the most recent lines
	fmt.Printf("\n%sRecent Badger Database Logs:%s\n\n", utils.BOLD, utils.RESET)

	numToPrint := numLines
	if totalLines < numLines {
		numToPrint = totalLines
	}

	for i := 0; i < numToPrint; i++ {
		index := (currentIndex - numToPrint + i + numLines) % numLines
		if lines[index] != "" {
			fmt.Println(lines[index])
		}
	}

	return nil
}

func handleWorkflowCommands() {
	if len(os.Args) < 3 {
		utils.LogError("Insufficient arguments for workflow command")
		fmt.Println("\nUsage:")
		fmt.Println("  mg workflow new <path_to_template_file>      Create a new workflow template")
		fmt.Println("  mg workflow templates                        List all templates")
		fmt.Println("  mg workflow templates delete <template_name> Delete a template")
		fmt.Println("  mg workflow kv logs                         Display recent Badger database logs")
		return
	}

	command := os.Args[2]

	// Initialize KV store
	kvDB, err := kv.InitDB("migraine")
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to initialize kv store: %v", err))
		return
	}
	defer kvDB.Close()

	store := kv.New(kvDB)
	templateStore := kv.NewTemplateStoreManager(store)

	switch command {
	case "new":
		if len(os.Args) < 4 {
			utils.LogError("Template file path is required")
			fmt.Println("\nUsage:")
			fmt.Println("  mg workflow new <path_to_template_file>")
			return
		}

		templatePath := os.Args[3]

		// Read and validate template file
		content, err := os.ReadFile(templatePath)
		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to read template file: %v", err))
			return
		}

		// Extract slug from filename
		slug := filepath.Base(templatePath)
		slug = slug[:len(slug)-len(filepath.Ext(slug))]

		// Clean slug - remove special characters and spaces
		slug = utils.FormatString(slug)

		// Check if template already exists
		existing, err := templateStore.GetTemplate(slug)
		if err == nil && existing != nil {
			utils.LogError(fmt.Sprintf("Template with name '%s' already exists", slug))
			return
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

		err = templateStore.CreateTemplate(template)
		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to create template: %v", err))
			return
		}

		utils.LogSuccess(fmt.Sprintf("Template '%s' created successfully", slug))

	case "templates":
		if len(os.Args) > 3 && os.Args[3] == "delete" {
			if len(os.Args) < 5 {
				utils.LogError("Template name is required for delete operation")
				fmt.Println("\nUsage:")
				fmt.Println("  mg workflow templates delete <template_name>")
				return
			}

			templateName := os.Args[4]
			err := templateStore.DeleteTemplate(templateName)
			if err != nil {
				utils.LogError(fmt.Sprintf("Failed to delete template: %v", err))
				return
			}

			utils.LogSuccess(fmt.Sprintf("Template '%s' deleted successfully", templateName))
			return
		}

		// List templates
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

	case "kv":
		if len(os.Args) < 4 || os.Args[3] != "logs" {
			utils.LogError("Invalid kv command")
			fmt.Println("\nUsage:")
			fmt.Println("  mg workflow kv logs    Display recent Badger database logs")
			return
		}

		err := displayRecentLogs(20)
		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to display logs: %v", err))
			return
		}

	default:
		utils.LogError(fmt.Sprintf("Unknown workflow command: %s", command))
		fmt.Println("\nAvailable commands:")
		fmt.Println("  mg workflow new <path_to_template_file>      Create a new workflow template")
		fmt.Println("  mg workflow templates                        List all templates")
		fmt.Println("  mg workflow templates delete <template_name> Delete a template")
		fmt.Println("  mg workflow kv logs                         Display recent Badger database logs")
	}
}
