package oauth

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	slackyOAuth "github.com/jcserv/slacky/internal/oauth"
	tuiKeys "github.com/jcserv/slacky/internal/tui/keys"
)

// Step represents the current step in the OAuth wizard
type Step int

const (
	StepWelcome Step = iota
	StepClientID
	StepClientSecret
	StepRunningOAuth
	StepSuccess
	StepError
)

// Model represents the OAuth wizard state
type Model struct {
	step         Step
	clientID     textinput.Model
	clientSecret textinput.Model
	spinner      spinner.Model
	err          error
	width        int
	height       int
	version      string
	logoRendered string

	// OAuth flow results
	tokenResp *slackyOAuth.TokenResponse

	// i18n
	localizer *i18n.Localizer

	// Keybindings
	keyMap *tuiKeys.ScopedKeyMap

	// Control flow
	shouldContinue bool
	quitting       bool
}

// OAuthResult represents the result of the OAuth wizard
type OAuthResult struct {
	ShouldContinue bool
	TokenResponse  *slackyOAuth.TokenResponse
}
