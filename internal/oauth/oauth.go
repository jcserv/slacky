package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// SlackOAuthURL is the Slack OAuth authorization endpoint
	SlackOAuthURL = "https://slack.com/oauth/v2/authorize"
	// SlackTokenURL is the Slack OAuth token exchange endpoint
	SlackTokenURL = "https://slack.com/api/oauth.v2.access"
)

var (
	ErrTimeout       = errors.New("OAuth flow timed out")
	ErrAuthFailed    = errors.New("OAuth authorization failed")
	ErrServerFailed  = errors.New("failed to start OAuth callback server")
	ErrInvalidState  = errors.New("invalid OAuth state parameter")
	ErrAccessDenied  = errors.New("user denied access")
)

// FlowOptions contains options for the OAuth flow
type FlowOptions struct {
	// ClientID is your Slack app's client ID
	ClientID string
	// ClientSecret is your Slack app's client secret
	ClientSecret string
	// Scopes are the OAuth scopes to request
	Scopes []string
	// Port is the local port for the callback server (0 for random port)
	Port int
	// Timeout is how long to wait for the OAuth flow to complete
	Timeout time.Duration
	// OnURL is called with the authorization URL that should be opened in a browser
	OnURL func(authURL string) error
	// WriteSuccessHTML is called to write the success page HTML
	WriteSuccessHTML func(w io.Writer)
}

// TokenResponse contains the OAuth token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	BotUserID   string `json:"bot_user_id"`
	AppID       string `json:"app_id"`
	Team        struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"team"`
	AuthedUser struct {
		ID          string `json:"id"`
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	} `json:"authed_user"`
}

// Flow executes the OAuth authorization flow
func Flow(ctx context.Context, opts FlowOptions) (*TokenResponse, error) {
	if opts.ClientID == "" {
		return nil, errors.New("ClientID is required")
	}
	if opts.ClientSecret == "" {
		return nil, errors.New("ClientSecret is required")
	}
	if len(opts.Scopes) == 0 {
		return nil, errors.New("at least one scope is required")
	}
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Minute
	}

	// Generate state for CSRF protection
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Start local callback server (HTTP for localhost)
	server, port, codeChan, errChan, err := startCallbackServer(opts.Port, state, opts.WriteSuccessHTML)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrServerFailed, err)
	}
	defer server.Close()

	// Build redirect URI (HTTP is allowed for localhost by Slack)
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	// Build authorization URL
	authURL := buildAuthURL(opts.ClientID, redirectURI, state, opts.Scopes)

	// Call the OnURL callback to open the browser
	if opts.OnURL != nil {
		if err := opts.OnURL(authURL); err != nil {
			return nil, err
		}
	}

	// Wait for callback or timeout
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	select {
	case code := <-codeChan:
		// Exchange code for token
		return exchangeCode(ctx, opts.ClientID, opts.ClientSecret, code, redirectURI)
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ErrTimeout
	}
}

// generateState generates a random state parameter for CSRF protection
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// buildAuthURL builds the authorization URL
func buildAuthURL(clientID, redirectURI, state string, scopes []string) string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	// For user tokens, only use user_scope (not scope which is for bot tokens)
	// Slack expects space-separated scopes, not comma-separated
	params.Set("user_scope", strings.Join(scopes, " "))

	return SlackOAuthURL + "?" + params.Encode()
}

// startCallbackServer starts a local HTTP server to receive the OAuth callback
func startCallbackServer(port int, expectedState string, writeSuccess func(io.Writer)) (*http.Server, int, chan string, chan error, error) {
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Use specified port or find available port
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	actualPort := listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Check for errors
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			errChan <- fmt.Errorf("%w: %s", ErrAuthFailed, errParam)
			http.Error(w, "Authorization failed", http.StatusBadRequest)
			return
		}

		// Verify state
		state := r.URL.Query().Get("state")
		if state != expectedState {
			errChan <- ErrInvalidState
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			return
		}

		// Get authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("%w: no code in response", ErrAuthFailed)
			http.Error(w, "No authorization code received", http.StatusBadRequest)
			return
		}

		// Send code to channel
		codeChan <- code

		// Write success page
		w.Header().Set("Content-Type", "text/html")
		if writeSuccess != nil {
			writeSuccess(w)
		} else {
			fmt.Fprint(w, defaultSuccessHTML)
		}
	})

	server := &http.Server{
		Handler: mux,
	}

	// Start server in background
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	return server, actualPort, codeChan, errChan, nil
}

// exchangeCode exchanges the authorization code for an access token
func exchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", SlackTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
		TokenResponse
	}

	if err := parseJSON(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("token exchange failed: %s", result.Error)
	}

	// For user tokens, we want the authed_user access token
	if result.AuthedUser.AccessToken != "" {
		result.TokenResponse.AccessToken = result.AuthedUser.AccessToken
		result.TokenResponse.TokenType = result.AuthedUser.TokenType
		result.TokenResponse.Scope = result.AuthedUser.Scope
	}

	return &result.TokenResponse, nil
}

// parseJSON is a simple JSON parser helper
func parseJSON(r io.Reader, v interface{}) error {
	// Use encoding/json
	decoder := json.NewDecoder(r)
	return decoder.Decode(v)
}

var defaultSuccessHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Slacky - Authentication Successful</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .container {
            background: white;
            padding: 3rem;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
            text-align: center;
            max-width: 400px;
        }
        h1 {
            color: #333;
            margin-bottom: 1rem;
        }
        p {
            color: #666;
            line-height: 1.6;
        }
        .checkmark {
            font-size: 4rem;
            color: #4CAF50;
            margin-bottom: 1rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="checkmark">✓</div>
        <h1>Authentication Successful!</h1>
        <p>You can now close this window and return to your terminal.</p>
    </div>
</body>
</html>`
