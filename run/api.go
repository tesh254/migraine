package run

import (
	"io"
	"os"
	"os/exec"
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

func ExecuteCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)

	// Create formatted writers for stdout and stderr
	stdoutWriter := NewFormattedWriter(os.Stdout)
	stderrWriter := NewFormattedWriter(os.Stderr)

	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	return cmd.Run()
}
