package constants

import "time"

// Slack API limits
const (
	// MaxMessagesPerRequest is the maximum number of messages to fetch in a single API call
	MaxMessagesPerRequest = 100

	// MaxReactionsPerRequest is the maximum number of reactions to fetch
	MaxReactionsPerRequest = 50

	// DefaultMessagesPerRequest is the default number of messages to fetch
	DefaultMessagesPerRequest = 100

	// DefaultReactionsPerRequest is the default number of reactions to fetch
	DefaultReactionsPerRequest = 50
)

// Polling intervals
const (
	// DefaultCurrentChannelInterval is how often to poll the current channel for new messages
	DefaultCurrentChannelInterval = 5 * time.Second

	// DefaultSidebarInterval is how often to poll for sidebar unread counts
	DefaultSidebarInterval = 30 * time.Second
)

// UI dimensions and spacing
const (
	// UIHeaderHeight is the height reserved for the header area
	UIHeaderHeight = 6

	// UIMinHeight is the minimum terminal height required
	UIMinHeight = 10

	// UIMinWidth is the minimum terminal width required
	UIMinWidth = 40
)
