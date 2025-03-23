package utils

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")).Bold(true)
)

// SuccessReport reports a success message to the console.
func SuccessReport(msg string) {
	_, err := fmt.Println(successStyle.Render("[SUCCESS] | ") + msg)
	if err != nil {
		panic(err)
	}
}

// ErrorReport reports an error message to the console.
func ErrorReport(msg string) {
	_, err := fmt.Println(errorStyle.Render("[ERROR] | ") + msg)
	if err != nil {
		panic(err)
	}
}
