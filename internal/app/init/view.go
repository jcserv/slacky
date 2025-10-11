package init

import (
	"fmt"
	"strings"

	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/tui/components"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// view renders the initialization wizard UI
func view(m Model) string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(m.logoRendered)
	s.WriteString("\n\n")

	s.WriteString(styles.Title.Render("Thanks for trying out Slacky!"))
	s.WriteString("\n")
	s.WriteString(styles.Subtitle.Render("Let's set up your Slack workspace connection"))
	s.WriteString("\n\n")

	if m.step == StepAuthFailed {
		return renderAuthFailedView(m, s)
	}

	if m.step >= StepWelcome && m.step < StepPreferences {
		s.WriteString(styles.Highlight.Render("📋 Get your tokens from https://api.slack.com/apps"))
		s.WriteString("\n")
		s.WriteString(styles.Subtitle.Render("  • Bot Token: OAuth & Permissions → Bot User OAuth Token"))
		s.WriteString("\n")
		s.WriteString(styles.Subtitle.Render("  • Socket Token: Socket Mode → App-Level Token"))
		s.WriteString("\n\n")
	}

	if m.step == StepWelcome {
		s.WriteString(styles.Highlight.Render("Press enter to continue"))
		s.WriteString("\n")
		return s.String()
	}

	s.WriteString(styles.Border.Render("───────────────────────────────────────────────────────────────"))
	s.WriteString("\n\n")

	// Render bot token step
	if m.step >= StepBotToken {
		if m.step > StepBotToken {
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

	// Render socket token step
	if m.step >= StepSocketToken {
		if m.step > StepSocketToken {
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

	// Render testing step
	if m.step >= StepTesting {
		if m.step > StepTesting {
			s.WriteString(styles.Completed.Render(fmt.Sprintf("✓ Connected to %s as %s", m.teamName, m.userName)))
			s.WriteString("\n\n")
		} else {
			s.WriteString(fmt.Sprintf("%s Testing connection...", m.spinner.View()))
			s.WriteString("\n\n")
		}
	}

	// Render preferences step
	if m.step >= StepPreferences && m.step != StepComplete && m.step != StepError {
		s.WriteString(styles.Label.Render("Preferences"))
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

	// Render completion step
	if m.step == StepComplete {
		configPath, _ := config.ConfigPath()
		s.WriteString(styles.Success.Render("✓ Configuration saved!"))
		s.WriteString("\n\n")
		s.WriteString(styles.Subtitle.Render(fmt.Sprintf("Config file: %s", configPath)))
		s.WriteString("\n\n")
		s.WriteString(styles.Info.Render("Starting Slacky..."))
		s.WriteString("\n")
		return s.String()
	}

	// Render error step
	if m.step == StepError {
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

// renderAuthFailedView renders the authentication failed view
func renderAuthFailedView(m Model, s strings.Builder) string {
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
