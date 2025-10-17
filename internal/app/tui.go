package app

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/jcserv/slacky/internal/app/views"
	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/models"
	"github.com/jcserv/slacky/internal/tui"
	"github.com/jcserv/slacky/internal/tui/actions"
	"github.com/jcserv/slacky/internal/tui/components/statusbar"
	"github.com/jcserv/slacky/internal/tui/components/tabs"
	tuiKeys "github.com/jcserv/slacky/internal/tui/keys"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// TUIModel holds the main application state
type TUIModel struct {
	app          *App
	spinner      spinner.Model
	keyMap       *tuiKeys.ScopedKeyMap
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
	showHelp     bool
}

// NewTUI creates a new TUI model with the given app
func (app *App) NewTUI() TUIModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Label

	// Create localizer with detected locale
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	// Load keybindings from config with localizer
	keyMap, err := tuiKeys.LoadKeybindings(app.config, localizer)
	if err != nil {
		slog.Warn("Failed to load keybindings, using defaults", "error", err)
		// Fallback to defaults
		keyMap, _ = tuiKeys.LoadKeybindings(nil, localizer)
	}

	chatView := views.NewChatModel()
	chatView.SetKeyMap(keyMap)

	return TUIModel{
		app:          app,
		spinner:      s,
		keyMap:       keyMap,
		tabs:         tabs.NewModel(),
		statusBar:    statusbar.NewModel(),
		chatView:     chatView,
		activityView: views.NewActivityModel(),
		userView:     views.NewUserModel(),
		version:      "v0.1.0-dev",
		localizer:    localizer,
		showHelp:     true, // Show help by default for new users
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

		// Calculate content height (total - tabs - status bar - borders - newlines)
		// Tabs: 2 lines (content + bottom border)
		// Status bar: 2 lines (top border + content)
		// Newlines: 2 lines (after tabs, before status bar)
		// Total: 6 lines
		contentHeight := msg.Height - 6

		m.chatView.SetSize(msg.Width, contentHeight)
		m.activityView.SetSize(msg.Width, contentHeight)
		m.userView.SetSize(msg.Width, contentHeight)

		return m, nil

	case tea.KeyMsg:
		// Handle quit (always available)
		if m.keyMap.MatchesAction(msg, actions.ActionQuit, actions.ScopeGlobal) {
			m.quitting = true
			return m, tea.Quit
		}

		// Handle help toggle (always available)
		if m.keyMap.MatchesAction(msg, actions.ActionToggleHelp, actions.ScopeGlobal) {
			m.showHelp = !m.showHelp
			return m, nil
		}

		// Only handle other actions after auth success
		if m.authSuccess {
			// Handle tab navigation
			if m.keyMap.MatchesAction(msg, actions.ActionNextTab, actions.ScopeGlobal) {
				m.tabs.NextTab()
				return m, nil
			}
			if m.keyMap.MatchesAction(msg, actions.ActionPrevTab, actions.ScopeGlobal) {
				m.tabs.PrevTab()
				return m, nil
			}

			// Handle direct view navigation
			if m.keyMap.MatchesAction(msg, actions.ActionGoToChat, actions.ScopeGlobal) {
				m.tabs.SetCurrentTab(tabs.ChatTab)
				return m, nil
			}
			if m.keyMap.MatchesAction(msg, actions.ActionGoToActivity, actions.ScopeGlobal) {
				m.tabs.SetCurrentTab(tabs.ActivityTab)
				return m, nil
			}
			if m.keyMap.MatchesAction(msg, actions.ActionGoToUser, actions.ScopeGlobal) {
				m.tabs.SetCurrentTab(tabs.UserTab)
				return m, nil
			}
		}

		// If we didn't handle the key, let it fall through to the active view

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

		// Load channels after successful auth
		return m, loadChannels(m.app.SlackClient)

	case channelsLoadedMsg:
		// Set channels in the chat view
		m.chatView.SetChannels(msg.channels)
		// Load starred conversations after channels are loaded
		return m, loadStarredConversations(m.app.SlackClient)

	case starredConversationsLoadedMsg:
		// Mark channels as starred and update the chat view
		currentChannels := m.chatView.GetChannels()
		updatedChannels := models.MarkAsStarred(currentChannels, msg.starredIDs)
		m.chatView.SetChannels(updatedChannels)
		return m, nil

	case messagesLoadedMsg:
		// Set messages in the chat view
		m.chatView.SetMessages(msg.messages)
		return m, nil

	case messageSentMsg:
		// Reload messages for the channel after sending
		selectedCh := m.chatView.GetSelectedChannel()
		if selectedCh != nil && selectedCh.ID == msg.channelID {
			return m, loadMessages(m.app.SlackClient, msg.channelID, 100)
		}
		return m, nil

	case views.SendChatMessageMsg:
		// Handle message send request from chat view
		return m, sendMessage(m.app.SlackClient, msg.ChannelID, msg.Text)

	case views.LoadChannelMessagesMsg:
		// Update status bar with selected channel name
		selectedCh := m.chatView.GetSelectedChannel()
		if selectedCh != nil {
			m.statusBar.SetCurrentChannel(selectedCh.Name)
		}
		// Load messages for the selected channel
		return m, loadMessages(m.app.SlackClient, msg.ChannelID, 100)

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
		quitKey, _ := m.keyMap.GetBinding(actions.ActionQuit, actions.ScopeGlobal)
		s.WriteString(tui.RenderKeyBindings(quitKey))
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
		quitKey, _ := m.keyMap.GetBinding(actions.ActionQuit, actions.ScopeGlobal)
		s.WriteString(tui.RenderKeyBindings(quitKey))
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

	// Update status bar with help keybindings if help is enabled
	if m.showHelp {
		var currentScope actions.ActionScope
		switch m.tabs.GetCurrentTab() {
		case tabs.ChatTab:
			currentScope = actions.ScopeChat
		case tabs.ActivityTab:
			currentScope = actions.ScopeActivity
		case tabs.UserTab:
			currentScope = actions.ScopeUser
		default:
			currentScope = actions.ScopeGlobal
		}

		// Get only essential bindings to avoid overwhelming help text
		bindings := m.keyMap.GetEssentialBindings(currentScope)
		helpText := tui.RenderKeyBindings(bindings...)
		m.statusBar.SetHelpText(helpText)
	} else {
		m.statusBar.SetHelpText("")
	}

	// Render status bar at the bottom
	s.WriteString("\n")
	s.WriteString(m.statusBar.View())

	return s.String()
}
