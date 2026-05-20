package execution

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetDefaultShell_EnvVar(t *testing.T) {
	origShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", origShell)

	os.Setenv("SHELL", "/bin/zsh")
	shell := getDefaultShell()
	if shell != "/bin/zsh" {
		t.Errorf("expected /bin/zsh, got %s", shell)
	}
}

func TestGetDefaultShell_CustomPath(t *testing.T) {
	origShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", origShell)

	os.Setenv("SHELL", "/usr/local/bin/fish")
	shell := getDefaultShell()
	if shell != "/usr/local/bin/fish" {
		t.Errorf("expected /usr/local/bin/fish, got %s", shell)
	}
}

func TestExecuteCommand_Echo(t *testing.T) {
	if os.Getenv("GO_TEST_WAS_RUN") != "" {
		return
	}

	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("skipping PTY test in CI environment")
	}

	tmpScript := filepath.Join(t.TempDir(), "test_script.sh")
	content := []byte("#!/bin/sh\necho hello\n")
	if err := os.WriteFile(tmpScript, content, 0755); err != nil {
		t.Fatal(err)
	}

	err := ExecuteCommand("sh " + tmpScript)
	if err != nil {
		t.Errorf("ExecuteCommand failed: %v", err)
	}
}

func TestGetDefaultShell_Fallback(t *testing.T) {
	origShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", origShell)

	os.Setenv("SHELL", "")

	shell := getDefaultShell()
	if shell == "" {
		t.Error("expected a fallback shell, got empty string")
	}
	if !strings.HasPrefix(shell, "/") {
		t.Errorf("expected absolute path, got %s", shell)
	}
}