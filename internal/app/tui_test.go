package app

import (
	"context"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jcserv/slacky/internal/config"
	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/tui/actions"
	"github.com/jcserv/slacky/internal/tui/components/tabs"
)

func init() {
	// Initialize i18n for tests
	if err := slackyI18n.Init(); err != nil {
		panic(err)
	}
}

func TestNewTUI(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.Workspace{
			BotToken:    "xoxb-test",
			SocketToken: "xapp-test",
		},
	}

	app, err := New(nil, cfg)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	m := app.NewTUI()

	if m.app != app {
		t.Error("Expected app to be set")
	}

	if m.keyMap == nil {
		t.Error("Expected key map to be initialized")
	}

	if m.localizer == nil {
		t.Error("Expected localizer to be initialized")
	}

	// Check that showHelp is initialized to true (default for new users)
	if !m.showHelp {
		t.Error("Expected showHelp to be true initially (default for new users)")
	}

	// Check that components are initialized
	if m.tabs.GetCurrentTab() != tabs.ChatTab {
		t.Error("Expected tabs to be initialized with ChatTab")
	}
}

func TestHelpToggle(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.Workspace{
			BotToken:    "xoxb-test",
			SocketToken: "xapp-test",
		},
	}

	app, err := New(context.TODO(), cfg)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	m := app.NewTUI()

	// Initially help should be visible (default for new users)
	if !m.showHelp {
		t.Error("Expected help to be visible initially (default for new users)")
	}

	// Simulate help toggle key press
	helpKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updated, _ := m.Update(helpKey)

	// Help should now be hidden
	if updated.(TUIModel).showHelp {
		t.Error("Expected help to be hidden after toggle")
	}

	// Toggle again to show
	updated, _ = updated.(TUIModel).Update(helpKey)

	// Help should be visible again
	if !updated.(TUIModel).showHelp {
		t.Error("Expected help to be visible after second toggle")
	}
}

func TestTUIModelInit(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.Workspace{
			BotToken:    "xoxb-test",
			SocketToken: "xapp-test",
		},
	}

	app, _ := New(nil, cfg)
	m := app.NewTUI()

	cmd := m.Init()

	if cmd == nil {
		t.Error("Expected Init to return a batch command")
	}
}

func TestTUIModelUpdateWindowSize(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.Workspace{
			BotToken:    "xoxb-test",
			SocketToken: "xapp-test",
		},
	}

	app, _ := New(nil, cfg)
	m := app.NewTUI()
	m.authSuccess = true // Set to authenticated state

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := m.Update(msg)

	tuiModel := updatedModel.(TUIModel)

	if tuiModel.width != 120 {
		t.Errorf("Expected width 120, got %d", tuiModel.width)
	}

	if tuiModel.height != 40 {
		t.Errorf("Expected height 40, got %d", tuiModel.height)
	}
}

func TestTUIModelUpdateQuit(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.Workspace{
			BotToken:    "xoxb-test",
			SocketToken: "xapp-test",
		},
	}

	app, _ := New(nil, cfg)
	m := app.NewTUI()

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedModel, cmd := m.Update(msg)

	tuiModel := updatedModel.(TUIModel)

	if !tuiModel.quitting {
		t.Error("Expected quitting to be true")
	}

	if cmd == nil {
		t.Error("Expected quit command to be returned")
	}
}

func TestTUIModelUpdateTabNavigation(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.Workspace{
			BotToken:    "xoxb-test",
			SocketToken: "xapp-test",
		},
	}

	app, _ := New(nil, cfg)
	m := app.NewTUI()
	m.authSuccess = true // Must be authenticated to navigate tabs

	t.Run("Next tab", func(t *testing.T) {
		initialTab := m.tabs.GetCurrentTab()

		// Simulate tab key press
		msg := tea.KeyMsg{Type: tea.KeyTab}
		updatedModel, _ := m.Update(msg)
		tuiModel := updatedModel.(TUIModel)

		if tuiModel.tabs.GetCurrentTab() == initialTab {
			t.Error("Expected tab to change after Tab key")
		}
	})

	t.Run("Previous tab", func(t *testing.T) {
		m.tabs.SetCurrentTab(tabs.ActivityTab)

		// Simulate shift+tab key press
		msg := tea.KeyMsg{Type: tea.KeyShiftTab}
		updatedModel, _ := m.Update(msg)
		tuiModel := updatedModel.(TUIModel)

		if tuiModel.tabs.GetCurrentTab() != tabs.ChatTab {
			t.Error("Expected to navigate to ChatTab")
		}
	})

	t.Run("User key shortcut", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}
		updatedModel, _ := m.Update(msg)
		tuiModel := updatedModel.(TUIModel)

		if tuiModel.tabs.GetCurrentTab() != tabs.UserTab {
			t.Error("Expected to navigate to UserTab with 'u' key")
		}
	})
}

func TestTUIModelUpdateTabNavigationWithoutAuth(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.Workspace{
			BotToken:    "xoxb-test",
			SocketToken: "xapp-test",
		},
	}

	app, _ := New(nil, cfg)
	m := app.NewTUI()
	m.authSuccess = false // Not authenticated

	initialTab := m.tabs.GetCurrentTab()

	// Simulate tab key press
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModel, _ := m.Update(msg)
	tuiModel := updatedModel.(TUIModel)

	// Tab should not change when not authenticated
	if tuiModel.tabs.GetCurrentTab() != initialTab {
		t.Error("Expected tab to not change when not authenticated")
	}
}

func TestTUIModelUpdateAuthSuccess(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.Workspace{
			BotToken:    "xoxb-test",
			SocketToken: "xapp-test",
		},
	}

	app, _ := New(nil, cfg)
	m := app.NewTUI()

	msg := authSuccessMsg{
		teamName: "Test Team",
		userName: "testuser",
	}

	updatedModel, _ := m.Update(msg)
	tuiModel := updatedModel.(TUIModel)

	if !tuiModel.authSuccess {
		t.Error("Expected authSuccess to be true")
	}

	if tuiModel.teamName != "Test Team" {
		t.Errorf("Expected teamName 'Test Team', got %s", tuiModel.teamName)
	}

	if tuiModel.userName != "testuser" {
		t.Errorf("Expected userName 'testuser', got %s", tuiModel.userName)
	}
}

func TestTUIModelUpdateError(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.Workspace{
			BotToken:    "xoxb-test",
			SocketToken: "xapp-test",
		},
	}

	app, _ := New(nil, cfg)
	m := app.NewTUI()

	testErr := errMsg(tea.ErrProgramKilled)
	updatedModel, _ := m.Update(testErr)
	tuiModel := updatedModel.(TUIModel)

	if tuiModel.err == nil {
		t.Error("Expected error to be set")
	}
}

func TestTUIModelViewStates(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.Workspace{
			BotToken:    "xoxb-test",
			SocketToken: "xapp-test",
		},
	}

	app, _ := New(nil, cfg)
	m := app.NewTUI()

	t.Run("Loading state", func(t *testing.T) {
		m.authSuccess = false
		m.err = nil

		view := m.View()

		if !strings.Contains(view, "Connecting") {
			t.Error("View should show connecting message in loading state")
		}
	})

	t.Run("Error state", func(t *testing.T) {
		m.err = errMsg(tea.ErrProgramKilled)

		view := m.View()

		if !strings.Contains(view, "Error") {
			t.Error("View should show error message in error state")
		}
	})

	t.Run("Success state with tabs", func(t *testing.T) {
		m.err = nil
		m.authSuccess = true
		m.userName = "testuser"
		m.width = 100
		m.height = 40
		m.tabs.SetWidth(100)
		m.tabs.SetUserName("testuser")
		m.statusBar.SetWidth(100)
		m.statusBar.SetConnected(true)
		m.chatView.SetSize(100, 36)

		view := m.View()

		if view == "" {
			t.Error("View should not be empty in success state")
		}

		// Should contain tabs
		if !strings.Contains(view, "Chat") {
			t.Error("View should contain Chat tab")
		}

		// Should contain active view content
		if !strings.Contains(view, "Chat View") {
			t.Error("View should contain chat view content")
		}

		// Should contain status bar
		if !strings.Contains(view, "Connected") {
			t.Error("View should contain status bar")
		}
	})
}

func TestTUIModelViewSwitching(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.Workspace{
			BotToken:    "xoxb-test",
			SocketToken: "xapp-test",
		},
	}

	app, _ := New(nil, cfg)
	m := app.NewTUI()
	m.authSuccess = true
	m.width = 100
	m.height = 40
	m.tabs.SetWidth(100)
	m.statusBar.SetWidth(100)

	t.Run("Chat view", func(t *testing.T) {
		m.tabs.SetCurrentTab(tabs.ChatTab)
		m.chatView.SetSize(100, 36)

		view := m.View()

		if !strings.Contains(view, "Chat View") {
			t.Error("Should display chat view content")
		}
	})

	t.Run("Activity view", func(t *testing.T) {
		m.tabs.SetCurrentTab(tabs.ActivityTab)
		m.activityView.SetSize(100, 36)

		view := m.View()

		if !strings.Contains(view, "Activity View") {
			t.Error("Should display activity view content")
		}
	})

	t.Run("User view", func(t *testing.T) {
		m.tabs.SetCurrentTab(tabs.UserTab)
		m.userView.SetSize(100, 36)
		m.userView.SetUserInfo("testuser", "Test Team")

		view := m.View()

		if !strings.Contains(view, "User Information") {
			t.Error("Should display user view content")
		}
	})
}

func TestTUIKeyBindings(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.Workspace{
			BotToken:    "xoxb-test",
			SocketToken: "xapp-test",
		},
	}

	app, _ := New(nil, cfg)
	m := app.NewTUI()

	t.Run("Quit key binding", func(t *testing.T) {
		quitKey, ok := m.keyMap.GetBinding(actions.ActionQuit, actions.ScopeGlobal)
		if !ok {
			t.Error("Expected Quit binding to exist")
		}

		// Test Ctrl+C key which is the new quit key
		if !key.Matches(tea.KeyMsg{Type: tea.KeyCtrlC}, quitKey) {
			t.Error("Ctrl+C should match Quit binding")
		}
	})

	t.Run("Next tab key binding", func(t *testing.T) {
		nextTabKey, ok := m.keyMap.GetBinding(actions.ActionNextTab, actions.ScopeGlobal)
		if !ok {
			t.Error("Expected NextTab binding to exist")
		}

		if !key.Matches(tea.KeyMsg{Type: tea.KeyTab}, nextTabKey) {
			t.Error("Tab should match NextTab binding")
		}

		if !key.Matches(tea.KeyMsg{Type: tea.KeyRight}, nextTabKey) {
			t.Error("Right arrow should match NextTab binding")
		}
	})

	t.Run("Previous tab key binding", func(t *testing.T) {
		prevTabKey, ok := m.keyMap.GetBinding(actions.ActionPrevTab, actions.ScopeGlobal)
		if !ok {
			t.Error("Expected PrevTab binding to exist")
		}

		if !key.Matches(tea.KeyMsg{Type: tea.KeyShiftTab}, prevTabKey) {
			t.Error("Shift+Tab should match PrevTab binding")
		}

		if !key.Matches(tea.KeyMsg{Type: tea.KeyLeft}, prevTabKey) {
			t.Error("Left arrow should match PrevTab binding")
		}
	})

	t.Run("User key binding", func(t *testing.T) {
		userKey, ok := m.keyMap.GetBinding(actions.ActionGoToUser, actions.ScopeGlobal)
		if !ok {
			t.Error("Expected GoToUser binding to exist")
		}

		if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}, userKey) {
			t.Error("'u' key should match GoToUser binding")
		}
	})
}
