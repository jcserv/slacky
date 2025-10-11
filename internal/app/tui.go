package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jcserv/slacky/internal/tui/styles"
)

var quitKeys = key.NewBinding(
	key.WithKeys("esc", "ctrl+c"),
	key.WithHelp("", "ctrl+c to quit"),
)

// TUIModel holds the main application state
type TUIModel struct {
	app         *App
	spinner     spinner.Model
	quitting    bool
	err         error
	authSuccess bool
	teamName    string
	userName    string
}

// NewTUI creates a new TUI model with the given app
func (app *App) NewTUI() TUIModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Label
	return TUIModel{
		app:     app,
		spinner: s,
	}
}

// Init initializes the application
func (m TUIModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, loadConfig(m.app))
}

// Update handles messages and updates the model
func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil

	case authSuccessMsg:
		m.authSuccess = true
		m.teamName = msg.teamName
		m.userName = msg.userName
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

// View renders the application UI
func (m TUIModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  %s %v\n\n  %s\n\n",
			styles.Error.Render("Error:"),
			m.err,
			quitKeys.Help().Desc,
		)
	}

	if m.authSuccess {
		return fmt.Sprintf("\n  %s Connected to %s as %s\n\n  %s\n\n",
			styles.Success.Render("✓"),
			styles.Info.Render(m.teamName),
			styles.Highlight.Render(m.userName),
			quitKeys.Help().Desc,
		)
	}

	str := fmt.Sprintf("\n\n   %s %s %s\n\n",
		m.spinner.View(),
		styles.Subtitle.Render("Connecting to Slack..."),
		quitKeys.Help().Desc,
	)
	if m.quitting {
		return str + "\n"
	}
	return str
}
