package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/models"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// ActivityModel represents the activity view
type ActivityModel struct {
	width      int
	height     int
	localizer  *i18n.Localizer
	list       list.Model
	activities []models.Activity
	filter     models.ActivityFilter
	focused    bool
}

// ActivitySelectedMsg is sent when the user selects an activity to view
type ActivitySelectedMsg struct {
	ChannelID string
	ThreadTS  string
}

// activityItem implements list.Item interface for the Bubbles list
type activityItem struct {
	activity models.Activity
}

func (i activityItem) Title() string {
	icon := i.activity.GetIcon()
	channelDisplay := i.activity.GetChannelDisplay()
	timeDisplay := i.activity.FormatTime()

	// Format: [icon] @user in #channel • 2h ago
	return fmt.Sprintf("%s %s in %s • %s",
		icon,
		i.activity.UserName,
		channelDisplay,
		timeDisplay,
	)
}

func (i activityItem) Description() string {
	// Show message preview or reaction info
	if i.activity.Type == models.ActivityTypeReaction {
		return i.activity.GetReactionDisplay() + " " + i.activity.GetMessagePreview(60)
	}
	return i.activity.GetMessagePreview(80)
}

func (i activityItem) FilterValue() string {
	return i.activity.UserName + " " + i.activity.ChannelName + " " + i.activity.MessageText
}

// NewActivityModel creates a new activity view model
func NewActivityModel() ActivityModel {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	// Create custom list delegate for styling
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetHeight(2)  // Two lines: title + description
	delegate.SetSpacing(0) // No spacing between items

	// Style the list items
	delegate.Styles.NormalTitle = styles.Label
	delegate.Styles.NormalDesc = styles.Dim
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

	// Disable vim-style navigation - users can configure via keybindings if desired
	l.KeyMap.CursorUp.SetKeys("up")     // Only arrow up
	l.KeyMap.CursorDown.SetKeys("down") // Only arrow down
	l.KeyMap.GoToStart.SetKeys("home")  // Only home
	l.KeyMap.GoToEnd.SetKeys("end")     // Only end

	// Style the list
	l.Styles.Title = styles.Title
	l.Styles.FilterPrompt = styles.Label
	l.Styles.FilterCursor = styles.Label
	// Remove default top padding from list
	l.Styles.NoItems = lipgloss.NewStyle().PaddingTop(0)

	return ActivityModel{
		width:      80,
		height:     24,
		localizer:  localizer,
		list:       l,
		activities: []models.Activity{},
		filter:     models.ActivityFilterAll,
		focused:    false,
	}
}

// Init initializes the activity view
func (m ActivityModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the activity view
func (m ActivityModel) Update(msg tea.Msg) (ActivityModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Only handle keys if focused
		if !m.focused {
			return m, nil
		}

		switch msg.String() {
		case "tab":
			// Tab: Cycle to next filter
			m.cycleNextFilter()
			m.refreshList()
			return m, nil
		case "shift+tab":
			// Shift+Tab: Cycle to previous filter
			m.cyclePrevFilter()
			m.refreshList()
			return m, nil
		case "left":
			// Left arrow: Cycle to previous filter (legacy support)
			m.cyclePrevFilter()
			m.refreshList()
			return m, nil
		case "right":
			// Right arrow: Cycle to next filter (legacy support)
			m.cycleNextFilter()
			m.refreshList()
			return m, nil
		case " ", "enter":
			// Space or Enter to select activity
			selected := m.GetSelectedActivity()
			if selected != nil {
				return m, func() tea.Msg {
					return ActivitySelectedMsg{
						ChannelID: selected.ChannelID,
						ThreadTS:  selected.ThreadTS,
					}
				}
			}
			return m, nil
		}
	}

	// Update the list
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// SetSize sets the dimensions of the activity view
func (m *ActivityModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate list dimensions (account for filter line and left/right padding only)
	listWidth := width - 4   // Account for left/right padding (2 chars each side)
	listHeight := height - 3 // Account for filter line (1 line) and borders (2 lines)
	if listWidth < 1 {
		listWidth = 1
	}
	if listHeight < 1 {
		listHeight = 1
	}

	m.list.SetSize(listWidth, listHeight)
}

// SetActivities sets the list of activities
func (m *ActivityModel) SetActivities(activities []models.Activity) {
	m.activities = activities
	m.refreshList()
}

// AppendActivities appends new activities to the existing list
// This is used for live updates from polling
func (m *ActivityModel) AppendActivities(newActivities []models.Activity) {
	if len(newActivities) == 0 {
		return
	}

	// Create a map of existing activities by MessageID to avoid duplicates
	existingMap := make(map[string]bool)
	for _, activity := range m.activities {
		existingMap[activity.MessageID] = true
	}

	// Only append activities that don't already exist
	for _, newActivity := range newActivities {
		if !existingMap[newActivity.MessageID] {
			m.activities = append(m.activities, newActivity)
		}
	}

	// Refresh the list (will re-filter and re-sort)
	m.refreshList()
}

// refreshList updates the list based on current filter
func (m *ActivityModel) refreshList() {
	// Filter and sort activities
	filtered := models.FilterActivities(m.activities, m.filter)
	sorted := models.SortActivitiesByTime(filtered)

	// Convert to list items
	items := make([]list.Item, len(sorted))
	for i, activity := range sorted {
		items[i] = activityItem{activity: activity}
	}

	m.list.SetItems(items)
}

// GetSelectedActivity returns the currently selected activity
func (m ActivityModel) GetSelectedActivity() *models.Activity {
	if len(m.activities) == 0 {
		return nil
	}

	selectedIdx := m.list.Index()
	filtered := models.FilterActivities(m.activities, m.filter)
	sorted := models.SortActivitiesByTime(filtered)

	if selectedIdx < 0 || selectedIdx >= len(sorted) {
		return nil
	}

	return &sorted[selectedIdx]
}

// SetFocused sets the focus state
func (m *ActivityModel) SetFocused(focused bool) {
	m.focused = focused
}

// IsFocused returns whether the view is focused
func (m ActivityModel) IsFocused() bool {
	return m.focused
}

// cycleNextFilter cycles to the next filter (All -> Mentions -> Reactions -> All)
func (m *ActivityModel) cycleNextFilter() {
	switch m.filter {
	case models.ActivityFilterAll:
		m.filter = models.ActivityFilterMentions
	case models.ActivityFilterMentions:
		m.filter = models.ActivityFilterReactions
	case models.ActivityFilterReactions:
		m.filter = models.ActivityFilterAll
	default:
		m.filter = models.ActivityFilterAll
	}
}

// cyclePrevFilter cycles to the previous filter (All -> Reactions -> Mentions -> All)
func (m *ActivityModel) cyclePrevFilter() {
	switch m.filter {
	case models.ActivityFilterAll:
		m.filter = models.ActivityFilterReactions
	case models.ActivityFilterReactions:
		m.filter = models.ActivityFilterMentions
	case models.ActivityFilterMentions:
		m.filter = models.ActivityFilterAll
	default:
		m.filter = models.ActivityFilterAll
	}
}

// View renders the activity view
func (m ActivityModel) View() string {
	var parts []string

	// Render filter tabs (no empty line after)
	filterTabs := m.renderFilterTabs()
	parts = append(parts, filterTabs)

	// Render activity list or empty state
	if len(m.activities) == 0 {
		emptyMsg := m.localize("activity.empty", "No activity yet. Mentions, reactions, and thread replies will appear here.")
		parts = append(parts, styles.Dim.Render(emptyMsg))
	} else {
		filtered := models.FilterActivities(m.activities, m.filter)
		if len(filtered) == 0 {
			emptyMsg := m.localize("activity.empty_filter", "No activities match this filter.")
			parts = append(parts, styles.Dim.Render(emptyMsg))
		} else {
			// Get list view and trim any leading newlines/whitespace
			listView := m.list.View()
			listView = strings.TrimLeft(listView, "\n")
			parts = append(parts, listView)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// Create border style with horizontal padding only
	// Width needs to account for content + padding (2) + borders (2) = content + 4
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Width(m.width - 4). // Subtract border (2) and padding (2)
		Height(m.height).
		PaddingLeft(1).
		PaddingRight(1)

	if m.focused {
		borderStyle = borderStyle.BorderForeground(styles.ColourSuccess)
	}

	return borderStyle.Render(content)
}

// renderFilterTabs renders a subtle filter status line
func (m ActivityModel) renderFilterTabs() string {
	var filterName string
	switch m.filter {
	case models.ActivityFilterAll:
		filterName = "All"
	case models.ActivityFilterMentions:
		filterName = "Mentions"
	case models.ActivityFilterReactions:
		filterName = "Reactions"
	default:
		filterName = "All"
	}

	// Minimal status line: "[All] (←/→)"
	current := styles.Label.Bold(true).Render("[" + filterName + "]")
	hint := styles.Dim.Render(" (←/→)")

	return current + hint
}

// localize is a helper function to localize a message by ID with an optional fallback
func (m ActivityModel) localize(messageID string, fallback string) string {
	cfg := &i18n.LocalizeConfig{
		MessageID: messageID,
	}
	msg, err := m.localizer.Localize(cfg)
	if err != nil && fallback != "" {
		return fallback
	}
	return msg
}
