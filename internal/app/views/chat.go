package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// ChatModel represents the chat view
type ChatModel struct {
	width     int
	height    int
	localizer *i18n.Localizer
}

// NewChatModel creates a new chat view model
func NewChatModel() ChatModel {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	return ChatModel{
		width:     80,
		height:    24,
		localizer: localizer,
	}
}

// Init initializes the chat view
func (m ChatModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the chat view
func (m ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	return m, nil
}

// SetSize sets the dimensions of the chat view
func (m *ChatModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// View renders the chat view
func (m ChatModel) View() string {
	// Placeholder content
	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(2).
		Render(
			styles.Title.Render("💬 Chat View") + "\n\n" +
				styles.Subtitle.Render("This is where chat messages will appear.") + "\n\n" +
				styles.Dim.Render("Coming soon: Channel list, message history, and message input."),
		)

	return content
}

// localize is a helper function to localize a message by ID with an optional fallback
func (m ChatModel) localize(messageID string, fallback string) string {
	cfg := &i18n.LocalizeConfig{
		MessageID: messageID,
	}
	msg, err := m.localizer.Localize(cfg)
	if err != nil && fallback != "" {
		return fallback
	}
	return msg
}
