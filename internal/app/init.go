package app

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/jcserv/slacky/internal/styles"
)

// New creates a new application model with proper initialization
func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Label
	return Model{spinner: s}
}
