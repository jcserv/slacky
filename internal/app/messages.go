package app

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/models"
	slackClient "github.com/jcserv/slacky/internal/slack"
)

// Message types for the application
type (
	errMsg         error
	authSuccessMsg struct {
		teamName string
		userName string
	}
	channelsLoadedMsg struct {
		channels []models.Channel
	}
	messagesLoadedMsg struct {
		channelID string
		messages  []models.Message
	}
	messageSentMsg struct {
		channelID string
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
		var client *slackClient.Client
		if cfg.Workspace.UserToken != "" {
			client = slackClient.New(cfg.Workspace.UserToken)
		} else {
			client = slackClient.NewWithSocketMode(cfg.Workspace.BotToken, cfg.Workspace.SocketToken)
		}

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

// loadChannels fetches all channels from Slack
func loadChannels(client *slackClient.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		slackChannels, err := client.GetChannels(ctx)
		if err != nil {
			return errMsg(fmt.Errorf("failed to load channels: %w", err))
		}

		// Convert Slack channels to our model
		channels := make([]models.Channel, 0, len(slackChannels))
		for _, sc := range slackChannels {
			channels = append(channels, models.FromSlackChannel(sc))
		}

		return channelsLoadedMsg{channels: channels}
	}
}

// loadMessages fetches message history for a channel
func loadMessages(client *slackClient.Client, channelID string, limit int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		slackMessages, err := client.GetConversationHistory(ctx, channelID, limit)
		if err != nil {
			return errMsg(fmt.Errorf("failed to load messages for channel %s: %w", channelID, err))
		}

		// Convert Slack messages to our model (reverse order - oldest first)
		messages := make([]models.Message, 0, len(slackMessages))
		for i := len(slackMessages) - 1; i >= 0; i-- {
			msg := models.FromSlackMessage(slackMessages[i], channelID)

			// Fetch user info for the message
			if msg.UserID != "" {
				user, err := client.GetUserInfo(ctx, msg.UserID)
				if err == nil {
					// Use real name if available, otherwise use display name
					if user.RealName != "" {
						msg.UserName = user.RealName
					} else if user.Profile.DisplayName != "" {
						msg.UserName = user.Profile.DisplayName
					} else {
						msg.UserName = user.Name
					}
				}
			}

			messages = append(messages, msg)
		}

		return messagesLoadedMsg{
			channelID: channelID,
			messages:  messages,
		}
	}
}

// sendMessage sends a message to a channel
func sendMessage(client *slackClient.Client, channelID, text string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := client.SendMessage(ctx, channelID, text)
		if err != nil {
			return errMsg(fmt.Errorf("failed to send message: %w", err))
		}

		return messageSentMsg{channelID: channelID}
	}
}
