package core

import (
	"fmt"

	"github.com/tesh254/migraine/utils"
)

func (c *Core) verify(envPath string) {
	// Check for postgresql service running
	exists, err := utils.CheckPostgresRunning()
	if err != nil {
		utils.LogError(fmt.Sprintf("Error verifying postgresql service: %v", err))
	} else if exists {
		utils.ColorPrint("green", "✓ PostgreSQL service is running\n")
	} else {
		utils.LogWarning("PostgreSQL service is not running")
	}

	// Check for DATABASE_URL
	pathToEnv := ""
	if envPath != "" {
		pathToEnv = envPath
	}
	checklist := []string{
		"DATABASE_URL",
	}
	for _, check := range checklist {
		exists, err := utils.CheckEnvVarExists(check, pathToEnv)

		if err != nil {
			utils.LogError(fmt.Sprintf("Error verifying %s: %v", check, err))
		} else if exists {
			utils.ColorPrint("green", fmt.Sprintf("✓ %s is set\n", check))
		} else {
			utils.LogWarning(fmt.Sprintf("%s is not set", check))
		}
	}
}
