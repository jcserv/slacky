package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkAsStarred(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		channels   []Channel
		starredIDs []string
		verify     func(t *testing.T, result []Channel)
	}{
		{
			name: "Marks single channel as starred",
			channels: []Channel{
				{ID: "C123", Name: "general"},
				{ID: "C456", Name: "random"},
			},
			starredIDs: []string{"C123"},
			verify: func(t *testing.T, result []Channel) {
				assert.Len(t, result, 2)
				assert.True(t, result[0].IsStarred, "Channel C123 should be starred")
				assert.False(t, result[1].IsStarred, "Channel C456 should not be starred")
			},
		},
		{
			name: "Marks multiple channels as starred",
			channels: []Channel{
				{ID: "C123", Name: "general"},
				{ID: "C456", Name: "random"},
				{ID: "C789", Name: "dev"},
			},
			starredIDs: []string{"C123", "C789"},
			verify: func(t *testing.T, result []Channel) {
				assert.Len(t, result, 3)
				assert.True(t, result[0].IsStarred, "Channel C123 should be starred")
				assert.False(t, result[1].IsStarred, "Channel C456 should not be starred")
				assert.True(t, result[2].IsStarred, "Channel C789 should be starred")
			},
		},
		{
			name: "Handles no starred channels",
			channels: []Channel{
				{ID: "C123", Name: "general"},
				{ID: "C456", Name: "random"},
			},
			starredIDs: []string{},
			verify: func(t *testing.T, result []Channel) {
				assert.Len(t, result, 2)
				assert.False(t, result[0].IsStarred, "No channels should be starred")
				assert.False(t, result[1].IsStarred, "No channels should be starred")
			},
		},
		{
			name: "Handles all channels starred",
			channels: []Channel{
				{ID: "C123", Name: "general"},
				{ID: "C456", Name: "random"},
			},
			starredIDs: []string{"C123", "C456"},
			verify: func(t *testing.T, result []Channel) {
				assert.Len(t, result, 2)
				assert.True(t, result[0].IsStarred, "Channel C123 should be starred")
				assert.True(t, result[1].IsStarred, "Channel C456 should be starred")
			},
		},
		{
			name: "Handles starred ID not in channel list",
			channels: []Channel{
				{ID: "C123", Name: "general"},
			},
			starredIDs: []string{"C999"}, // ID not in channels
			verify: func(t *testing.T, result []Channel) {
				assert.Len(t, result, 1)
				assert.False(t, result[0].IsStarred, "Channel C123 should not be starred")
			},
		},
		{
			name:       "Handles empty channel list",
			channels:   []Channel{},
			starredIDs: []string{"C123"},
			verify: func(t *testing.T, result []Channel) {
				assert.Len(t, result, 0)
			},
		},
		{
			name: "Preserves channel data",
			channels: []Channel{
				{
					ID:      "C123",
					Name:    "general",
					Type:    ChannelTypePublic,
					Topic:   "General discussion",
					Purpose: "Company-wide announcements",
				},
			},
			starredIDs: []string{"C123"},
			verify: func(t *testing.T, result []Channel) {
				assert.Len(t, result, 1)
				assert.True(t, result[0].IsStarred)
				assert.Equal(t, "C123", result[0].ID)
				assert.Equal(t, "general", result[0].Name)
				assert.Equal(t, ChannelTypePublic, result[0].Type)
				assert.Equal(t, "General discussion", result[0].Topic)
				assert.Equal(t, "Company-wide announcements", result[0].Purpose)
			},
		},
		{
			name: "Works with DMs and groups",
			channels: []Channel{
				{ID: "D123", Name: "alice", Type: ChannelTypeDM},
				{ID: "G456", Name: "project-team", Type: ChannelTypePrivate},
			},
			starredIDs: []string{"D123", "G456"},
			verify: func(t *testing.T, result []Channel) {
				assert.Len(t, result, 2)
				assert.True(t, result[0].IsStarred, "DM should be starred")
				assert.True(t, result[1].IsStarred, "Group should be starred")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := MarkAsStarred(tt.channels, tt.starredIDs)

			tt.verify(t, result)
		})
	}
}

func TestMarkAsStarred_DoesNotModifyOriginal(t *testing.T) {
	t.Parallel()

	original := []Channel{
		{ID: "C123", Name: "general", IsStarred: false},
	}

	result := MarkAsStarred(original, []string{"C123"})

	// Original should be unchanged
	assert.False(t, original[0].IsStarred, "Original channel should not be modified")
	// Result should have the change
	assert.True(t, result[0].IsStarred, "Result channel should be starred")
}
