package core

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/tesh254/go-migraine/utils"
)

func (c *Core) getDatabaseURL(envFile string, dbVar *string) {
	var fs FS
	var config Config

	cwd := fs.getCurrentDirectory()
	pathToEnv := fmt.Sprintf("%s/%s", cwd, envFile)
	fs.checkIfEnvFileExists(pathToEnv)

	projectEnv, err := godotenv.Read(pathToEnv)
	if err != nil {
		log.Fatalf(":::env::: failed to load env: %v", err)
	}

	key := "DATABASE_URL"

	if config.getConfig().DbUrl != nil {
		key = *config.getConfig().DbUrl
	}

	if dbVar != nil {
		key = *dbVar
	}

	dbUrl, exists := projectEnv[key]

	if !exists {
		log.Fatalf(":::env::: var %s`%s=`%s does not exist in your %s`%s`%s file, please countercheck\n", utils.BOLD, key, utils.RESET, utils.BOLD, pathToEnv, utils.RESET)
	}

	if len(dbUrl) == 0 {
		log.Fatalf(":::env::: var `%s=` provided but it's empty\n", key)
	}

	config.updateEnvConfig(envFile, *dbVar, dbUrl, true)
}
