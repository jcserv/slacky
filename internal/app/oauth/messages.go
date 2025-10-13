package oauth

import (
	slackyOAuth "github.com/jcserv/slacky/internal/oauth"
)

// OAuthCompleteMsg is sent when the OAuth flow completes successfully
type OAuthCompleteMsg struct {
	Response *slackyOAuth.TokenResponse
}

// OAuthErrorMsg is sent when the OAuth flow fails
type OAuthErrorMsg struct {
	Err error
}
