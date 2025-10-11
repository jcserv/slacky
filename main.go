package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jcserv/slacky/internal/cmd"
	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/slack"
	"github.com/jcserv/slacky/internal/styles"
)

type errMsg error
type authSuccessMsg struct {
	teamName string
	userName string
}

type model struct {
	spinner     spinner.Model
	quitting    bool
	err         error
	authSuccess bool
	teamName    string
	userName    string
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Label
	return model{spinner: s}
}

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

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, loadConfig())
}

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

func mustGetConfigPath() string {
	path, _ := config.ConfigPath()
	return path
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m model) View() string {
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

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			runInit()
			return
		case "help", "-h", "--help":
			printHelp()
			return
		case "version", "-v", "--version":
			printVersion()
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
			printHelp()
			os.Exit(1)
		}
	}

	// Check if config exists, if not run init wizard
	if _, err := config.Load(); err != nil {
		if errors.Is(err, config.ErrConfigNotFound) {
			fmt.Println(styles.Subtitle.Render("No configuration found. Let's set up Slacky!"))
			fmt.Println()
			runInit()
			return
		}
		// Other errors (permissions, invalid yaml, etc)
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Run the main TUI app
	runMainApp()
}

func runInit() {
	fmt.Println(styles.Dim.Render("Starting setup wizard..."))
	result, err := cmd.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	// If init completed successfully, continue to main app
	if result.ShouldContinue {
		runMainApp()
	}
}

func runMainApp() {
	// Run without alternate screen so it stays inline with terminal history
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Slacky - A terminal client for Slack")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  slacky           Start the Slack client")
	fmt.Println("  slacky init      Run setup wizard to create config file")
	fmt.Println("  slacky help      Show this help message")
	fmt.Println("  slacky version   Show version information")
	fmt.Println()
}

func printVersion() {
	fmt.Println("Slacky v0.1.0-dev")
}
