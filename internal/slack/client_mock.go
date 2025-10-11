package slack

import (
	"context"
	"errors"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// MockClient is a mock implementation of the Slack client for testing
type MockClient struct {
	// TestAuthFunc allows tests to control the response of TestAuth
	TestAuthFunc func(ctx context.Context) (*slack.AuthTestResponse, error)

	// GetChannelsFunc allows tests to control the response of GetChannels
	GetChannelsFunc func(ctx context.Context) ([]slack.Channel, error)

	// GetConversationHistoryFunc allows tests to control the response of GetConversationHistory
	GetConversationHistoryFunc func(ctx context.Context, channelID string, limit int) ([]slack.Message, error)

	// SendMessageFunc allows tests to control the response of SendMessage
	SendMessageFunc func(ctx context.Context, channelID, text string) error

	// EventsChannel is used to simulate Socket Mode events
	EventsChannel chan socketmode.Event

	// RunFunc allows tests to control the Run behavior
	RunFunc func(ctx context.Context) error
}

// NewMockClient creates a new mock Slack client
func NewMockClient() *MockClient {
	return &MockClient{
		EventsChannel: make(chan socketmode.Event, 10),
	}
}

// NewMockClientWithDefaults creates a mock client with default successful responses
func NewMockClientWithDefaults() *MockClient {
	return &MockClient{
		TestAuthFunc: func(ctx context.Context) (*slack.AuthTestResponse, error) {
			return &slack.AuthTestResponse{
				User:   "test-user",
				UserID: "U12345",
				Team:   "test-team",
				TeamID: "T12345",
			}, nil
		},
		GetChannelsFunc: func(ctx context.Context) ([]slack.Channel, error) {
			return []slack.Channel{
				{
					GroupConversation: slack.GroupConversation{
						Name: "general",
						Conversation: slack.Conversation{
							ID: "C12345",
						},
					},
				},
			}, nil
		},
		GetConversationHistoryFunc: func(ctx context.Context, channelID string, limit int) ([]slack.Message, error) {
			return []slack.Message{
				{
					Msg: slack.Msg{
						Text: "Test message",
						User: "U12345",
					},
				},
			}, nil
		},
		SendMessageFunc: func(ctx context.Context, channelID, text string) error {
			return nil
		},
		RunFunc: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
		EventsChannel: make(chan socketmode.Event, 10),
	}
}

// TestAuth implements the Client interface
func (m *MockClient) TestAuth(ctx context.Context) (*slack.AuthTestResponse, error) {
	if m.TestAuthFunc != nil {
		return m.TestAuthFunc(ctx)
	}
	return nil, errors.New("TestAuthFunc not set")
}

// GetChannels implements the Client interface
func (m *MockClient) GetChannels(ctx context.Context) ([]slack.Channel, error) {
	if m.GetChannelsFunc != nil {
		return m.GetChannelsFunc(ctx)
	}
	return nil, errors.New("GetChannelsFunc not set")
}

// GetConversationHistory implements the Client interface
func (m *MockClient) GetConversationHistory(ctx context.Context, channelID string, limit int) ([]slack.Message, error) {
	if m.GetConversationHistoryFunc != nil {
		return m.GetConversationHistoryFunc(ctx, channelID, limit)
	}
	return nil, errors.New("GetConversationHistoryFunc not set")
}

// SendMessage implements the Client interface
func (m *MockClient) SendMessage(ctx context.Context, channelID, text string) error {
	if m.SendMessageFunc != nil {
		return m.SendMessageFunc(ctx, channelID, text)
	}
	return errors.New("SendMessageFunc not set")
}

// Run implements the Client interface
func (m *MockClient) Run(ctx context.Context) error {
	if m.RunFunc != nil {
		return m.RunFunc(ctx)
	}
	<-ctx.Done()
	return ctx.Err()
}

// Events implements the Client interface
func (m *MockClient) Events() chan socketmode.Event {
	return m.EventsChannel
}

// API returns nil for the mock (not used in tests)
func (m *MockClient) API() *slack.Client {
	return nil
}

// Socket returns nil for the mock (not used in tests)
func (m *MockClient) Socket() *socketmode.Client {
	return nil
}
