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
	"github.com/jcserv/slacky/internal/tui/keys"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// FocusedComponent represents which component is currently focused
type FocusedComponent int

const (
	FocusSidebar FocusedComponent = iota
	FocusMessages
	FocusInput
)

// SendChatMessageMsg is sent when the user wants to send a message
type SendChatMessageMsg struct {
	ChannelID string
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

// ChatModel represents the chat view
type ChatModel struct {
	sidebar  sidebar.Model
	messages messages.Model
	input    input.Model

	width     int
	height    int
	localizer *i18n.Localizer
	keyMap    *keys.ScopedKeyMap

	// State
	selectedChannel *models.Channel
	focused         FocusedComponent
}

// NewChatModel creates a new chat view model
func NewChatModel() ChatModel {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	return ChatModel{
		sidebar:   sidebar.NewModel(),
		messages:  messages.NewModel(),
		input:     input.NewModel(),
		width:     80,
		height:    24,
		localizer: localizer,
		focused:   FocusSidebar, // Start with sidebar focused
	}
}

// Init initializes the chat view
func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(
		m.sidebar.Init(),
		m.messages.Init(),
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
			// Space key: jump to input
			if m.keyMap.MatchesAction(msg, actions.ActionBeginInput, actions.ScopeChat) && m.focused != FocusInput {
				m.setFocus(FocusInput)
				return m, nil
			}

			// Arrow keys for focus cycling (when not in input mode)
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
			// Escape key: return to sidebar (navigation mode)
			if m.focused != FocusSidebar {
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
		if m.selectedChannel != nil {
			m.input.Reset()
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
	}

	// Update components based on focus
	var cmd tea.Cmd

	// Always update all components to handle window resize, etc.
	m.sidebar, cmd = m.sidebar.Update(msg)
	cmds = append(cmds, cmd)

	m.messages, cmd = m.messages.Update(msg)
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

	// Input: fixed height of 2 lines (minimal)
	inputHeight := 2

	// Messages: remaining height minus input, divider, and box border
	// height - 2 (input) - 1 (divider) - 2 (box borders) = height - 5
	messagesHeight := height - inputHeight - 3

	// Set component sizes
	m.sidebar.SetSize(sidebarWidth, height)
	m.messages.SetSize(contentWidth, messagesHeight)
	m.input.SetSize(contentWidth, inputHeight)
}

// View renders the chat view
func (m ChatModel) View() string {
	// Render sidebar
	sidebarView := m.sidebar.View()

	// Measure actual rendered sidebar width
	actualSidebarWidth := lipgloss.Width(sidebarView)

	// Calculate box dimensions using actual sidebar width
	// The remaining width after sidebar is the total space for the content box (including borders)
	remainingWidth := m.width - actualSidebarWidth
	// contentWidth is the inner width (Box.Width sets inner content width in lipgloss)
	contentWidth := remainingWidth - 2

	// Create divider between messages and input
	dividerLine := ""
	for i := 0; i < contentWidth; i++ {
		dividerLine += "─"
	}
	divider := styles.Border.Render(dividerLine)

	// Render messages view or empty state
	messagesView := m.messages.View()
	if m.selectedChannel == nil {
		// Show empty state when no channel is selected
		// Use same height calculation as messagesHeight: height - inputHeight - 3
		emptyStateHeight := m.height - 2 - 3 // height - inputHeight - (divider + borders)
		emptyState := lipgloss.NewStyle().
			Width(contentWidth).
			Height(emptyStateHeight).
			Align(lipgloss.Center, lipgloss.Center).
			Render(styles.Dim.Render(m.localize("chat.no_channel_selected", "Select a channel to start chatting")))
		messagesView = emptyState
	}

	// Wrap input view with fixed height to prevent flickering
	// Input height is always 2 lines (matching inputHeight in SetSize)
	inputView := lipgloss.NewStyle().
		Height(2).
		Render(m.input.View())

	// Render messages and input (stacked vertically with divider)
	messagesAndInput := lipgloss.JoinVertical(
		lipgloss.Left,
		messagesView,
		divider,
		inputView,
	)

	// Wrap content in a bordered box that matches sidebar height
	// Change border color based on focus
	boxStyle := styles.Box.
		Width(contentWidth). // Inner content width (borders are added by lipgloss)
		Height(m.height)     // Match sidebar height

	// Highlight border when messages or input is focused
	if m.focused == FocusMessages || m.focused == FocusInput {
		boxStyle = boxStyle.BorderForeground(styles.ColourSuccess)
	}

	contentView := boxStyle.Render(messagesAndInput)

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
	m.input.SetFocused(component == FocusInput)
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

// GetSelectedChannel returns the currently selected channel
func (m ChatModel) GetSelectedChannel() *models.Channel {
	return m.selectedChannel
}

// SetKeyMap sets the keybinding map for the chat view
func (m *ChatModel) SetKeyMap(keyMap *keys.ScopedKeyMap) {
	m.keyMap = keyMap
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
