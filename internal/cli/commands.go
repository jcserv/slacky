package cli

import (
	"fmt"
	"os"

	"github.com/jcserv/slacky/internal/config"
	initpkg "github.com/jcserv/slacky/internal/init"
	"github.com/jcserv/slacky/internal/styles"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Handler     func() error
}

// Commands holds all available CLI commands
var Commands = map[string]Command{
	"init": {
		Name:        "init",
		Description: "Run setup wizard to create config file",
		Handler:     handleInit,
	},
	"help": {
		Name:        "help",
		Description: "Show help message",
		Handler:     handleHelp,
	},
	"version": {
		Name:        "version",
		Description: "Show version information",
		Handler:     handleVersion,
	},
}

// handleInit runs the initialization wizard
func handleInit() error {
	result, err := initpkg.InitWithVersion(GetVersion())
	if err != nil {
		return fmt.Errorf("init failed: %w", err)
	}

	// If init completed successfully, continue to main app
	if result.ShouldContinue {
		return RunMainApp()
	}

	return nil
}

// handleHelp displays the help message
func handleHelp() error {
	PrintHelp()
	return nil
}

// handleVersion displays version information
func handleVersion() error {
	PrintVersion()
	return nil
}

// handleUnknownCommand handles unknown commands
func handleUnknownCommand(cmd string) error {
	fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
	PrintHelp()
	os.Exit(1)
	return nil // This will never be reached due to os.Exit(1)
}

// validateConfig checks if configuration exists and is valid
func validateConfig() error {
	_, err := config.Load()
	if err != nil {
		if err == config.ErrConfigNotFound {
			fmt.Println(styles.Subtitle.Render("No configuration found. Let's set up Slacky!"))
			fmt.Println()
			return handleInit()
		}
		return fmt.Errorf("error loading config: %w", err)
	}
	return nil
}
