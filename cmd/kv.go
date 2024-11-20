package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tesh254/migraine/utils"
)

var kvCmd = &cobra.Command{
	Use:     "kv",
	Aliases: []string{"store"},
	Short:   "Manage KV store",
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Returns recent kv store logs",
	Run: func(cmd *cobra.Command, args []string) {
		err := displayRecentLogs(20)

		if err != nil {
			utils.LogError(fmt.Sprintf("Failed to display logs: %v", err))
		}
	},
}

func init() {
	rootCmd.AddCommand(kvCmd)
	kvCmd.AddCommand(logsCmd)
}

func displayRecentLogs(numLines int) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	logPath := filepath.Join(homeDir, ".migraine_db", "logs", "badger.log")
	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer file.Close()

	// Create a scanner to read the file
	scanner := bufio.NewScanner(file)

	// Store the lines in a circular buffer
	lines := make([]string, numLines)
	currentIndex := 0
	totalLines := 0

	// Read all lines
	for scanner.Scan() {
		lines[currentIndex] = scanner.Text()
		currentIndex = (currentIndex + 1) % numLines
		totalLines++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log file: %v", err)
	}

	// Print the most recent lines
	fmt.Printf("\n%sRecent Badger Database Logs:%s\n\n", utils.BOLD, utils.RESET)

	numToPrint := numLines
	if totalLines < numLines {
		numToPrint = totalLines
	}

	for i := 0; i < numToPrint; i++ {
		index := (currentIndex - numToPrint + i + numLines) % numLines
		if lines[index] != "" {
			fmt.Println(lines[index])
		}
	}

	return nil
}
