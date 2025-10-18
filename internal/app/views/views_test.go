package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/models"
	"github.com/jcserv/slacky/internal/tui/keys"
)

func init() {
	// Initialize i18n for tests
	if err := slackyI18n.Init(); err != nil {
		panic(err)
	}
}

// Test ChatModel
func TestNewChatModel(t *testing.T) {
	m := NewChatModel()

	if m.width != 80 {
		t.Errorf("Expected default width 80, got %d", m.width)
	}

	if m.height != 24 {
		t.Errorf("Expected default height 24, got %d", m.height)
	}

	if m.localizer == nil {
		t.Error("Expected localizer to be initialized")
	}
}

func TestChatModelSetSize(t *testing.T) {
	m := NewChatModel()

	m.SetSize(120, 40)

	if m.width != 120 {
		t.Errorf("Expected width 120, got %d", m.width)
	}

	if m.height != 40 {
		t.Errorf("Expected height 40, got %d", m.height)
	}
}

func TestChatModelInit(t *testing.T) {
	m := NewChatModel()

	cmd := m.Init()

	if cmd != nil {
		t.Error("Expected Init to return nil command")
	}
}

func TestChatModelUpdate(t *testing.T) {
	m := NewChatModel()

	updatedModel, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if cmd != nil {
		t.Error("Expected Update to return nil command")
	}

	// Model should be returned
	if updatedModel.width != m.width {
		t.Error("Model should maintain state through Update")
	}
}

func TestChatModelView(t *testing.T) {
	m := NewChatModel()
	m.SetSize(100, 30)

	view := m.View()

	if view == "" {
		t.Error("View should not return empty string")
	}

	// Check for sidebar component presence
	if !strings.Contains(view, "Channels") {
		t.Error("View should contain 'Channels' sidebar")
	}
}

// Test ActivityModel
func TestNewActivityModel(t *testing.T) {
	m := NewActivityModel()

	if m.width != 80 {
		t.Errorf("Expected default width 80, got %d", m.width)
	}

	if m.height != 24 {
		t.Errorf("Expected default height 24, got %d", m.height)
	}

	if m.localizer == nil {
		t.Error("Expected localizer to be initialized")
	}
}

func TestActivityModelSetSize(t *testing.T) {
	m := NewActivityModel()

	m.SetSize(120, 40)

	if m.width != 120 {
		t.Errorf("Expected width 120, got %d", m.width)
	}

	if m.height != 40 {
		t.Errorf("Expected height 40, got %d", m.height)
	}
}

func TestActivityModelInit(t *testing.T) {
	m := NewActivityModel()

	cmd := m.Init()

	if cmd != nil {
		t.Error("Expected Init to return nil command")
	}
}

func TestActivityModelUpdate(t *testing.T) {
	m := NewActivityModel()

	updatedModel, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if cmd != nil {
		t.Error("Expected Update to return nil command")
	}

	// Model should be returned
	if updatedModel.width != m.width {
		t.Error("Model should maintain state through Update")
	}
}

func TestActivityModelView(t *testing.T) {
	m := NewActivityModel()
	m.SetSize(100, 30)

	view := m.View()

	if view == "" {
		t.Error("View should not return empty string")
	}

	// Check for filter status line
	if !strings.Contains(view, "[All]") {
		t.Error("View should contain filter status showing current filter")
	}
}

// Test UserModel
func TestNewUserModel(t *testing.T) {
	m := NewUserModel()

	if m.width != 80 {
		t.Errorf("Expected default width 80, got %d", m.width)
	}

	if m.height != 24 {
		t.Errorf("Expected default height 24, got %d", m.height)
	}

	if m.localizer == nil {
		t.Error("Expected localizer to be initialized")
	}

	if m.userName != "" {
		t.Error("Expected userName to be empty initially")
	}

	if m.teamName != "" {
		t.Error("Expected teamName to be empty initially")
	}
}

func TestUserModelSetSize(t *testing.T) {
	m := NewUserModel()

	m.SetSize(120, 40)

	if m.width != 120 {
		t.Errorf("Expected width 120, got %d", m.width)
	}

	if m.height != 40 {
		t.Errorf("Expected height 40, got %d", m.height)
	}
}

func TestUserModelSetUserInfo(t *testing.T) {
	m := NewUserModel()

	m.SetUserInfo("testuser", "Test Workspace")

	if m.userName != "testuser" {
		t.Errorf("Expected userName 'testuser', got %s", m.userName)
	}

	if m.teamName != "Test Workspace" {
		t.Errorf("Expected teamName 'Test Workspace', got %s", m.teamName)
	}
}

func TestUserModelInit(t *testing.T) {
	m := NewUserModel()

	cmd := m.Init()

	if cmd != nil {
		t.Error("Expected Init to return nil command")
	}
}

func TestUserModelUpdate(t *testing.T) {
	m := NewUserModel()

	updatedModel, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if cmd != nil {
		t.Error("Expected Update to return nil command")
	}

	// Model should be returned
	if updatedModel.width != m.width {
		t.Error("Model should maintain state through Update")
	}
}

func TestUserModelView(t *testing.T) {
	m := NewUserModel()
	m.SetSize(100, 30)
	m.SetUserInfo("testuser", "Test Workspace")

	view := m.View()

	if view == "" {
		t.Error("View should not return empty string")
	}

	if !strings.Contains(view, "User Information") {
		t.Error("View should contain 'User Information' title")
	}

	if !strings.Contains(view, "testuser") {
		t.Error("View should contain username")
	}

	if !strings.Contains(view, "Test Workspace") {
		t.Error("View should contain team name")
	}
}

func TestUserModelViewWithoutUserInfo(t *testing.T) {
	m := NewUserModel()
	m.SetSize(100, 30)

	view := m.View()

	if view == "" {
		t.Error("View should not return empty string even without user info")
	}

	// Should still contain the title
	if !strings.Contains(view, "User Information") {
		t.Error("View should contain title even without user info")
	}
}

// Test tab cycling behavior in ChatModel
func TestChatModelSpaceKeyWithNoChannel(t *testing.T) {
	m := NewChatModel()

	// Load keybindings
	localizer := slackyI18n.NewLocalizer("en")
	keyMap, err := loadTestKeyMap(localizer)
	if err != nil {
		t.Fatalf("Failed to load keybindings: %v", err)
	}
	m.SetKeyMap(keyMap)

	// Enter the view (simulating user pressing Space at tab level)
	m.EnterView()

	// Start with focus on messages (not sidebar)
	m.focused = FocusMessages

	// Press tab to cycle focus to input
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel

	// Should cycle to input
	if m.focused != FocusInput {
		t.Errorf("Expected focus to be on input (FocusInput), got %v", m.focused)
	}
}

func TestChatModelSpaceKeyWithChannel(t *testing.T) {
	m := NewChatModel()

	// Load keybindings
	localizer := slackyI18n.NewLocalizer("en")
	keyMap, err := loadTestKeyMap(localizer)
	if err != nil {
		t.Fatalf("Failed to load keybindings: %v", err)
	}
	m.SetKeyMap(keyMap)

	// Enter the view (simulating user pressing Space at tab level)
	m.EnterView()

	// Set a selected channel
	m.selectedChannel = &models.Channel{
		ID:   "C123",
		Name: "general",
	}

	// Start with focus on sidebar (default when entering view)
	// Press tab to cycle forward
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel

	// Should cycle to messages
	if m.focused != FocusMessages {
		t.Errorf("Expected focus to be on messages (FocusMessages), got %v", m.focused)
	}
}

func TestChatModelSpaceKeyAlreadyOnSidebar(t *testing.T) {
	m := NewChatModel()

	// Load keybindings
	localizer := slackyI18n.NewLocalizer("en")
	keyMap, err := loadTestKeyMap(localizer)
	if err != nil {
		t.Fatalf("Failed to load keybindings: %v", err)
	}
	m.SetKeyMap(keyMap)

	// No channel selected, already on sidebar
	m.focused = FocusSidebar

	// Press space - should still focus sidebar (to ensure highlighting)
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updatedModel

	// Should remain on sidebar
	if m.focused != FocusSidebar {
		t.Errorf("Expected focus to remain on sidebar (FocusSidebar), got %v", m.focused)
	}
}

func TestChatModelSpaceKeyAlreadyOnInput(t *testing.T) {
	m := NewChatModel()

	// Load keybindings
	localizer := slackyI18n.NewLocalizer("en")
	keyMap, err := loadTestKeyMap(localizer)
	if err != nil {
		t.Fatalf("Failed to load keybindings: %v", err)
	}
	m.SetKeyMap(keyMap)

	// Set a selected channel
	m.selectedChannel = &models.Channel{
		ID:   "C123",
		Name: "general",
	}

	// Already on input
	m.focused = FocusInput

	// Press space - should not change focus (already on input)
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updatedModel

	// Should remain on input (space condition prevents changing when already on input)
	if m.focused != FocusInput {
		t.Errorf("Expected focus to remain on input (FocusInput), got %v", m.focused)
	}
}

// Helper function to load keybindings for tests
func loadTestKeyMap(localizer *i18n.Localizer) (*keys.ScopedKeyMap, error) {
	return keys.LoadKeybindings(nil, localizer)
}

// Test thread exit behavior - simulates the full user flow
func TestChatModelThreadExitReturnsToChannel(t *testing.T) {
	m := NewChatModel()

	// Load keybindings
	localizer := slackyI18n.NewLocalizer("en")
	keyMap, err := loadTestKeyMap(localizer)
	if err != nil {
		t.Fatalf("Failed to load keybindings: %v", err)
	}
	m.SetKeyMap(keyMap)

	// Setup: Add channels and select one
	channels := []models.Channel{
		{ID: "C123", Name: "general", Type: models.ChannelTypePublic},
		{ID: "C456", Name: "random", Type: models.ChannelTypePublic},
	}
	m.SetChannels(channels)
	m.SelectChannel("C123")
	m.EnterView()

	// Add messages to the channel with a thread
	messages := []models.Message{
		{ID: "1234.5678", Text: "First message", UserName: "user1", ChannelID: "C123", ReplyCount: 2},
		{ID: "1234.5679", Text: "Second message", UserName: "user2", ChannelID: "C123"},
	}
	m.SetMessages(messages)

	// Verify initial state
	if m.threadActive {
		t.Error("Thread should not be active initially")
	}
	if m.selectedChannel == nil || m.selectedChannel.ID != "C123" {
		t.Errorf("Expected selected channel to be C123, got %v", m.selectedChannel)
	}

	// Simulate entering a thread
	threadParentMsg := messages[0]
	m.SetThreadReplies("C123", "#general", "1234.5678", &threadParentMsg, []models.Message{
		threadParentMsg,
		{ID: "1234.5680", Text: "Reply 1", UserName: "user3", ChannelID: "C123", ThreadTS: "1234.5678"},
	})
	m.threadActive = true
	m.setFocus(FocusThread)

	// Verify thread is active
	if !m.threadActive {
		t.Error("Thread should be active after entering")
	}

	// Press escape to exit thread
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel

	// Verify thread is no longer active
	if m.threadActive {
		t.Error("Thread should not be active after pressing escape")
	}

	// Verify we're back to the correct channel
	if m.selectedChannel == nil || m.selectedChannel.ID != "C123" {
		t.Errorf("Expected to return to channel C123, got %v", m.selectedChannel)
	}

	// Verify focus is back on messages
	if m.focused != FocusMessages {
		t.Errorf("Expected focus to be on messages, got %v", m.focused)
	}

	// Verify a command was returned (if we need to reload messages)
	// In the case where the thread's channel is the same as selected channel,
	// no command should be returned
	if cmd != nil {
		// If a command is returned, execute it to see what message it produces
		msg := cmd()
		if _, ok := msg.(ChannelSelectedMsg); !ok {
			t.Errorf("Expected ChannelSelectedMsg or nil command, got %T: %+v", msg, msg)
		}
	}
}
