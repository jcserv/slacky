package app

import (
	"context"
	"fmt"

	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/slack"
)

// App represents the main application instance
type App struct {
	config      *config.Config
	SlackClient *slack.Client
	ctx         context.Context

	// cleanupFuncs holds functions to call during shutdown
	cleanupFuncs []func() error
}

// New creates a new application instance
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	// Create Slack client
	var client *slack.Client
	if cfg.Workspace.UserToken != "" {
		// New format: user token
		client = slack.New(cfg.Workspace.UserToken)
	} else if cfg.Workspace.BotToken != "" && cfg.Workspace.SocketToken != "" {
		// Legacy format: bot + socket tokens
		client = slack.NewWithSocketMode(cfg.Workspace.BotToken, cfg.Workspace.SocketToken)
	} else {
		return nil, fmt.Errorf("no valid Slack credentials found in config")
	}

	app := &App{
		config:       cfg,
		SlackClient:  client,
		ctx:          ctx,
		cleanupFuncs: []func() error{},
	}

	// Add cleanup function for Slack client
	app.cleanupFuncs = append(app.cleanupFuncs, func() error {
		// Close any Slack connections if needed
		// For now, the slack client doesn't have explicit cleanup
		return nil
	})

	return app, nil
}

// Config returns the application configuration
func (app *App) Config() *config.Config {
	return app.config
}

// Shutdown performs a graceful shutdown of the application
func (app *App) Shutdown() {
	// Call all cleanup functions
	for _, cleanup := range app.cleanupFuncs {
		if cleanup != nil {
			if err := cleanup(); err != nil {
				// Log error but don't fail shutdown
				fmt.Printf("Warning: cleanup error: %v\n", err)
			}
		}
	}
}
