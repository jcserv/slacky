package oauth

import (
	"io"
	"log/slog"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	slackyI18n "github.com/jcserv/slacky/internal/i18n"
	tuiKeys "github.com/jcserv/slacky/internal/tui/keys"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// RunWithVersion runs the OAuth wizard with a specific version
func RunWithVersion(version string) (*OAuthResult, error) {
	m := initialModel()
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

	return &OAuthResult{
		ShouldContinue: fm.shouldContinue,
		TokenResponse:  fm.tokenResp,
	}, nil
}

// Run runs the OAuth wizard with default version
func Run() (*OAuthResult, error) {
	return RunWithVersion("dev")
}

// initialModel creates a new OAuth wizard model
func initialModel() Model {
	clientIDInput := textinput.New()
	clientIDInput.Placeholder = "1234567890.1234567890123"
	clientIDInput.CharLimit = 200
	clientIDInput.Width = 60
	clientIDInput.Focus()

	clientSecretInput := textinput.New()
	clientSecretInput.Placeholder = "********************************"
	clientSecretInput.CharLimit = 200
	clientSecretInput.Width = 60
	clientSecretInput.EchoMode = textinput.EchoPassword
	clientSecretInput.EchoCharacter = '•'

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Label

	// Create localizer with detected locale
	locale := slackyI18n.DetectLocale()
	localizer := slackyI18n.NewLocalizer(locale)

	// Load keybindings
	keyMap, err := tuiKeys.LoadKeybindings(nil, localizer)
	if err != nil {
		slog.Warn("Failed to load OAuth keybindings, using defaults", "error", err)
		keyMap, _ = tuiKeys.LoadKeybindings(nil, localizer)
	}

	return Model{
		step:         StepWelcome,
		clientID:     clientIDInput,
		clientSecret: clientSecretInput,
		spinner:      s,
		localizer:    localizer,
		keyMap:       keyMap,
	}
}

// writeSuccessPage writes the OAuth success HTML page
func writeSuccessPage(w io.Writer) {
	io.WriteString(w, `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Slacky - Authentication Successful</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .container {
            background: white;
            padding: 3rem;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
            text-align: center;
            max-width: 400px;
        }
        h1 {
            color: #333;
            margin-bottom: 1rem;
        }
        p {
            color: #666;
            line-height: 1.6;
        }
        .checkmark {
            font-size: 4rem;
            color: #4CAF50;
            margin-bottom: 1rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="checkmark">✓</div>
        <h1>Authentication Successful!</h1>
        <p>You can now close this window and return to your terminal.</p>
    </div>
</body>
</html>`)
}
