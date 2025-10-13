package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jcserv/slacky/internal/app"
	oauthWizard "github.com/jcserv/slacky/internal/app/oauth"
	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/version"
	"github.com/spf13/cobra"
)

// oauthCmd represents the oauth command
var oauthCmd = &cobra.Command{
	Use:   "oauth",
	Short: "Authenticate with Slack using OAuth",
	Long: `Authenticate with Slack using OAuth 2.0.

This command will guide you through:
1. Entering your Slack app credentials (Client ID and Client Secret)
2. Starting a local OAuth callback server
3. Opening your browser to Slack's authorization page
4. Waiting for you to approve the app
5. Saving your user token to ~/.config/slacky/config.yaml

You can find your Client ID and Client Secret in your Slack app's Basic Information page:
https://api.slack.com/apps`,
	RunE: runOAuth,
}

func init() {
	rootCmd.AddCommand(oauthCmd)
}

func runOAuth(cmd *cobra.Command, args []string) error {
	// Run the OAuth wizard
	result, err := oauthWizard.RunWithVersion(version.Version)
	if err != nil {
		return fmt.Errorf("oauth wizard failed: %w", err)
	}

	// If user cancelled, exit gracefully
	if !result.ShouldContinue || result.TokenResponse == nil {
		fmt.Println("OAuth setup cancelled.")
		return nil
	}

	tokenResp := result.TokenResponse

	// Load or create config
	cfg, err := config.Load()
	if err != nil {
		if err == config.ErrConfigNotFound {
			cfg = config.DefaultConfig()
		} else {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Update with OAuth token
	cfg.Workspace.UserToken = tokenResp.AccessToken
	cfg.Workspace.TeamName = tokenResp.Team.Name
	cfg.Workspace.TeamID = tokenResp.Team.ID
	cfg.Workspace.UserID = tokenResp.AuthedUser.ID

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// If OAuth completed successfully, continue to main app
	if result.ShouldContinue {
		fmt.Println("Setup complete! Starting Slacky...")
		fmt.Println()

		// Validate config
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}

		// Create and run the main app
		return runMainAppOAuth(cmd, cfg)
	}

	return nil
}

// runMainAppOAuth creates and runs the main application TUI after OAuth
func runMainAppOAuth(cmd *cobra.Command, cfg *config.Config) error {
	ctx := cmd.Context()

	// Create app instance
	appInstance, err := app.New(ctx, cfg)
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
