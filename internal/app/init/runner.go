package init

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// InitWithVersion runs the initialization wizard with a specific version
func InitWithVersion(version string) (*InitResult, error) {
	existingCfg, _ := config.Load()

	m := initialModel(existingCfg)
	m.version = version

	width, _, err := term.GetSize(0)
	if err != nil {
		width = 80
	}
	m.width = width
	m.logoRendered = renderLogo(m)

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(Model)

	return &InitResult{
		ShouldContinue: fm.continueToApp,
	}, nil
}

// Init runs the initialization wizard with default version
func Init() (*InitResult, error) {
	return InitWithVersion("dev")
}

// initialModel creates a new initialization model
func initialModel(existingCfg *config.Config) Model {
	botInput := textinput.New()
	botInput.Placeholder = "xoxb-..."
	botInput.CharLimit = 200
	botInput.Width = 60
	botInput.Focus()

	socketInput := textinput.New()
	socketInput.Placeholder = "xapp-..."
	socketInput.CharLimit = 200
	socketInput.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Label

	vimMode := true
	showTimestamps := true
	if existingCfg != nil {
		vimMode = existingCfg.UI.VimMode
		showTimestamps = existingCfg.UI.ShowTimestamps
	}

	return Model{
		step:           StepWelcome,
		botToken:       botInput,
		socketToken:    socketInput,
		spinner:        s,
		vimMode:        vimMode,
		showTimestamps: showTimestamps,
		existingConfig: existingCfg,
	}
}
