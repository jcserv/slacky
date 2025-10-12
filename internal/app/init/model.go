package init

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/jcserv/slacky/internal/config"
	tuiKeys "github.com/jcserv/slacky/internal/tui/keys"
)

// Step represents the current step in the initialization wizard
type Step int

const (
	StepWelcome Step = iota
	StepBotToken
	StepBotTokenTesting
	StepSocketToken
	StepTesting
	StepPreferences
	StepComplete
	StepError
	StepAuthFailed
)

// Model holds the initialization wizard state
type Model struct {
	step           Step
	botToken       textinput.Model
	socketToken    textinput.Model
	vimMode        bool
	showTimestamps bool
	spinner        spinner.Model
	err            error
	teamName       string
	userName       string
	existingConfig *config.Config
	quitting       bool
	continueToApp  bool
	prefCursor     int
	authFailCursor int
	width          int
	logoRendered   string
	version        string
	localizer      *i18n.Localizer
	keyMap         *tuiKeys.ScopedKeyMap
}

// InitResult represents the result of the initialization process
type InitResult struct {
	ShouldContinue bool
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return update(m, msg)
}

// View renders the application UI
func (m Model) View() string {
	return view(m)
}
