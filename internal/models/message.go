package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

// Message represents a Slack message
type Message struct {
	ID        string
	ChannelID string
	UserID    string
	UserName  string
	Text      string
	Timestamp time.Time
	ThreadTS  string // Thread timestamp (if part of a thread)
	IsEdited  bool
	IsPinned  bool

	// Reactions
	Reactions []Reaction

	// Files/Attachments
	Files       []slack.File
	Attachments []slack.Attachment
}

// Reaction represents a message reaction
type Reaction struct {
	Name  string
	Count int
	Users []string
}

// FromSlackMessage converts a slack.Message to our Message model
func FromSlackMessage(sm slack.Message, channelID string) Message {
	timestamp, _ := parseSlackTimestamp(sm.Timestamp)

	msg := Message{
		ID:          sm.Timestamp, // Use timestamp as ID
		ChannelID:   channelID,
		UserID:      sm.User,
		Text:        sm.Text,
		Timestamp:   timestamp,
		ThreadTS:    sm.ThreadTimestamp,
		IsEdited:    sm.Edited != nil,
		Files:       sm.Files,
		Attachments: sm.Attachments,
	}

	// Convert reactions
	for _, r := range sm.Reactions {
		msg.Reactions = append(msg.Reactions, Reaction{
			Name:  r.Name,
			Count: r.Count,
			Users: r.Users,
		})
	}

	return msg
}

// parseSlackTimestamp converts Slack's timestamp format to time.Time
// Slack timestamps are in format "1234567890.123456"
func parseSlackTimestamp(ts string) (time.Time, error) {
	parts := strings.Split(ts, ".")
	if len(parts) == 0 {
		return time.Time{}, fmt.Errorf("invalid timestamp format")
	}

	// Parse the seconds part
	var seconds int64
	_, _ = fmt.Sscanf(parts[0], "%d", &seconds)

	return time.Unix(seconds, 0), nil
}

// ParseSlackTimestampToTime is a public version of parseSlackTimestamp
// for use by other packages
func ParseSlackTimestampToTime(ts string) time.Time {
	t, _ := parseSlackTimestamp(ts)
	return t
}

// FormatTime returns a formatted timestamp for display
func (m Message) FormatTime() string {
	now := time.Now()
	diff := now.Sub(m.Timestamp)

	// If today, show time
	if diff < 24*time.Hour && now.Day() == m.Timestamp.Day() {
		return m.Timestamp.Format("15:04")
	}

	// If this week, show day and time
	if diff < 7*24*time.Hour {
		return m.Timestamp.Format("Mon 15:04")
	}

	// Otherwise show date
	return m.Timestamp.Format("Jan 2 15:04")
}

// GetDisplayText returns the formatted text for display
func (m Message) GetDisplayText() string {
	// TODO: Handle Slack markdown formatting, user mentions, etc.
	return m.Text
}

// IsThreadReply returns true if this message is a reply in a thread
func (m Message) IsThreadReply() bool {
	return m.ThreadTS != "" && m.ThreadTS != m.ID
}

// IsNewerThan returns true if this message is newer than the given timestamp
func (m Message) IsNewerThan(timestamp string) bool {
	if timestamp == "" {
		return true
	}
	// Compare timestamps (Slack timestamps are in format "1234567890.123456")
	return m.ID > timestamp
}

// GetTimestamp returns the message timestamp as a string (same as ID)
func (m Message) GetTimestamp() string {
	return m.ID
}
