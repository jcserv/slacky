package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/jcserv/slacky/internal/util"
	"github.com/slack-go/slack"
)

// Message represents a Slack message
type Message struct {
	ID         string
	ChannelID  string
	UserID     string
	UserName   string
	Text       string
	Timestamp  time.Time
	ThreadTS   string // Thread timestamp (if part of a thread)
	ReplyCount int    // Number of replies in thread (only for parent messages)
	IsEdited   bool
	IsPinned   bool

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
		ReplyCount:  sm.ReplyCount,
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

	var seconds int64
	n, err := fmt.Sscanf(parts[0], "%d", &seconds)
	if err != nil || n != 1 {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

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
	var parts []string

	// Add message text if present
	if m.Text != "" {
		parts = append(parts, util.ConvertEmojiInText(m.Text))
	}

	// Add file attachments (each file is already formatted with potential newlines)
	if len(m.Files) > 0 {
		for _, file := range m.Files {
			parts = append(parts, formatFileInfo(file))
		}
	}

	// Add attachment information (for things like link previews, app messages)
	// Only show if there's no text and no files
	if len(parts) == 0 && len(m.Attachments) > 0 {
		for _, att := range m.Attachments {
			if att.Title != "" {
				attachmentText := "📎 " + att.Title
				if att.TitleLink != "" {
					attachmentText += "\n   🔗 " + att.TitleLink
				}
				parts = append(parts, attachmentText)
			} else if att.Fallback != "" {
				parts = append(parts, att.Fallback)
			}
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "\n")
}

// IsThreadReply returns true if this message is a reply in a thread
func (m Message) IsThreadReply() bool {
	return m.ThreadTS != "" && m.ThreadTS != m.ID
}

// IsThreadParent returns true if this message is the parent of a thread
func (m Message) IsThreadParent() bool {
	return m.ThreadTS != "" && m.ThreadTS == m.ID && m.ReplyCount > 0
}

// HasThread returns true if this message has thread replies
func (m Message) HasThread() bool {
	return m.ReplyCount > 0
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

// getFileIcon returns an emoji icon based on the file type
func getFileIcon(filetype string) string {
	switch strings.ToLower(filetype) {
	// Images
	case "png", "jpg", "jpeg", "gif", "bmp", "svg", "webp", "ico":
		return "📷"
	// Videos
	case "mp4", "mov", "avi", "mkv", "webm", "flv", "wmv":
		return "📹"
	// Audio
	case "mp3", "wav", "flac", "aac", "ogg", "m4a", "wma":
		return "🎵"
	// Documents
	case "pdf", "doc", "docx", "txt", "rtf", "odt":
		return "📄"
	// Spreadsheets
	case "xlsx", "xls", "csv", "ods":
		return "📊"
	// Presentations
	case "ppt", "pptx", "key", "odp":
		return "📊"
	// Archives
	case "zip", "tar", "gz", "rar", "7z", "bz2":
		return "🗜️"
	// Code
	case "go", "py", "js", "ts", "java", "c", "cpp", "h", "rs", "rb", "php", "html", "css", "json", "xml", "yaml", "yml":
		return "💻"
	default:
		return "📎"
	}
}

// formatFileSize formats a file size in bytes to a human-readable string
func formatFileSize(bytes int) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// makeClickableURL creates a terminal hyperlink using OSC 8 escape sequences
// Format: \e]8;;URL\e\\TEXT\e]8;;\e\\
func makeClickableURL(url, text string) string {
	if url == "" {
		return text
	}
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}

// formatFileInfo formats a file attachment for display
func formatFileInfo(file slack.File) string {
	icon := getFileIcon(file.Filetype)
	name := file.Name
	if name == "" {
		name = file.Title
	}
	if name == "" {
		name = "file"
	}

	size := formatFileSize(file.Size)

	// Build the first line with icon, name, and size
	var firstLine string
	if file.OriginalW > 0 && file.OriginalH > 0 {
		firstLine = fmt.Sprintf("%s %s (%s, %dx%d)", icon, name, size, file.OriginalW, file.OriginalH)
	} else {
		firstLine = fmt.Sprintf("%s %s (%s)", icon, name, size)
	}

	// Add URL on a new line with indentation
	url := file.URLPrivate
	if url == "" {
		url = file.Permalink
	}
	if url != "" {
		return fmt.Sprintf("%s\n   🔗 %s", firstLine, url)
	}

	return firstLine
}
