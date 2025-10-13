package input

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// Model represents the message input component
type Model struct {
	textarea textarea.Model
	width    int
	height   int
	focused  bool

	// i18n
	localizer *i18n.Localizer
}

// SendMessageMsg is sent when the user wants to send a message
type SendMessageMsg struct {
	Text string
}

// NewModel creates a new input model
func NewModel() Model {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	ta := textarea.New()
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.CharLimit = 4000 // Slack message limit

	// Style the textarea (no border since we have an outer box)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.Prompt = styles.Label.Bold(true)
	ta.BlurredStyle.Prompt = styles.Dim
	ta.Prompt = "> "

	m := Model{
		textarea:  ta,
		width:     80,
		height:    5,
		focused:   false,
		localizer: localizer,
	}

	// Set localized placeholder
	m.textarea.Placeholder = m.localize("chat.message_placeholder", "Type a message...")

	return m
}

// Init initializes the input component
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the input
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Only handle keys if focused
		if !m.focused {
			return m, nil
		}

		switch msg.String() {
		case "enter":
			// Send message on Enter (without modifier)
			// Textarea handles multiline with Alt+Enter
			text := strings.TrimSpace(m.textarea.Value())
			if text != "" {
				// Clear the input
				m.textarea.Reset()
				// Return SendMessageMsg
				return m, func() tea.Msg {
					return SendMessageMsg{Text: text}
				}
			}
			return m, nil
		case "ctrl+n":
			// Ctrl+N for new line (alternative to Shift+Enter)
			m.textarea.InsertRune('\n')
			return m, nil
		}
	}

	// Let textarea handle the input
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// View renders the input component
func (m Model) View() string {
	return m.textarea.View()
}

// SetSize sets the dimensions of the input
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Set textarea width (account for borders and padding)
	textareaWidth := width - 4
	if textareaWidth < 10 {
		textareaWidth = 10
	}
	m.textarea.SetWidth(textareaWidth)

	// Height should be fixed for input
	inputHeight := 3
	if height > 5 {
		inputHeight = 4
	}
	m.textarea.SetHeight(inputHeight)
}

// SetFocused sets the focus state of the input
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
	if focused {
		m.textarea.Focus()
	} else {
		m.textarea.Blur()
	}
}

// IsFocused returns whether the input is focused
func (m Model) IsFocused() bool {
	return m.focused
}

// SetPlaceholder sets the placeholder text
func (m *Model) SetPlaceholder(placeholder string) {
	m.textarea.Placeholder = placeholder
}

// GetValue returns the current input value
func (m Model) GetValue() string {
	return m.textarea.Value()
}

// SetValue sets the input value
func (m *Model) SetValue(value string) {
	m.textarea.SetValue(value)
}

// Reset clears the input
func (m *Model) Reset() {
	m.textarea.Reset()
}

// IsEmpty returns whether the input is empty
func (m Model) IsEmpty() bool {
	return strings.TrimSpace(m.textarea.Value()) == ""
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
