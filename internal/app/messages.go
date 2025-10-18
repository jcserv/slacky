package app

import (
	"context"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/constants"
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
	// Polling-related messages
	newMessagesPolledMsg struct {
		channelID string
		messages  []models.Message
	}
	channelUnreadUpdatedMsg struct {
		channelID   string
		unreadCount int
		hasUnread   bool
	}
	activitiesPolledMsg struct {
		activities []models.Activity
	}
	threadRepliesLoadedMsg struct {
		channelID     string
		threadTS      string
		parentMessage models.Message
		messages      []models.Message
	}
	threadReplySentMsg struct {
		channelID string
		threadTS  string
	}
	// Reaction messages
	reactionAddedMsg struct {
		channelID string
		timestamp string
		emojiName string
	}
	reactionRemovedMsg struct {
		channelID string
		timestamp string
		emojiName string
	}
	reactionErrorMsg struct {
		err error
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
		if cfg.Workspace.UserToken == "" {
			return errMsg(fmt.Errorf("authentication failed: no user token configured"))
		}

		client := slackClient.New(cfg.Workspace.UserToken)
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

			if ch.Type == models.ChannelTypeDM && ch.UserID != "" {
				if name, ok := globalUserCache.get(ch.UserID); ok {
					ch.UserName = name
				} else {
					user, err := client.GetUserInfo(ctx, ch.UserID)
					if err == nil {
						ch.UserName = extractUserName(user)
						ch.IsBot = user.IsBot || user.IsAppUser || user.ID == "USLACKBOT"
						globalUserCache.set(ch.UserID, ch.UserName)
					}
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

		messages := enrichMessagesWithUserInfo(ctx, client, slackMessages, channelID)

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
		mentionMessages, err := client.SearchMessages(ctx, mentionQuery, constants.MaxMessagesPerRequest)
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
		reactionItems, err := client.GetUserReactions(ctx, constants.MaxReactionsPerRequest)
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
				ch.UserName = getUserNameWithCache(ctx, client, ch.UserID)
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

// pollCurrentChannelMessages polls for new messages in the current channel
func pollCurrentChannelMessages(client *slackClient.Client, channelID, lastTimestamp string) tea.Cmd {
	return func() tea.Msg {
		if channelID == "" || lastTimestamp == "" {
			return nil
		}

		ctx := context.Background()
		slackMessages, err := client.GetConversationHistorySince(ctx, channelID, lastTimestamp, constants.MaxMessagesPerRequest)
		if err != nil {
			slog.Warn("failed to poll current channel messages",
				"channel_id", channelID,
				"error", err)
			return nil
		}

		if len(slackMessages) == 0 {
			return nil
		}

		messages := enrichMessagesWithUserInfo(ctx, client, slackMessages, channelID)

		return newMessagesPolledMsg{
			channelID: channelID,
			messages:  messages,
		}
	}
}

// pollSidebarUnreads polls for unread counts across all channels
func pollSidebarUnreads(client *slackClient.Client, channelIDs []string) tea.Cmd {
	return func() tea.Msg {
		if len(channelIDs) == 0 {
			return nil
		}

		ctx := context.Background()
		unreads, err := client.GetMultipleChannelUnreads(ctx, channelIDs)
		if err != nil {
			slog.Warn("failed to poll sidebar unreads",
				"channel_count", len(channelIDs),
				"error", err)
			return nil
		}

		// Return the first unread update (we'll batch these later if needed)
		// For now, we'll return all updates and handle them in the TUI
		var cmds []tea.Cmd
		for _, unread := range unreads {
			cmds = append(cmds, func() tea.Msg {
				return channelUnreadUpdatedMsg{
					channelID:   unread.ChannelID,
					unreadCount: unread.UnreadCount,
					hasUnread:   unread.HasUnread,
				}
			})
		}

		return tea.Batch(cmds...)()
	}
}

// pollActivitiesUpdates polls for new activities (mentions, reactions, DMs)
// This is a separate implementation from loadActivities to handle errors gracefully during polling
func pollActivitiesUpdates(client *slackClient.Client, userID string, channels []models.Channel) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		activities := []models.Activity{}

		// Create channel lookup map for quick access
		channelMap := make(map[string]models.Channel)
		for _, ch := range channels {
			channelMap[ch.ID] = ch
		}

		// 1. Get mentions (skip on error, don't break polling)
		mentionQuery := fmt.Sprintf("@%s", userID)
		mentionMessages, err := client.SearchMessages(ctx, mentionQuery, constants.MaxMessagesPerRequest)
		if err == nil {
			for _, msg := range mentionMessages {
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
					ThreadTS:    msg.Timestamp,
				}

				if exists {
					activity.ChannelType = ch.Type
				}

				activities = append(activities, activity)
			}
		}

		// 2. Get reactions (skip on error, don't break polling)
		reactionItems, err := client.GetUserReactions(ctx, constants.MaxReactionsPerRequest)
		if err == nil {
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
		}

		// 3. Get unread DMs (skip on error, don't break polling)
		unreadChannels, err := client.GetUnreadConversations(ctx)
		if err == nil {
			for _, slackCh := range unreadChannels {
				// Only include DMs, not channels
				if !slackCh.IsIM && !slackCh.IsMpIM {
					continue
				}

				ch := models.FromSlackChannel(slackCh)
				if ch.Type == models.ChannelTypeDM && ch.UserID != "" {
					ch.UserName = getUserNameWithCache(ctx, client, ch.UserID)
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
		}

		// Always return activities, even if empty (don't break polling on errors)
		return activitiesPolledMsg{
			activities: activities,
		}
	}
}

// loadThreadReplies fetches all replies in a thread
func loadThreadReplies(client *slackClient.Client, channelID, threadTS string, parentMessage models.Message) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		slackMessages, err := client.GetThreadReplies(ctx, channelID, threadTS)
		if err != nil {
			return errMsg(fmt.Errorf("failed to load thread replies for %s: %w", threadTS, err))
		}

		messages := make([]models.Message, 0, len(slackMessages))
		for _, sm := range slackMessages {
			msg := models.FromSlackMessage(sm, channelID)

			if msg.UserID != "" {
				msg.UserName = getUserNameWithCache(ctx, client, msg.UserID)
			}

			messages = append(messages, msg)
		}

		return threadRepliesLoadedMsg{
			channelID:     channelID,
			threadTS:      threadTS,
			parentMessage: parentMessage,
			messages:      messages,
		}
	}
}

// sendThreadReply sends a message as a thread reply
func sendThreadReply(client *slackClient.Client, channelID, threadTS, text string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := client.SendThreadMessage(ctx, channelID, threadTS, text)
		if err != nil {
			return errMsg(fmt.Errorf("failed to send thread reply: %w", err))
		}

		return threadReplySentMsg{
			channelID: channelID,
			threadTS:  threadTS,
		}
	}
}

// toggleReaction toggles a reaction on a message (adds if not present, removes if present)
func toggleReaction(client *slackClient.Client, channelID, timestamp, emojiName string, userID string, currentReactions []models.Reaction) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Check if the user has already reacted with this emoji
		hasReacted := false
		for _, reaction := range currentReactions {
			if reaction.Name == emojiName {
				// Check if the current user is in the list
				for _, uid := range reaction.Users {
					if uid == userID {
						hasReacted = true
						break
					}
				}
				break
			}
		}

		var err error
		if hasReacted {
			// Remove the reaction
			err = client.RemoveReaction(ctx, channelID, timestamp, emojiName)
			if err != nil {
				return reactionErrorMsg{err: fmt.Errorf("failed to remove reaction: %w", err)}
			}
			return reactionRemovedMsg{
				channelID: channelID,
				timestamp: timestamp,
				emojiName: emojiName,
			}
		} else {
			// Add the reaction
			err = client.AddReaction(ctx, channelID, timestamp, emojiName)
			if err != nil {
				return reactionErrorMsg{err: fmt.Errorf("failed to add reaction: %w", err)}
			}
			return reactionAddedMsg{
				channelID: channelID,
				timestamp: timestamp,
				emojiName: emojiName,
			}
		}
	}
}
