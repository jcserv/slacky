package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jcserv/slacky/internal/app"
)

// RunMainApp starts the main TUI application
func RunMainApp() error {
	p := tea.NewProgram(app.New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("application error: %w", err)
	}
	return nil
}

// Run handles the main application flow
func Run() error {
	// Parse command line arguments
	if len(os.Args) > 1 {
		command := os.Args[1]

		// Handle special help flags
		if command == "-h" || command == "--help" {
			command = "help"
		}
		if command == "-v" || command == "--version" {
			command = "version"
		}

		// Execute command if it exists
		if cmd, exists := Commands[command]; exists {
			return cmd.Handler()
		}

		// Handle unknown command
		return handleUnknownCommand(command)
	}

	// No command specified, validate config and run main app
	if err := validateConfig(); err != nil {
		return err
	}

	return RunMainApp()
}
