package cmd

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/oauth"
	"github.com/jcserv/slacky/internal/tui/styles"
	"github.com/spf13/cobra"
)

// oauthCmd represents the oauth command
var oauthCmd = &cobra.Command{
	Use:   "oauth",
	Short: "Authenticate with Slack using OAuth",
	Long: `Authenticate with Slack using OAuth 2.0.

This command will:
1. Start a local OAuth callback server
2. Open your browser to Slack's authorization page
3. Wait for you to approve the app
4. Save your user token to ~/.config/slacky/config.yaml

Prerequisites:
- Set SLACK_CLIENT_ID environment variable
- Set SLACK_CLIENT_SECRET environment variable

These can be found in your Slack app's Basic Information page.`,
	RunE: runOAuth,
}

func init() {
	rootCmd.AddCommand(oauthCmd)
}

func runOAuth(cmd *cobra.Command, args []string) error {
	fmt.Println(styles.Title.Render("🔐 Slacky OAuth Authentication"))
	fmt.Println()

	// Load OAuth config from environment
	oauthCfg, err := oauth.LoadFromEnv()
	if err != nil {
		fmt.Println(styles.Error.Render("✗ Missing OAuth credentials"))
		fmt.Println()
		fmt.Println("Please set the following environment variables:")
		fmt.Println(styles.Label.Render("  SLACK_CLIENT_ID") + "     - Your Slack app's Client ID")
		fmt.Println(styles.Label.Render("  SLACK_CLIENT_SECRET") + " - Your Slack app's Client Secret")
		fmt.Println()
		fmt.Println("Find these in your Slack app's " + styles.Label.Render("Basic Information") + " page:")
		fmt.Println("  https://api.slack.com/apps")
		fmt.Println()
		return err
	}

	fmt.Println(styles.Success.Render("✓ OAuth credentials loaded"))
	fmt.Println()

	// Scopes we need
	scopes := []string{
		"channels:history",
		"channels:read",
		"channels:write",
		"chat:write",
		"groups:history",
		"groups:read",
		"groups:write",
		"im:history",
		"im:read",
		"im:write",
		"mpim:history",
		"mpim:read",
		"mpim:write",
		"users:read",
	}

	// Run OAuth flow
	fmt.Println(styles.Subtitle.Render("Starting OAuth flow..."))
	fmt.Println()

	ctx := context.Background()
	tokenResp, err := oauth.Flow(ctx, oauth.FlowOptions{
		ClientID:     oauthCfg.ClientID,
		ClientSecret: oauthCfg.ClientSecret,
		Scopes:       scopes,
		Port:         8080, // Fixed port for Slack redirect URI
		Timeout:      5 * time.Minute,
		OnURL: func(authURL string) error {
			fmt.Println(styles.Info.Render("→ Opening browser for authorization..."))
			fmt.Println()
			fmt.Println("If your browser doesn't open automatically, visit:")
			fmt.Println(styles.Dim.Render("  " + authURL))
			fmt.Println()

			// Try to open browser
			if err := oauth.OpenBrowser(authURL); err != nil {
				fmt.Println(styles.Warning.Render("⚠ Failed to open browser automatically"))
				fmt.Println("Please open the URL above manually.")
			}

			return nil
		},
		WriteSuccessHTML: writeSuccessPage,
	})

	if err != nil {
		fmt.Println()
		fmt.Println(styles.Error.Render("✗ OAuth failed: " + err.Error()))
		return err
	}

	// Save config
	fmt.Println(styles.Success.Render("✓ Successfully authenticated!"))
	fmt.Println()
	fmt.Println(styles.Label.Render("Team:") + " " + tokenResp.Team.Name)
	fmt.Println(styles.Label.Render("User:") + " " + tokenResp.AuthedUser.ID)
	fmt.Println()

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

	// Clear legacy tokens if present
	cfg.Workspace.BotToken = ""
	cfg.Workspace.SocketToken = ""

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configPath, _ := config.ConfigPath()
	fmt.Println(styles.Success.Render("✓ Configuration saved"))
	fmt.Println(styles.Dim.Render("  " + configPath))
	fmt.Println()
	fmt.Println(styles.Info.Render("You can now run") + " " + styles.Label.Bold(true).Render("slacky") + " " + styles.Info.Render("to start the app!"))
	fmt.Println()

	return nil
}

func writeSuccessPage(w io.Writer) {
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Slacky - Authentication Successful</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .container {
            background: white;
            padding: 3rem;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
            text-align: center;
            max-width: 400px;
        }
        h1 {
            color: #333;
            margin-bottom: 1rem;
        }
        p {
            color: #666;
            line-height: 1.6;
        }
        .checkmark {
            font-size: 4rem;
            color: #4CAF50;
            margin-bottom: 1rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="checkmark">✓</div>
        <h1>Authentication Successful!</h1>
        <p>You can now close this window and return to your terminal.</p>
    </div>
</body>
</html>`)
}
