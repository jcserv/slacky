package oauth

import (
	"errors"
	"os"
)

var (
	ErrMissingClientID     = errors.New("SLACK_CLIENT_ID environment variable is required")
	ErrMissingClientSecret = errors.New("SLACK_CLIENT_SECRET environment variable is required")
)

// Config holds OAuth configuration
type Config struct {
	ClientID     string
	ClientSecret string
}

// LoadFromEnv loads OAuth config from environment variables
func LoadFromEnv() (*Config, error) {
	clientID := os.Getenv("SLACK_CLIENT_ID")
	if clientID == "" {
		return nil, ErrMissingClientID
	}

	clientSecret := os.Getenv("SLACK_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, ErrMissingClientSecret
	}

	return &Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}
