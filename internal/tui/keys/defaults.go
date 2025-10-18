package keys

import "github.com/jcserv/slacky/internal/tui/actions"

// KeybindingDef defines a keybinding with its action and keys
type KeybindingDef struct {
	Action actions.Action
	Keys   []string // Multiple keys can trigger the same action
}

// DefaultGlobalKeybindings returns the default keybindings available globally
func DefaultGlobalKeybindings() []KeybindingDef {
	return []KeybindingDef{
		{Action: actions.ActionQuit, Keys: []string{"ctrl+c"}},
		{Action: actions.ActionHelp, Keys: []string{"?"}},
		{Action: actions.ActionToggleHelp, Keys: []string{"?"}},
		{Action: actions.ActionRefresh, Keys: []string{"r"}},
		{Action: actions.ActionSearch, Keys: []string{"{mod}+k", "/"}},
		{Action: actions.ActionCommandBar, Keys: []string{"{mod}+p"}},

		// Tab navigation
		{Action: actions.ActionNextTab, Keys: []string{"tab"}},
		{Action: actions.ActionPrevTab, Keys: []string{"shift+tab"}},
		{Action: actions.ActionEnterView, Keys: []string{" ", "space", "enter"}},
		{Action: actions.ActionExitView, Keys: []string{"esc"}},

		// View navigation (disabled by default, can be enabled in config)
		// {Action: actions.ActionGoToChat, Keys: []string{"1"}},
		// {Action: actions.ActionGoToActivity, Keys: []string{"2"}},
		// {Action: actions.ActionGoToUser, Keys: []string{"3", "u"}},

		// General navigation
		{Action: actions.ActionUp, Keys: []string{"up", "k"}},
		{Action: actions.ActionDown, Keys: []string{"down", "j"}},
		{Action: actions.ActionLeft, Keys: []string{"left", "h"}},
		{Action: actions.ActionRight, Keys: []string{"right", "l"}},
		{Action: actions.ActionPageUp, Keys: []string{"{mod}+u", "pgup"}},
		{Action: actions.ActionPageDown, Keys: []string{"{mod}+d", "pgdown"}},
		{Action: actions.ActionFirstLine, Keys: []string{"g", "home"}},
		{Action: actions.ActionLastLine, Keys: []string{"G", "end"}},

		// Selection
		{Action: actions.ActionContinue, Keys: []string{"enter"}},
		{Action: actions.ActionCancel, Keys: []string{"esc"}},
	}
}

// DefaultChatKeybindings returns the default keybindings for chat view
func DefaultChatKeybindings() []KeybindingDef {
	return []KeybindingDef{
		// Focus navigation
		{Action: actions.ActionBeginInput, Keys: []string{" ", "space"}},

		// Selection actions
		{Action: actions.ActionSelectConversation, Keys: []string{" ", "space", "enter"}},
		{Action: actions.ActionOpenThread, Keys: []string{" ", "space", "enter"}},

		// Message actions
		{Action: actions.ActionSendMessage, Keys: []string{"enter"}},
		{Action: actions.ActionEditMessage, Keys: []string{"e"}},
		{Action: actions.ActionDeleteMessage, Keys: []string{"d"}},
		{Action: actions.ActionReplyInThread, Keys: []string{"t"}},
		{Action: actions.ActionReact, Keys: []string{"r"}},
		{Action: actions.ActionSaveMessage, Keys: []string{"s"}},
		{Action: actions.ActionPinMessage, Keys: []string{"p"}},
		{Action: actions.ActionCopyLink, Keys: []string{"y"}},
		{Action: actions.ActionMarkUnread, Keys: []string{"u"}},

		// Channel actions
		{Action: actions.ActionStarChannel, Keys: []string{"*"}},
		{Action: actions.ActionCopyChannelName, Keys: []string{"{mod}+shift+c"}},
		{Action: actions.ActionCopyChannelLink, Keys: []string{"{mod}+shift+l"}},
	}
}

// DefaultMessageKeybindings returns the default keybindings for message input
func DefaultMessageKeybindings() []KeybindingDef {
	return []KeybindingDef{
		// Text formatting (matching Slack's defaults)
		{Action: actions.ActionBold, Keys: []string{"{mod}+b"}},
		{Action: actions.ActionItalic, Keys: []string{"{mod}+i"}},
		{Action: actions.ActionStrikethrough, Keys: []string{"{mod}+shift+x"}},
		{Action: actions.ActionLink, Keys: []string{"{mod}+shift+u"}},
		{Action: actions.ActionOrderedList, Keys: []string{"{mod}+shift+7"}},
		{Action: actions.ActionBulletList, Keys: []string{"{mod}+shift+8"}},
		{Action: actions.ActionBlockquote, Keys: []string{"{mod}+shift+9"}},
		{Action: actions.ActionCode, Keys: []string{"{mod}+shift+c"}},
		{Action: actions.ActionCodeBlock, Keys: []string{"{mod}+{alt}+shift+c"}},

		// Message actions
		{Action: actions.ActionSendMessage, Keys: []string{"{mod}+enter"}},
		{Action: actions.ActionScheduleMessage, Keys: []string{"{mod}+shift+enter"}},
		{Action: actions.ActionAttachFile, Keys: []string{"{mod}+u"}},
	}
}

// DefaultInitKeybindings returns the default keybindings for init wizard
func DefaultInitKeybindings() []KeybindingDef {
	return []KeybindingDef{
		{Action: actions.ActionContinue, Keys: []string{"enter"}},
		{Action: actions.ActionUp, Keys: []string{"up"}},
		{Action: actions.ActionDown, Keys: []string{"down"}},
		{Action: actions.ActionToggle, Keys: []string{" ", "space"}},
		{Action: actions.ActionQuit, Keys: []string{"ctrl+c"}},
	}
}

// DefaultUserKeybindings returns the default keybindings for user view
func DefaultUserKeybindings() []KeybindingDef {
	return []KeybindingDef{
		{Action: actions.ActionSetStatus, Keys: []string{"s"}},
		{Action: actions.ActionOpenDM, Keys: []string{"m"}},
	}
}

// DefaultActivityKeybindings returns the default keybindings for activity view
func DefaultActivityKeybindings() []KeybindingDef {
	return []KeybindingDef{
		// Activity-specific actions can be added here
		// For now, it mostly uses global keybindings
	}
}

// DefaultKeybindingsByScope returns all default keybindings organized by scope
func DefaultKeybindingsByScope() map[actions.ActionScope][]KeybindingDef {
	return map[actions.ActionScope][]KeybindingDef{
		actions.ScopeGlobal:   DefaultGlobalKeybindings(),
		actions.ScopeChat:     DefaultChatKeybindings(),
		actions.ScopeMessage:  DefaultMessageKeybindings(),
		actions.ScopeInit:     DefaultInitKeybindings(),
		actions.ScopeUser:     DefaultUserKeybindings(),
		actions.ScopeActivity: DefaultActivityKeybindings(),
	}
}
