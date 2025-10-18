package messages

import "github.com/jcserv/slacky/internal/util"

// ConvertEmoji converts a Slack emoji shortcode to Unicode emoji
// If the emoji is not found in the map, returns the original shortcode with colons
func ConvertEmoji(shortcode string) string {
	return util.ConvertEmoji(shortcode)
}

// ConvertEmojiInText converts all emoji shortcodes in a text string to Unicode emojis
// Example: "Hello :wave: :smile:" becomes "Hello 👋 😄"
func ConvertEmojiInText(text string) string {
	return util.ConvertEmojiInText(text)
}
