package statusbar

import (
	"strings"
	"testing"
	"time"

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

	if m.localizer == nil {
		t.Error("Expected localizer to be initialized")
	}

	if m.isConnected {
		t.Error("Expected isConnected to be false initially")
	}

	if m.currentChannel != "" {
		t.Error("Expected currentChannel to be empty initially")
	}

	// Check that time is initialized
	if m.currentTime.IsZero() {
		t.Error("Expected currentTime to be initialized")
	}
}

func TestInit(t *testing.T) {
	m := NewModel()

	cmd := m.Init()

	if cmd == nil {
		t.Error("Expected Init to return a tick command")
	}
}

func TestSetWidth(t *testing.T) {
	m := NewModel()

	m.SetWidth(120)

	if m.statusbar.Width != 120 {
		t.Errorf("Expected width 120, got %d", m.statusbar.Width)
	}
}

func TestSetCurrentChannel(t *testing.T) {
	m := NewModel()

	m.SetCurrentChannel("general")

	if m.currentChannel != "general" {
		t.Errorf("Expected currentChannel 'general', got %s", m.currentChannel)
	}

	// Verify content was updated
	if !strings.Contains(m.statusbar.FirstColumn, "general") {
		t.Error("Expected FirstColumn to contain channel name")
	}
}

func TestSetConnected(t *testing.T) {
	m := NewModel()

	t.Run("Set connected to true", func(t *testing.T) {
		m.SetConnected(true)

		if !m.isConnected {
			t.Error("Expected isConnected to be true")
		}

		// Verify content was updated
		if !strings.Contains(m.statusbar.ThirdColumn, "Connected") {
			t.Error("Expected ThirdColumn to show connected status")
		}
	})

	t.Run("Set connected to false", func(t *testing.T) {
		m.SetConnected(false)

		if m.isConnected {
			t.Error("Expected isConnected to be false")
		}

		// Verify content was updated
		if !strings.Contains(m.statusbar.ThirdColumn, "Disconnected") {
			t.Error("Expected ThirdColumn to show disconnected status")
		}
	})
}

func TestUpdateWithWindowSize(t *testing.T) {
	m := NewModel()

	msg := tea.WindowSizeMsg{Width: 150, Height: 50}
	updatedModel, _ := m.Update(msg)

	if updatedModel.statusbar.Width != 150 {
		t.Errorf("Expected width to be updated to 150, got %d", updatedModel.statusbar.Width)
	}
}

func TestUpdateWithTickMsg(t *testing.T) {
	m := NewModel()
	m.SetConnected(true)
	m.SetCurrentChannel("test")

	// Create a TickMsg with a specific time
	testTime := time.Date(2024, 1, 1, 12, 30, 45, 0, time.UTC)
	msg := TickMsg(testTime)

	updatedModel, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("Expected Update to return a tick command")
	}

	if !updatedModel.currentTime.Equal(testTime) {
		t.Errorf("Expected currentTime to be updated to %v, got %v", testTime, updatedModel.currentTime)
	}

	// Verify the time is reflected in the fourth column
	expectedTime := testTime.Format("15:04:05")
	if !strings.Contains(updatedModel.statusbar.FourthColumn, expectedTime) {
		t.Errorf("Expected FourthColumn to contain %s, got %s", expectedTime, updatedModel.statusbar.FourthColumn)
	}
}

func TestView(t *testing.T) {
	m := NewModel()
	m.SetWidth(100)
	m.SetCurrentChannel("general")
	m.SetConnected(true)

	view := m.View()

	if view == "" {
		t.Error("View should not return empty string")
	}

	// View should contain channel info
	if !strings.Contains(view, "general") {
		t.Error("View should contain channel name")
	}

	// View should contain connection status
	if !strings.Contains(view, "Connected") {
		t.Error("View should contain connection status")
	}
}

func TestUpdateContent(t *testing.T) {
	m := NewModel()

	t.Run("With channel name", func(t *testing.T) {
		m.SetCurrentChannel("random")
		m.SetConnected(true)

		if !strings.Contains(m.statusbar.FirstColumn, "#random") {
			t.Error("FirstColumn should contain channel with # prefix")
		}

		if !strings.Contains(m.statusbar.ThirdColumn, "Connected") {
			t.Error("ThirdColumn should show connected")
		}
	})

	t.Run("Without channel name", func(t *testing.T) {
		m.SetCurrentChannel("")
		m.SetConnected(false)

		if !strings.Contains(m.statusbar.FirstColumn, "No channel") {
			t.Error("FirstColumn should show 'No channel' message")
		}

		if !strings.Contains(m.statusbar.ThirdColumn, "Disconnected") {
			t.Error("ThirdColumn should show disconnected")
		}
	})

	t.Run("Time formatting", func(t *testing.T) {
		testTime := time.Date(2024, 12, 25, 15, 30, 45, 0, time.UTC)
		m.currentTime = testTime
		m.updateContent()

		expectedTime := "15:30:45"
		if !strings.Contains(m.statusbar.FourthColumn, expectedTime) {
			t.Errorf("FourthColumn should contain time %s, got %s", expectedTime, m.statusbar.FourthColumn)
		}
	})
}

func TestStatusBarColumns(t *testing.T) {
	m := NewModel()
	m.SetWidth(100)
	m.SetCurrentChannel("general")
	m.SetConnected(true)

	// Verify all four columns are set
	if m.statusbar.FirstColumn == "" {
		t.Error("FirstColumn should not be empty")
	}

	// SecondColumn is intentionally empty for now
	if m.statusbar.SecondColumn != "" {
		t.Error("SecondColumn should be empty")
	}

	if m.statusbar.ThirdColumn == "" {
		t.Error("ThirdColumn should not be empty")
	}

	if m.statusbar.FourthColumn == "" {
		t.Error("FourthColumn should not be empty")
	}
}
