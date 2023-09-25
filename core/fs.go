package core

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/tesh254/go-migraine/constants"
	"github.com/tesh254/go-migraine/utils"
)

type FS struct{}

func (f *FS) getCurrentDirectory() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(":::current_dir:::| error loading current directory:", err)
	}

	return cwd
}

func (f *FS) addConfigFileToGitignore() {
	cwd := f.getCurrentDirectory()

	filePath := fmt.Sprintf("%s/%s", cwd, ".gitignore")

	if _, err := os.Stat(filePath); err != nil {
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

		if err != nil {
			log.Printf(":::fs::: error opening .gitignore file please add `%s%s%s` to your `%s.gitignore%s`\n", utils.BOLD, constants.CONFIG, utils.RESET, utils.BOLD, utils.RESET)
			return
		}
		log.Printf(":::fs::: found `%s.gitignore%s` updating...\n", utils.BOLD, utils.RESET)

		defer f.Close()

		newLine := fmt.Sprintf("%s\n", constants.CONFIG)
		if _, err := f.WriteString(newLine); err != nil {
			log.Printf(":::fs::: error opening .gitignore file please add `%s%s%s` to your `%s.gitignore%s`\n", utils.BOLD, constants.CONFIG, utils.RESET, utils.BOLD, utils.RESET)
			return
		}
		log.Printf(":::fs::: successfully updated your `%s.gitignore%s %s`\n", utils.BOLD, utils.RESET, utils.CHECK)
		return
	}
}

func (f *FS) checkIfConfigFileExistsCreateIfNot() {
	var config Config
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
			DbVar:                     nil,
			HasDBUrl:                  false,
		}

		f.writeJSONToFile(filePath, config)
		f.addConfigFileToGitignore()
	} else if err != nil {
		log.Fatalf(":::fs::: error while checking file: %v\n", err)
	} else {
		log.Printf(":::fs::: `%s%s%s` exists %s\n", utils.BOLD, constants.CONFIG, utils.RESET, utils.CHECK)
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
