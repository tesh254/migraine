package core

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/tesh254/migraine/constants"
	"github.com/tesh254/migraine/utils"
)

type VersionControl struct {
}

func (vc *VersionControl) addConfigFileToGitignore() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf(":::fs::: error getting current directory: %v\n", err)
		return
	}

	filePath := fmt.Sprintf("%s/%s", cwd, ".gitignore")

	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		if err := os.WriteFile(filePath, []byte{}, os.ModePerm); err != nil {
			log.Printf(":::fs::: error creating .gitignore file: %v\n", err)
			return
		}
		log.Printf(":::fs::: created `%s.gitignore%s`\n", utils.BOLD, utils.RESET)
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf(":::fs::: error reading .gitignore file: %v\n", err)
		return
	}

	newLine := fmt.Sprintf("%s\n", constants.CONFIG)
	if !vc.containsLine(content, constants.CONFIG) {
		content = append(content, []byte(newLine)...)

		if err := os.WriteFile(filePath, content, os.ModePerm); err != nil {
			log.Printf(":::fs::: error writing to .gitignore file: %v\n", err)
			return
		}

		log.Printf(":::fs::: successfully updated `%s.gitignore%s %s`\n", utils.BOLD, utils.RESET, utils.CHECK)
	} else {
		log.Printf(":::fs::: line already exists in `%s.gitignore%s`\n", utils.BOLD, utils.RESET)
	}
}

func (vc *VersionControl) containsLine(content []byte, line string) bool {
	for _, existingLine := range content {
		if string(existingLine) == line {
			return true
		}
	}

	return false
}
