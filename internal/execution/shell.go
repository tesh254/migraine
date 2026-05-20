package execution

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

func getDefaultShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}

	if currentUser, err := user.Current(); err == nil {
		file, err := os.Open("/etc/passwd")
		if err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				fields := strings.Split(scanner.Text(), ":")
				if len(fields) >= 7 && fields[0] == currentUser.Username {
					if shell := fields[6]; shell != "" {
						return shell
					}
				}
			}
		}
	}

	return "/bin/sh"
}

func ExecuteCommand(command string) error {
	shell := getDefaultShell()

	cmd := exec.Command(shell, "-c", command)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}
