package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/models"
	"github.com/jcserv/slacky/internal/tui/actions"
	"github.com/jcserv/slacky/internal/tui/components/input"
	"github.com/jcserv/slacky/internal/tui/components/messages"
	"github.com/jcserv/slacky/internal/tui/components/sidebar"
	"github.com/jcserv/slacky/internal/tui/components/thread"
	"github.com/jcserv/slacky/internal/tui/keys"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// FocusedComponent represents which component is currently focused
type FocusedComponent int

const (
	FocusSidebar FocusedComponent = iota
	FocusMessages
	FocusThread
	FocusInput
)

// SendChatMessageMsg is sent when the user wants to send a message
type SendChatMessageMsg struct {
	ChannelID string
	Text      string
}

// SendThreadMessageMsg is sent when the user wants to send a thread reply
type SendThreadMessageMsg struct {
	ChannelID string
	ThreadTS  string
	Text      string
}

// ChannelSelectedMsg is sent when a channel is selected
type ChannelSelectedMsg struct {
	ChannelID string
}

// LoadChannelMessagesMsg is sent to request loading messages for a channel
type LoadChannelMessagesMsg struct {
	ChannelID string
}

// LoadThreadRepliesMsg is sent to request loading thread replies
type LoadThreadRepliesMsg struct {
	ChannelID     string
	ThreadTS      string
	ParentMessage models.Message
}

// ChatModel represents the chat view
type ChatModel struct {
	sidebar  sidebar.Model
	messages messages.Model
	thread   thread.Model
	input    input.Model

	width     int
	height    int
	localizer *i18n.Localizer
	keyMap    *keys.ScopedKeyMap

	// State
	selectedChannel *models.Channel
	focused         FocusedComponent
	threadActive    bool // Whether thread view is active
}

// NewChatModel creates a new chat view model
func NewChatModel() ChatModel {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	return ChatModel{
		sidebar:      sidebar.NewModel(),
		messages:     messages.NewModel(),
		thread:       thread.NewModel(),
		input:        input.NewModel(),
		width:        80,
		height:       24,
		localizer:    localizer,
		focused:      FocusSidebar, // Start with sidebar focused
		threadActive: false,
	}
}

// Init initializes the chat view
func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(
		m.sidebar.Init(),
		m.messages.Init(),
		m.thread.Init(),
		m.input.Init(),
	)
}

// Update handles messages for the chat view
func (m ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle focus switching with keyMap if available
		if m.keyMap != nil {
			// Space key: jump to input if channel selected, otherwise focus sidebar
			if m.keyMap.MatchesAction(msg, actions.ActionBeginInput, actions.ScopeChat) {
				if m.selectedChannel != nil && m.focused != FocusInput {
					// Channel selected: jump to input
					m.setFocus(FocusInput)
					return m, nil
				} else if m.selectedChannel == nil {
					// No channel selected: focus sidebar (always ensure it's highlighted)
					m.setFocus(FocusSidebar)
					return m, nil
				}
			}

			// Vertical navigation (down/up) for messages <-> input
			if m.focused == FocusMessages {
				if m.keyMap.MatchesAction(msg, actions.ActionDown, actions.ScopeGlobal) {
					m.setFocus(FocusInput)
					return m, nil
				}
			}
			if m.focused == FocusInput {
				if m.keyMap.MatchesAction(msg, actions.ActionUp, actions.ScopeGlobal) {
					m.setFocus(FocusMessages)
					return m, nil
				}
			}

			// Horizontal navigation (left/right) for sidebar <-> content
			if m.focused != FocusInput {
				if m.keyMap.MatchesAction(msg, actions.ActionRight, actions.ScopeGlobal) {
					m.cycleFocusForward()
					return m, nil
				}
				if m.keyMap.MatchesAction(msg, actions.ActionLeft, actions.ScopeGlobal) {
					m.cycleFocusBackward()
					return m, nil
				}
			}
		}

		// Handle focus switching with hardcoded keys (fallback)
		switch msg.String() {
		case "esc":
			// Escape key: exit thread if active, otherwise return to sidebar
			if m.threadActive {
				m.exitThread()
				return m, nil
			} else if m.focused != FocusSidebar {
				m.setFocus(FocusSidebar)
				return m, nil
			}
		case "tab":
			// Cycle focus forward: sidebar -> messages -> input -> sidebar
			m.cycleFocusForward()
			return m, nil
		case "shift+tab":
			// Cycle focus backward: sidebar -> input -> messages -> sidebar
			m.cycleFocusBackward()
			return m, nil
		case "enter":
			// If on sidebar, select channel (keep focus on sidebar)
			if m.focused == FocusSidebar {
				selectedCh := m.sidebar.GetSelectedChannel()
				if selectedCh != nil {
					m.selectedChannel = selectedCh
					m.messages.SetChannel(selectedCh.ID, selectedCh.GetDisplayName())
					// Keep focus on sidebar so users can cycle through channels
					// Emit channel selection message to trigger loading
					return m, func() tea.Msg {
						return ChannelSelectedMsg{ChannelID: selectedCh.ID}
					}
				}
				return m, nil
			}
		}

	case input.SendMessageMsg:
		// Handle message sending - bubble up to parent with channel info
		m.input.Reset()
		if m.threadActive {
			// Send as thread reply
			return m, func() tea.Msg {
				return SendThreadMessageMsg{
					ChannelID: m.thread.GetChannelID(),
					ThreadTS:  m.thread.GetThreadTS(),
					Text:      msg.Text,
				}
			}
		} else if m.selectedChannel != nil {
			// Send as regular message
			return m, func() tea.Msg {
				return SendChatMessageMsg{
					ChannelID: m.selectedChannel.ID,
					Text:      msg.Text,
				}
			}
		}
		return m, nil

	case ChannelSelectedMsg:
		// Load messages for the newly selected channel
		return m, func() tea.Msg {
			return LoadChannelMessagesMsg(msg)
		}

	case messages.ThreadOpenRequestMsg:
		// User wants to open a thread
		m.enterThread()
		return m, func() tea.Msg {
			return LoadThreadRepliesMsg{
				ChannelID:     msg.ChannelID,
				ThreadTS:      msg.ThreadTS,
				ParentMessage: msg.ParentMessage,
			}
		}
	}

	// Update components based on focus
	var cmd tea.Cmd

	// Always update all components to handle window resize, etc.
	m.sidebar, cmd = m.sidebar.Update(msg)
	cmds = append(cmds, cmd)

	m.messages, cmd = m.messages.Update(msg)
	cmds = append(cmds, cmd)

	m.thread, cmd = m.thread.Update(msg)
	cmds = append(cmds, cmd)

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// SetSize sets the dimensions of the chat view
func (m *ChatModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate dimensions
	sidebarWidth := m.getSidebarWidth()

	// Content box outer width (including borders)
	boxOuterWidth := width - sidebarWidth

	// Content area: box width minus borders (2 for left/right)
	contentWidth := boxOuterWidth - 2

	// Input: fixed height of 6 lines content (to fill available space)
	inputHeight := 6
	// Input box total height: 6 (content) + 2 (borders) = 8
	inputBoxTotalHeight := 8

	// Messages box total height: remaining height after input box
	messagesBoxTotalHeight := height - inputBoxTotalHeight
	// Messages content height: total - borders
	messagesHeight := messagesBoxTotalHeight - 2

	// Set component sizes
	// Sidebar height accounts for borders (inner height = height - 2, total = height)
	m.sidebar.SetSize(sidebarWidth, height-2)
	m.messages.SetSize(contentWidth, messagesHeight)
	m.thread.SetSize(contentWidth, messagesHeight)
	m.input.SetSize(contentWidth, inputHeight)
}

// View renders the chat view
func (m ChatModel) View() string {
	// Render sidebar
	sidebarView := m.sidebar.View()

	// Measure actual rendered sidebar width
	actualSidebarWidth := lipgloss.Width(sidebarView)

	// Calculate box dimensions using actual sidebar width
	// The remaining width after sidebar is the total space for the content boxes (including borders)
	remainingWidth := m.width - actualSidebarWidth
	// contentWidth is the inner width (Box.Width sets inner content width in lipgloss)
	contentWidth := remainingWidth - 2

	// Input box total height: 6 (content) + 2 (borders) = 8
	inputBoxTotalHeight := 8
	// Messages box total height: remaining height after input box
	messagesBoxTotalHeight := m.height - inputBoxTotalHeight

	// Render messages or thread view
	var messagesView string
	if m.threadActive {
		// Show thread view when thread is active
		messagesView = m.thread.View()
	} else {
		// Show regular messages view
		messagesView = m.messages.View()
		if m.selectedChannel == nil {
			// Show empty state when no channel is selected
			// Height should match messages content: messagesBoxTotalHeight - 2 (for borders)
			emptyStateHeight := messagesBoxTotalHeight - 2
			emptyState := lipgloss.NewStyle().
				Width(contentWidth).
				Height(emptyStateHeight).
				Align(lipgloss.Center, lipgloss.Center).
				Render(styles.Dim.Render(m.localize("chat.no_channel_selected", "Select a channel to start chatting")))
			messagesView = emptyState
		}
	}

	// Wrap messages/thread in a bordered box with focus-aware styling
	// Height() sets INNER content size, borders are added on top
	// Inner height: messagesBoxTotalHeight - 2 = (height - 8) - 2 = height - 10
	messagesBoxStyle := styles.Box.
		Width(contentWidth).
		Height(messagesBoxTotalHeight - 2) // Inner content height (borders added by lipgloss)

	if m.focused == FocusMessages || m.focused == FocusThread {
		messagesBoxStyle = messagesBoxStyle.BorderForeground(styles.ColourSuccess)
	}

	messagesBoxView := messagesBoxStyle.Render(messagesView)

	// Wrap input view with fixed height to prevent flickering
	inputContentHeight := inputBoxTotalHeight - 2 // 8 - 2 = 6 lines
	inputView := lipgloss.NewStyle().
		Height(inputContentHeight).
		Render(m.input.View())

	// Wrap input in a bordered box with focus-aware styling
	// Height() sets INNER content size, borders are added on top
	// Inner height: 6 lines (total will be 6 + 2 borders = 8)
	inputBoxStyle := styles.Box.
		Width(contentWidth).
		Height(inputContentHeight) // Inner content height (borders added by lipgloss)

	if m.focused == FocusInput {
		inputBoxStyle = inputBoxStyle.BorderForeground(styles.ColourSuccess)
	}

	inputBoxView := inputBoxStyle.Render(inputView)

	// Stack the two boxes vertically - should now equal m.height total
	contentView := lipgloss.JoinVertical(
		lipgloss.Left,
		messagesBoxView,
		inputBoxView,
	)

	// Combine sidebar and content horizontally
	fullView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebarView,
		contentView,
	)

	return fullView
}

// getSidebarWidth returns the width of the sidebar
func (m ChatModel) getSidebarWidth() int {
	sidebarWidth := m.width / 4
	if sidebarWidth < 20 {
		sidebarWidth = 20
	}
	if sidebarWidth > 40 {
		sidebarWidth = 40
	}
	return sidebarWidth
}

// cycleFocusForward cycles focus to the next component
func (m *ChatModel) cycleFocusForward() {
	switch m.focused {
	case FocusSidebar:
		m.setFocus(FocusMessages)
	case FocusMessages:
		m.setFocus(FocusInput)
	case FocusInput:
		m.setFocus(FocusSidebar)
	default:
		m.setFocus(FocusSidebar)
	}
}

// cycleFocusBackward cycles focus to the previous component
func (m *ChatModel) cycleFocusBackward() {
	switch m.focused {
	case FocusSidebar:
		m.setFocus(FocusInput)
	case FocusInput:
		m.setFocus(FocusMessages)
	case FocusMessages:
		m.setFocus(FocusSidebar)
	default:
		m.setFocus(FocusSidebar)
	}
}

// setFocus sets the focused component
func (m *ChatModel) setFocus(component FocusedComponent) {
	m.focused = component

	// Update component focus states
	m.sidebar.SetFocused(component == FocusSidebar)
	m.messages.SetFocused(component == FocusMessages)
	m.thread.SetFocused(component == FocusThread)
	m.input.SetFocused(component == FocusInput)
}

// enterThread enters thread view mode
func (m *ChatModel) enterThread() {
	m.threadActive = true
	m.messages.EnableSelection()
	m.setFocus(FocusThread)
}

// exitThread exits thread view mode and returns to channel view
func (m *ChatModel) exitThread() {
	m.threadActive = false
	m.messages.DisableSelection()
	m.thread.Clear()
	m.setFocus(FocusMessages)
}

// SetChannels sets the list of channels in the sidebar
func (m *ChatModel) SetChannels(channels []models.Channel) {
	m.sidebar.SetChannels(channels)
}

// SetMessages sets the messages in the viewport
func (m *ChatModel) SetMessages(msgs []models.Message) {
	m.messages.SetMessages(msgs)
}

// AddMessage adds a new message to the viewport
func (m *ChatModel) AddMessage(msg models.Message) {
	m.messages.AddMessage(msg)
}

// SetThreadReplies sets the thread replies in the thread view
func (m *ChatModel) SetThreadReplies(channelID, channelName, threadTS string, parentMessage *models.Message, messages []models.Message) {
	m.thread.SetThread(channelID, channelName, threadTS, parentMessage, messages)
}

// AddThreadReply adds a new reply to the thread
func (m *ChatModel) AddThreadReply(msg models.Message) {
	m.thread.AddMessage(msg)
}

// GetSelectedChannel returns the currently selected channel
func (m ChatModel) GetSelectedChannel() *models.Channel {
	return m.selectedChannel
}

// GetChannels returns all channels from the sidebar
func (m ChatModel) GetChannels() []models.Channel {
	return m.sidebar.GetChannels()
}

// SelectChannel selects a channel by ID in the sidebar
func (m *ChatModel) SelectChannel(channelID string) {
	m.sidebar.SelectChannel(channelID)
	// Also update the selected channel
	selectedCh := m.sidebar.GetSelectedChannel()
	if selectedCh != nil {
		m.selectedChannel = selectedCh
		m.messages.SetChannel(selectedCh.ID, selectedCh.GetDisplayName())
	}
}

// SetKeyMap sets the keybinding map for the chat view
func (m *ChatModel) SetKeyMap(keyMap *keys.ScopedKeyMap) {
	m.keyMap = keyMap
}

// UpdateChannelUnread updates the unread status of a channel in the sidebar
func (m *ChatModel) UpdateChannelUnread(channelID string, unreadCount int, hasUnread bool) {
	// Get current channels from sidebar
	channels := m.sidebar.GetChannels()

	// Update the matching channel
	for i := range channels {
		if channels[i].ID == channelID {
			channels[i].UpdateUnreadStatus(unreadCount)
			break
		}
	}

	// Update sidebar with modified channels
	m.sidebar.SetChannels(channels)
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
