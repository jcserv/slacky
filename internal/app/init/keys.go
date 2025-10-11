package init

import (
	"github.com/charmbracelet/bubbles/key"

	"github.com/jcserv/slacky/internal/tui"
)

// keyMap defines all key bindings for the init wizard
type keyMap struct {
	tui.KeyMap // Embed global keys (Quit, Help)
	Enter      key.Binding
	Up         key.Binding
	Down       key.Binding
	Toggle     key.Binding
}

// defaultKeyMap returns the default key bindings for the init wizard
func defaultKeyMap() keyMap {
	return keyMap{
		KeyMap: tui.DefaultKeyMap(),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "continue"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" ", "space"),
			key.WithHelp("space", "toggle"),
		),
	}
}

// keys holds all key bindings for the init wizard
var keys = defaultKeyMap()
