package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// PollingConfig holds configuration for different polling intervals
type PollingConfig struct {
	CurrentChannelInterval time.Duration // How often to poll the current channel for new messages
	SidebarInterval        time.Duration // How often to poll for sidebar unread counts
	ActivityInterval       time.Duration // How often to poll for new activities
	Enabled                bool          // Whether polling is enabled
}

// DefaultPollingConfig returns the default polling configuration
func DefaultPollingConfig() PollingConfig {
	return PollingConfig{
		CurrentChannelInterval: 5 * time.Second,  // Poll current channel every 5 seconds
		SidebarInterval:        30 * time.Second, // Poll sidebar every 30 seconds
		ActivityInterval:       60 * time.Second, // Poll activities every 60 seconds
		Enabled:                true,
	}
}

// pollTickMsg signals that it's time to poll for updates
type pollTickMsg struct {
	pollType PollType
}

// PollType represents the type of polling to perform
type PollType int

const (
	PollTypeCurrentChannel PollType = iota
	PollTypeSidebar
	PollTypeActivity
)

// startPolling returns a batch of commands to start all polling intervals
func startPolling(config PollingConfig) tea.Cmd {
	if !config.Enabled {
		return nil
	}

	return tea.Batch(
		pollCurrentChannel(config.CurrentChannelInterval),
		pollSidebar(config.SidebarInterval),
		pollActivity(config.ActivityInterval),
	)
}

// pollCurrentChannel schedules periodic polling for the current channel
func pollCurrentChannel(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return pollTickMsg{pollType: PollTypeCurrentChannel}
	})
}

// pollSidebar schedules periodic polling for sidebar unread counts
func pollSidebar(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return pollTickMsg{pollType: PollTypeSidebar}
	})
}

// pollActivity schedules periodic polling for activity updates
func pollActivity(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return pollTickMsg{pollType: PollTypeActivity}
	})
}
