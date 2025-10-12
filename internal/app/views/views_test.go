package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
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

	if !strings.Contains(view, "Chat View") {
		t.Error("View should contain 'Chat View' title")
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

	if !strings.Contains(view, "Activity View") {
		t.Error("View should contain 'Activity View' title")
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
