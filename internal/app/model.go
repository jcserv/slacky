package app

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Model holds the main application state
type Model struct {
	spinner     spinner.Model
	quitting    bool
	err         error
	authSuccess bool
	teamName    string
	userName    string
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, loadConfig())
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return update(m, msg)
}

// View renders the application UI
func (m Model) View() string {
	return view(m)
}
