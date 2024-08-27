package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

func CheckServiceRunning(commands []string, serviceName string) (bool, error) {
	for _, cmd := range commands {
		parts := strings.Fields(cmd)
		var output []byte
		var err error

		if len(parts) > 1 && parts[0] == "ps" {
			// Special case for the ps command
			psCmd := exec.Command("ps", "aux")
			psOutput, err := psCmd.Output()
			if err != nil {
				continue
			}
			grepCmd := exec.Command("grep", serviceName)
			grepCmd.Stdin = strings.NewReader(string(psOutput))
			output, err = grepCmd.Output()
		} else {
			// For other commands
			output, err = exec.Command(parts[0], parts[1:]...).Output()
		}

		if err == nil && len(output) > 0 {
			return true, nil
		}
	}

	return false, fmt.Errorf("%s process not found", serviceName)
}

func CheckPostgresRunning() (bool, error) {
	commands := []string{
		"pgrep -x postgres",
		"pgrep -x postmaster",
		"pgrep -f postgresql",
		"ps aux | grep postgres | grep -v grep",
		"systemctl is-active postgresql",
		"service postgresql status",
	}
	return CheckServiceRunning(commands, "postgres")
}
