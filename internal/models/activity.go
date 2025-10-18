package models

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityTypeMention     ActivityType = "mention"
	ActivityTypeReaction    ActivityType = "reaction"
	ActivityTypeUnreadDM    ActivityType = "unread_dm"
	ActivityTypeThreadReply ActivityType = "thread_reply"
)

// Activity represents a user activity notification
type Activity struct {
	Type          ActivityType
	ChannelID     string
	ChannelName   string
	ChannelType   ChannelType // Reuse from channel.go
	UserID        string      // User who triggered the activity
	UserName      string
	MessageID     string // Message timestamp
	MessageText   string // Preview of the message
	Timestamp     time.Time
	ThreadTS      string // Thread timestamp (if this is a thread activity)
	ReactionName  string // For reaction activities
	ReactionCount int    // For reaction activities
}

// GetIcon returns an icon representing the activity type
func (a Activity) GetIcon() string {
	switch a.Type {
	case ActivityTypeMention:
		return "@"
	case ActivityTypeReaction:
		return "👍"
	case ActivityTypeUnreadDM:
		return "💬"
	case ActivityTypeThreadReply:
		return "🧵"
	default:
		return "•"
	}
}

// GetTypeLabel returns a human-readable label for the activity type
func (a Activity) GetTypeLabel() string {
	switch a.Type {
	case ActivityTypeMention:
		return "Mention"
	case ActivityTypeReaction:
		return "Reaction"
	case ActivityTypeUnreadDM:
		return "Unread DM"
	case ActivityTypeThreadReply:
		return "Thread Reply"
	default:
		return "Activity"
	}
}

// GetChannelDisplay returns the formatted channel name for display
func (a Activity) GetChannelDisplay() string {
	switch a.ChannelType {
	case ChannelTypeDM:
		return "@" + a.ChannelName
	case ChannelTypeMPDM:
		return a.ChannelName
	case ChannelTypePrivate:
		return a.ChannelName
	case ChannelTypePublic:
		return "#" + a.ChannelName
	default:
		return a.ChannelName
	}
}

// GetMessagePreview returns a truncated message preview for display
func (a Activity) GetMessagePreview(maxLength int) string {
	text := a.MessageText

	// Remove newlines for preview
	text = strings.ReplaceAll(text, "\n", " ")

	// Truncate if too long
	if len(text) > maxLength {
		return text[:maxLength-3] + "..."
	}

	return text
}

// FormatTime returns a formatted timestamp for display
func (a Activity) FormatTime() string {
	now := time.Now()
	diff := now.Sub(a.Timestamp)

	// Less than a minute
	if diff < time.Minute {
		return "just now"
	}

	// Less than an hour - show minutes
	if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", minutes)
	}

	// Less than a day - show hours
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	}

	// If today, show time
	if now.Day() == a.Timestamp.Day() && now.Month() == a.Timestamp.Month() && now.Year() == a.Timestamp.Year() {
		return a.Timestamp.Format("15:04")
	}

	// If this week, show day and time
	if diff < 7*24*time.Hour {
		return a.Timestamp.Format("Mon 15:04")
	}

	// Otherwise show date
	return a.Timestamp.Format("Jan 2 15:04")
}

// GetReactionDisplay returns a formatted reaction string (e.g., ":thumbsup: 3")
func (a Activity) GetReactionDisplay() string {
	if a.Type != ActivityTypeReaction {
		return ""
	}

	if a.ReactionCount > 1 {
		return fmt.Sprintf(":%s: %d", a.ReactionName, a.ReactionCount)
	}

	return fmt.Sprintf(":%s:", a.ReactionName)
}

// ActivityFilter represents the filter type for activities
type ActivityFilter string

const (
	ActivityFilterAll       ActivityFilter = "all"
	ActivityFilterMentions  ActivityFilter = "mentions"
	ActivityFilterReactions ActivityFilter = "reactions"
)

// FilterActivities filters a slice of activities based on the filter type
func FilterActivities(activities []Activity, filter ActivityFilter) []Activity {
	if filter == ActivityFilterAll {
		return activities
	}

	filtered := make([]Activity, 0)
	for _, activity := range activities {
		switch filter {
		case ActivityFilterMentions:
			if activity.Type == ActivityTypeMention {
				filtered = append(filtered, activity)
			}
		case ActivityFilterReactions:
			if activity.Type == ActivityTypeReaction {
				filtered = append(filtered, activity)
			}
		}
	}

	return filtered
}

// SortActivitiesByTime sorts activities by timestamp (most recent first)
func SortActivitiesByTime(activities []Activity) []Activity {
	sorted := make([]Activity, len(activities))
	copy(sorted, activities)

	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.After(sorted[j].Timestamp)
	})

	return sorted
}
