package core

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/tesh254/migraine/constants"
	"github.com/tesh254/migraine/utils"
)

func (cli *CLI) StartREPL() {
	fmt.Print(constants.MIGRAINE_ASCII)
	utils.LogInfo("Welcome to the Migraine REPL! Type 'help' for available commands or 'exit' to quit.")

	scanner := bufio.NewScanner(os.Stdin)
	var core Core

	for {
		utils.ColorPrint("blue", "▲ ")
		scanner.Scan()
		input := scanner.Text()

		args := strings.Fields(input)
		if len(args) == 0 {
			continue
		}

		command := args[0]

		switch command {
		case "exit", "quit":
			utils.LogSuccess("Goodbye!")
			return
		case "help":
			utils.ColorPrint("blue", "Available commands:")
			fmt.Println("  init                  - Initialize migraine")
			fmt.Println("  create <name>         - Create a new migration")
			fmt.Println("  run                   - Run all migrations")
			fmt.Println("  rollback              - Rollback the last migration")
			fmt.Println("  version               - Show migraine version")
			fmt.Println("  exit, quit            - Exit the REPL")
		case "init":
			core.getDatabaseURL(".env", nil)
			fs := FS{}
			fs.checkIfMigrationFolderExists()
			db := core.connection()
			core.createMigrationsTable()
			db.Close()
		case "verify":
			envPath := ""
			if len(args) > 1 {
				envPath = args[1]
			}
			exists, err := utils.CheckEnvVarExists("DATABASE_URL", envPath)
			if err != nil {
				utils.LogError(fmt.Sprintf("Error verifying DATABASE_URL: %v", err))
			} else if exists {
				utils.ColorPrint("green", "✓ DATABASE_URL is set\n")
			} else {
				utils.LogWarning("DATABASE_URL is not set")
			}
		case "create":
			if len(args) < 2 {
				utils.LogError("Please provide a name for the migration")
				continue
			}
			migrationName := strings.Join(args[1:], " ")
			db := core.connection()
			core.createMigration(migrationName)
			db.Close()
		case "run":
			db := core.connection()
			core.runAllMigrations()
			db.Close()
		case "rollback":
			db := core.connection()
			core.rollbackLastMigration()
			db.Close()
		case "version":
			utils.ColorPrint("green", constants.CurrentOSWithVersion())
		default:
			utils.LogWarning(fmt.Sprintf("Unknown command: %s", command))
			utils.LogInfo("Type 'help' for available commands")
		}
	}
}
