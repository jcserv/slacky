package init

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/slack"
)

// Message types for the initialization wizard
type (
	authTestMsg struct {
		teamName string
		userName string
		err      error
	}

	botTokenTestMsg struct {
		teamName string
		userName string
		err      error
	}

	configSavedMsg struct{}

	errMsg struct {
		err error
	}
)

// testBotToken tests bot token authentication with Slack
func testBotToken(botToken string) tea.Cmd {
	return func() tea.Msg {
		// Use a dummy socket token for client creation (not used in auth test)
		client := slack.New(botToken, "dummy-socket-token")

		ctx := context.Background()
		authResp, err := client.TestAuth(ctx)
		if err != nil {
			return botTokenTestMsg{err: err}
		}
		return botTokenTestMsg{
			teamName: authResp.Team,
			userName: authResp.User,
		}
	}
}

// testAuth tests authentication with Slack
func testAuth(botToken, socketToken string) tea.Cmd {
	return func() tea.Msg {
		client := slack.New(botToken, socketToken)

		ctx := context.Background()
		authResp, err := client.TestAuth(ctx)
		if err != nil {
			return authTestMsg{err: err}
		}
		return authTestMsg{
			teamName: authResp.Team,
			userName: authResp.User,
		}
	}
}

// saveConfig saves the configuration to disk
func saveConfig(botToken, socketToken string, vimMode, showTimestamps bool) tea.Cmd {
	return func() tea.Msg {
		cfg := &config.Config{
			Workspace: config.Workspace{
				BotToken:    botToken,
				SocketToken: socketToken,
			},
			UI: config.UI{
				Theme:          "default",
				VimMode:        vimMode,
				ShowTimestamps: showTimestamps,
			},
		}

		if err := config.Save(cfg); err != nil {
			return errMsg{err: fmt.Errorf("failed to save config: %w", err)}
		}

		return configSavedMsg{}
	}
}
