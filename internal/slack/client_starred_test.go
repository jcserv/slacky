package slack

import (
	"context"
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestGetStarredConversations_FiltersCorrectly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		items    []slack.Item
		expected []string
	}{
		{
			name: "Only includes direct channel stars",
			items: []slack.Item{
				{Type: slack.TYPE_CHANNEL, Channel: "C123"},
				{Type: slack.TYPE_IM, Channel: "D456"},
				{Type: slack.TYPE_GROUP, Channel: "G789"},
			},
			expected: []string{"C123", "D456", "G789"},
		},
		{
			name: "Excludes message stars",
			items: []slack.Item{
				{Type: slack.TYPE_CHANNEL, Channel: "C123"},
				{Type: slack.TYPE_MESSAGE, Channel: "C456"}, // Should be excluded
			},
			expected: []string{"C123"},
		},
		{
			name: "Excludes file and file comment stars",
			items: []slack.Item{
				{Type: slack.TYPE_CHANNEL, Channel: "C123"},
				{Type: slack.TYPE_FILE, Channel: "C456"},         // Should be excluded
				{Type: slack.TYPE_FILE_COMMENT, Channel: "C789"}, // Should be excluded
			},
			expected: []string{"C123"},
		},
		{
			name: "Deduplicates channel IDs",
			items: []slack.Item{
				{Type: slack.TYPE_CHANNEL, Channel: "C123"},
				{Type: slack.TYPE_CHANNEL, Channel: "C123"}, // Duplicate
				{Type: slack.TYPE_IM, Channel: "D456"},
			},
			expected: []string{"C123", "D456"},
		},
		{
			name: "Handles empty channel field",
			items: []slack.Item{
				{Type: slack.TYPE_CHANNEL, Channel: "C123"},
				{Type: slack.TYPE_CHANNEL, Channel: ""}, // Empty channel
			},
			expected: []string{"C123"},
		},
		{
			name:     "Returns empty list when no stars",
			items:    []slack.Item{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a mock client with custom behavior
			mockClient := NewMockClient()
			mockClient.ListAllStarsFunc = func(ctx context.Context) ([]slack.Item, error) {
				return tt.items, nil
			}

			result, err := mockClient.GetStarredConversations(context.Background())

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, result, "Starred channel IDs should match expected")
		})
	}
}

func TestGetStarredConversations_HandlesAPIError(t *testing.T) {
	t.Parallel()

	mockClient := NewMockClient()
	mockClient.ListAllStarsFunc = func(ctx context.Context) ([]slack.Item, error) {
		return nil, assert.AnError
	}

	result, err := mockClient.GetStarredConversations(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
}
