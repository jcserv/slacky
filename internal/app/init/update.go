package init

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// update handles messages and updates the model
func update(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.logoRendered = renderLogo(m)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, keys.Enter):
			return handleEnter(m)

		case key.Matches(msg, keys.Up):
			// if m.step == StepPreferences && m.prefCursor > 0 {
			// 	m.prefCursor--
			// }
			if m.step == StepAuthFailed && m.authFailCursor > 0 {
				m.authFailCursor--
			}

		case key.Matches(msg, keys.Down):
			// if m.step == StepPreferences && m.prefCursor < 1 {
			// 	m.prefCursor++
			// }
			if m.step == StepAuthFailed && m.authFailCursor < 2 {
				m.authFailCursor++
			}

		case key.Matches(msg, keys.Toggle):
			if m.step == StepPreferences {
				switch m.prefCursor {
				case 0:
					m.showTimestamps = !m.showTimestamps
				}
			}
		}

	case botTokenTestMsg:
		if msg.err != nil {
			m.err = msg.err
			m.step = StepAuthFailed
			return m, nil
		}
		m.teamName = msg.teamName
		m.userName = msg.userName
		m.step = StepSocketToken
		m.socketToken.Focus()
		return m, textinput.Blink

	case authTestMsg:
		if msg.err != nil {
			m.err = msg.err
			m.step = StepAuthFailed
			return m, nil
		}
		m.teamName = msg.teamName
		m.userName = msg.userName
		m.step = StepPreferences
		return m, nil

	case spinner.TickMsg:
		if m.step == StepBotTokenTesting || m.step == StepTesting {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case configSavedMsg:
		m.step = StepComplete
		m.continueToApp = true
		return m, tea.Quit

	case errMsg:
		m.err = msg.err
		m.step = StepError
		return m, nil
	}

	var cmd tea.Cmd
	switch m.step {
	case StepBotToken:
		m.botToken, cmd = m.botToken.Update(msg)
	case StepSocketToken:
		m.socketToken, cmd = m.socketToken.Update(msg)
	}

	return m, cmd
}

// handleEnter processes the enter key based on current step
func handleEnter(m Model) (tea.Model, tea.Cmd) {
	switch m.step {
	case StepWelcome:
		m.step = StepBotToken
		m.botToken.Focus()
		return m, textinput.Blink

	case StepBotToken:
		value := strings.TrimSpace(m.botToken.Value())
		if value == "" {
			m.err = errors.New("bot token is required")
			return m, nil
		}
		m.err = nil
		m.step = StepBotTokenTesting
		m.botToken.Blur()
		return m, tea.Batch(m.spinner.Tick, testBotToken(m.botToken.Value()))

	case StepSocketToken:
		value := strings.TrimSpace(m.socketToken.Value())
		if value == "" {
			m.err = errors.New("socket token is required")
			return m, nil
		}
		m.err = nil
		m.step = StepTesting
		m.socketToken.Blur()
		return m, tea.Batch(m.spinner.Tick, testAuth(m.botToken.Value(), m.socketToken.Value()))

	case StepPreferences:
		return m, saveConfig(m.botToken.Value(), m.socketToken.Value(), m.vimMode, m.showTimestamps)

	case StepAuthFailed:
		switch m.authFailCursor {
		case 0:
			m.step = StepBotToken
			m.err = nil
			m.botToken.Focus()
			return m, textinput.Blink
		case 1:
			m.step = StepSocketToken
			m.err = nil
			m.socketToken.Focus()
			return m, textinput.Blink
		case 2:
			m.step = StepTesting
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, testAuth(m.botToken.Value(), m.socketToken.Value()))
		}

	case StepComplete:
		m.continueToApp = true
		return m, tea.Quit

	case StepError:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}
