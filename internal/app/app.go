package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/slack"
	"github.com/jcserv/slacky/internal/styles"
)

// Message types
type (
	errMsg         error
	authSuccessMsg struct {
		teamName string
		userName string
	}
)

// Model holds the main application state
type Model struct {
	spinner     spinner.Model
	quitting    bool
	err         error
	authSuccess bool
	teamName    string
	userName    string
}

var quitKeys = key.NewBinding(
	key.WithKeys("esc", "ctrl+c"),
	key.WithHelp("", "ctrl+c to quit"),
)

// New creates a new application model
func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Label
	return Model{spinner: s}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, loadConfig())
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m Model) View() string {
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

// loadConfig loads and validates the configuration
func loadConfig() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load()
		if err != nil {
			if errors.Is(err, config.ErrConfigNotFound) {
				return errMsg(fmt.Errorf("config not found. Run 'slacky init' to set up your configuration"))
			}
			return errMsg(err)
		}

		if err := cfg.Validate(); err != nil {
			return errMsg(fmt.Errorf("invalid config: %w", err))
		}

		return checkAuth(cfg)()
	}
}

// checkAuth tests authentication with Slack
func checkAuth(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		client := slack.New(cfg.Workspace.BotToken, cfg.Workspace.SocketToken)

		ctx := context.Background()
		authResp, err := client.TestAuth(ctx)
		if err != nil {
			return errMsg(fmt.Errorf("authentication failed: %w", err))
		}

		return authSuccessMsg{
			teamName: authResp.Team,
			userName: authResp.User,
		}
	}
}
