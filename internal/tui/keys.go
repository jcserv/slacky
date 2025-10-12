package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines global application key bindings
type KeyMap struct {
	Quit       key.Binding
	Help       key.Binding
	NextTab    key.Binding
	PrevTab    key.Binding
	SelectUser key.Binding
}

// DefaultKeyMap returns the default key bindings for the application
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("tab", "right"),
			key.WithHelp("tab/→", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab", "left"),
			key.WithHelp("shift+tab/←", "previous tab"),
		),
		SelectUser: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "user info"),
		),
	}
}

// ShortHelp returns key bindings to be shown in the short help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextTab, k.PrevTab, k.SelectUser, k.Help, k.Quit}
}

// FullHelp returns key bindings to be shown in the full help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextTab, k.PrevTab, k.SelectUser},
		{k.Help, k.Quit},
	}
}
