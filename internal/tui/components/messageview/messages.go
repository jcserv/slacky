package messageview

// ReactionToggleRequestMsg is sent when user wants to toggle a reaction
type ReactionToggleRequestMsg struct {
	ChannelID string
	Timestamp string
	EmojiName string
}
