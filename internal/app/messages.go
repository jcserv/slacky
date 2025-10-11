package app

import (
	"context"
	"errors"
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
func loadConfig() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load()
		if err != nil {
			if errors.Is(err, config.ErrConfigNotFound) {
				return errMsg(fmt.Errorf("config not found. Run 'slacky init' to set up your configuration"))
			}
			return errMsg(err)
		}

		if err := cfg.Validate(); err != nil {
			return errMsg(fmt.Errorf("invalid config: %w", err))
		}

		return checkAuth(cfg)()
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
