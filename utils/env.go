package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func CheckEnvVarExists(varName string, envFilePath string) (bool, error) {
	if _, exists := os.LookupEnv(varName); exists {
		return true, nil
	}

	if envFilePath == "" {
		envFilePath = ".env"
	}

	file, err := os.Open(envFilePath)
	if err != nil {
		return false, fmt.Errorf("error opening .env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, varName+"=") {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading .env file: %w", err)
	}

	return false, nil
}
