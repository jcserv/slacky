package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// TabType represents the type of tab
type TabType int

const (
	ChatTab TabType = iota
	ActivityTab
	UserTab
)

// Model represents the tabs component
type Model struct {
	currentTab TabType
	width      int
	userName   string
	localizer  *i18n.Localizer
}

// NewModel creates a new tabs model
func NewModel() Model {
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	return Model{
		currentTab: ChatTab,
		width:      80,
		localizer:  localizer,
	}
}

// Update handles messages for the tabs component
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// SetCurrentTab sets the currently active tab
func (m *Model) SetCurrentTab(tab TabType) {
	m.currentTab = tab
}

// GetCurrentTab returns the currently active tab
func (m Model) GetCurrentTab() TabType {
	return m.currentTab
}

// SetWidth sets the width of the tabs component
func (m *Model) SetWidth(width int) {
	m.width = width
}

// SetUserName sets the user name to display
func (m *Model) SetUserName(userName string) {
	m.userName = userName
}

// NextTab switches to the next tab
func (m *Model) NextTab() {
	switch m.currentTab {
	case ChatTab:
		m.currentTab = ActivityTab
	case ActivityTab:
		m.currentTab = UserTab
	case UserTab:
		m.currentTab = ChatTab
	}
}

// PrevTab switches to the previous tab
func (m *Model) PrevTab() {
	switch m.currentTab {
	case ChatTab:
		m.currentTab = UserTab
	case ActivityTab:
		m.currentTab = ChatTab
	case UserTab:
		m.currentTab = ActivityTab
	}
}

// View renders the tabs component
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	// Tab labels
	chatLabel := m.localize("tabs.chat", "Chat")
	activityLabel := m.localize("tabs.activity", "Activity")

	// Render tabs with active/inactive styles
	var tabs []string

	// Chat tab
	if m.currentTab == ChatTab {
		tabs = append(tabs, styles.ActiveTab.Render(chatLabel))
	} else {
		tabs = append(tabs, styles.Tab.Render(chatLabel))
	}

	// Separator
	tabs = append(tabs, styles.TabSeparator.Render(" │ "))

	// Activity tab
	if m.currentTab == ActivityTab {
		tabs = append(tabs, styles.ActiveTab.Render(activityLabel))
	} else {
		tabs = append(tabs, styles.Tab.Render(activityLabel))
	}

	// Join main tabs
	mainTabs := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// User info on the right (only if userName is set)
	var userInfo string
	if m.userName != "" {
		userLabel := "@" + m.userName
		var userStyle lipgloss.Style
		if m.currentTab == UserTab {
			userStyle = styles.ActiveTab
		} else {
			userStyle = styles.Tab
		}
		userInfo = userStyle.Render(userLabel)
	}

	// Calculate spacing to push user to the right
	mainTabsWidth := lipgloss.Width(mainTabs)
	userInfoWidth := lipgloss.Width(userInfo)
	spacing := m.width - mainTabsWidth - userInfoWidth - 4 // -4 for padding

	if spacing < 0 {
		spacing = 0
	}

	spacer := strings.Repeat(" ", spacing)

	// Join everything horizontally
	var row string
	if userInfo != "" {
		row = lipgloss.JoinHorizontal(lipgloss.Top, mainTabs, spacer, userInfo)
	} else {
		row = mainTabs
	}

	// Apply TabsRow style
	return styles.TabsRow.Width(m.width).Render(row)
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
