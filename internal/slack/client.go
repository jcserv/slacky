package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// Client wraps the Slack API client
type Client struct {
	api    *slack.Client
	socket *socketmode.Client // Optional, only for bot mode
}

// New creates a new Slack client with a user token
func New(userToken string) *Client {
	api := slack.New(userToken)

	return &Client{
		api:    api,
		socket: nil, // User tokens don't use Socket Mode
	}
}

// TestAuth tests the connection and returns auth info
func (c *Client) TestAuth(ctx context.Context) (*slack.AuthTestResponse, error) {
	resp, err := c.api.AuthTestContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth test failed: %w", err)
	}
	return resp, nil
}

// GetChannels retrieves all channels and DMs the user has access to
func (c *Client) GetChannels(ctx context.Context) ([]slack.Channel, error) {
	var allChannels []slack.Channel
	params := &slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel", "im", "mpim"},
		Limit: 100,
	}

	for {
		channels, nextCursor, err := c.api.GetConversationsContext(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("failed to get channels: %w", err)
		}

		allChannels = append(allChannels, channels...)

		if nextCursor == "" {
			break
		}
		params.Cursor = nextCursor
	}

	return allChannels, nil
}

// GetConversationHistory retrieves messages from a channel
func (c *Client) GetConversationHistory(ctx context.Context, channelID string, limit int) ([]slack.Message, error) {
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     limit,
	}

	history, err := c.api.GetConversationHistoryContext(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}

	return history.Messages, nil
}

// SendMessage sends a message to a channel
func (c *Client) SendMessage(ctx context.Context, channelID, text string) error {
	_, _, err := c.api.PostMessageContext(ctx, channelID, slack.MsgOptionText(text, false))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// Run starts the Socket Mode client and processes events
// Only works if client was created with NewWithSocketMode
func (c *Client) Run(ctx context.Context) error {
	if c.socket == nil {
		return fmt.Errorf("Socket Mode not available (client created with user token)")
	}
	return c.socket.RunContext(ctx)
}

// Events returns the channel for receiving Socket Mode events
// Only works if client was created with NewWithSocketMode
func (c *Client) Events() chan socketmode.Event {
	if c.socket == nil {
		return nil
	}
	return c.socket.Events
}

// HasSocketMode returns true if this client has Socket Mode enabled
func (c *Client) HasSocketMode() bool {
	return c.socket != nil
}

// API returns the underlying Slack API client
func (c *Client) API() *slack.Client {
	return c.api
}

// Socket returns the underlying Socket Mode client
func (c *Client) Socket() *socketmode.Client {
	return c.socket
}

// GetUserInfo retrieves information about a user
func (c *Client) GetUserInfo(ctx context.Context, userID string) (*slack.User, error) {
	user, err := c.api.GetUserInfoContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info for %s: %w", userID, err)
	}
	return user, nil
}

// GetStarredConversations retrieves channel IDs for all starred conversations (channels, DMs, groups)
func (c *Client) GetStarredConversations(ctx context.Context) ([]string, error) {
	items, err := c.api.ListAllStarsContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get starred items: %w", err)
	}

	// Use a map to deduplicate channel IDs
	starredChannelSet := make(map[string]bool)
	for _, item := range items {
		// Only include direct conversation stars (not messages within channels)
		// This ensures we only show channels that are explicitly starred
		switch item.Type {
		case slack.TYPE_CHANNEL, slack.TYPE_IM, slack.TYPE_GROUP:
			if item.Channel != "" {
				starredChannelSet[item.Channel] = true
			}
		}
	}

	// Convert set to slice
	starredChannelIDs := make([]string, 0, len(starredChannelSet))
	for channelID := range starredChannelSet {
		starredChannelIDs = append(starredChannelIDs, channelID)
	}

	return starredChannelIDs, nil
}

// SearchMessages searches for messages matching a query
// Query syntax: https://api.slack.com/methods/search.messages
// Example queries:
//   - "from:@username" - messages from a user
//   - "in:#channel" - messages in a channel
//   - "@username" or "mentions:me" - messages mentioning a user
func (c *Client) SearchMessages(ctx context.Context, query string, limit int) ([]slack.SearchMessage, error) {
	params := slack.SearchParameters{
		Count:         limit,
		Sort:          "timestamp",
		SortDirection: "desc",
		Highlight:     false,
	}

	messages, err := c.api.SearchMessagesContext(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	return messages.Matches, nil
}

// GetUserReactions retrieves reactions made by or to the authenticated user
func (c *Client) GetUserReactions(ctx context.Context, limit int) ([]slack.ReactedItem, error) {
	params := slack.ListReactionsParameters{
		Count: limit,
		Full:  true, // Get full message details
	}

	items, _, err := c.api.ListReactionsContext(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get reactions: %w", err)
	}

	return items, nil
}

// GetUnreadConversations retrieves conversations with unread messages
func (c *Client) GetUnreadConversations(ctx context.Context) ([]slack.Channel, error) {
	var unreadChannels []slack.Channel
	params := &slack.GetConversationsParameters{
		Types:           []string{"public_channel", "private_channel", "im", "mpim"},
		Limit:           100,
		ExcludeArchived: true,
	}

	for {
		channels, nextCursor, err := c.api.GetConversationsContext(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversations: %w", err)
		}

		// Filter for channels with unread messages
		for _, channel := range channels {
			// Get conversation info to check for unread messages
			info, err := c.api.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
				ChannelID:         channel.ID,
				IncludeNumMembers: false,
			})
			if err != nil {
				continue // Skip on error
			}

			// Check if there are unread messages
			if info.UnreadCount > 0 {
				channel.UnreadCount = info.UnreadCount
				unreadChannels = append(unreadChannels, channel)
			}
		}

		if nextCursor == "" {
			break
		}
		params.Cursor = nextCursor
	}

	return unreadChannels, nil
}
