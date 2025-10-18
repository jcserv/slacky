package thread

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

// Model represents the thread viewport component
type Model struct {
	viewport viewport.Model
	messages []models.Message
	width    int
	height   int
	focused  bool

	// Thread context
	channelName   string
	channelID     string
	threadTS      string
	parentMessage *models.Message

	// i18n
	localizer *i18n.Localizer
}

// NewModel creates a new thread model
func NewModel() Model {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	vp := viewport.New(80, 20)
	vp.SetContent("No thread selected")

	return Model{
		viewport:  vp,
		messages:  []models.Message{},
		width:     80,
		height:    20,
		localizer: localizer,
	}
}

// Init initializes the thread component
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

// View renders the thread viewport
func (m Model) View() string {
	// Render the thread header
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

// renderHeader renders the thread header
func (m Model) renderHeader() string {
	if m.threadTS == "" {
		return styles.Subtitle.Render(m.localize("thread.no_thread", "No thread selected"))
	}

	// Show thread context: "← channel-name"
	// Note: channelName already includes the # prefix from GetDisplayName()
	threadTitle := fmt.Sprintf("← %s", m.channelName)

	titleDisplay := styles.Title.Render(threadTitle)

	return lipgloss.NewStyle().
		Width(m.width-2).
		Padding(0, 1).
		Render(titleDisplay)
}

// SetSize sets the dimensions of the thread viewport
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

// SetThread sets the thread to display
func (m *Model) SetThread(channelID, channelName, threadTS string, parentMessage *models.Message, messages []models.Message) {
	m.channelID = channelID
	m.channelName = channelName
	m.threadTS = threadTS
	m.parentMessage = parentMessage
	m.messages = messages
	m.renderMessages()
}

// AddMessage adds a new message to the thread
func (m *Model) AddMessage(msg models.Message) {
	m.messages = append(m.messages, msg)

	// Store current scroll position before re-rendering
	currentYOffset := m.viewport.YOffset

	m.renderMessages()

	// Restore scroll position
	m.viewport.SetYOffset(currentYOffset)
}

// renderMessages renders all thread messages to the viewport content
func (m *Model) renderMessages() {
	if len(m.messages) == 0 {
		m.viewport.SetContent(styles.Dim.Render(m.localize("thread.no_replies", "No replies yet")))
		return
	}

	var lines []string
	contentWidth := m.viewport.Width

	for i, msg := range m.messages {
		if i == 0 {
			// First message is the parent - render differently
			lines = append(lines, m.formatParentMessage(msg, contentWidth))
		} else {
			// Subsequent messages are replies
			lines = append(lines, m.formatReplyMessage(msg, contentWidth))
		}
	}

	content := strings.Join(lines, "\n")
	m.viewport.SetContent(content)
}

// formatParentMessage formats the parent message for display
func (m *Model) formatParentMessage(msg models.Message, width int) string {
	timeStr := styles.Dim.Render(fmt.Sprintf("[%s]", msg.FormatTime()))
	username := styles.Label.Bold(true).Render(msg.UserName)

	if msg.UserName == "" {
		username = styles.Dim.Render(m.localize("chat.unknown_user", "Unknown"))
	}

	messageText := msg.GetDisplayText()
	if messageText == "" {
		messageText = styles.Dim.Italic(true).Render(fmt.Sprintf("(%s)", m.localize("chat.no_content", "no content")))
	}

	wrapWidth := width - 20
	if wrapWidth < 20 {
		wrapWidth = 20
	}

	wrappedText := wordwrap.String(messageText, wrapWidth)

	// Format: [12:34] username: message
	line := fmt.Sprintf("%s %s: %s", timeStr, username, wrappedText)

	if msg.IsEdited {
		line += styles.Dim.Render(fmt.Sprintf(" (%s)", m.localize("chat.edited", "edited")))
	}

	// Add reactions if any
	if len(msg.Reactions) > 0 {
		reactionStrs := make([]string, 0, len(msg.Reactions))
		for _, r := range msg.Reactions {
			reactionStrs = append(reactionStrs, fmt.Sprintf(":%s: %d", r.Name, r.Count))
		}
		reactionLine := styles.Dim.Render("    " + strings.Join(reactionStrs, " "))
		line += "\n" + reactionLine
	}

	return line
}

// formatReplyMessage formats a reply message for display
func (m *Model) formatReplyMessage(msg models.Message, width int) string {
	timeStr := styles.Dim.Render(fmt.Sprintf("[%s]", msg.FormatTime()))
	username := styles.Label.Bold(true).Render(msg.UserName)

	if msg.UserName == "" {
		username = styles.Dim.Render(m.localize("chat.unknown_user", "Unknown"))
	}

	messageText := msg.GetDisplayText()
	if messageText == "" {
		messageText = styles.Dim.Italic(true).Render(fmt.Sprintf("(%s)", m.localize("chat.no_content", "no content")))
	}

	wrapWidth := width - 20
	if wrapWidth < 20 {
		wrapWidth = 20
	}

	wrappedText := wordwrap.String(messageText, wrapWidth)

	// Format: [12:34] username: message
	line := fmt.Sprintf("%s %s: %s", timeStr, username, wrappedText)

	if msg.IsEdited {
		line += styles.Dim.Render(fmt.Sprintf(" (%s)", m.localize("chat.edited", "edited")))
	}

	// Add reactions if any
	if len(msg.Reactions) > 0 {
		reactionStrs := make([]string, 0, len(msg.Reactions))
		for _, r := range msg.Reactions {
			reactionStrs = append(reactionStrs, fmt.Sprintf(":%s: %d", r.Name, r.Count))
		}
		reactionLine := styles.Dim.Render("    " + strings.Join(reactionStrs, " "))
		line += "\n" + reactionLine
	}

	return line
}

// SetFocused sets the focus state of the viewport
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// IsFocused returns whether the viewport is focused
func (m Model) IsFocused() bool {
	return m.focused
}

// Clear clears the thread state
func (m *Model) Clear() {
	m.channelID = ""
	m.channelName = ""
	m.threadTS = ""
	m.parentMessage = nil
	m.messages = []models.Message{}
	m.viewport.SetContent("")
}

// GetThreadTS returns the current thread timestamp
func (m Model) GetThreadTS() string {
	return m.threadTS
}

// GetChannelID returns the current channel ID
func (m Model) GetChannelID() string {
	return m.channelID
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
