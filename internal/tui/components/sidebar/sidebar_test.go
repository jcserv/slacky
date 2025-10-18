package sidebar

import (
	"strings"
	"testing"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/models"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Initialize i18n for tests
	if err := slackyI18n.Init(); err != nil {
		panic(err)
	}
}

func TestSetChannels_SplitsStarredAndUnstarred(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		channels           []models.Channel
		expectedStarred    int
		expectedUnstarred  int
		expectedTotalOrder []string // Expected order of channel IDs
	}{
		{
			name: "Splits starred and unstarred correctly",
			channels: []models.Channel{
				{ID: "C1", Name: "general", IsStarred: true},
				{ID: "C2", Name: "random", IsStarred: false},
				{ID: "C3", Name: "dev", IsStarred: true},
			},
			expectedStarred:    2,
			expectedUnstarred:  1,
			expectedTotalOrder: []string{"C1", "C3", "C2"}, // Starred first, then unstarred
		},
		{
			name: "Handles all starred",
			channels: []models.Channel{
				{ID: "C1", Name: "general", IsStarred: true},
				{ID: "C2", Name: "random", IsStarred: true},
			},
			expectedStarred:    2,
			expectedUnstarred:  0,
			expectedTotalOrder: []string{"C1", "C2"},
		},
		{
			name: "Handles no starred",
			channels: []models.Channel{
				{ID: "C1", Name: "general", IsStarred: false},
				{ID: "C2", Name: "random", IsStarred: false},
			},
			expectedStarred:    0,
			expectedUnstarred:  2,
			expectedTotalOrder: []string{"C1", "C2"},
		},
		{
			name:               "Handles empty list",
			channels:           []models.Channel{},
			expectedStarred:    0,
			expectedUnstarred:  0,
			expectedTotalOrder: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := NewModel()
			m.SetChannels(tt.channels)

			assert.Len(t, m.starred, tt.expectedStarred, "Starred count should match")
			assert.Len(t, m.unstarred, tt.expectedUnstarred, "Unstarred count should match")
			assert.Len(t, m.channels, len(tt.expectedTotalOrder), "Total count should match")

			// Verify order
			for i, expectedID := range tt.expectedTotalOrder {
				assert.Equal(t, expectedID, m.channels[i].ID, "Channel order should be correct")
			}
		})
	}
}

func TestView_RendersStarredSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		channels       []models.Channel
		expectStarred  bool
		expectChannels bool
	}{
		{
			name: "Renders both sections when starred exist",
			channels: []models.Channel{
				{ID: "C1", Name: "general", IsStarred: true},
				{ID: "C2", Name: "random", IsStarred: false},
			},
			expectStarred:  true,
			expectChannels: true,
		},
		{
			name: "Renders only channels section when no starred",
			channels: []models.Channel{
				{ID: "C1", Name: "general", IsStarred: false},
				{ID: "C2", Name: "random", IsStarred: false},
			},
			expectStarred:  false,
			expectChannels: true,
		},
		{
			name: "Renders both sections when all starred",
			channels: []models.Channel{
				{ID: "C1", Name: "general", IsStarred: true},
			},
			expectStarred:  true,
			expectChannels: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := NewModel()
			m.SetSize(30, 20)
			m.SetChannels(tt.channels)

			view := m.View()

			if tt.expectStarred {
				assert.Contains(t, view, "Starred", "View should contain 'Starred' header")
			}

			if tt.expectChannels {
				assert.Contains(t, view, "Channels", "View should contain 'Channels' header")
			}

			// Verify channel names appear
			for _, ch := range tt.channels {
				assert.Contains(t, view, ch.Name, "View should contain channel name: %s", ch.Name)
			}
		})
	}
}

func TestGetSelectedChannel_WorksAcrossSections(t *testing.T) {
	t.Parallel()

	channels := []models.Channel{
		{ID: "C1", Name: "general", IsStarred: true},
		{ID: "C2", Name: "dev", IsStarred: true},
		{ID: "C3", Name: "random", IsStarred: false},
		{ID: "C4", Name: "watercooler", IsStarred: false},
	}

	tests := []struct {
		name        string
		selectIndex int
		expectedID  string
	}{
		{
			name:        "Select first starred channel",
			selectIndex: 0,
			expectedID:  "C1",
		},
		{
			name:        "Select second starred channel",
			selectIndex: 1,
			expectedID:  "C2",
		},
		{
			name:        "Select first unstarred channel",
			selectIndex: 2,
			expectedID:  "C3",
		},
		{
			name:        "Select second unstarred channel",
			selectIndex: 3,
			expectedID:  "C4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := NewModel()
			m.SetChannels(channels)
			m.list.Select(tt.selectIndex)
			selected := m.GetSelectedChannel()

			assert.NotNil(t, selected, "Selected channel should not be nil")
			assert.Equal(t, tt.expectedID, selected.ID, "Selected channel ID should match")
		})
	}
}

func TestGetChannels_ReturnsCombinedList(t *testing.T) {
	t.Parallel()

	channels := []models.Channel{
		{ID: "C1", Name: "general", IsStarred: true},
		{ID: "C2", Name: "random", IsStarred: false},
		{ID: "C3", Name: "dev", IsStarred: true},
	}

	m := NewModel()
	m.SetChannels(channels)

	result := m.GetChannels()

	assert.Len(t, result, 3, "Should return all channels")

	// Verify order: starred first, then unstarred
	assert.Equal(t, "C1", result[0].ID)
	assert.Equal(t, "C3", result[1].ID)
	assert.Equal(t, "C2", result[2].ID)
}

func TestSelectChannel_WorksAcrossSections(t *testing.T) {
	t.Parallel()

	channels := []models.Channel{
		{ID: "C1", Name: "general", IsStarred: true},
		{ID: "C2", Name: "random", IsStarred: false},
	}

	tests := []struct {
		name       string
		selectID   string
		shouldFind bool
	}{
		{
			name:       "Select starred channel",
			selectID:   "C1",
			shouldFind: true,
		},
		{
			name:       "Select unstarred channel",
			selectID:   "C2",
			shouldFind: true,
		},
		{
			name:       "Select non-existent channel",
			selectID:   "C999",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := NewModel()
			m.SetChannels(channels)
			m.SelectChannel(tt.selectID)
			selected := m.GetSelectedChannel()

			if tt.shouldFind {
				assert.NotNil(t, selected, "Should find channel")
				assert.Equal(t, tt.selectID, selected.ID, "Selected channel ID should match")
			}
		})
	}
}

func TestView_OrderStarredBeforeChannels(t *testing.T) {
	t.Parallel()

	channels := []models.Channel{
		{ID: "C1", Name: "random", IsStarred: false},
		{ID: "C2", Name: "general", IsStarred: true},
	}

	m := NewModel()
	m.SetSize(30, 20)
	m.SetChannels(channels)

	view := m.View()

	// Find positions of headers and channel names
	starredPos := strings.Index(view, "Starred")
	channelsPos := strings.Index(view, "Channels")
	generalPos := strings.Index(view, "general")
	randomPos := strings.Index(view, "random")

	assert.NotEqual(t, -1, starredPos, "Should contain Starred header")
	assert.NotEqual(t, -1, channelsPos, "Should contain Channels header")

	// Verify order: Starred header -> general -> Channels header -> random
	assert.Less(t, starredPos, generalPos, "Starred header should come before general")
	assert.Less(t, generalPos, channelsPos, "general should come before Channels header")
	assert.Less(t, channelsPos, randomPos, "Channels header should come before random")
}

func TestNextChannel_PrevChannel_WorkAcrossSections(t *testing.T) {
	t.Parallel()

	channels := []models.Channel{
		{ID: "C1", Name: "general", IsStarred: true},
		{ID: "C2", Name: "dev", IsStarred: true},
		{ID: "C3", Name: "random", IsStarred: false},
	}

	m := NewModel()
	m.SetChannels(channels)

	// Start at first channel
	m.list.Select(0)
	assert.Equal(t, "C1", m.GetSelectedChannel().ID)

	// Move next through starred section
	m.NextChannel()
	assert.Equal(t, "C2", m.GetSelectedChannel().ID, "Should move to next starred channel")

	// Move next into unstarred section
	m.NextChannel()
	assert.Equal(t, "C3", m.GetSelectedChannel().ID, "Should move from starred to unstarred section")

	// Move prev back into starred section
	m.PrevChannel()
	assert.Equal(t, "C2", m.GetSelectedChannel().ID, "Should move back to starred section")

	// Move prev to first
	m.PrevChannel()
	assert.Equal(t, "C1", m.GetSelectedChannel().ID, "Should move to first channel")
}
