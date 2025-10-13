package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	ErrConfigNotFound = errors.New("config file not found")
	ErrInvalidConfig  = errors.New("invalid config file")
)

// Config represents the application configuration
type Config struct {
	Workspace   Workspace    `yaml:"workspace"`
	UI          UI           `yaml:"ui"`
	Keybindings *Keybindings `yaml:"keybindings,omitempty"`
}

// Workspace contains Slack workspace credentials
type Workspace struct {
	UserToken  string `yaml:"user_token"`
	TeamName   string `yaml:"team_name,omitempty"`
	TeamID     string `yaml:"team_id,omitempty"`
	UserID     string `yaml:"user_id,omitempty"`

	// Legacy fields (deprecated)
	BotToken    string `yaml:"bot_token,omitempty"`
	SocketToken string `yaml:"socket_token,omitempty"`
}

// UI contains UI preferences
type UI struct {
	Theme          string `yaml:"theme"`
	VimMode        bool   `yaml:"vim_mode"`
	ShowTimestamps bool   `yaml:"show_timestamps"`
}

// Keybinding represents a single keybinding configuration
type Keybinding struct {
	Key    string `yaml:"key"`
	Action string `yaml:"action"`
}

// Keybindings contains keybinding configurations organized by scope
type Keybindings struct {
	Global   []Keybinding `yaml:"global,omitempty"`
	Chat     []Keybinding `yaml:"chat,omitempty"`
	Message  []Keybinding `yaml:"message,omitempty"`
	Init     []Keybinding `yaml:"init,omitempty"`
	User     []Keybinding `yaml:"user,omitempty"`
	Activity []Keybinding `yaml:"activity,omitempty"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		Workspace: Workspace{},
		UI: UI{
			Theme:          "default",
			VimMode:        true,
			ShowTimestamps: true,
		},
	}
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "slacky")
	return filepath.Join(configDir, "config.yaml"), nil
}

// Load reads the config file from the default location
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrConfigNotFound
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	return &cfg, nil
}

// Save writes the config to the default location
func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Validate checks if the config has required fields
func (c *Config) Validate() error {
	// Check for user token (new format)
	if c.Workspace.UserToken != "" {
		return nil
	}

	// Fallback to legacy bot token format
	if c.Workspace.BotToken != "" && c.Workspace.SocketToken != "" {
		return nil
	}

	return errors.New("user_token is required (or legacy bot_token and socket_token)")
}
