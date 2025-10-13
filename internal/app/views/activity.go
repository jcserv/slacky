package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// ActivityModel represents the activity view
type ActivityModel struct {
	width     int
	height    int
	localizer *i18n.Localizer
}

// NewActivityModel creates a new activity view model
func NewActivityModel() ActivityModel {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	return ActivityModel{
		width:     80,
		height:    24,
		localizer: localizer,
	}
}

// Init initializes the activity view
func (m ActivityModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the activity view
func (m ActivityModel) Update(msg tea.Msg) (ActivityModel, tea.Cmd) {
	return m, nil
}

// SetSize sets the dimensions of the activity view
func (m *ActivityModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// View renders the activity view
func (m ActivityModel) View() string {
	// Placeholder content
	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(2).
		Render(
			styles.Title.Render(m.localize("activity.title", "📊 Activity View")) + "\n\n" +
				styles.Subtitle.Render(m.localize("activity.subtitle", "This is where activity notifications will appear.")) + "\n\n" +
				styles.Dim.Render(m.localize("activity.coming_soon", "Coming soon: Mentions, reactions, and thread updates.")),
		)

	return content
}

// localize is a helper function to localize a message by ID with an optional fallback
func (m ActivityModel) localize(messageID string, fallback string) string {
	cfg := &i18n.LocalizeConfig{
		MessageID: messageID,
	}
	msg, err := m.localizer.Localize(cfg)
	if err != nil && fallback != "" {
		return fallback
	}
	return msg
}
