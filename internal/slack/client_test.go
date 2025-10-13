package slack_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jcserv/slacky/internal/slack"
	slackapi "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	userToken := "xoxb-test-token"
	client := slack.New(userToken)

	assert.NotNil(t, client)
	assert.NotNil(t, client.API())
}

func TestMockClient_TestAuth_Success(t *testing.T) {
	t.Parallel()

	mock := slack.NewMockClientWithDefaults()
	ctx := context.Background()

	resp, err := mock.TestAuth(ctx)

	require.NoError(t, err)
	assert.Equal(t, "test-user", resp.User)
	assert.Equal(t, "U12345", resp.UserID)
	assert.Equal(t, "test-team", resp.Team)
	assert.Equal(t, "T12345", resp.TeamID)
}

func TestMockClient_TestAuth_Failure(t *testing.T) {
	t.Parallel()

	mock := slack.NewMockClient()
	mock.TestAuthFunc = func(ctx context.Context) (*slackapi.AuthTestResponse, error) {
		return nil, errors.New("auth failed")
	}

	ctx := context.Background()

	resp, err := mock.TestAuth(ctx)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "auth failed")
}

func TestMockClient_GetChannels_Success(t *testing.T) {
	t.Parallel()

	mock := slack.NewMockClientWithDefaults()
	ctx := context.Background()

	channels, err := mock.GetChannels(ctx)

	require.NoError(t, err)
	assert.Len(t, channels, 1)
	assert.Equal(t, "general", channels[0].Name)
	assert.Equal(t, "C12345", channels[0].ID)
}

func TestMockClient_GetChannels_Empty(t *testing.T) {
	t.Parallel()

	mock := slack.NewMockClient()
	mock.GetChannelsFunc = func(ctx context.Context) ([]slackapi.Channel, error) {
		return []slackapi.Channel{}, nil
	}

	ctx := context.Background()

	channels, err := mock.GetChannels(ctx)

	require.NoError(t, err)
	assert.Empty(t, channels)
}

func TestMockClient_GetChannels_Failure(t *testing.T) {
	t.Parallel()

	mock := slack.NewMockClient()
	mock.GetChannelsFunc = func(ctx context.Context) ([]slackapi.Channel, error) {
		return nil, errors.New("failed to get channels")
	}

	ctx := context.Background()

	channels, err := mock.GetChannels(ctx)

	require.Error(t, err)
	assert.Nil(t, channels)
	assert.Contains(t, err.Error(), "failed to get channels")
}

func TestMockClient_GetConversationHistory_Success(t *testing.T) {
	t.Parallel()

	mock := slack.NewMockClientWithDefaults()
	ctx := context.Background()

	messages, err := mock.GetConversationHistory(ctx, "C12345", 10)

	require.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "Test message", messages[0].Text)
	assert.Equal(t, "U12345", messages[0].User)
}

func TestMockClient_SendMessage_Success(t *testing.T) {
	t.Parallel()

	mock := slack.NewMockClientWithDefaults()
	ctx := context.Background()

	err := mock.SendMessage(ctx, "C12345", "Hello, world!")

	assert.NoError(t, err)
}

func TestMockClient_SendMessage_Failure(t *testing.T) {
	t.Parallel()

	mock := slack.NewMockClient()
	mock.SendMessageFunc = func(ctx context.Context, channelID, text string) error {
		return errors.New("failed to send message")
	}

	ctx := context.Background()

	err := mock.SendMessage(ctx, "C12345", "Hello, world!")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send message")
}

func TestMockClient_CustomBehavior(t *testing.T) {
	t.Parallel()

	mock := slack.NewMockClient()

	// Set up custom behavior
	callCount := 0
	mock.TestAuthFunc = func(ctx context.Context) (*slackapi.AuthTestResponse, error) {
		callCount++
		if callCount == 1 {
			return nil, errors.New("first call fails")
		}
		return &slackapi.AuthTestResponse{
			User:   "retry-user",
			UserID: "U99999",
		}, nil
	}

	ctx := context.Background()

	// First call should fail
	resp, err := mock.TestAuth(ctx)
	require.Error(t, err)
	assert.Nil(t, resp)

	// Second call should succeed
	resp, err = mock.TestAuth(ctx)
	require.NoError(t, err)
	assert.Equal(t, "retry-user", resp.User)
	assert.Equal(t, "U99999", resp.UserID)
}

func TestMockClient_Events(t *testing.T) {
	t.Parallel()

	mock := slack.NewMockClient()

	events := mock.Events()

	assert.NotNil(t, events)
	assert.Equal(t, 0, len(events), "events channel should start empty")
}
