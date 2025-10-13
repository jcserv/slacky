package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jcserv/slacky/internal/app"
	oauthWizard "github.com/jcserv/slacky/internal/app/oauth"
	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Flags().BoolP("help", "h", false, "Help")
	rootCmd.Flags().BoolP("version", "v", false, "Version")

	rootCmd.AddCommand(
		oauthCmd,
		versionCmd,
	)
}

var rootCmd = &cobra.Command{
	Use:   "slacky",
	Short: "A terminal client for Slack",
	Long: `Slacky is a terminal-based Slack client that provides an interactive interface
for managing Slack workspaces, channels, and messages directly from your terminal.`,
	Example: `
# Start the Slack client
slacky

# Authenticate with OAuth
slacky oauth

# Show version
slacky version
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := setupApp(cmd)
		if err != nil {
			return err
		}
		defer app.Shutdown()

		// Set up the TUI.
		program := tea.NewProgram(
			app.NewTUI(),
			tea.WithAltScreen(),
		)

		if _, err := program.Run(); err != nil {
			return fmt.Errorf("TUI run error: %w", err)
		}
		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// setupApp handles the common setup logic for both interactive and non-interactive modes.
func setupApp(cmd *cobra.Command) (*app.App, error) {
	ctx := cmd.Context()

	// Try to load existing config
	cfg, err := config.Load()
	if err != nil {
		if err == config.ErrConfigNotFound {
			// Auto-run OAuth wizard if config is missing
			fmt.Println("No configuration found. Let's set up Slacky with OAuth!")
			fmt.Println()

			// Run OAuth wizard
			result, err := oauthWizard.RunWithVersion(version.Version)
			if err != nil {
				return nil, fmt.Errorf("oauth setup failed: %w", err)
			}

			// If user cancelled, exit gracefully
			if !result.ShouldContinue || result.TokenResponse == nil {
				fmt.Println("OAuth setup cancelled.")
				os.Exit(0)
			}

			// Create config with OAuth token
			cfg = config.DefaultConfig()
			tokenResp := result.TokenResponse
			cfg.Workspace.UserToken = tokenResp.AccessToken
			cfg.Workspace.TeamName = tokenResp.Team.Name
			cfg.Workspace.TeamID = tokenResp.Team.ID
			cfg.Workspace.UserID = tokenResp.AuthedUser.ID

			// Save the config
			if err := config.Save(cfg); err != nil {
				return nil, fmt.Errorf("failed to save config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create app instance
	appInstance, err := app.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create app instance: %w", err)
	}

	return appInstance, nil
}
