package run

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

type FormattedWriter struct {
	w io.Writer
}

func (fw *FormattedWriter) Write(p []byte) (n int, err error) {
	// Gray color code and small font
	const (
		colorGray  = "\033[90m"
		fontSmall  = "\033[2m"
		resetCodes = "\033[0m"
	)

	// Write the formatting prefix first
	_, err = fw.w.Write([]byte(colorGray + fontSmall))
	if err != nil {
		return 0, err
	}

	// Write the actual content
	n, err = fw.w.Write(p)
	if err != nil {
		return n, err
	}

	// Write the reset codes
	_, err = fw.w.Write([]byte(resetCodes))
	if err != nil {
		return n, err
	}

	// Return the length of the original content
	return n, nil
}

// NewFormattedWriter creates a new FormattedWriter
func NewFormattedWriter(w io.Writer) *FormattedWriter {
	return &FormattedWriter{w: w}
}

func getDefaultShell() string {
	// First try to get the shell from SHELL environment variable
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}

	// Fallback to checking /etc/passwd for the user's shell
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

	// Default to sh as the most compatible shell
	return "/bin/sh"
}

func ExecuteCommand(command string) error {
	shell := getDefaultShell()

	// Extract the base name of the shell
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
		// Default behavior for other shells (sh, etc.)
		cmd = exec.Command(shell, "-c", command)
	}

	stdoutWriter := NewFormattedWriter(os.Stdout)
	stderrWriter := NewFormattedWriter(os.Stderr)

	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	cmd.Env = os.Environ()

	return cmd.Run()
}
