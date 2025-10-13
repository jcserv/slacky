package models

import (
	"github.com/slack-go/slack"
)

// ChannelType represents the type of channel
type ChannelType string

const (
	ChannelTypePublic  ChannelType = "public"
	ChannelTypePrivate ChannelType = "private"
	ChannelTypeDM      ChannelType = "dm"
	ChannelTypeMPDM    ChannelType = "mpdm" // Multi-person DM
)

// Channel represents a Slack channel or DM
type Channel struct {
	ID          string
	Name        string
	Type        ChannelType
	Topic       string
	Purpose     string
	IsMember    bool
	UnreadCount int
	HasUnread   bool
	IsArchived  bool
	IsStarred   bool

	// For DMs
	UserID   string // User ID for DMs
	UserName string // User name for DMs
}

// FromSlackChannel converts a slack.Channel to our Channel model
func FromSlackChannel(sc slack.Channel) Channel {
	channelType := ChannelTypePublic
	if sc.IsPrivate {
		channelType = ChannelTypePrivate
	}

	return Channel{
		ID:         sc.ID,
		Name:       sc.Name,
		Type:       channelType,
		Topic:      sc.Topic.Value,
		Purpose:    sc.Purpose.Value,
		IsMember:   sc.IsMember,
		IsArchived: sc.IsArchived,
	}
}

// GetDisplayName returns the display name for the channel
func (c Channel) GetDisplayName() string {
	switch c.Type {
	case ChannelTypeDM:
		if c.UserName != "" {
			return c.UserName
		}
		return c.Name
	case ChannelTypeMPDM:
		return c.Name
	case ChannelTypePrivate:
		return c.Name
	case ChannelTypePublic:
		return "#" + c.Name
	default:
		return c.Name
	}
}

// GetIcon returns the icon for the channel type
func (c Channel) GetIcon() string {
	switch c.Type {
	case ChannelTypePublic:
		return "#"
	case ChannelTypePrivate:
		return "🔒"
	case ChannelTypeDM:
		return "@"
	case ChannelTypeMPDM:
		return "👥"
	default:
		return "#"
	}
}

// GetDescription returns a description for the channel
func (c Channel) GetDescription() string {
	if c.Topic != "" {
		return c.Topic
	}
	if c.Purpose != "" {
		return c.Purpose
	}
	return ""
}
