package utils

import (
	"fmt"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

// ColorPrint prints text in the specified color
func ColorPrint(color string, text string) {
	var colorCode string
	switch color {
	case "red":
		colorCode = colorRed
	case "green":
		colorCode = colorGreen
	case "yellow":
		colorCode = colorYellow
	case "blue":
		colorCode = colorBlue
	default:
		colorCode = colorReset
	}
	fmt.Printf("%s%s%s", colorCode, text, colorReset)
}

// LogInfo prints an info message in blue
func LogInfo(message string) {
	fmt.Printf("%s[INFO]%s %s\n", colorBlue, colorReset, message)
}

// LogSuccess prints a success message in green
func LogSuccess(message string) {
	fmt.Printf("%s[SUCCESS]%s %s\n", colorGreen, colorReset, message)
}

// LogWarning prints a warning message in yellow
func LogWarning(message string) {
	fmt.Printf("%s[WARNING]%s %s\n", colorYellow, colorReset, message)
}

// LogError prints an error message in red
func LogError(message string) {
	fmt.Printf("%s[ERROR]%s %s\n", colorRed, colorReset, message)
}
