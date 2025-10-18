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

// Test space key behavior in ChatModel
func TestChatModelSpaceKeyWithNoChannel(t *testing.T) {
	m := NewChatModel()

	// Load keybindings
	localizer := slackyI18n.NewLocalizer("en")
	keyMap, err := loadTestKeyMap(localizer)
	if err != nil {
		t.Fatalf("Failed to load keybindings: %v", err)
	}
	m.SetKeyMap(keyMap)

	// Start with focus on messages (not sidebar)
	m.focused = FocusMessages

	// Press space when no channel is selected
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updatedModel

	// Should focus sidebar
	if m.focused != FocusSidebar {
		t.Errorf("Expected focus to be on sidebar (FocusSidebar), got %v", m.focused)
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

	// Set a selected channel
	m.selectedChannel = &models.Channel{
		ID:   "C123",
		Name: "general",
	}

	// Start with focus on sidebar
	m.focused = FocusSidebar

	// Press space when channel is selected
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updatedModel

	// Should focus input
	if m.focused != FocusInput {
		t.Errorf("Expected focus to be on input (FocusInput), got %v", m.focused)
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
