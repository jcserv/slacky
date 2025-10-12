package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/jcserv/slacky/internal/app/views"
	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/tui"
	"github.com/jcserv/slacky/internal/tui/components/statusbar"
	"github.com/jcserv/slacky/internal/tui/components/tabs"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// TUIModel holds the main application state
type TUIModel struct {
	app          *App
	spinner      spinner.Model
	keys         tui.KeyMap
	tabs         tabs.Model
	statusBar    statusbar.Model
	chatView     views.ChatModel
	activityView views.ActivityModel
	userView     views.UserModel
	width        int
	height       int
	version      string
	quitting     bool
	err          error
	authSuccess  bool
	teamName     string
	userName     string
	localizer    *i18n.Localizer
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
		app:          app,
		spinner:      s,
		keys:         tui.DefaultKeyMap(),
		tabs:         tabs.NewModel(),
		statusBar:    statusbar.NewModel(),
		chatView:     views.NewChatModel(),
		activityView: views.NewActivityModel(),
		userView:     views.NewUserModel(),
		version:      "v0.1.0-dev", // TODO: Get from build info
		localizer:    localizer,
	}
}

// Init initializes the application
func (m TUIModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.statusBar.Init(), loadConfig(m.app))
}

// Update handles messages and updates the model
func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update component sizes
		m.tabs.SetWidth(msg.Width)
		m.statusBar.SetWidth(msg.Width)

		// Calculate content height (total - tabs - status bar - borders)
		contentHeight := msg.Height - 4 // Reserve space for tabs and status bar

		m.chatView.SetSize(msg.Width, contentHeight)
		m.activityView.SetSize(msg.Width, contentHeight)
		m.userView.SetSize(msg.Width, contentHeight)

		return m, nil

	case tea.KeyMsg:
		// Handle quit
		if key.Matches(msg, m.keys.Quit) {
			m.quitting = true
			return m, tea.Quit
		}

		// Only handle tab navigation after auth success
		if m.authSuccess {
			// Handle tab navigation
			if key.Matches(msg, m.keys.NextTab) {
				m.tabs.NextTab()
				return m, nil
			}
			if key.Matches(msg, m.keys.PrevTab) {
				m.tabs.PrevTab()
				return m, nil
			}
			if key.Matches(msg, m.keys.SelectUser) {
				m.tabs.SetCurrentTab(tabs.UserTab)
				return m, nil
			}
		}

		return m, nil

	case errMsg:
		m.err = msg
		return m, nil

	case authSuccessMsg:
		m.authSuccess = true
		m.teamName = msg.teamName
		m.userName = msg.userName

		// Update components with user info
		m.tabs.SetUserName(msg.userName)
		m.userView.SetUserInfo(msg.userName, msg.teamName)
		m.statusBar.SetConnected(true)

		return m, nil

	case statusbar.TickMsg:
		// Update status bar with tick
		var cmd tea.Cmd
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update spinner for loading state
	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)
	cmds = append(cmds, spinnerCmd)

	// Update active view
	if m.authSuccess {
		switch m.tabs.GetCurrentTab() {
		case tabs.ChatTab:
			var cmd tea.Cmd
			m.chatView, cmd = m.chatView.Update(msg)
			cmds = append(cmds, cmd)
		case tabs.ActivityTab:
			var cmd tea.Cmd
			m.activityView, cmd = m.activityView.Update(msg)
			cmds = append(cmds, cmd)
		case tabs.UserTab:
			var cmd tea.Cmd
			m.userView, cmd = m.userView.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
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

	// Error state - simple centered error
	if m.err != nil {
		s.WriteString("\n")
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

	// Loading state - simple centered spinner
	if !m.authSuccess {
		s.WriteString("\n")
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

	// Main application state with tabs
	// Render tabs at the top
	s.WriteString(m.tabs.View())
	s.WriteString("\n")

	// Render the active view content
	var content string
	switch m.tabs.GetCurrentTab() {
	case tabs.ChatTab:
		content = m.chatView.View()
	case tabs.ActivityTab:
		content = m.activityView.View()
	case tabs.UserTab:
		content = m.userView.View()
	default:
		content = m.chatView.View()
	}
	s.WriteString(content)

	// Render status bar at the bottom
	s.WriteString("\n")
	s.WriteString(m.statusBar.View())

	return s.String()
}
