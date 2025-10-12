package tabs

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

func TestNewModel(t *testing.T) {
	m := NewModel()

	if m.currentTab != ChatTab {
		t.Errorf("Expected initial tab to be ChatTab, got %v", m.currentTab)
	}

	if m.width != 80 {
		t.Errorf("Expected default width 80, got %d", m.width)
	}

	if m.localizer == nil {
		t.Error("Expected localizer to be initialized")
	}
}

func TestSetCurrentTab(t *testing.T) {
	m := NewModel()

	m.SetCurrentTab(ActivityTab)
	if m.currentTab != ActivityTab {
		t.Errorf("Expected ActivityTab, got %v", m.currentTab)
	}

	m.SetCurrentTab(UserTab)
	if m.currentTab != UserTab {
		t.Errorf("Expected UserTab, got %v", m.currentTab)
	}
}

func TestGetCurrentTab(t *testing.T) {
	m := NewModel()

	if m.GetCurrentTab() != ChatTab {
		t.Errorf("Expected ChatTab, got %v", m.GetCurrentTab())
	}

	m.currentTab = ActivityTab
	if m.GetCurrentTab() != ActivityTab {
		t.Errorf("Expected ActivityTab, got %v", m.GetCurrentTab())
	}
}

func TestSetWidth(t *testing.T) {
	m := NewModel()

	m.SetWidth(120)
	if m.width != 120 {
		t.Errorf("Expected width 120, got %d", m.width)
	}
}

func TestSetUserName(t *testing.T) {
	m := NewModel()

	m.SetUserName("testuser")
	if m.userName != "testuser" {
		t.Errorf("Expected userName 'testuser', got %s", m.userName)
	}
}

func TestNextTab(t *testing.T) {
	tests := []struct {
		name        string
		currentTab  TabType
		expectedTab TabType
	}{
		{"Chat to Activity", ChatTab, ActivityTab},
		{"Activity to User", ActivityTab, UserTab},
		{"User to Chat", UserTab, ChatTab},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.currentTab = tt.currentTab

			m.NextTab()

			if m.currentTab != tt.expectedTab {
				t.Errorf("Expected %v, got %v", tt.expectedTab, m.currentTab)
			}
		})
	}
}

func TestPrevTab(t *testing.T) {
	tests := []struct {
		name        string
		currentTab  TabType
		expectedTab TabType
	}{
		{"Chat to User", ChatTab, UserTab},
		{"Activity to Chat", ActivityTab, ChatTab},
		{"User to Activity", UserTab, ActivityTab},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.currentTab = tt.currentTab

			m.PrevTab()

			if m.currentTab != tt.expectedTab {
				t.Errorf("Expected %v, got %v", tt.expectedTab, m.currentTab)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	m := NewModel()

	// Test that Update returns the model unchanged for now
	updatedModel, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if cmd != nil {
		t.Error("Expected no command from Update")
	}

	// Model should be returned (even if unchanged)
	if updatedModel.width != m.width {
		t.Error("Model should maintain state through Update")
	}
}

func TestView(t *testing.T) {
	m := NewModel()
	m.SetWidth(100)
	m.SetUserName("testuser")

	t.Run("View returns content", func(t *testing.T) {
		view := m.View()

		if view == "" {
			t.Error("View should not return empty string")
		}
	})

	t.Run("View contains tab labels", func(t *testing.T) {
		view := m.View()

		// Should contain Chat and Activity labels
		if !strings.Contains(view, "Chat") {
			t.Error("View should contain 'Chat' label")
		}

		if !strings.Contains(view, "Activity") {
			t.Error("View should contain 'Activity' label")
		}
	})

	t.Run("View contains username when set", func(t *testing.T) {
		view := m.View()

		if !strings.Contains(view, "@testuser") {
			t.Error("View should contain username with @ prefix")
		}
	})

	t.Run("View handles zero width", func(t *testing.T) {
		m.SetWidth(0)
		view := m.View()

		if view != "" {
			t.Error("View should return empty string when width is 0")
		}
	})

	t.Run("View renders without username", func(t *testing.T) {
		m := NewModel()
		m.SetWidth(100)
		// Don't set username

		view := m.View()

		if view == "" {
			t.Error("View should render even without username")
		}

		// Should still contain tabs
		if !strings.Contains(view, "Chat") {
			t.Error("View should contain tabs even without username")
		}
	})
}

func TestTabCycling(t *testing.T) {
	m := NewModel()

	// Cycle through tabs with NextTab
	if m.GetCurrentTab() != ChatTab {
		t.Error("Should start at ChatTab")
	}

	m.NextTab()
	if m.GetCurrentTab() != ActivityTab {
		t.Error("NextTab from Chat should go to Activity")
	}

	m.NextTab()
	if m.GetCurrentTab() != UserTab {
		t.Error("NextTab from Activity should go to User")
	}

	m.NextTab()
	if m.GetCurrentTab() != ChatTab {
		t.Error("NextTab from User should cycle back to Chat")
	}

	// Cycle through tabs with PrevTab
	m.PrevTab()
	if m.GetCurrentTab() != UserTab {
		t.Error("PrevTab from Chat should go to User")
	}

	m.PrevTab()
	if m.GetCurrentTab() != ActivityTab {
		t.Error("PrevTab from User should go to Activity")
	}

	m.PrevTab()
	if m.GetCurrentTab() != ChatTab {
		t.Error("PrevTab from Activity should go to Chat")
	}
}
