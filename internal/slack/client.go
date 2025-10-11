package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// Client wraps the Slack API client and Socket Mode client
type Client struct {
	api    *slack.Client
	socket *socketmode.Client
}

// New creates a new Slack client with the provided tokens
func New(botToken, appToken string) *Client {
	api := slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
	)

	socket := socketmode.New(
		api,
		socketmode.OptionDebug(false),
	)

	return &Client{
		api:    api,
		socket: socket,
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
func (c *Client) Run(ctx context.Context) error {
	return c.socket.RunContext(ctx)
}

// Events returns the channel for receiving Socket Mode events
func (c *Client) Events() chan socketmode.Event {
	return c.socket.Events
}

// API returns the underlying Slack API client
func (c *Client) API() *slack.Client {
	return c.api
}

// Socket returns the underlying Socket Mode client
func (c *Client) Socket() *socketmode.Client {
	return c.socket
}
