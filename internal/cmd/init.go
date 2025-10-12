package cmd

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jcserv/slacky/internal/app"
	_init "github.com/jcserv/slacky/internal/app/init"
	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/version"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Run setup wizard to create config file",
	Long: `Run the interactive setup wizard to configure Slacky.
This will guide you through setting up your Slack bot token and app token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := runInitWizard()
		if err != nil {
			return fmt.Errorf("init failed: %w", err)
		}

		// If init completed successfully, continue to main app
		if result.ShouldContinue {
			fmt.Println("Setup complete! Starting Slacky...")
			fmt.Println()

			// Load the newly created config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config after init: %w", err)
			}

			// Validate config
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("invalid config: %w", err)
			}

			// Create and run the main app
			return runMainApp(cmd, cfg)
		}

		return nil
	},
}

// runInitWizard runs the initialization wizard
func runInitWizard() (*InitResult, error) {
	// Call the internal/init package
	result, err := _init.InitWithVersion(version.Version)
	if err != nil {
		return nil, err
	}
	return &InitResult{
		ShouldContinue: result.ShouldContinue,
	}, nil
}

// InitResult represents the result of the initialization process
type InitResult struct {
	ShouldContinue bool
}

// runMainApp creates and runs the main application TUI
func runMainApp(cmd *cobra.Command, cfg *config.Config) error {
	ctx := cmd.Context()

	// Create app instance
	appInstance, err := setupAppWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create app instance: %w", err)
	}
	defer appInstance.Shutdown()

	// Set up the TUI
	program := tea.NewProgram(
		appInstance.NewTUI(),
		tea.WithAltScreen(),
	)

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("TUI run error: %w", err)
	}
	return nil
}

// setupAppWithConfig creates an app instance with the given config
func setupAppWithConfig(ctx context.Context, cfg *config.Config) (*app.App, error) {
	appInstance, err := app.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create app instance: %w", err)
	}
	return appInstance, nil
}
