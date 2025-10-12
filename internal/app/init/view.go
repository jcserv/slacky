package init

import (
	"fmt"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/jcserv/slacky/internal/tui"
	"github.com/jcserv/slacky/internal/tui/components"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// localize is a helper function to localize a message by ID with an optional fallback
func (m Model) localize(messageID string, fallback string, templateData ...map[string]interface{}) string {
	cfg := &i18n.LocalizeConfig{
		MessageID: messageID,
	}
	if len(templateData) > 0 {
		cfg.TemplateData = templateData[0]
	}
	msg, err := m.localizer.Localize(cfg)
	if err != nil && fallback != "" {
		return fallback
	}
	return msg
}

// view renders the initialization wizard UI
func view(m Model) string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(m.logoRendered)
	s.WriteString("\n\n")

	s.WriteString(styles.Title.Render(m.localize("init.title", "Thanks for trying out Slacky!")))
	s.WriteString("\n")
	s.WriteString(styles.Subtitle.Render(m.localize("init.subtitle", "Let's set up your Slack workspace connection")))
	s.WriteString("\n\n")

	if m.step == StepAuthFailed {
		return renderAuthFailedView(m, &s)
	}

	if m.step >= StepWelcome && m.step < StepPreferences && m.step != StepBotTokenTesting && m.step != StepTesting {
		s.WriteString(styles.Highlight.Render(m.localize("init.tokens_help", "📋 Get your tokens from https://api.slack.com/apps")))
		s.WriteString("\n")
		s.WriteString(styles.Subtitle.Render(m.localize("init.bot_token_help", "  • Bot Token: OAuth & Permissions → Bot User OAuth Token")))
		s.WriteString("\n")
		s.WriteString(styles.Subtitle.Render(m.localize("init.socket_token_help", "  • Socket Token: Socket Mode → App-Level Token")))
		s.WriteString("\n\n")
	}

	if m.step == StepWelcome {
		s.WriteString(styles.Highlight.Render(m.localize("init.welcome_prompt", "Press enter to continue")))
		s.WriteString("\n")
		return s.String()
	}

	s.WriteString(tui.RenderBorder(63))
	s.WriteString("\n\n")

	// Render bot token step
	if m.step >= StepBotToken {
		if m.step > StepBotTokenTesting {
			s.WriteString(styles.Completed.Render("✓ " + m.localize("init.bot_token_label", "Bot Token")))
			s.WriteString("\n")
			s.WriteString(styles.Dim.Render("  " + maskToken(m.botToken.Value())))
			s.WriteString("\n\n")
		} else if m.step == StepBotTokenTesting {
			s.WriteString(styles.Completed.Render("✓ " + m.localize("init.bot_token_label", "Bot Token")))
			s.WriteString("\n")
			s.WriteString(styles.Dim.Render("  " + maskToken(m.botToken.Value())))
			s.WriteString("\n\n")
			s.WriteString(fmt.Sprintf("%s %s", m.spinner.View(), m.localize("init.testing_bot_token", "Testing bot token...")))
			s.WriteString("\n\n")
		} else {
			s.WriteString(styles.Label.Render(m.localize("init.bot_token_label", "Bot Token")))
			s.WriteString("\n")
			s.WriteString(styles.Subtitle.Render(m.localize("init.bot_token_prompt", "Enter your Bot User OAuth Token (starts with xoxb-)")))
			s.WriteString("\n\n")
			s.WriteString("  " + m.botToken.View())
			s.WriteString("\n\n")
			if m.err != nil {
				s.WriteString("  " + styles.Error.Render("✗ "+m.err.Error()))
				s.WriteString("\n\n")
			}
		}
	}

	// Render socket token step
	if m.step >= StepSocketToken {
		if m.step > StepSocketToken {
			s.WriteString(styles.Completed.Render("✓ " + m.localize("init.socket_token_label", "Socket Token")))
			s.WriteString("\n")
			s.WriteString(styles.Dim.Render("  " + maskToken(m.socketToken.Value())))
			s.WriteString("\n\n")
		} else {
			s.WriteString(styles.Label.Render(m.localize("init.socket_token_label", "Socket Token")))
			s.WriteString("\n")
			s.WriteString(styles.Subtitle.Render(m.localize("init.socket_token_prompt", "Enter your App-Level Token for Socket Mode (starts with xapp-)")))
			s.WriteString("\n\n")
			s.WriteString("  " + m.socketToken.View())
			s.WriteString("\n\n")
			if m.err != nil {
				s.WriteString("  " + styles.Error.Render("✗ "+m.err.Error()))
				s.WriteString("\n\n")
			}
		}
	}

	// Render testing step
	if m.step >= StepTesting {
		if m.step > StepTesting {
			connectedMsg := m.localize("init.connected",
				fmt.Sprintf("✓ Connected to %s as %s", m.teamName, m.userName),
				map[string]interface{}{
					"TeamName": m.teamName,
					"UserName": m.userName,
				})
			s.WriteString(styles.Completed.Render(connectedMsg))
			s.WriteString("\n\n")
		} else {
			s.WriteString(fmt.Sprintf("%s %s", m.spinner.View(), m.localize("init.testing_connection", "Testing connection...")))
			s.WriteString("\n\n")
		}
	}

	// Render preferences step
	if m.step >= StepPreferences && m.step != StepError {
		s.WriteString(styles.Label.Render(m.localize("init.preferences_label", "Preferences")))
		s.WriteString("\n\n")

		timestampIcon := "☐"
		if m.showTimestamps {
			timestampIcon = "☑"
		}
		timestampLine := fmt.Sprintf("  %s %s", timestampIcon, m.localize("init.show_timestamps", "Show timestamps on messages"))
		if m.prefCursor == 0 {
			s.WriteString(styles.Highlight.Render("> ") + timestampLine + "\n")
		} else {
			s.WriteString(styles.Dim.Render("  ") + timestampLine + "\n")
		}

		s.WriteString("\n")
		s.WriteString(tui.RenderKeyBindings(keys.Up, keys.Down, keys.Toggle))
		s.WriteString("\n")
	}

	// Render error step
	if m.step == StepError {
		s.WriteString(styles.Error.Render("✗ " + m.localize("error.general", "Error") + ": " + m.err.Error()))
		s.WriteString("\n\n")
		s.WriteString(styles.Subtitle.Render(m.localize("error.please_check_tokens", "Please check your tokens and try again.")))
		s.WriteString("\n\n")
		s.WriteString(styles.Help.Render(m.localize("error.press_enter_exit", "Press Enter to exit")))
		s.WriteString("\n")
		return s.String()
	}

	s.WriteString(tui.RenderBorder(63))
	s.WriteString("\n")
	s.WriteString(tui.RenderKeyBindings(keys.Enter, keys.Quit))
	s.WriteString("\n")

	return s.String()
}

// renderAuthFailedView renders the authentication failed view
func renderAuthFailedView(m Model, s *strings.Builder) string {
	s.WriteString(tui.RenderBorder(63))
	s.WriteString("\n\n")
	s.WriteString(styles.Error.Render(m.localize("init.auth_failed", "✗ Authentication failed")))
	s.WriteString("\n\n")
	s.WriteString(styles.Subtitle.Render(m.err.Error()))
	s.WriteString("\n\n")
	s.WriteString(styles.Label.Render(m.localize("init.what_to_do", "What would you like to do?")))
	s.WriteString("\n\n")

	editBotLine := "  " + m.localize("init.edit_bot_token", "Edit Bot Token")
	if m.authFailCursor == 0 {
		s.WriteString(styles.Highlight.Render("> ") + editBotLine + "\n")
	} else {
		s.WriteString(styles.Dim.Render("  ") + editBotLine + "\n")
	}

	// editSocketLine := "  " + m.localize("init.edit_socket_token", "Edit Socket Token")
	// if m.authFailCursor == 1 {
	// 	s.WriteString(styles.Highlight.Render("> ") + editSocketLine + "\n")
	// } else {
	// 	s.WriteString(styles.Dim.Render("  ") + editSocketLine + "\n")
	// }

	retryLine := "  " + m.localize("init.retry_connection", "Retry Connection")
	if m.authFailCursor == 2 {
		s.WriteString(styles.Highlight.Render("> ") + retryLine + "\n")
	} else {
		s.WriteString(styles.Dim.Render("  ") + retryLine + "\n")
	}

	s.WriteString("\n")
	s.WriteString(tui.RenderKeyBindings(keys.Up, keys.Down, keys.Enter, keys.Quit))
	s.WriteString("\n")
	return s.String()
}

// renderLogo renders the application logo
func renderLogo(m Model) string {
	if m.width < components.MinWidth() {
		return components.SmallRender(m.version, m.width)
	}

	return components.Render(components.Opts{
		Version:      m.version,
		Width:        m.width,
		FillColor:    styles.ColourDim,
		VersionColor: styles.Tertiary,
	})
}

// maskToken masks sensitive token information
func maskToken(token string) string {
	if len(token) <= 10 {
		return token
	}
	prefix := token[:8]
	suffix := token[len(token)-4:]
	return prefix + "•••" + suffix
}
