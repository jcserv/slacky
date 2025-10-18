package app

import (
	"context"
	"sync"

	"github.com/jcserv/slacky/internal/models"
	slackClient "github.com/jcserv/slacky/internal/slack"
	"github.com/slack-go/slack"
)

// userCache provides thread-safe caching of user information
type userCache struct {
	mu    sync.RWMutex
	users map[string]string // userID -> userName
}

// newUserCache creates a new user cache
func newUserCache() *userCache {
	return &userCache{
		users: make(map[string]string),
	}
}

// get retrieves a username from cache
func (c *userCache) get(userID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	name, ok := c.users[userID]
	return name, ok
}

// set stores a username in cache
func (c *userCache) set(userID, userName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.users[userID] = userName
}

// Global user cache instance
var globalUserCache = newUserCache()

// extractUserName extracts the best available name for a user.
// Priority: RealName > DisplayName > Name
func extractUserName(user *slack.User) string {
	if user.RealName != "" {
		return user.RealName
	}
	if user.Profile.DisplayName != "" {
		return user.Profile.DisplayName
	}
	return user.Name
}

// getUserNameWithCache fetches a user's name, using cache when available
func getUserNameWithCache(ctx context.Context, client *slackClient.Client, userID string) string {
	// Check cache first
	if name, ok := globalUserCache.get(userID); ok {
		return name
	}

	// Fetch from API
	user, err := client.GetUserInfo(ctx, userID)
	if err != nil {
		return ""
	}

	name := extractUserName(user)
	globalUserCache.set(userID, name)
	return name
}

// enrichMessagesWithUserInfo fetches user info for each message and populates UserName.
// Messages are processed in reverse order (newest first from Slack API).
// Uses caching to reduce API calls for repeated users.
func enrichMessagesWithUserInfo(ctx context.Context, client *slackClient.Client, slackMessages []slack.Message, channelID string) []models.Message {
	messages := make([]models.Message, 0, len(slackMessages))

	// Process in reverse order to get oldest-first
	for i := len(slackMessages) - 1; i >= 0; i-- {
		msg := models.FromSlackMessage(slackMessages[i], channelID)

		if msg.UserID != "" {
			msg.UserName = getUserNameWithCache(ctx, client, msg.UserID)
		}

		messages = append(messages, msg)
	}

	return messages
}
