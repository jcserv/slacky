package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jcserv/slacky/internal/styles"
)

const asciiArt = `███████╗██╗      ██████╗  ██████╗██╗  ██╗██╗   ██╗
██╔════╝██║     ██╔═══██╗██╔════╝██║ ██╔╝╚██╗ ██╔╝
███████╗██║     ████████║██║     █████╔╝  ╚████╔╝
╚════██║██║     ██╔═══██║██║     ██╔═██╗   ╚██╔╝
███████║███████╗██║   ██║╚██████╗██║  ██╗   ██║
╚══════╝╚══════╝╚═╝   ╚═╝ ╚═════╝╚═╝  ╚═╝   ╚═╝`

const fillCharacter = "»"

type Opts struct {
	Version      string
	Width        int
	FillColor    lipgloss.Color
	VersionColor lipgloss.Color
}

// Render renders the Slacky logo with version string.
// Layout: left slashes, version aligned to logo end, right slashes extend to width.
func Render(opts Opts) string {
	var b strings.Builder

	lines := strings.Split(asciiArt, "\n")
	logoWidth := lipgloss.Width(lines[0])

	const leftWidth = 4
	leftField := strings.Repeat(fillCharacter, leftWidth)
	leftFieldStyle := lipgloss.NewStyle().Foreground(opts.FillColor)
	versionStyle := lipgloss.NewStyle().Foreground(opts.VersionColor)

	rightWidth := opts.Width - leftWidth - 1 - logoWidth - 1
	if rightWidth < 0 {
		rightWidth = 0
	}

	versionEndPos := leftWidth + 1 + logoWidth
	versionStartPos := versionEndPos - lipgloss.Width(opts.Version)
	leftSlashes := leftFieldStyle.Render(leftField)

	spacesBeforeVersion := versionStartPos - leftWidth - 1
	if spacesBeforeVersion < 0 {
		spacesBeforeVersion = 0
	}

	rightSlashesCount := opts.Width - versionEndPos - 1
	if rightSlashesCount < 0 {
		rightSlashesCount = 0
	}
	rightSlashes := strings.Repeat(fillCharacter, rightSlashesCount)

	versionLine := leftSlashes + " " + strings.Repeat(" ", spacesBeforeVersion) +
		versionStyle.Render(opts.Version) + " " + leftFieldStyle.Render(rightSlashes)

	b.WriteString(versionLine)
	b.WriteString("\n")

	for i, line := range lines {
		currentRightWidth := rightWidth
		if i > 0 {
			currentRightWidth = rightWidth - i
			if currentRightWidth < 0 {
				currentRightWidth = 0
			}
		}
		rightField := strings.Repeat(fillCharacter, currentRightWidth)

		b.WriteString(leftFieldStyle.Render(leftField))
		b.WriteString(" ")
		b.WriteString(styles.Logo.Render(line))
		b.WriteString(" ")
		b.WriteString(leftFieldStyle.Render(rightField))
		if i < len(lines)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func Height() int {
	return len(strings.Split(asciiArt, "\n")) + 1
}

func LogoWidth() int {
	lines := strings.Split(asciiArt, "\n")
	if len(lines) > 0 {
		return lipgloss.Width(lines[0])
	}
	return 0
}

func MinWidth() int {
	return LogoWidth() + 6
}

func SmallRender(version string, width int) string {
	title := styles.Logo.Render("Slacky")
	versionStyle := lipgloss.NewStyle().Foreground(styles.Dim.GetForeground())

	titleWithVersion := fmt.Sprintf("%s %s", title, versionStyle.Render(version))
	remainingWidth := width - lipgloss.Width(titleWithVersion) - 1
	if remainingWidth > 0 {
		diagStyle := lipgloss.NewStyle().Foreground(styles.Dim.GetForeground())
		slashes := strings.Repeat(fillCharacter, remainingWidth)
		return fmt.Sprintf("%s %s", titleWithVersion, diagStyle.Render(slashes))
	}
	return titleWithVersion
}
