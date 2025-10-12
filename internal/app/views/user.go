package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// UserModel represents the user info view
type UserModel struct {
	width     int
	height    int
	userName  string
	teamName  string
	localizer *i18n.Localizer
}

// NewUserModel creates a new user info view model
func NewUserModel() UserModel {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	return UserModel{
		width:     80,
		height:    24,
		localizer: localizer,
	}
}

// Init initializes the user info view
func (m UserModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the user info view
func (m UserModel) Update(msg tea.Msg) (UserModel, tea.Cmd) {
	return m, nil
}

// SetSize sets the dimensions of the user info view
func (m *UserModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetUserInfo sets the user and team information
func (m *UserModel) SetUserInfo(userName, teamName string) {
	m.userName = userName
	m.teamName = teamName
}

// View renders the user info view
func (m UserModel) View() string {
	// User information
	userInfo := styles.Title.Render("👤 User Information") + "\n\n" +
		styles.Label.Render("User: ") + styles.Highlight.Render(m.userName) + "\n" +
		styles.Label.Render("Workspace: ") + styles.Info.Render(m.teamName) + "\n\n" +
		styles.Dim.Render("Coming soon: User preferences, status, and account settings.")

	// Placeholder content
	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(2).
		Render(userInfo)

	return content
}

// localize is a helper function to localize a message by ID with an optional fallback
func (m UserModel) localize(messageID string, fallback string) string {
	cfg := &i18n.LocalizeConfig{
		MessageID: messageID,
	}
	msg, err := m.localizer.Localize(cfg)
	if err != nil && fallback != "" {
		return fallback
	}
	return msg
}
