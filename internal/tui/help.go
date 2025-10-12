package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/jcserv/slacky/internal/tui/styles"
)

// RenderKeyBindings renders a list of key bindings in a consistent format
// Example: "[enter] continue • [ctrl+c] quit"
func RenderKeyBindings(bindings ...key.Binding) string {
	if len(bindings) == 0 {
		return ""
	}

	var parts []string
	for _, b := range bindings {
		help := b.Help()
		if help.Key == "" && help.Desc == "" {
			continue
		}

		keyPart := help.Key
		descPart := help.Desc

		// Format as [key] description
		parts = append(parts, "["+keyPart+"] "+descPart)
	}

	return styles.Help.Render(strings.Join(parts, " • "))
}

// RenderBorder renders a horizontal border line
func RenderBorder(width int) string {
	if width <= 0 {
		width = 63 // Default width
	}
	return styles.Border.Render(strings.Repeat("─", width))
}

// RenderKeyBindingsWithBorder renders keybinding instructions with a border above them
func RenderKeyBindingsWithBorder(bindings ...key.Binding) string {
	if len(bindings) == 0 {
		return ""
	}

	var s strings.Builder
	s.WriteString(RenderBorder(63))
	s.WriteString("\n")
	s.WriteString(RenderKeyBindings(bindings...))
	s.WriteString("\n")
	return s.String()
}
