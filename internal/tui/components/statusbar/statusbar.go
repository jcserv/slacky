package statusbar

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/statusbar"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
)

// TickMsg is sent when the clock should update
type TickMsg time.Time

// Model wraps the teacup statusbar with additional state
type Model struct {
	statusbar      statusbar.Model
	currentTime    time.Time
	currentChannel string
	isConnected    bool
	helpText       string // Help keybindings to display in second column
	localizer      *i18n.Localizer
}

// NewModel creates a new status bar model with default colors
func NewModel() Model {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	// Create teacup statusbar with colors matching our theme
	sb := statusbar.New(
		// First column - channel (white on dark)
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#444a73", Dark: "#444a73"},
		},
		// Second column - empty/filler (white on dark)
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#444a73", Dark: "#444a73"},
		},
		// Third column - connection status (green on dark)
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Light: "#86e1b3", Dark: "#86e1b3"},
			Background: lipgloss.AdaptiveColor{Light: "#444a73", Dark: "#444a73"},
		},
		// Fourth column - time (dim on dark)
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Light: "#828bb8", Dark: "#828bb8"},
			Background: lipgloss.AdaptiveColor{Light: "#444a73", Dark: "#444a73"},
		},
	)

	return Model{
		statusbar:   sb,
		currentTime: time.Now(),
		localizer:   localizer,
	}
}

// Init initializes the status bar with a tick command for the clock
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// SetWidth sets the width of the status bar
func (m *Model) SetWidth(width int) {
	m.statusbar.SetSize(width)
}

// Update updates the statusbar and handles tick messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.statusbar.SetSize(msg.Width)
	case TickMsg:
		m.currentTime = time.Time(msg)
		m.updateContent()
		return m, tickCmd()
	}

	m.statusbar, cmd = m.statusbar.Update(msg)
	return m, cmd
}

// SetCurrentChannel sets the current channel name
func (m *Model) SetCurrentChannel(channel string) {
	m.currentChannel = channel
	m.updateContent()
}

// SetConnected sets the connection status
func (m *Model) SetConnected(connected bool) {
	m.isConnected = connected
	m.updateContent()
}

// SetHelpText sets the help text to display in the second column
func (m *Model) SetHelpText(helpText string) {
	m.helpText = helpText
	m.updateContent()
}

// updateContent updates the statusbar content based on current state
func (m *Model) updateContent() {
	// First column: channel
	var firstCol string
	if m.currentChannel != "" {
		firstCol = "#" + m.currentChannel
	} else {
		firstCol = m.localize("statusbar.no_channel", "No channel")
	}

	// Second column: help keybindings (if set)
	secondCol := m.helpText

	// Third column: connection status
	var thirdCol string
	if m.isConnected {
		thirdCol = "● " + m.localize("statusbar.connected", "Connected")
	} else {
		thirdCol = "○ " + m.localize("statusbar.disconnected", "Disconnected")
	}

	// Fourth column: time
	fourthCol := m.currentTime.Format("15:04:05")

	m.statusbar.SetContent(firstCol, secondCol, thirdCol, fourthCol)
}

// View renders the status bar
func (m Model) View() string {
	return m.statusbar.View()
}

// localize is a helper function to localize a message by ID with an optional fallback
func (m Model) localize(messageID string, fallback string) string {
	cfg := &i18n.LocalizeConfig{
		MessageID: messageID,
	}
	msg, err := m.localizer.Localize(cfg)
	if err != nil && fallback != "" {
		return fallback
	}
	return msg
}

// tickCmd returns a command that sends a TickMsg every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
