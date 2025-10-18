package messages

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/models"
	"github.com/jcserv/slacky/internal/tui/components/messageview"
	"github.com/jcserv/slacky/internal/tui/styles"
	"github.com/muesli/reflow/wordwrap"
)

// Model represents the message viewport component
type Model struct {
	view        messageview.Model
	channelName string
	channelID   string
	localizer   *i18n.Localizer
}

// NewModel creates a new messages model
func NewModel() Model {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	config := messageview.Config{
		HeaderFormatter:  headerFormatter,
		MessageFormatter: messageFormatter(localizer),
		EmptyStateText:   localize(localizer, "chat.no_messages", "No messages yet. Start the conversation!"),
	}

	return Model{
		view:      messageview.New(config),
		localizer: localizer,
	}
}

// Init initializes the messages component
func (m Model) Init() tea.Cmd {
	return m.view.Init()
}

// ThreadOpenRequestMsg is sent when user wants to open a thread
type ThreadOpenRequestMsg struct {
	ChannelID     string
	ThreadTS      string
	ParentMessage models.Message
}

// Update handles messages for the viewport
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.view, cmd = m.view.Update(msg)
	return m, cmd
}

// View renders the messages viewport
func (m Model) View() string {
	// Render the channel header
	header := m.renderHeader()

	// Render the viewport
	viewportContent := m.view.View()

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

	config := m.view.GetConfig()
	width := 80 // Default, will be overridden by SetSize
	if config.ChannelName != "" {
		// Get actual width from viewport if available
		width = 80 // This will be set properly via SetSize
	}

	return lipgloss.NewStyle().
		Width(width-2).
		Padding(0, 1).
		Render(channelDisplay)
}

// SetSize sets the dimensions of the messages viewport
func (m *Model) SetSize(width, height int) {
	// Account for header (2 lines: title + padding)
	viewportHeight := height - 2
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	m.view.SetSize(width-2, viewportHeight)
}

// SetMessages sets the list of messages to display
func (m *Model) SetMessages(messages []models.Message) {
	m.view.SetMessages(messages)
}

// AddMessage adds a new message to the viewport
func (m *Model) AddMessage(msg models.Message) {
	m.view.AddMessage(msg)
}

// SetChannel sets the current channel information
func (m *Model) SetChannel(channelID, channelName string) {
	m.channelID = channelID
	m.channelName = channelName

	// Update config with new channel info
	config := m.view.GetConfig()
	config.ChannelID = channelID
	config.ChannelName = channelName
	m.view.UpdateConfig(config)
}

// SetCurrentUserID sets the current user ID for checking reactions
func (m *Model) SetCurrentUserID(userID string) {
	m.view.SetCurrentUserID(userID)
}

// SetFocused sets the focus state of the viewport
func (m *Model) SetFocused(focused bool) {
	m.view.SetFocused(focused)
}

// IsFocused returns whether the viewport is focused
func (m Model) IsFocused() bool {
	return m.view.IsFocused()
}

// ScrollUp scrolls the viewport up
func (m *Model) ScrollUp(lines int) {
	// Note: This functionality would need to be added to messageview if needed
}

// ScrollDown scrolls the viewport down
func (m *Model) ScrollDown(lines int) {
	// Note: This functionality would need to be added to messageview if needed
}

// PageUp scrolls up by half a page
func (m *Model) PageUp() {
	// Note: This functionality would need to be added to messageview if needed
}

// PageDown scrolls down by half a page
func (m *Model) PageDown() {
	// Note: This functionality would need to be added to messageview if needed
}

// GotoTop scrolls to the top
func (m *Model) GotoTop() {
	// Note: This functionality would need to be added to messageview if needed
}

// GotoBottom scrolls to the bottom
func (m *Model) GotoBottom() {
	// Note: This functionality would need to be added to messageview if needed
}

// EnableSelection enables message selection mode
func (m *Model) EnableSelection() {
	m.view.EnableSelection()
}

// GetSelectedMessage returns the currently selected message, or nil if none selected
func (m Model) GetSelectedMessage() *models.Message {
	return m.view.GetSelectedMessage()
}

// IsSelectionEnabled returns whether message selection is currently enabled
func (m Model) IsSelectionEnabled() bool {
	return m.view.IsSelectionEnabled()
}

// DisableSelection disables message selection mode
func (m *Model) DisableSelection() {
	m.view.DisableSelection()
}

// EnableReactionMode enables reaction navigation mode for the selected message
func (m *Model) EnableReactionMode() {
	m.view.EnableReactionMode()
}

// DisableReactionMode disables reaction navigation mode
func (m *Model) DisableReactionMode() {
	m.view.DisableReactionMode()
}

// IsReactionMode returns whether reaction navigation is currently active
func (m Model) IsReactionMode() bool {
	return m.view.IsReactionMode()
}

// GetSelectedReaction returns the currently selected reaction, or nil if none selected
func (m Model) GetSelectedReaction() *models.Reaction {
	return m.view.GetSelectedReaction()
}

// NavigateReactionLeft moves the reaction cursor to the left
func (m *Model) NavigateReactionLeft() {
	m.view.NavigateReactionLeft()
}

// NavigateReactionRight moves the reaction cursor to the right
func (m *Model) NavigateReactionRight() {
	m.view.NavigateReactionRight()
}

// localize is a helper function to localize a message by ID with an optional fallback
func (m Model) localize(messageID string, fallback string) string {
	return localize(m.localizer, messageID, fallback)
}

// localize is a package-level helper function
func localize(localizer *i18n.Localizer, messageID string, fallback string) string {
	cfg := &i18n.LocalizeConfig{
		MessageID: messageID,
	}
	msg, err := localizer.Localize(cfg)
	if err != nil && fallback != "" {
		return fallback
	}
	return msg
}

// headerFormatter formats the header for the messages view
func headerFormatter(channelName string, isThread bool, threadTS string) string {
	if channelName == "" {
		return ""
	}
	return channelName
}

// wrapMessageText wraps message text intelligently, preserving URLs on single lines
func wrapMessageText(text string, width int) string {
	if text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		// Check if this line contains a URL (starts with http:// or https://)
		// or is a link line (starts with spaces and 🔗)
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "http://") ||
			strings.HasPrefix(trimmedLine, "https://") ||
			strings.HasPrefix(trimmedLine, "🔗") {
			// Don't wrap URL lines
			wrappedLines = append(wrappedLines, line)
		} else {
			// Wrap regular text lines
			wrapped := wordwrap.String(line, width)
			wrappedLines = append(wrappedLines, wrapped)
		}
	}

	return strings.Join(wrappedLines, "\n")
}

// messageFormatter returns a function that formats messages for display
func messageFormatter(localizer *i18n.Localizer) messageview.MessageFormatter {
	return func(msg models.Message, width int, index int, state messageview.RenderState) string {
		// Format: [HH:MM] username: message text
		timeStr := styles.Dim.Render(fmt.Sprintf("[%s]", msg.FormatTime()))
		username := styles.Label.Bold(true).Render(msg.UserName)

		// Handle empty username (system messages, etc.)
		if msg.UserName == "" {
			username = styles.Dim.Render(localize(localizer, "chat.unknown_user", "Unknown"))
		}

		// Wrap the message text
		messageText := msg.GetDisplayText()
		if messageText == "" {
			messageText = styles.Dim.Italic(true).Render(fmt.Sprintf("(%s)", localize(localizer, "chat.no_content", "no content")))
		}

		// Calculate width for wrapping (total - timestamp - username - separators)
		// Format: "[12:34] username: " = ~20 chars typically
		wrapWidth := width - 20
		if wrapWidth < 20 {
			wrapWidth = 20
		}

		wrappedText := wrapMessageText(messageText, wrapWidth)

		// For thread replies, add indent
		indent := ""
		if msg.IsThreadReply() {
			indent = "  ↳ "
		}

		// Add cursor/selection indicator if selection is enabled
		prefix := ""
		if state.SelectionEnabled && index == state.Cursor {
			prefix = styles.Success.Render("▸ ")
		} else if state.SelectionEnabled {
			prefix = "  "
		}

		// First line with timestamp and username
		firstLine := fmt.Sprintf("%s%s%s %s: %s", prefix, indent, timeStr, username, wrappedText)

		// Show edited indicator
		if msg.IsEdited {
			firstLine += styles.Dim.Render(fmt.Sprintf(" (%s)", localize(localizer, "chat.edited", "edited")))
		}

		// Add reactions if any (show before thread indicator)
		if len(msg.Reactions) > 0 {
			reactionBubbles := make([]string, 0, len(msg.Reactions))
			for i, r := range msg.Reactions {
				// Convert emoji shortcode to Unicode emoji
				emoji := ConvertEmoji(r.Name)

				// Format: [emoji count]
				bubble := fmt.Sprintf("[%s %d]", emoji, r.Count)

				// Check if current user has reacted with this emoji
				userHasReacted := false
				if state.CurrentUserID != "" {
					for _, uid := range r.Users {
						if uid == state.CurrentUserID {
							userHasReacted = true
							break
						}
					}
				}

				// Apply style based on selection and user reaction status
				if state.ReactionMode && state.Cursor == index && state.ReactionCursor == i {
					// Selected reaction (has background)
					if userHasReacted {
						// Selected + user reacted: background + success color
						bubble = styles.ReactionBubbleSelected.Bold(true).Foreground(styles.ColourSuccess).Render(bubble)
					} else {
						// Selected only: background
						bubble = styles.ReactionBubbleSelected.Render(bubble)
					}
				} else if userHasReacted {
					// User reacted but not selected: success color + bold
					bubble = styles.ReactionBubbleUserReacted.Render(bubble)
				} else {
					// Default: dim
					bubble = styles.ReactionBubble.Render(bubble)
				}

				reactionBubbles = append(reactionBubbles, bubble)
			}
			reactionLine := prefix + "    " + strings.Join(reactionBubbles, " ")
			firstLine += "\n" + reactionLine
		}

		// Show thread indicator if message has thread (after reactions)
		if msg.HasThread() {
			// Format: "💬 X replies"
			threadIndicator := fmt.Sprintf("💬 %d replies", msg.ReplyCount)
			if msg.ReplyCount == 1 {
				threadIndicator = "💬 1 reply"
			}
			firstLine += "\n" + prefix + "  " + styles.Success.Render(threadIndicator)
		}

		return firstLine
	}
}
