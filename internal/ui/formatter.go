package ui

import (
	"fmt"
	"strings"
	"time"
)

// WorkflowHeader displays the workflow header
func WorkflowHeader(workflowName, action string) {
	timestamp := time.Now().Format("2006-01-02 15:04")
	actionText := fmt.Sprintf("ACTION: %s", action)
	startedText := fmt.Sprintf("STARTED: %s", timestamp)

	// Calculate padding to make the header look balanced
	totalWidth := 80

	workflowText := fmt.Sprintf("WORKFLOW: %s", workflowName)

	headerLine := strings.Repeat("-", totalWidth-2) // -2 for the "+"
	header := fmt.Sprintf("+%s+\n| %s | %s | %s |\n+%s+",
		headerLine,
		padRight(workflowText, 30),
		padRight(actionText, 15),
		startedText,
		headerLine)
	fmt.Println(header)
}

// PrecheckResult displays a precheck result
func PrecheckResult(name, status string, duration time.Duration, message string) {
	statusIcon := "✓"
	statusText := "ok"
	if status == "warn" {
		statusIcon = "!"
		statusText = "warn"
	} else if status == "fail" {
		statusIcon = "✗"
		statusText = "fail"
	}

	durationStr := formatDuration(duration)
	fmt.Printf("  %s %s %s (%s)\n", statusIcon, padRight(name, 30), padRight(statusText, 4), durationStr)

	if message != "" {
		fmt.Printf("         -> %s\n", message)
	}
}

// ScriptProgress displays script progress
func ScriptProgress(current, total int, name string, duration time.Duration) {
	durationStr := formatDuration(duration)
	fmt.Printf("  (%d/%d) %s %s\n", current, total, padRight(name, 20), durationStr)
}

// ScriptOutput displays script output
func ScriptOutput(output string) {
	if output != "" {
		fmt.Printf("         -> %s\n", output)
	}
}

// Summary displays the workflow summary
func Summary(status string, duration time.Duration, prechecksPassed, prechecksFailed, prechecksWarn int, scriptsTotal, scriptsCompleted int, artifacts, endpoint string) {
	fmt.Println("\n[ SUMMARY ]")
	fmt.Printf("  Status: %s\n", status)
	fmt.Printf("  Duration: %s\n", formatDuration(duration))
	fmt.Printf("  Prechecks: %d passed, %d failed, %d warn\n", prechecksPassed, prechecksFailed, prechecksWarn)
	fmt.Printf("  Scripts: %d/%d completed\n", scriptsCompleted, scriptsTotal)
	if artifacts != "" {
		fmt.Printf("  Artifacts: %s\n", artifacts)
	}
	if endpoint != "" {
		fmt.Printf("  Endpoint: %s\n", endpoint)
	}
}

// ErrorHeader displays the error header
func ErrorHeader(action string) {
	actionText := fmt.Sprintf("ACTION: %s", action)
	totalWidth := 80
	headerLine := strings.Repeat("-", totalWidth-2) // -2 for the "+"

	header := fmt.Sprintf("+%s+\n| %s |\n+%s+",
		headerLine,
		actionText,
		headerLine)
	fmt.Println(header)
}

// ErrorBlock displays an error block
func ErrorBlock(title, command, reason, required, suggestion string) {
	fmt.Println("\n[ ERROR ]")
	fmt.Printf("  ✗ %s\n\n", title)

	contentWidth := 76 // Width inside the border (80 total width - 4 for borders and spaces)

	borderLine := strings.Repeat("─", contentWidth)
	fmt.Printf("  ┌%s┐\n", borderLine)
	fmt.Printf("  │ %-74s │\n", "COMMAND: "+command)
	fmt.Printf("  │ %-74s │\n", "REASON: "+reason)
	if required != "" {
		fmt.Printf("  │ %-74s │\n", "REQUIRED: "+required)
	}
	fmt.Printf("  │ %-74s │\n", "")
	if suggestion != "" {
		suggestionLines := strings.Split(suggestion, "\n")
		for _, line := range suggestionLines {
			if line != "" {
				fmt.Printf("  │ %-74s │\n", line)
			}
		}
	}
	fmt.Printf("  └%s┘\n", borderLine)
}

// Status displays the final status
func Status(message string) {
	fmt.Println("\n[ STATUS ]")
	fmt.Printf("  %s\n", message)
}

// padRight pads a string to the right with spaces
func padRight(str string, length int) string {
	if len(str) >= length {
		return str
	}
	return str + strings.Repeat(" ", length-len(str))
}

// formatDuration formats a duration nicely
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fμs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1000000)
	} else {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
}

// SectionHeader displays a section header
func SectionHeader(name string) {
	fmt.Printf("\n[ %s ]\n", strings.ToUpper(name))
}

// LogInfoBordered displays an info message in a bordered format
func LogInfoBordered(message string) {
	contentWidth := 76 // Width inside the border (80 total width - 4 for borders and spaces)
	borderLine := strings.Repeat("─", contentWidth)

	fmt.Printf("  ┌%s┐\n", borderLine)
	fmt.Printf("  │ %-74s │\n", "ℹ "+message)
	fmt.Printf("  └%s┘\n", borderLine)
}

// LogSuccessBordered displays a success message in a bordered format
func LogSuccessBordered(message string) {
	contentWidth := 76 // Width inside the border (80 total width - 4 for borders and spaces)
	borderLine := strings.Repeat("─", contentWidth)

	fmt.Printf("  ┌%s┐\n", borderLine)
	fmt.Printf("  │ %-74s │\n", "✓ "+message)
	fmt.Printf("  └%s┘\n", borderLine)
}

// LogWarningBordered displays a warning message in a bordered format
func LogWarningBordered(message string) {
	contentWidth := 76 // Width inside the border (80 total width - 4 for borders and spaces)
	borderLine := strings.Repeat("─", contentWidth)

	fmt.Printf("  ┌%s┐\n", borderLine)
	fmt.Printf("  │ %-74s │\n", "⚠ "+message)
	fmt.Printf("  └%s┘\n", borderLine)
}

// LogErrorBordered displays an error message in a bordered format
func LogErrorBordered(message string) {
	contentWidth := 76 // Width inside the border (80 total width - 4 for borders and spaces)
	borderLine := strings.Repeat("─", contentWidth)

	fmt.Printf("  ┌%s┐\n", borderLine)
	fmt.Printf("  │ %-74s │\n", "✗ "+message)
	fmt.Printf("  └%s┘\n", borderLine)
}
