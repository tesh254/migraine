package core

import (
	"flag"
	"fmt"
	"os"

	"github.com/tesh254/migraine/constants"
	"github.com/tesh254/migraine/kv"
	"github.com/tesh254/migraine/utils"
	"github.com/tesh254/migraine/workflow"
)

type CLI struct{}

func (cli *CLI) RunCLI() {
	var (
		workflowCommand = flag.String("workflow", "", "Workflow commands (new, execute)")
		wkCommand       = flag.String("wk", "", "Alias for workflow commands (new, execute)")
		help            = flag.Bool("help", false, "Show flag options for migraine")
		version         = flag.Bool("version", false, "Show migraine current installed version")
	)

	flag.Usage = func() {
		fmt.Print(constants.MIGRAINE_ASCII_V2)
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Println(constants.CurrentOSWithVersion())
		fmt.Print(constants.MIGRAINE_USAGE)
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

	kvDB, err := kv.InitDB("migraine")
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to initialize kv store: %v", err))
		return
	}
	defer kvDB.Close()

	store := kv.New(kvDB)
	templateStore := kv.NewTemplateStoreManager(store)
	// workflowStore := kv.NewWorkflowStore(store)

	switch {
	case *workflowCommand != "" || *wkCommand != "":
		command := *workflowCommand
		if command == "" {
			command = *wkCommand
		}
		args := flag.Args()
		if len(args) < 1 {
			utils.LogError("Insufficient arguments for workflow command")
			flag.Usage()
			return
		}
		switch command {
		case "new":
			if len(args) < 2 {
				utils.LogError("Template file path is required")
				return
			}
			workflow := workflow.WorkflowMapper{}
			templatePath := args[1]
			fmt.Println(args)
			err := workflow.CreateWorkflowTemplate(templatePath, templateStore)
			if err != nil {
				utils.LogError(fmt.Sprintf("Failed to create workflow template: %v", err))
			}
		}

	default:
		flag.Usage()
		return
	}
}
