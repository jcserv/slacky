package app

import (
	"fmt"

	"github.com/jcserv/slacky/internal/styles"
)

// view renders the application UI
func view(m Model) string {
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
