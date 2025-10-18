package messages

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/models"
	"github.com/jcserv/slacky/internal/tui/styles"
	"github.com/muesli/reflow/wordwrap"
)

// Model represents the message viewport component
type Model struct {
	viewport viewport.Model
	messages []models.Message
	width    int
	height   int
	focused  bool

	// Current channel info
	channelName string
	channelID   string

	// i18n
	localizer *i18n.Localizer
}

// NewModel creates a new messages model
func NewModel() Model {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	vp := viewport.New(80, 20)
	vp.SetContent("No messages to display")

	return Model{
		viewport:  vp,
		messages:  []models.Message{},
		width:     80,
		height:    20,
		localizer: localizer,
	}
}

// Init initializes the messages component
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

		// Let viewport handle scrolling
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the messages viewport
func (m Model) View() string {
	// Render the channel header
	header := m.renderHeader()

	// Render the viewport
	viewportContent := m.viewport.View()

	// Combine header and viewport
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		viewportContent,
	)

	return content
}

// renderHeader renders the channel header
func (m Model) renderHeader() string {
	if m.channelName == "" {
		return styles.Subtitle.Render(m.localize("chat.select_channel", "Select a channel"))
	}

	channelDisplay := styles.Title.Render(m.channelName)

	return lipgloss.NewStyle().
		Width(m.width-2).
		Padding(0, 1).
		Render(channelDisplay)
}

// SetSize sets the dimensions of the messages viewport
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Account for header (2 lines: title + padding)
	viewportHeight := height - 2
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	m.viewport.Width = width - 2 // Account for padding
	m.viewport.Height = viewportHeight
}

// SetMessages sets the list of messages to display
func (m *Model) SetMessages(messages []models.Message) {
	m.messages = messages
	m.renderMessages()
}

// AddMessage adds a new message to the viewport
// Preserves the current scroll position (doesn't auto-scroll)
func (m *Model) AddMessage(msg models.Message) {
	m.messages = append(m.messages, msg)

	// Store current scroll position before re-rendering
	currentYOffset := m.viewport.YOffset

	m.renderMessages()

	// Restore scroll position (preserve where user was viewing)
	m.viewport.SetYOffset(currentYOffset)
}

// renderMessages renders all messages to the viewport content
func (m *Model) renderMessages() {
	if len(m.messages) == 0 {
		m.viewport.SetContent(styles.Dim.Render(m.localize("chat.no_messages", "No messages yet. Start the conversation!")))
		return
	}

	var lines []string
	contentWidth := m.viewport.Width

	for _, msg := range m.messages {
		lines = append(lines, m.formatMessage(msg, contentWidth))
	}

	content := strings.Join(lines, "\n")
	m.viewport.SetContent(content)
}

// formatMessage formats a single message for display
func (m *Model) formatMessage(msg models.Message, width int) string {
	// Format: [HH:MM] username: message text
	timeStr := styles.Dim.Render(fmt.Sprintf("[%s]", msg.FormatTime()))
	username := styles.Label.Bold(true).Render(msg.UserName)

	// Handle empty username (system messages, etc.)
	if msg.UserName == "" {
		username = styles.Dim.Render(m.localize("chat.unknown_user", "Unknown"))
	}

	// Wrap the message text
	messageText := msg.GetDisplayText()
	if messageText == "" {
		messageText = styles.Dim.Italic(true).Render(fmt.Sprintf("(%s)", m.localize("chat.no_content", "no content")))
	}

	// Calculate width for wrapping (total - timestamp - username - separators)
	// Format: "[12:34] username: " = ~20 chars typically
	wrapWidth := width - 20
	if wrapWidth < 20 {
		wrapWidth = 20
	}

	wrappedText := wordwrap.String(messageText, wrapWidth)

	// For thread replies, add indent
	indent := ""
	if msg.IsThreadReply() {
		indent = "  ↳ "
	}

	// First line with timestamp and username
	firstLine := fmt.Sprintf("%s%s %s: %s", indent, timeStr, username, wrappedText)

	// Show edited indicator
	if msg.IsEdited {
		firstLine += styles.Dim.Render(fmt.Sprintf(" (%s)", m.localize("chat.edited", "edited")))
	}

	// Add reactions if any
	if len(msg.Reactions) > 0 {
		reactionStrs := make([]string, 0, len(msg.Reactions))
		for _, r := range msg.Reactions {
			reactionStrs = append(reactionStrs, fmt.Sprintf(":%s: %d", r.Name, r.Count))
		}
		reactionLine := styles.Dim.Render("    " + strings.Join(reactionStrs, " "))
		firstLine += "\n" + reactionLine
	}

	return firstLine
}

// SetChannel sets the current channel information
func (m *Model) SetChannel(channelID, channelName string) {
	m.channelID = channelID
	m.channelName = channelName
}

// SetFocused sets the focus state of the viewport
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// IsFocused returns whether the viewport is focused
func (m Model) IsFocused() bool {
	return m.focused
}

// ScrollUp scrolls the viewport up
func (m *Model) ScrollUp(lines int) {
	m.viewport.ScrollUp(lines)
}

// ScrollDown scrolls the viewport down
func (m *Model) ScrollDown(lines int) {
	m.viewport.ScrollDown(lines)
}

// PageUp scrolls up by half a page
func (m *Model) PageUp() {
	m.viewport.HalfPageUp()
}

// PageDown scrolls down by half a page
func (m *Model) PageDown() {
	m.viewport.HalfPageDown()
}

// GotoTop scrolls to the top
func (m *Model) GotoTop() {
	m.viewport.GotoTop()
}

// GotoBottom scrolls to the bottom
func (m *Model) GotoBottom() {
	m.viewport.GotoBottom()
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
