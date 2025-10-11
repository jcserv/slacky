package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/slack"
	"github.com/jcserv/slacky/internal/styles"
	"github.com/jcserv/slacky/internal/ui/components"
)

type step int

const (
	stepWelcome step = iota
	stepBotToken
	stepSocketToken
	stepTesting
	stepPreferences
	stepComplete
	stepError
	stepAuthFailed
)

type authTestMsg struct {
	teamName string
	userName string
	err      error
}

type configSavedMsg struct{}

type errMsg struct {
	err error
}

type initModel struct {
	step           step
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
}

type InitResult struct {
	ShouldContinue bool
}

func Init() (*InitResult, error) {
	return InitWithVersion("dev")
}

func InitWithVersion(version string) (*InitResult, error) {
	existingCfg, _ := config.Load()

	m := initialInitModel(existingCfg)
	m.version = version

	width, _, err := term.GetSize(0)
	if err != nil {
		width = 80
	}
	m.width = width
	m.logoRendered = m.renderLogo()

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(initModel)

	return &InitResult{
		ShouldContinue: fm.continueToApp,
	}, nil
}

func initialInitModel(existingCfg *config.Config) initModel {
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

	return initModel{
		step:           stepWelcome,
		botToken:       botInput,
		socketToken:    socketInput,
		spinner:        s,
		vimMode:        vimMode,
		showTimestamps: showTimestamps,
		existingConfig: existingCfg,
	}
}

func (m initModel) Init() tea.Cmd {
	return nil
}

func (m initModel) renderLogo() string {
	if m.width < components.MinWidth() {
		return components.SmallRender(m.version, m.width)
	}

	return components.Render(components.Opts{
		Version:      m.version,
		Width:        m.width,
		DiagColor:    styles.ColourDim,
		VersionColor: styles.Tertiary,
	})
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.logoRendered = m.renderLogo()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			return m.handleEnter()

		case "up", "k":
			if m.step == stepPreferences && m.prefCursor > 0 {
				m.prefCursor--
			}
			if m.step == stepAuthFailed && m.authFailCursor > 0 {
				m.authFailCursor--
			}

		case "down", "j":
			if m.step == stepPreferences && m.prefCursor < 1 {
				m.prefCursor++
			}
			if m.step == stepAuthFailed && m.authFailCursor < 2 {
				m.authFailCursor++
			}

		case " ", "space":
			if m.step == stepPreferences {
				switch m.prefCursor {
				case 0:
					m.vimMode = !m.vimMode
				case 1:
					m.showTimestamps = !m.showTimestamps
				}
			}
		}

	case authTestMsg:
		if msg.err != nil {
			m.err = msg.err
			m.step = stepAuthFailed
			return m, nil
		}
		m.teamName = msg.teamName
		m.userName = msg.userName
		m.step = stepPreferences
		return m, nil

	case spinner.TickMsg:
		if m.step == stepTesting {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case configSavedMsg:
		m.step = stepComplete
		m.continueToApp = true
		return m, tea.Quit

	case errMsg:
		m.err = msg.err
		m.step = stepError
		return m, nil
	}

	var cmd tea.Cmd
	switch m.step {
	case stepBotToken:
		m.botToken, cmd = m.botToken.Update(msg)
	case stepSocketToken:
		m.socketToken, cmd = m.socketToken.Update(msg)
	}

	return m, cmd
}

func (m initModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepWelcome:
		m.step = stepBotToken
		m.botToken.Focus()
		return m, textinput.Blink

	case stepBotToken:
		value := strings.TrimSpace(m.botToken.Value())
		if value == "" {
			m.err = errors.New("bot token is required")
			return m, nil
		}
		m.err = nil
		m.step = stepSocketToken
		m.botToken.Blur()
		m.socketToken.Focus()
		return m, textinput.Blink

	case stepSocketToken:
		value := strings.TrimSpace(m.socketToken.Value())
		if value == "" {
			m.err = errors.New("socket token is required")
			return m, nil
		}
		m.err = nil
		m.step = stepTesting
		m.socketToken.Blur()
		return m, tea.Batch(m.spinner.Tick, m.testAuth())

	case stepPreferences:
		return m.saveConfig()

	case stepAuthFailed:
		switch m.authFailCursor {
		case 0:
			m.step = stepBotToken
			m.err = nil
			m.botToken.Focus()
			return m, textinput.Blink
		case 1:
			m.step = stepSocketToken
			m.err = nil
			m.socketToken.Focus()
			return m, textinput.Blink
		case 2:
			m.step = stepTesting
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, m.testAuth())
		}

	case stepComplete:
		m.continueToApp = true
		return m, tea.Quit

	case stepError:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m initModel) testAuth() tea.Cmd {
	return func() tea.Msg {
		client := slack.New(m.botToken.Value(), m.socketToken.Value())
		ctx := context.Background()
		authResp, err := client.TestAuth(ctx)
		if err != nil {
			return authTestMsg{err: err}
		}
		return authTestMsg{
			teamName: authResp.Team,
			userName: authResp.User,
		}
	}
}

func (m initModel) saveConfig() (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		cfg := &config.Config{
			Workspace: config.Workspace{
				BotToken:    m.botToken.Value(),
				SocketToken: m.socketToken.Value(),
			},
			UI: config.UI{
				Theme:          "default",
				VimMode:        m.vimMode,
				ShowTimestamps: m.showTimestamps,
			},
		}

		if err := config.Save(cfg); err != nil {
			return errMsg{err: fmt.Errorf("failed to save config: %w", err)}
		}

		return configSavedMsg{}
	}
}

func (m initModel) View() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(m.logoRendered)
	s.WriteString("\n\n")

	s.WriteString(styles.Title.Render("Thanks for trying out Slacky!"))
	s.WriteString("\n")
	s.WriteString(styles.Subtitle.Render("Let's set up your Slack workspace connection"))
	s.WriteString("\n\n")

	if m.step == stepAuthFailed {
		s.WriteString(styles.Border.Render("───────────────────────────────────────────────────────────────"))
		s.WriteString("\n\n")
		s.WriteString(styles.Error.Render("✗ Authentication failed"))
		s.WriteString("\n\n")
		s.WriteString(styles.Subtitle.Render(m.err.Error()))
		s.WriteString("\n\n")
		s.WriteString(styles.Label.Render("What would you like to do?"))
		s.WriteString("\n\n")

		editBotLine := "  Edit Bot Token"
		if m.authFailCursor == 0 {
			s.WriteString(styles.Highlight.Render("> ") + editBotLine + "\n")
		} else {
			s.WriteString(styles.Dim.Render("  ") + editBotLine + "\n")
		}

		editSocketLine := "  Edit Socket Token"
		if m.authFailCursor == 1 {
			s.WriteString(styles.Highlight.Render("> ") + editSocketLine + "\n")
		} else {
			s.WriteString(styles.Dim.Render("  ") + editSocketLine + "\n")
		}

		retryLine := "  Retry Connection"
		if m.authFailCursor == 2 {
			s.WriteString(styles.Highlight.Render("> ") + retryLine + "\n")
		} else {
			s.WriteString(styles.Dim.Render("  ") + retryLine + "\n")
		}

		s.WriteString("\n")
		s.WriteString(styles.Help.Render("[↑/↓] navigate • [enter] select • [ctrl+c] quit"))
		s.WriteString("\n")
		return s.String()
	}

	if m.step >= stepWelcome && m.step < stepPreferences {
		s.WriteString(styles.Highlight.Render("📋 Get your tokens from https://api.slack.com/apps"))
		s.WriteString("\n")
		s.WriteString(styles.Subtitle.Render("  • Bot Token: OAuth & Permissions → Bot User OAuth Token"))
		s.WriteString("\n")
		s.WriteString(styles.Subtitle.Render("  • Socket Token: Socket Mode → App-Level Token"))
		s.WriteString("\n\n")
	}

	if m.step == stepWelcome {
		s.WriteString(styles.Highlight.Render("Press enter to continue"))
		s.WriteString("\n")
		return s.String()
	}

	s.WriteString(styles.Border.Render("───────────────────────────────────────────────────────────────"))
	s.WriteString("\n\n")

	if m.step >= stepBotToken {
		if m.step > stepBotToken {
			s.WriteString(styles.Completed.Render("✓ Bot Token"))
			s.WriteString("\n")
			s.WriteString(styles.Dim.Render("  " + maskToken(m.botToken.Value())))
			s.WriteString("\n\n")
		} else {
			s.WriteString(styles.Label.Render("Bot Token"))
			s.WriteString("\n")
			s.WriteString(styles.Subtitle.Render("Enter your Bot User OAuth Token (starts with xoxb-)"))
			s.WriteString("\n\n")
			s.WriteString("  " + m.botToken.View())
			s.WriteString("\n\n")
			if m.err != nil {
				s.WriteString("  " + styles.Error.Render("✗ "+m.err.Error()))
				s.WriteString("\n\n")
			}
		}
	}

	if m.step >= stepSocketToken {
		if m.step > stepSocketToken {
			s.WriteString(styles.Completed.Render("✓ Socket Token"))
			s.WriteString("\n")
			s.WriteString(styles.Dim.Render("  " + maskToken(m.socketToken.Value())))
			s.WriteString("\n\n")
		} else {
			s.WriteString(styles.Label.Render("Socket Token"))
			s.WriteString("\n")
			s.WriteString(styles.Subtitle.Render("Enter your App-Level Token for Socket Mode (starts with xapp-)"))
			s.WriteString("\n\n")
			s.WriteString("  " + m.socketToken.View())
			s.WriteString("\n\n")
			if m.err != nil {
				s.WriteString("  " + styles.Error.Render("✗ "+m.err.Error()))
				s.WriteString("\n\n")
			}
		}
	}

	if m.step >= stepTesting {
		if m.step > stepTesting {
			s.WriteString(styles.Completed.Render(fmt.Sprintf("✓ Connected to %s as %s", m.teamName, m.userName)))
			s.WriteString("\n\n")
		} else {
			s.WriteString(fmt.Sprintf("%s Testing connection...", m.spinner.View()))
			s.WriteString("\n\n")
		}
	}

	if m.step >= stepPreferences && m.step != stepComplete && m.step != stepError {
		s.WriteString(styles.Label.Render("UI Preferences"))
		s.WriteString("\n\n")

		vimIcon := "☐"
		if m.vimMode {
			vimIcon = "☑"
		}
		vimLine := fmt.Sprintf("  %s Enable vim-style keybindings", vimIcon)
		if m.prefCursor == 0 {
			s.WriteString(styles.Highlight.Render("> ") + vimLine + "\n")
		} else {
			s.WriteString(styles.Dim.Render("  ") + vimLine + "\n")
		}

		timestampIcon := "☐"
		if m.showTimestamps {
			timestampIcon = "☑"
		}
		timestampLine := fmt.Sprintf("  %s Show timestamps on messages", timestampIcon)
		if m.prefCursor == 1 {
			s.WriteString(styles.Highlight.Render("> ") + timestampLine + "\n")
		} else {
			s.WriteString(styles.Dim.Render("  ") + timestampLine + "\n")
		}

		s.WriteString("\n")
		s.WriteString(styles.Help.Render("[↑/↓] navigate • [space] select"))
		s.WriteString("\n")
	}

	if m.step == stepComplete {
		configPath, _ := config.ConfigPath()
		s.WriteString(styles.Success.Render("✓ Configuration saved!"))
		s.WriteString("\n\n")
		s.WriteString(styles.Subtitle.Render(fmt.Sprintf("Config file: %s", configPath)))
		s.WriteString("\n\n")
		s.WriteString(styles.Info.Render("Starting Slacky..."))
		s.WriteString("\n")
		return s.String()
	}

	if m.step == stepError {
		s.WriteString(styles.Error.Render("✗ Error: " + m.err.Error()))
		s.WriteString("\n\n")
		s.WriteString(styles.Subtitle.Render("Please check your tokens and try again."))
		s.WriteString("\n\n")
		s.WriteString(styles.Help.Render("Press Enter to exit"))
		s.WriteString("\n")
		return s.String()
	}

	s.WriteString(styles.Border.Render("───────────────────────────────────────────────────────────────"))
	s.WriteString("\n")
	s.WriteString(styles.Help.Render("[enter] continue • [ctrl+c] quit"))
	s.WriteString("\n")

	return s.String()
}

func maskToken(token string) string {
	if len(token) <= 10 {
		return token
	}
	prefix := token[:8]
	suffix := token[len(token)-4:]
	return prefix + "•••" + suffix
}
