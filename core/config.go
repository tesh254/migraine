package core

import (
	"fmt"
	"log"

	"github.com/tesh254/go-migraine/constants"
	"github.com/tesh254/go-migraine/utils"
)

type Config struct {
	Version                   string  `json:"version"`
	IsMigrationsFolderCreated bool    `json:"is_migrations_folder_created"`
	IsMigrationsTableCreated  bool    `json:"is_migrations_table_created"`
	EnvFile                   *string `json:"env_file"`
	DbUrl                     *string `json:"db_url"`
	DbVar                     *string `json:"db_var"`
	MigrationsPath            *string `json:"migrations_path"`
	HasDBUrl                  bool    `json:"has_db_url"`
}

func (config *Config) updateMigrationsConfig(IsMigrationsFolderCreated bool, isMigrationsTableCreated bool, migrationsPath *string) {
	var fs FS
	prevConfig := config.getConfig()
	configuration := Config{
		IsMigrationsFolderCreated: isMigrationsTableCreated,
		IsMigrationsTableCreated:  isMigrationsTableCreated,
		MigrationsPath:            migrationsPath,
		EnvFile:                   prevConfig.EnvFile,
		DbUrl:                     prevConfig.DbUrl,
		DbVar:                     prevConfig.DbVar,
		Version:                   constants.VERSION,
		HasDBUrl:                  prevConfig.HasDBUrl,
	}

	err := fs.writeJSONToFile(".migraine.config.json", configuration)

	if err != nil {
		log.Fatalf(":::config::: | unable to write to `%s.migraine.config.json%s`\n", utils.BOLD, utils.RESET)
	}
}

func (config *Config) updateEnvConfig(envFile string, dbVar string, dbUrl string, hasDBUrl bool) {
	var fs FS
	prevConfig := config.getConfig()

	configuration := Config{
		EnvFile:                   &envFile,
		DbUrl:                     &dbUrl,
		DbVar:                     &dbVar,
		IsMigrationsFolderCreated: prevConfig.IsMigrationsFolderCreated,
		IsMigrationsTableCreated:  prevConfig.IsMigrationsTableCreated,
		MigrationsPath:            prevConfig.MigrationsPath,
		Version:                   constants.VERSION,
		HasDBUrl:                  hasDBUrl,
	}

	err := fs.writeJSONToFile(fmt.Sprintf("%s/%s", fs.getCurrentDirectory(), constants.CONFIG), configuration)

	if err != nil {
		log.Fatalf(":::config::: | unable to write to `%s.migraine.config.json%s`\n", utils.BOLD, utils.RESET)
	}
}

func (config *Config) getConfig() Config {
	var fs FS
	prevConfig, err := fs.readJSONFromFile(fmt.Sprintf("%s/%s", fs.getCurrentDirectory(), constants.CONFIG))

	if err != nil {
		log.Fatalf(":::config::: | unable to read from `%s.migraine.config.json%s %v`\n", utils.BOLD, utils.RESET, err)
	}

	return prevConfig
}
