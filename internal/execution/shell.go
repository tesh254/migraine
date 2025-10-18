package execution

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/creack/pty"
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
	shellName := filepath.Base(shell)

	var cmd *exec.Cmd
	switch shellName {
	case "bash":
		cmd = exec.Command(shell, "--login", "-c", command)
	case "zsh":
		cmd = exec.Command(shell, "-l", "-c", command)
	case "fish":
		cmd = exec.Command(shell, "-l", "-c", command)
	default:
		cmd = exec.Command(shell, "-c", command)
	}

	cmd.Env = os.Environ()

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}
	defer ptmx.Close()

	const (
		colorGray  = "\033[90m"
		fontSmall  = "\033[2m"
		resetCodes = "\033[0m"
	)

	fmt.Fprint(os.Stdout, colorGray+fontSmall)

	// Create channels to handle both directions of data flow
	stdoutErrCh := make(chan error, 1)
	stdinErrCh := make(chan error, 1)

	// Copy pty output to stdout
	go func() {
		_, err := io.Copy(os.Stdout, ptmx)
		stdoutErrCh <- err
	}()

	// Copy stdin to pty input
	go func() {
		_, err := io.Copy(ptmx, os.Stdin)
		stdinErrCh <- err
	}()

	// Wait for command to finish
	err = cmd.Wait()

	// Close the PTY to terminate the copying goroutines
	_ = ptmx.Close()

	// Wait for copying to complete and capture errors
	stdoutErr := <-stdoutErrCh
	stdinErr := <-stdinErrCh

	fmt.Fprint(os.Stdout, resetCodes)

	// Return the first error encountered or nil if no error
	if err != nil {
		return err
	}
	if stdoutErr != nil {
		return stdoutErr
	}
	if stdinErr != nil {
		return stdinErr
	}

	return nil
}
