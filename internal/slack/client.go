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

// GetChannels retrieves all channels the bot has access to
func (c *Client) GetChannels(ctx context.Context) ([]slack.Channel, error) {
	var allChannels []slack.Channel
	params := &slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
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
