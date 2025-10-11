package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/tui"
	"github.com/jcserv/slacky/internal/tui/components"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// TUIModel holds the main application state
type TUIModel struct {
	app         *App
	spinner     spinner.Model
	keys        tui.KeyMap
	width       int
	version     string
	quitting    bool
	err         error
	authSuccess bool
	teamName    string
	userName    string
	localizer   *i18n.Localizer
}

// NewTUI creates a new TUI model with the given app
func (app *App) NewTUI() TUIModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Label

	// Create localizer with detected locale
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	return TUIModel{
		app:       app,
		spinner:   s,
		keys:      tui.DefaultKeyMap(),
		version:   "v0.1.0-dev", // TODO: Get from build info
		localizer: localizer,
	}
}

// Init initializes the application
func (m TUIModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, loadConfig(m.app))
}

// Update handles messages and updates the model
func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil

	case authSuccessMsg:
		m.authSuccess = true
		m.teamName = msg.teamName
		m.userName = msg.userName
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

// localize is a helper function to localize a message by ID with an optional fallback
func (m TUIModel) localize(messageID string, fallback string) string {
	cfg := &i18n.LocalizeConfig{
		MessageID: messageID,
	}
	msg, err := m.localizer.Localize(cfg)
	if err != nil && fallback != "" {
		return fallback
	}
	return msg
}

// View renders the application UI
func (m TUIModel) View() string {
	var s strings.Builder

	// Render logo
	s.WriteString("\n")
	s.WriteString(m.renderLogo())
	s.WriteString("\n\n")

	// Error state
	if m.err != nil {
		s.WriteString(tui.RenderBorder(63))
		s.WriteString("\n\n")
		s.WriteString(styles.Error.Render("✗ " + m.localize("error.general", "Error")))
		s.WriteString("\n\n")
		s.WriteString(styles.Subtitle.Render(fmt.Sprintf("  %v", m.err)))
		s.WriteString("\n\n")
		s.WriteString(tui.RenderBorder(63))
		s.WriteString("\n")
		s.WriteString(tui.RenderKeyBindings(m.keys.Quit))
		s.WriteString("\n")
		return s.String()
	}

	// Success state
	if m.authSuccess {
		s.WriteString(tui.RenderBorder(63))
		s.WriteString("\n\n")
		s.WriteString(styles.Success.Render(m.localize("tui.connected", "✓ Connected Successfully")))
		s.WriteString("\n\n")
		s.WriteString(styles.Label.Render(m.localize("tui.workspace_label", "Workspace:")+" ") + styles.Info.Render(m.teamName))
		s.WriteString("\n")
		s.WriteString(styles.Label.Render(m.localize("tui.user_label", "User:")+"      ") + styles.Highlight.Render(m.userName))
		s.WriteString("\n\n")
		s.WriteString(styles.Subtitle.Render(m.localize("tui.ready", "Ready to start messaging!")))
		s.WriteString("\n\n")
		s.WriteString(tui.RenderBorder(63))
		s.WriteString("\n")
		s.WriteString(tui.RenderKeyBindings(m.keys.Quit))
		s.WriteString("\n")
		return s.String()
	}

	// Loading state
	s.WriteString(tui.RenderBorder(63))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("%s %s",
		m.spinner.View(),
		styles.Label.Render(m.localize("tui.connecting", "Connecting to Slack..."))))
	s.WriteString("\n\n")
	s.WriteString(styles.Subtitle.Render(m.localize("tui.authenticating", "Authenticating with workspace")))
	s.WriteString("\n\n")
	s.WriteString(tui.RenderBorder(63))
	s.WriteString("\n")
	s.WriteString(tui.RenderKeyBindings(m.keys.Quit))
	s.WriteString("\n")

	return s.String()
}

// renderLogo renders the application logo
func (m TUIModel) renderLogo() string {
	if m.width < components.MinWidth() {
		return components.SmallRender(m.version, m.width)
	}

	return components.Render(components.Opts{
		Version:      m.version,
		Width:        m.width,
		FillColor:    styles.ColourDim,
		VersionColor: styles.Tertiary,
	})
}
