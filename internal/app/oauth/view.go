package oauth

import (
	"fmt"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/jcserv/slacky/internal/tui"
	"github.com/jcserv/slacky/internal/tui/actions"
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

// View renders the OAuth wizard UI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(m.logoRendered)
	s.WriteString("\n\n")

	s.WriteString(styles.Title.Render("Thanks for trying out Slacky!"))
	s.WriteString("\n")
	s.WriteString(styles.Subtitle.Render("Let's set up your Slack workspace connection"))
	s.WriteString("\n\n")

	if m.step == StepError {
		return renderErrorView(m, &s)
	}

	if m.step == StepSuccess {
		return renderSuccessView(m, &s)
	}

	// Show OAuth setup instructions on welcome and input steps
	if m.step >= StepWelcome && m.step < StepRunningOAuth {
		s.WriteString(styles.Highlight.Render("🔐 OAuth Setup"))
		s.WriteString("\n")
		s.WriteString(styles.Subtitle.Render("  • Get your Client ID and Client Secret from https://api.slack.com/apps"))
		s.WriteString("\n")
		s.WriteString(styles.Subtitle.Render("  • Go to your app → Basic Information → App Credentials"))
		s.WriteString("\n")
	}

	if m.step == StepWelcome {
		enterKey, _ := m.keyMap.GetBinding(actions.ActionContinue, actions.ScopeInit)
		quitKey, _ := m.keyMap.GetBinding(actions.ActionQuit, actions.ScopeInit)
		s.WriteString(tui.RenderKeyBindingsWithBorder(enterKey, quitKey))
		return s.String()
	}

	s.WriteString(tui.RenderBorder(63))
	s.WriteString("\n")

	// Render Client ID step
	if m.step >= StepClientID {
		if m.step > StepClientID {
			s.WriteString(styles.Completed.Render("✓ Client ID"))
			s.WriteString("\n")
			s.WriteString(styles.Dim.Render("  " + maskCredential(m.clientID.Value())))
			s.WriteString("\n\n")
		} else {
			s.WriteString(styles.Label.Render("Client ID"))
			s.WriteString("\n")
			s.WriteString(styles.Subtitle.Render("Enter your Slack app's Client ID"))
			s.WriteString("\n\n")
			s.WriteString("  " + m.clientID.View())
			s.WriteString("\n\n")
			if m.err != nil {
				s.WriteString("  " + styles.Error.Render("✗ "+m.err.Error()))
				s.WriteString("\n\n")
			}
		}
	}

	// Render Client Secret step
	if m.step >= StepClientSecret {
		if m.step > StepClientSecret {
			s.WriteString(styles.Completed.Render("✓ Client Secret"))
			s.WriteString("\n")
			s.WriteString(styles.Dim.Render("  " + maskCredential(m.clientSecret.Value())))
			s.WriteString("\n\n")
		} else {
			s.WriteString(styles.Label.Render("Client Secret"))
			s.WriteString("\n")
			s.WriteString(styles.Subtitle.Render("Enter your Slack app's Client Secret"))
			s.WriteString("\n\n")
			s.WriteString("  " + m.clientSecret.View())
			s.WriteString("\n\n")
			if m.err != nil {
				s.WriteString("  " + styles.Error.Render("✗ "+m.err.Error()))
				s.WriteString("\n\n")
			}
		}
	}

	// Render OAuth flow running
	if m.step == StepRunningOAuth {
		s.WriteString(fmt.Sprintf("%s %s", m.spinner.View(), "Running OAuth flow..."))
		s.WriteString("\n")
		s.WriteString(styles.Dim.Render("  Your browser will open for authorization"))
		s.WriteString("\n\n")
	}

	if m.step < StepRunningOAuth {
		enterKey, _ := m.keyMap.GetBinding(actions.ActionContinue, actions.ScopeInit)
		quitKey, _ := m.keyMap.GetBinding(actions.ActionQuit, actions.ScopeInit)
		s.WriteString(tui.RenderKeyBindingsWithBorder(enterKey, quitKey))
	}

	return s.String()
}

// renderErrorView renders the error state
func renderErrorView(m Model, s *strings.Builder) string {
	s.WriteString(tui.RenderBorder(63))
	s.WriteString("\n\n")
	s.WriteString(styles.Error.Render("✗ OAuth failed"))
	s.WriteString("\n\n")
	s.WriteString(styles.Subtitle.Render(m.err.Error()))
	s.WriteString("\n\n")
	s.WriteString(styles.Help.Render("Press Enter to exit"))
	s.WriteString("\n")
	return s.String()
}

// renderSuccessView renders the success state
func renderSuccessView(m Model, s *strings.Builder) string {
	s.WriteString(tui.RenderBorder(63))
	s.WriteString("\n\n")
	s.WriteString(styles.Success.Render("✓ Successfully authenticated!"))
	s.WriteString("\n\n")
	if m.tokenResp != nil {
		s.WriteString(styles.Label.Render("  Team: ") + m.tokenResp.Team.Name)
		s.WriteString("\n")
		s.WriteString(styles.Label.Render("  User: ") + m.tokenResp.AuthedUser.ID)
		s.WriteString("\n\n")
	}
	s.WriteString(styles.Subtitle.Render("Press Enter to continue"))
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

// maskCredential masks sensitive credential information
func maskCredential(cred string) string {
	if len(cred) <= 10 {
		return strings.Repeat("•", len(cred))
	}
	prefix := cred[:8]
	return prefix + strings.Repeat("•", len(cred)-8)
}
