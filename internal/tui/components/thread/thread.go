package thread

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/models"
	"github.com/jcserv/slacky/internal/tui/components/messages"
	"github.com/jcserv/slacky/internal/tui/components/messageview"
	"github.com/jcserv/slacky/internal/tui/styles"
	"github.com/muesli/reflow/wordwrap"
)

// Model represents the thread viewport component
type Model struct {
	view          messageview.Model
	threadTS      string
	parentMessage *models.Message
	channelID     string
	channelName   string
	localizer     *i18n.Localizer
}

// NewModel creates a new thread model
func NewModel() Model {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	config := messageview.Config{
		HeaderFormatter:  threadHeaderFormatter,
		MessageFormatter: threadMessageFormatter(localizer),
		EmptyStateText:   localize(localizer, "thread.no_replies", "No replies yet"),
	}

	return Model{
		view:      messageview.New(config),
		localizer: localizer,
	}
}

// Init initializes the thread component
func (m Model) Init() tea.Cmd {
	return m.view.Init()
}

// Update handles messages for the viewport
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.view, cmd = m.view.Update(msg)
	return m, cmd
}

// View renders the thread viewport
func (m Model) View() string {
	// Render the thread header
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

// renderHeader renders the thread header
func (m Model) renderHeader() string {
	if m.threadTS == "" {
		return styles.Subtitle.Render(m.localize("thread.no_thread", "No thread selected"))
	}

	// Show thread context: "← channel-name"
	// Note: channelName already includes the # prefix from GetDisplayName()
	threadTitle := fmt.Sprintf("← %s", m.channelName)

	titleDisplay := styles.Title.Render(threadTitle)

	config := m.view.GetConfig()
	width := 80 // Default, will be overridden by SetSize
	if config.ChannelName != "" {
		// Get actual width from viewport if available
		width = 80 // This will be set properly via SetSize
	}

	return lipgloss.NewStyle().
		Width(width-2).
		Padding(0, 1).
		Render(titleDisplay)
}

// SetSize sets the dimensions of the thread viewport
func (m *Model) SetSize(width, height int) {
	// Account for header (2 lines: title + padding)
	viewportHeight := height - 2
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	m.view.SetSize(width-2, viewportHeight)
}

// SetThread sets the thread to display
func (m *Model) SetThread(channelID, channelName, threadTS string, parentMessage *models.Message, msgs []models.Message) {
	m.channelID = channelID
	m.channelName = channelName
	m.threadTS = threadTS
	m.parentMessage = parentMessage

	// Update config with thread info
	config := m.view.GetConfig()
	config.ChannelID = channelID
	config.ChannelName = channelName
	config.ThreadTS = threadTS
	config.ParentMessage = parentMessage
	m.view.UpdateConfig(config)

	// Set messages
	m.view.SetMessages(msgs)
}

// AddMessage adds a new message to the thread
func (m *Model) AddMessage(msg models.Message) {
	m.view.AddMessage(msg)
}

// SetFocused sets the focus state of the viewport
func (m *Model) SetFocused(focused bool) {
	m.view.SetFocused(focused)
}

// IsFocused returns whether the viewport is focused
func (m Model) IsFocused() bool {
	return m.view.IsFocused()
}

// SetCurrentUserID sets the current user ID for checking reactions
func (m *Model) SetCurrentUserID(userID string) {
	m.view.SetCurrentUserID(userID)
}

// Clear clears the thread state
func (m *Model) Clear() {
	m.channelID = ""
	m.channelName = ""
	m.threadTS = ""
	m.parentMessage = nil

	// Update config
	config := m.view.GetConfig()
	config.ChannelID = ""
	config.ChannelName = ""
	config.ThreadTS = ""
	config.ParentMessage = nil
	m.view.UpdateConfig(config)

	// Clear messages
	m.view.SetMessages([]models.Message{})
}

// GetThreadTS returns the current thread timestamp
func (m Model) GetThreadTS() string {
	return m.threadTS
}

// GetChannelID returns the current channel ID
func (m Model) GetChannelID() string {
	return m.channelID
}

// EnableSelection enables message selection mode
func (m *Model) EnableSelection() {
	m.view.EnableSelection()
}

// DisableSelection disables message selection mode
func (m *Model) DisableSelection() {
	m.view.DisableSelection()
}

// GetSelectedMessage returns the currently selected message
func (m Model) GetSelectedMessage() *models.Message {
	return m.view.GetSelectedMessage()
}

// IsSelectionEnabled returns whether selection is enabled
func (m Model) IsSelectionEnabled() bool {
	return m.view.IsSelectionEnabled()
}

// EnableReactionMode enables reaction navigation mode
func (m *Model) EnableReactionMode() {
	m.view.EnableReactionMode()
}

// DisableReactionMode disables reaction navigation mode
func (m *Model) DisableReactionMode() {
	m.view.DisableReactionMode()
}

// IsReactionMode returns whether reaction mode is active
func (m Model) IsReactionMode() bool {
	return m.view.IsReactionMode()
}

// GetSelectedReaction returns the currently selected reaction
func (m Model) GetSelectedReaction() *models.Reaction {
	return m.view.GetSelectedReaction()
}

// NavigateReactionLeft moves the reaction cursor left
func (m *Model) NavigateReactionLeft() {
	m.view.NavigateReactionLeft()
}

// NavigateReactionRight moves the reaction cursor right
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

// threadHeaderFormatter formats the header for thread view
func threadHeaderFormatter(channelName string, isThread bool, threadTS string) string {
	if channelName == "" {
		return ""
	}
	return fmt.Sprintf("← %s", channelName)
}

// threadMessageFormatter returns a function that formats thread messages for display
func threadMessageFormatter(localizer *i18n.Localizer) messageview.MessageFormatter {
	return func(msg models.Message, width int, index int, state messageview.RenderState) string {
		timeStr := styles.Dim.Render(fmt.Sprintf("[%s]", msg.FormatTime()))
		username := styles.Label.Bold(true).Render(msg.UserName)

		if msg.UserName == "" {
			username = styles.Dim.Render(localize(localizer, "chat.unknown_user", "Unknown"))
		}

		messageText := msg.GetDisplayText()
		if messageText == "" {
			messageText = styles.Dim.Italic(true).Render(fmt.Sprintf("(%s)", localize(localizer, "chat.no_content", "no content")))
		}

		wrapWidth := width - 20
		if wrapWidth < 20 {
			wrapWidth = 20
		}

		wrappedText := wordwrap.String(messageText, wrapWidth)

		// Add cursor/selection indicator if selection is enabled
		prefix := ""
		if state.SelectionEnabled && index == state.Cursor {
			prefix = styles.Success.Render("▸ ")
		} else if state.SelectionEnabled {
			prefix = "  "
		}

		// Format: [12:34] username: message
		line := fmt.Sprintf("%s%s %s: %s", prefix, timeStr, username, wrappedText)

		if msg.IsEdited {
			line += styles.Dim.Render(fmt.Sprintf(" (%s)", localize(localizer, "chat.edited", "edited")))
		}

		// Add reactions if any
		if len(msg.Reactions) > 0 {
			reactionBubbles := make([]string, 0, len(msg.Reactions))
			for i, r := range msg.Reactions {
				emoji := messages.ConvertEmoji(r.Name)
				bubble := fmt.Sprintf("[%s %d]", emoji, r.Count)

				// Check if current user has reacted
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
			line += "\n" + reactionLine
		}

		return line
	}
}
