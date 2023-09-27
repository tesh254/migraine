package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tesh254/migraine/constants"
	"github.com/tesh254/migraine/utils"
)

type FS struct{}

func (f *FS) getCurrentDirectory() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(":::current_dir:::| error loading current directory:", err)
	}

	return cwd
}

func (f *FS) migrationSqlFileParser(migrationFilename string, isRollback bool) string {
	cwd := f.getCurrentDirectory()
	migrationPath := fmt.Sprintf("%s/migrations/%s", cwd, migrationFilename)
	file, err := os.Open(migrationPath)

	if err != nil {
		log.Fatalf(":::fs::: unable to read `%s%s%s` file: %v\n", utils.BOLD, migrationPath, utils.RESET, err)
	}

	var migrationLines []string
	var inMigrationBlock bool

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "--migraine-up") {
			inMigrationBlock = !isRollback
			continue
		}

		if strings.Contains(line, "--migraine-down") {
			inMigrationBlock = isRollback
			continue
		}

		if inMigrationBlock {
			migrationLines = append(migrationLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf(":::fs::: unable to file contents for `%s%s%s`: %v\n", utils.BOLD, migrationPath, utils.RESET, err)
	}

	sqlStmt := strings.Join(migrationLines, "\n")

	return sqlStmt
}

func (f *FS) checkIfConfigFileExistsCreateIfNot() {
	var config Config
	var vc VersionControl
	cwd := f.getCurrentDirectory()

	filePath := fmt.Sprintf("%s/%s", cwd, constants.CONFIG)

	_, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		log.Printf(":::fs::: `%s%s%s` doesn't exists, creating %s\n", utils.BOLD, constants.CONFIG, utils.RESET, utils.CHECK)
		file, err := os.Create(filePath)
		if err != nil {
			log.Fatalf(":::fs::: unable to create config file %v\n", err)
		}

		defer file.Close()
		config = Config{
			Version:                   constants.VERSION,
			IsMigrationsFolderCreated: false,
			IsMigrationsTableCreated:  false,
			EnvFile:                   nil,
			MigrationsPath:            nil,
			DbUrl:                     nil,
			HasDBUrl:                  false,
		}

		f.writeJSONToFile(filePath, config)
		vc.addConfigFileToGitignore()
	} else if err != nil {
		log.Fatalf(":::fs::: error while checking file: %v\n", err)
	} else {
		log.Printf(":::fs::: using existing `%s%s%s` %s\n", utils.BOLD, constants.CONFIG, utils.RESET, utils.CHECK)
	}
}

func (f *FS) checkIfMigrationFolderExists() bool {
	var fs FS
	migrationFolderName := "migrations"
	migrationPath := fmt.Sprintf("%s/%s", fs.getCurrentDirectory(), migrationFolderName)
	var config Config

	_, err := os.Stat(migrationPath)

	if os.IsNotExist(err) {
		log.Printf(":::migrations::: | %s`./migrations`%s folder does not exist; creating...\n", utils.BOLD, utils.RESET)

		if err := os.Mkdir(migrationPath, os.ModePerm); err != nil {
			log.Fatalf(":::migrations::: | unable to create %s`./migrations`%s folder: %v\n", utils.BOLD, utils.RESET, err)
		}

		log.Printf(":::migrations::: | folder created; ready for migrations; %s\n", utils.CHECK)
	} else if err != nil {
		log.Fatalf(":::migrations::: | error checking directory %v\n", err)
	} else {
		log.Println(":::migrations::: | ready for migrations;", utils.CHECK)
	}

	config.updateMigrationsConfig(true, false, &migrationFolderName)
	return true
}

func (f *FS) checkIfEnvFileExists(envPath string) {
	_, err := os.Stat(envPath)

	if err == nil {
		log.Printf(":::env::: | checking `%s%s%s`\n", utils.BOLD, envPath, utils.RESET)
	} else if os.IsNotExist(err) {
		log.Fatalf(":::env::: | file %s`%s`%s does not exist\n", utils.BOLD, utils.RESET, envPath)
	} else {
		log.Fatalf(":::env::: | error checking %s`%s`%s: %v\n", utils.BOLD, envPath, utils.RESET, err)
	}
}

func (f *FS) writeJSONToFile(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func (f *FS) readJSONFromFile(filename string) (Config, error) {
	var config Config

	file, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
