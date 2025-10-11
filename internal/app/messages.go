package app

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/slack"
)

// Message types for the application
type (
	errMsg         error
	authSuccessMsg struct {
		teamName string
		userName string
	}
)

// loadConfig loads and validates the configuration
func loadConfig(app *App) tea.Cmd {
	return func() tea.Msg {
		// Use the app's config directly since it's already loaded and validated
		return checkAuth(app.Config())()
	}
}

// checkAuth tests authentication with Slack
func checkAuth(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		client := slack.New(cfg.Workspace.BotToken, cfg.Workspace.SocketToken)

		ctx := context.Background()
		authResp, err := client.TestAuth(ctx)
		if err != nil {
			return errMsg(fmt.Errorf("authentication failed: %w", err))
		}

		return authSuccessMsg{
			teamName: authResp.Team,
			userName: authResp.User,
		}
	}
}
