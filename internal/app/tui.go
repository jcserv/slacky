package app

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/jcserv/slacky/internal/app/views"
	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/constants"
	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/models"
	"github.com/jcserv/slacky/internal/tui"
	"github.com/jcserv/slacky/internal/tui/actions"
	"github.com/jcserv/slacky/internal/tui/components/messageview"
	"github.com/jcserv/slacky/internal/tui/components/statusbar"
	"github.com/jcserv/slacky/internal/tui/components/tabs"
	tuiKeys "github.com/jcserv/slacky/internal/tui/keys"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// FocusLevel represents whether the user is navigating at tab level or inside a view.
// FocusLevelTab allows switching tabs with Tab key, while FocusLevelView
// allows Tab to cycle through elements within the active view.
type FocusLevel int

const (
	FocusLevelTab FocusLevel = iota
	FocusLevelView
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
	userID       string
	localizer    *i18n.Localizer
	showHelp     bool
	focusLevel   FocusLevel

	// Polling state
	pollingConfig    PollingConfig
	lastMessageTS    map[string]string // Map of channelID -> last message timestamp
	currentChannelID string            // Currently selected channel ID for polling
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

	// Load polling config from app config or use defaults
	pollingConfig := DefaultPollingConfig()
	if app.config.Polling != nil {
		pollingConfig = loadPollingConfigFromConfig(app.config.Polling)
	}

	return TUIModel{
		app:           app,
		spinner:       s,
		keyMap:        keyMap,
		tabs:          tabs.NewModel(),
		statusBar:     statusbar.NewModel(),
		chatView:      chatView,
		activityView:  views.NewActivityModel(),
		userView:      views.NewUserModel(),
		version:       "v0.1.0-dev",
		localizer:     localizer,
		showHelp:      true,          // Show help by default for new users
		focusLevel:    FocusLevelTab, // Start at tab level
		pollingConfig: pollingConfig,
		lastMessageTS: make(map[string]string),
	}
}

// loadPollingConfigFromConfig converts config.Polling to PollingConfig
func loadPollingConfigFromConfig(p *config.Polling) PollingConfig {
	return PollingConfig{
		Enabled:                p.Enabled,
		CurrentChannelInterval: time.Duration(p.CurrentChannelInterval) * time.Second,
		SidebarInterval:        time.Duration(p.SidebarInterval) * time.Second,
		ActivityInterval:       time.Duration(p.ActivityInterval) * time.Second,
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

		m.tabs.SetWidth(msg.Width)
		m.statusBar.SetWidth(msg.Width)

		// Content height accounts for tabs (2 lines), status bar (2 lines), and spacing (2 lines)
		contentHeight := msg.Height - constants.UIHeaderHeight

		m.chatView.SetSize(msg.Width, contentHeight)
		m.activityView.SetSize(msg.Width, contentHeight)
		m.userView.SetSize(msg.Width, contentHeight)

		return m, nil

	case tea.KeyMsg:
		if m.keyMap.MatchesAction(msg, actions.ActionQuit, actions.ScopeGlobal) {
			m.quitting = true
			return m, tea.Quit
		}

		if m.keyMap.MatchesAction(msg, actions.ActionToggleHelp, actions.ScopeGlobal) {
			m.showHelp = !m.showHelp
			return m, nil
		}

		// Focus management: Two-level navigation system
		// 1. Tab level (FocusLevelTab): Navigate between tabs (Chat, Activity, User)
		// 2. View level (FocusLevelView): Navigate within the active view
		//
		// User flow:
		// - Start at tab level, use Tab/Shift+Tab to switch tabs
		// - Press Space to enter current view (FocusLevelView)
		// - Press Escape to exit back to tab level
		// - Special case: Chat view has nested navigation (sidebar → messages → threads)
		//   so Escape only exits to tab level when sidebar is focused
		if m.authSuccess {
			if m.focusLevel == FocusLevelTab {
				if m.keyMap.MatchesAction(msg, actions.ActionNextTab, actions.ScopeGlobal) {
					m.tabs.NextTab()
					return m, nil
				}
				if m.keyMap.MatchesAction(msg, actions.ActionPrevTab, actions.ScopeGlobal) {
					m.tabs.PrevTab()
					return m, nil
				}

				if m.keyMap.MatchesAction(msg, actions.ActionEnterView, actions.ScopeGlobal) {
					m.focusLevel = FocusLevelView
					m.enterCurrentView()
					return m, nil
				}

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
			} else {
				// In chat view, only exit to tab level when sidebar is focused.
				// This allows Escape to be used for exiting threads first.
				if m.keyMap.MatchesAction(msg, actions.ActionExitView, actions.ScopeGlobal) {
					if m.tabs.GetCurrentTab() == tabs.ChatTab {
						if m.chatView.IsSidebarFocused() {
							m.focusLevel = FocusLevelTab
							m.exitCurrentView()
							return m, nil
						}
					} else {
						m.focusLevel = FocusLevelTab
						m.exitCurrentView()
						return m, nil
					}
				}
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
		m.userID = msg.userID

		// Update components with user info
		m.tabs.SetUserName(msg.userName)
		m.userView.SetUserInfo(msg.userName, msg.teamName)
		m.chatView.SetUserID(msg.userID)
		m.statusBar.SetConnected(true)

		// Set focus on the initial tab
		m.updateViewFocus()

		// Load channels and start polling after successful auth
		return m, tea.Batch(
			loadChannels(m.app.SlackClient),
			startPolling(m.pollingConfig),
		)

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
		// Load activities after channels and starred conversations are ready
		return m, loadActivities(m.app.SlackClient, m.userID, updatedChannels)

	case activitiesLoadedMsg:
		// Set activities in the activity view
		m.activityView.SetActivities(msg.activities)
		return m, nil

	case activitySelectedMsg:
		// Switch to chat tab and select the channel
		m.tabs.SetCurrentTab(tabs.ChatTab)
		m.updateViewFocus()
		m.chatView.SelectChannel(msg.channelID)
		// Load messages for the selected channel
		return m, loadMessages(m.app.SlackClient, msg.channelID)

	case messagesLoadedMsg:
		// Set messages in the chat view
		m.chatView.SetMessages(msg.messages)

		// Track last message timestamp for polling
		if len(msg.messages) > 0 {
			lastMsg := msg.messages[len(msg.messages)-1]
			m.lastMessageTS[msg.channelID] = lastMsg.GetTimestamp()
		}

		// Update current channel ID for polling
		m.currentChannelID = msg.channelID
		return m, nil

	case threadRepliesLoadedMsg:
		// Set thread replies in the chat view
		// Find the channel by ID to get its display name
		var channelName string
		for _, ch := range m.chatView.GetChannels() {
			if ch.ID == msg.channelID {
				channelName = ch.GetDisplayName()
				break
			}
		}
		// If we couldn't find the channel, use a fallback but still show the thread
		if channelName == "" {
			slog.Warn("Could not find channel name for thread", "channelID", msg.channelID)
			channelName = "unknown"
		}
		m.chatView.SetThreadReplies(msg.channelID, channelName, msg.threadTS, &msg.parentMessage, msg.messages)
		return m, nil

	case messageSentMsg:
		// Reload messages for the channel after sending
		selectedCh := m.chatView.GetSelectedChannel()
		if selectedCh != nil && selectedCh.ID == msg.channelID {
			return m, loadMessages(m.app.SlackClient, msg.channelID)
		}
		return m, nil

	case threadReplySentMsg:
		// Reload thread replies after sending
		// We need to reload to get the new reply with correct metadata
		return m, loadThreadReplies(m.app.SlackClient, msg.channelID, msg.threadTS, models.Message{})

	case reactionAddedMsg:
		// Reaction was successfully added - reload messages to show the update
		selectedCh := m.chatView.GetSelectedChannel()
		if selectedCh != nil && selectedCh.ID == msg.channelID {
			return m, loadMessages(m.app.SlackClient, msg.channelID)
		}
		return m, nil

	case reactionRemovedMsg:
		// Reaction was successfully removed - reload messages to show the update
		selectedCh := m.chatView.GetSelectedChannel()
		if selectedCh != nil && selectedCh.ID == msg.channelID {
			return m, loadMessages(m.app.SlackClient, msg.channelID)
		}
		return m, nil

	case reactionErrorMsg:
		// Handle reaction error
		slog.Error("reaction error", "error", msg.err)
		return m, nil

	case views.SendChatMessageMsg:
		// Handle message send request from chat view
		return m, sendMessage(m.app.SlackClient, msg.ChannelID, msg.Text)

	case views.SendThreadMessageMsg:
		// Handle thread reply send request from chat view
		return m, sendThreadReply(m.app.SlackClient, msg.ChannelID, msg.ThreadTS, msg.Text)

	case views.LoadChannelMessagesMsg:
		// Update status bar with selected channel name
		selectedCh := m.chatView.GetSelectedChannel()
		if selectedCh != nil {
			m.statusBar.SetCurrentChannel(selectedCh.Name)
		}
		// Load messages for the selected channel
		return m, loadMessages(m.app.SlackClient, msg.ChannelID)

	case views.LoadThreadRepliesMsg:
		// Load thread replies
		return m, loadThreadReplies(m.app.SlackClient, msg.ChannelID, msg.ThreadTS, msg.ParentMessage)

	case messageview.ReactionToggleRequestMsg:
		// User wants to toggle a reaction on a message
		// Get the selected message to access its current reactions
		selectedMsg := m.chatView.GetSelectedMessage()
		if selectedMsg != nil {
			return m, toggleReaction(
				m.app.SlackClient,
				msg.ChannelID,
				msg.Timestamp,
				msg.EmojiName,
				m.userID,
				selectedMsg.Reactions,
			)
		}
		return m, nil

	case views.ActivitySelectedMsg:
		// Convert to internal activitySelectedMsg
		return m, func() tea.Msg {
			return activitySelectedMsg{
				channelID: msg.ChannelID,
				threadTS:  msg.ThreadTS,
			}
		}

	case pollTickMsg:
		// Handle polling ticks
		switch msg.pollType {
		case PollTypeCurrentChannel:
			// Poll current channel for new messages
			lastTS := m.lastMessageTS[m.currentChannelID]
			cmd := pollCurrentChannelMessages(m.app.SlackClient, m.currentChannelID, lastTS)
			// Schedule next poll
			nextPoll := pollCurrentChannel(m.pollingConfig.CurrentChannelInterval)
			return m, tea.Batch(cmd, nextPoll)

		case PollTypeSidebar:
			// Poll sidebar for unread counts
			channelIDs := make([]string, 0)
			for _, ch := range m.chatView.GetChannels() {
				channelIDs = append(channelIDs, ch.ID)
			}
			cmd := pollSidebarUnreads(m.app.SlackClient, channelIDs)
			// Schedule next poll
			nextPoll := pollSidebar(m.pollingConfig.SidebarInterval)
			return m, tea.Batch(cmd, nextPoll)

		case PollTypeActivity:
			// Poll for new activities
			cmd := pollActivitiesUpdates(m.app.SlackClient, m.userID, m.chatView.GetChannels())
			// Schedule next poll
			nextPoll := pollActivity(m.pollingConfig.ActivityInterval)
			return m, tea.Batch(cmd, nextPoll)
		}

	case newMessagesPolledMsg:
		// Handle new messages from polling
		if msg.channelID == m.currentChannelID && len(msg.messages) > 0 {
			// Add new messages to the current view
			for _, newMsg := range msg.messages {
				m.chatView.AddMessage(newMsg)
			}

			// Update last message timestamp
			lastMsg := msg.messages[len(msg.messages)-1]
			m.lastMessageTS[msg.channelID] = lastMsg.GetTimestamp()
		}
		return m, nil

	case channelUnreadUpdatedMsg:
		// Update sidebar unread indicator
		m.chatView.UpdateChannelUnread(msg.channelID, msg.unreadCount, msg.hasUnread)
		return m, nil

	case activitiesPolledMsg:
		// Update activity view with new activities
		m.activityView.AppendActivities(msg.activities)
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

// enterCurrentView is called when the user presses Space at tab level to enter a view
func (m *TUIModel) enterCurrentView() {
	currentTab := m.tabs.GetCurrentTab()

	switch currentTab {
	case tabs.ChatTab:
		// Enter chat view - start with sidebar focused
		m.chatView.EnterView()
	case tabs.ActivityTab:
		// Enter activity view - focus on the activity list
		m.activityView.SetFocused(true)
	case tabs.UserTab:
		// User view doesn't have interactive elements, but we still mark as entered
		// No specific action needed
	}
}

// exitCurrentView is called when the user presses Esc at view level to return to tab level
func (m *TUIModel) exitCurrentView() {
	currentTab := m.tabs.GetCurrentTab()

	switch currentTab {
	case tabs.ChatTab:
		// Exit chat view
		m.chatView.ExitView()
	case tabs.ActivityTab:
		// Exit activity view
		m.activityView.SetFocused(false)
	case tabs.UserTab:
		// User view has no special exit logic
	}
}

// updateViewFocus sets focus on the active view and removes focus from others
// Deprecated: This is kept for compatibility but should be phased out in favor of enterCurrentView/exitCurrentView
func (m *TUIModel) updateViewFocus() {
	currentTab := m.tabs.GetCurrentTab()

	// Set focus based on current tab
	switch currentTab {
	case tabs.ChatTab:
		// Chat view doesn't have a SetFocused method, it manages its own focus
		m.activityView.SetFocused(false)
	case tabs.ActivityTab:
		m.activityView.SetFocused(true)
	case tabs.UserTab:
		// User view doesn't need focus (mostly static info)
		m.activityView.SetFocused(false)
	}
}
