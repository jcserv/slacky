package messageview

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/models"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// MessageFormatter is a function that formats a message for display
// Parameters: message, width, index, isParent (for thread), cursor position, reactionCursor, selectionEnabled, reactionMode, currentUserID
type MessageFormatter func(msg models.Message, width int, index int, state RenderState) string

// HeaderFormatter is a function that formats the header for display
// Parameters: channelName, isThread, threadTS
type HeaderFormatter func(channelName string, isThread bool, threadTS string) string

// RenderState contains the state needed for rendering messages
type RenderState struct {
	Cursor           int
	ReactionCursor   int
	SelectionEnabled bool
	ReactionMode     bool
	CurrentUserID    string
	IsParent         bool // For thread view: is this the parent message?
}

// Config holds configuration for the message view
type Config struct {
	// Required formatters
	HeaderFormatter  HeaderFormatter
	MessageFormatter MessageFormatter

	// Display settings
	EmptyStateText string

	// Context identifiers
	ChannelID   string
	ChannelName string

	// Thread-specific (optional)
	ThreadTS      string
	ParentMessage *models.Message
}

// Model represents a shared message viewport component
type Model struct {
	viewport viewport.Model
	messages []models.Message
	width    int
	height   int
	focused  bool

	// Message selection
	cursor           int
	selectionEnabled bool

	// Reaction navigation
	reactionMode   bool
	reactionCursor int

	// Current user ID (for checking if user has reacted)
	currentUserID string

	// Configuration
	config Config

	// i18n
	localizer *i18n.Localizer
}

// New creates a new message view model
func New(config Config) Model {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	vp := viewport.New(80, 20)
	if config.EmptyStateText != "" {
		vp.SetContent(config.EmptyStateText)
	}

	return Model{
		viewport:  vp,
		messages:  []models.Message{},
		width:     80,
		height:    20,
		config:    config,
		localizer: localizer,
	}
}

// Init initializes the message view component
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the viewport
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Only handle keys if focused
		if !m.focused {
			return m, nil
		}

		// Handle reaction navigation when in reaction mode
		if m.reactionMode {
			switch msg.String() {
			case "left":
				m.NavigateReactionLeft()
				return m, nil
			case "right":
				m.NavigateReactionRight()
				return m, nil
			case "esc":
				m.DisableReactionMode()
				return m, nil
			case " ", "space", "enter":
				// Return a message to toggle the selected reaction
				if reaction := m.GetSelectedReaction(); reaction != nil {
					selectedMsg := m.GetSelectedMessage()
					if selectedMsg != nil {
						return m, func() tea.Msg {
							return ReactionToggleRequestMsg{
								ChannelID: m.config.ChannelID,
								Timestamp: selectedMsg.ID,
								EmojiName: reaction.Name,
							}
						}
					}
				}
				return m, nil
			}
		}

		// Handle navigation when selection is enabled
		if m.selectionEnabled {
			switch msg.String() {
			case "up":
				if m.cursor > 0 {
					m.cursor--
					m.DisableReactionMode() // Exit reaction mode when moving messages
					m.renderMessages()
				}
				return m, nil
			case "down":
				if m.cursor < len(m.messages)-1 {
					m.cursor++
					m.DisableReactionMode() // Exit reaction mode when moving messages
					m.renderMessages()
				}
				return m, nil
			case "left", "right":
				// Enter reaction mode when pressing left/right on a message
				m.EnableReactionMode()
				return m, nil
			}
		}

		// Let viewport handle scrolling
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the message viewport
func (m Model) View() string {
	return m.viewport.View()
}

// GetHeader renders the header using the configured formatter
func (m Model) GetHeader() string {
	if m.config.HeaderFormatter == nil {
		return ""
	}
	isThread := m.config.ThreadTS != ""
	return m.config.HeaderFormatter(m.config.ChannelName, isThread, m.config.ThreadTS)
}

// SetSize sets the dimensions of the message viewport
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height
}

// SetMessages sets the list of messages to display
func (m *Model) SetMessages(messages []models.Message) {
	m.messages = messages

	// If focused and we now have messages, ensure selection is enabled
	if m.focused && len(messages) > 0 && !m.selectionEnabled {
		m.EnableSelection()
	}

	m.renderMessages()
}

// AddMessage adds a new message to the viewport
func (m *Model) AddMessage(msg models.Message) {
	m.messages = append(m.messages, msg)

	// Store current scroll position before re-rendering
	currentYOffset := m.viewport.YOffset

	m.renderMessages()

	// Restore scroll position
	m.viewport.SetYOffset(currentYOffset)
}

// renderMessages renders all messages to the viewport content
func (m *Model) renderMessages() {
	if len(m.messages) == 0 {
		emptyText := m.config.EmptyStateText
		if emptyText == "" {
			emptyText = m.localize("chat.no_messages", "No messages")
		}
		m.viewport.SetContent(styles.Dim.Render(emptyText))
		return
	}

	if m.config.MessageFormatter == nil {
		m.viewport.SetContent("No message formatter configured")
		return
	}

	var lines []string
	contentWidth := m.viewport.Width

	for i, msg := range m.messages {
		state := RenderState{
			Cursor:           m.cursor,
			ReactionCursor:   m.reactionCursor,
			SelectionEnabled: m.selectionEnabled,
			ReactionMode:     m.reactionMode,
			CurrentUserID:    m.currentUserID,
			IsParent:         i == 0 && m.config.ThreadTS != "", // First message in thread is parent
		}
		formatted := m.config.MessageFormatter(msg, contentWidth, i, state)
		lines = append(lines, formatted)
	}

	content := ""
	for i, line := range lines {
		if i > 0 {
			content += "\n"
		}
		content += line
	}
	m.viewport.SetContent(content)
}

// SetCurrentUserID sets the current user ID for checking reactions
func (m *Model) SetCurrentUserID(userID string) {
	m.currentUserID = userID
}

// SetFocused sets the focus state of the viewport
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
	// Enable selection when focused, disable when not focused
	if focused {
		// Always enable selection when focused, even if no messages yet
		// EnableSelection will handle the empty case gracefully
		m.EnableSelection()
	} else {
		m.DisableSelection()
	}
}

// IsFocused returns whether the viewport is focused
func (m Model) IsFocused() bool {
	return m.focused
}

// EnableSelection enables message selection mode
func (m *Model) EnableSelection() {
	m.selectionEnabled = true
	m.cursor = 0
	if m.cursor >= len(m.messages) {
		m.cursor = len(m.messages) - 1
	}
	m.renderMessages()
}

// DisableSelection disables message selection mode
func (m *Model) DisableSelection() {
	m.selectionEnabled = false
	m.reactionMode = false
	m.cursor = 0
	m.reactionCursor = 0
	m.renderMessages()
}

// GetSelectedMessage returns the currently selected message, or nil if none selected
func (m Model) GetSelectedMessage() *models.Message {
	if !m.selectionEnabled || m.cursor < 0 || m.cursor >= len(m.messages) {
		return nil
	}
	return &m.messages[m.cursor]
}

// IsSelectionEnabled returns whether message selection is currently enabled
func (m Model) IsSelectionEnabled() bool {
	return m.selectionEnabled
}

// EnableReactionMode enables reaction navigation mode for the selected message
func (m *Model) EnableReactionMode() {
	if !m.selectionEnabled || m.cursor < 0 || m.cursor >= len(m.messages) {
		return
	}

	msg := m.messages[m.cursor]
	if len(msg.Reactions) == 0 {
		return // No reactions to navigate
	}

	m.reactionMode = true
	m.reactionCursor = 0
	m.renderMessages()
}

// DisableReactionMode disables reaction navigation mode
func (m *Model) DisableReactionMode() {
	m.reactionMode = false
	m.reactionCursor = 0
	m.renderMessages()
}

// IsReactionMode returns whether reaction navigation is currently active
func (m Model) IsReactionMode() bool {
	return m.reactionMode
}

// GetSelectedReaction returns the currently selected reaction, or nil if none selected
func (m Model) GetSelectedReaction() *models.Reaction {
	if !m.reactionMode || m.cursor < 0 || m.cursor >= len(m.messages) {
		return nil
	}

	msg := m.messages[m.cursor]
	if m.reactionCursor < 0 || m.reactionCursor >= len(msg.Reactions) {
		return nil
	}

	return &msg.Reactions[m.reactionCursor]
}

// NavigateReactionLeft moves the reaction cursor to the left
func (m *Model) NavigateReactionLeft() {
	if !m.reactionMode {
		return
	}

	if m.reactionCursor > 0 {
		m.reactionCursor--
		m.renderMessages()
	}
}

// NavigateReactionRight moves the reaction cursor to the right
func (m *Model) NavigateReactionRight() {
	if !m.reactionMode {
		return
	}

	if m.cursor >= 0 && m.cursor < len(m.messages) {
		msg := m.messages[m.cursor]
		if m.reactionCursor < len(msg.Reactions)-1 {
			m.reactionCursor++
			m.renderMessages()
		}
	}
}

// UpdateConfig updates the configuration
func (m *Model) UpdateConfig(config Config) {
	m.config = config
	m.renderMessages()
}

// GetConfig returns the current configuration
func (m Model) GetConfig() Config {
	return m.config
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
