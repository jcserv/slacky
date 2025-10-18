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
		userID   string
	}
	channelsLoadedMsg struct {
		channels []models.Channel
	}
	starredConversationsLoadedMsg struct {
		starredIDs []string
	}
	messagesLoadedMsg struct {
		channelID string
		messages  []models.Message
	}
	messageSentMsg struct {
		channelID string
	}
	activitiesLoadedMsg struct {
		activities []models.Activity
	}
	activitySelectedMsg struct {
		channelID string
		threadTS  string // Optional: for thread replies
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
		}

		ctx := context.Background()
		authResp, err := client.TestAuth(ctx)
		if err != nil {
			return errMsg(fmt.Errorf("authentication failed: %w", err))
		}

		return authSuccessMsg{
			teamName: authResp.Team,
			userName: authResp.User,
			userID:   authResp.UserID,
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
			ch := models.FromSlackChannel(sc)

			// For DMs, fetch the user's name and check if it's a bot
			if ch.Type == models.ChannelTypeDM && ch.UserID != "" {
				user, err := client.GetUserInfo(ctx, ch.UserID)
				if err == nil {
					// Use real name if available, otherwise use display name
					if user.RealName != "" {
						ch.UserName = user.RealName
					} else if user.Profile.DisplayName != "" {
						ch.UserName = user.Profile.DisplayName
					} else {
						ch.UserName = user.Name
					}
					// Check if the user is a bot or app
					// Special case: Slackbot has user ID "USLACKBOT"
					ch.IsBot = user.IsBot || user.IsAppUser || user.ID == "USLACKBOT"
				}
			}

			channels = append(channels, ch)
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

// loadStarredConversations fetches starred conversation IDs from Slack
func loadStarredConversations(client *slackClient.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		starredIDs, err := client.GetStarredConversations(ctx)
		if err != nil {
			return errMsg(fmt.Errorf("failed to load starred conversations: %w", err))
		}

		return starredConversationsLoadedMsg{starredIDs: starredIDs}
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

// loadActivities fetches user activities from Slack
func loadActivities(client *slackClient.Client, userID string, channels []models.Channel) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		activities := []models.Activity{}

		// Create channel lookup map for quick access
		channelMap := make(map[string]models.Channel)
		for _, ch := range channels {
			channelMap[ch.ID] = ch
		}

		// 1. Get mentions
		mentionQuery := fmt.Sprintf("@%s", userID)
		mentionMessages, err := client.SearchMessages(ctx, mentionQuery, 50)
		if err != nil {
			return errMsg(fmt.Errorf("failed to load mentions: %w", err))
		}
		for _, msg := range mentionMessages {
			// Convert search message to activity
			ch, exists := channelMap[msg.Channel.ID]
			timestamp := models.ParseSlackTimestampToTime(msg.Timestamp)

			activity := models.Activity{
				Type:        models.ActivityTypeMention,
				ChannelID:   msg.Channel.ID,
				ChannelName: msg.Channel.Name,
				UserID:      msg.User,
				UserName:    msg.Username,
				MessageID:   msg.Timestamp,
				MessageText: msg.Text,
				Timestamp:   timestamp,
				ThreadTS:    msg.Timestamp, // TODO: handle actual thread TS
			}

			if exists {
				activity.ChannelType = ch.Type
			}

			activities = append(activities, activity)
		}

		// 2. Get reactions
		reactionItems, err := client.GetUserReactions(ctx, 50)
		if err != nil {
			return errMsg(fmt.Errorf("failed to load reactions: %w", err))
		}
		for _, item := range reactionItems {
			if item.Type == "message" && item.Message != nil {
				ch, exists := channelMap[item.Channel]
				timestamp := models.ParseSlackTimestampToTime(item.Message.Timestamp)

				for _, reaction := range item.Message.Reactions {
					activity := models.Activity{
						Type:          models.ActivityTypeReaction,
						ChannelID:     item.Channel,
						ChannelName:   item.Channel,
						UserID:        item.Message.User,
						UserName:      item.Message.Username,
						MessageID:     item.Message.Timestamp,
						MessageText:   item.Message.Text,
						Timestamp:     timestamp,
						ReactionName:  reaction.Name,
						ReactionCount: reaction.Count,
					}

					if exists {
						activity.ChannelName = ch.Name
						activity.ChannelType = ch.Type
					}

					activities = append(activities, activity)
				}
			}
		}

		// 3. Get unread DMs
		unreadChannels, err := client.GetUnreadConversations(ctx)
		if err != nil {
			return errMsg(fmt.Errorf("failed to load unread DMs: %w", err))
		}
		for _, slackCh := range unreadChannels {
			// Only include DMs, not channels
			if !slackCh.IsIM && !slackCh.IsMpIM {
				continue
			}

			ch := models.FromSlackChannel(slackCh)
			if ch.Type == models.ChannelTypeDM && ch.UserID != "" {
				user, err := client.GetUserInfo(ctx, ch.UserID)
				if err == nil {
					if user.RealName != "" {
						ch.UserName = user.RealName
					} else if user.Profile.DisplayName != "" {
						ch.UserName = user.Profile.DisplayName
					} else {
						ch.UserName = user.Name
					}
				}
			}

			// Get the last message as a preview
			messages, err := client.GetConversationHistory(ctx, ch.ID, 1)
			if err == nil && len(messages) > 0 {
				lastMsg := messages[0]
				timestamp := models.ParseSlackTimestampToTime(lastMsg.Timestamp)

				activity := models.Activity{
					Type:        models.ActivityTypeUnreadDM,
					ChannelID:   ch.ID,
					ChannelName: ch.UserName,
					ChannelType: ch.Type,
					UserID:      lastMsg.User,
					UserName:    ch.UserName,
					MessageID:   lastMsg.Timestamp,
					MessageText: lastMsg.Text,
					Timestamp:   timestamp,
				}

				activities = append(activities, activity)
			}
		}

		// TODO: Add thread replies (requires tracking user's thread participation)
		return activitiesLoadedMsg{activities: activities}
	}
}
