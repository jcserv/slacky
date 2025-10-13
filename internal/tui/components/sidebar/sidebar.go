package sidebar

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jcserv/slacky/internal/models"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// Model represents the sidebar component
type Model struct {
	list     list.Model
	channels []models.Channel
	width    int
	height   int
	focused  bool
}

// channelItem implements list.Item interface for the Bubbles list
type channelItem struct {
	channel models.Channel
}

func (i channelItem) Title() string {
	icon := i.channel.GetIcon()
	name := i.channel.Name

	// Add indicator for unread messages
	indicator := " "
	if i.channel.HasUnread {
		indicator = "●"
	}

	return fmt.Sprintf("%s %s %s", indicator, icon, name)
}

func (i channelItem) Description() string {
	return i.channel.GetDescription()
}

func (i channelItem) FilterValue() string {
	return i.channel.Name
}

// NewModel creates a new sidebar model
func NewModel() Model {
	// Create custom list delegate for styling
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetHeight(2)
	delegate.SetSpacing(0)

	// Style the list items
	delegate.Styles.NormalTitle = styles.Label
	delegate.Styles.NormalDesc = styles.Dim.Width(0) // Will be set dynamically
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(styles.ColourSuccess).
		Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(styles.ColourSuccess).
		Faint(true)

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)

	// Style the list
	l.Styles.Title = styles.Title
	l.Styles.FilterPrompt = styles.Label
	l.Styles.FilterCursor = styles.Label

	return Model{
		list:     l,
		channels: []models.Channel{},
		width:    20,
		height:   24,
		focused:  false,
	}
}

// Init initializes the sidebar
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the sidebar
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
			// Channel selection will be handled by parent
			return m, nil
		}
	}

	// Update the list
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the sidebar
func (m Model) View() string {
	// Create border style
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Width(m.width).
		Height(m.height)

	if m.focused {
		borderStyle = borderStyle.BorderForeground(styles.ColourSuccess)
	}

	// Render list with title
	title := styles.Subtitle.Render("Channels")
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		m.list.View(),
	)

	return borderStyle.Render(content)
}

// SetSize sets the dimensions of the sidebar
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate list dimensions (account for border and title)
	listWidth := width - 4  // 2 for border, 2 for padding
	listHeight := height - 5 // 2 for border, 1 for title, 2 for padding

	if listWidth < 1 {
		listWidth = 1
	}
	if listHeight < 1 {
		listHeight = 1
	}

	m.list.SetSize(listWidth, listHeight)
}

// SetChannels sets the list of channels
func (m *Model) SetChannels(channels []models.Channel) {
	m.channels = channels

	// Convert channels to list items
	items := make([]list.Item, len(channels))
	for i, ch := range channels {
		items[i] = channelItem{channel: ch}
	}

	m.list.SetItems(items)
}

// GetSelectedChannel returns the currently selected channel
func (m Model) GetSelectedChannel() *models.Channel {
	if len(m.channels) == 0 {
		return nil
	}

	selectedIdx := m.list.Index()
	if selectedIdx < 0 || selectedIdx >= len(m.channels) {
		return nil
	}

	return &m.channels[selectedIdx]
}

// SetFocused sets the focus state of the sidebar
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// IsFocused returns whether the sidebar is focused
func (m Model) IsFocused() bool {
	return m.focused
}

// SelectChannel sets the selected channel by ID
func (m *Model) SelectChannel(channelID string) {
	for i, ch := range m.channels {
		if ch.ID == channelID {
			m.list.Select(i)
			return
		}
	}
}

// NextChannel moves selection to the next channel
func (m *Model) NextChannel() {
	if len(m.channels) == 0 {
		return
	}
	idx := m.list.Index()
	if idx < len(m.channels)-1 {
		m.list.Select(idx + 1)
	}
}

// PrevChannel moves selection to the previous channel
func (m *Model) PrevChannel() {
	if len(m.channels) == 0 {
		return
	}
	idx := m.list.Index()
	if idx > 0 {
		m.list.Select(idx - 1)
	}
}
