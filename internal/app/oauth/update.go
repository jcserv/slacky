package oauth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	slackyOAuth "github.com/jcserv/slacky/internal/oauth"
	"github.com/jcserv/slacky/internal/tui/actions"
)

// Init initializes the OAuth wizard
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle action keys only when not actively typing in an input
		if m.step == StepWelcome || m.step == StepRunningOAuth ||
			m.step == StepSuccess || m.step == StepError {
			return m.handleKeyPress(msg)
		}

		// For input steps, check for action keys first, then let input handle the rest
		switch {
		case m.keyMap.MatchesAction(msg, actions.ActionQuit, actions.ScopeInit):
			m.quitting = true
			m.shouldContinue = false
			return m, tea.Quit

		case m.keyMap.MatchesAction(msg, actions.ActionContinue, actions.ScopeInit):
			return m.handleContinue()
		}

		// Let the input field handle the key
		// (fall through to input update below)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.logoRendered = renderLogo(m)
		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case OAuthCompleteMsg:
		m.tokenResp = msg.Response
		m.step = StepSuccess
		return m, nil

	case OAuthErrorMsg:
		m.err = msg.Err
		m.step = StepError
		return m, nil
	}

	// Update the active input field
	switch m.step {
	case StepClientID:
		m.clientID, cmd = m.clientID.Update(msg)
		cmds = append(cmds, cmd)
	case StepClientSecret:
		m.clientSecret, cmd = m.clientSecret.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case m.keyMap.MatchesAction(msg, actions.ActionQuit, actions.ScopeInit):
		if m.step == StepRunningOAuth {
			// Don't allow quitting during OAuth flow
			return m, nil
		}
		m.quitting = true
		m.shouldContinue = false
		return m, tea.Quit

	case m.keyMap.MatchesAction(msg, actions.ActionContinue, actions.ScopeInit):
		return m.handleContinue()
	}

	return m, nil
}

// handleContinue handles the continue/enter action
func (m Model) handleContinue() (tea.Model, tea.Cmd) {
	switch m.step {
	case StepWelcome:
		m.step = StepClientID
		m.clientID.Focus()
		return m, nil

	case StepClientID:
		// Validate Client ID
		clientID := strings.TrimSpace(m.clientID.Value())
		if clientID == "" {
			m.err = fmt.Errorf("Client ID cannot be empty")
			return m, nil
		}
		m.err = nil
		m.step = StepClientSecret
		m.clientID.Blur()
		m.clientSecret.Focus()
		return m, nil

	case StepClientSecret:
		// Validate Client Secret
		clientSecret := strings.TrimSpace(m.clientSecret.Value())
		if clientSecret == "" {
			m.err = fmt.Errorf("Client Secret cannot be empty")
			return m, nil
		}
		m.err = nil
		m.step = StepRunningOAuth
		m.clientSecret.Blur()
		return m, m.runOAuthFlow()

	case StepSuccess:
		m.shouldContinue = true
		m.quitting = true
		return m, tea.Quit

	case StepError:
		m.quitting = true
		m.shouldContinue = false
		return m, tea.Quit
	}

	return m, nil
}

// runOAuthFlow executes the OAuth flow
func (m Model) runOAuthFlow() tea.Cmd {
	return func() tea.Msg {
		clientID := strings.TrimSpace(m.clientID.Value())
		clientSecret := strings.TrimSpace(m.clientSecret.Value())

		scopes := []string{
			"channels:history",
			"channels:read",
			"channels:write",
			"chat:write",
			"groups:history",
			"groups:read",
			"groups:write",
			"im:history",
			"im:read",
			"im:write",
			"mpim:history",
			"mpim:read",
			"mpim:write",
			"users:read",
		}

		ctx := context.Background()
		tokenResp, err := slackyOAuth.Flow(ctx, slackyOAuth.FlowOptions{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       scopes,
			Port:         8080,
			Timeout:      5 * time.Minute,
			OnURL: func(authURL string) error {
				// Try to open browser
				return slackyOAuth.OpenBrowser(authURL)
			},
			WriteSuccessHTML: writeSuccessPage,
		})

		if err != nil {
			return OAuthErrorMsg{Err: err}
		}

		return OAuthCompleteMsg{Response: tokenResp}
	}
}
