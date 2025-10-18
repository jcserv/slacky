package actions

// Action represents a user action in the application
type Action string

// Global actions - available throughout the app
const (
	ActionQuit       Action = "quit"
	ActionHelp       Action = "help"
	ActionToggleHelp Action = "toggle_help"
	ActionRefresh    Action = "refresh"
	ActionSearch     Action = "search"
	ActionCommandBar Action = "command_bar" // Open command palette
)

// Navigation actions
const (
	ActionNextTab      Action = "next_tab"
	ActionPrevTab      Action = "prev_tab"
	ActionEnterView    Action = "enter_view" // Enter the current tab's view (Space at tab level)
	ActionExitView     Action = "exit_view"  // Exit view and return to tab level (Esc at view level)
	ActionUp           Action = "up"
	ActionDown         Action = "down"
	ActionLeft         Action = "left"
	ActionRight        Action = "right"
	ActionPageUp       Action = "page_up"
	ActionPageDown     Action = "page_down"
	ActionFirstLine    Action = "first_line"
	ActionLastLine     Action = "last_line"
	ActionGoToChat     Action = "go_to_chat"
	ActionGoToActivity Action = "go_to_activity"
	ActionGoToUser     Action = "go_to_user"
	ActionBeginInput   Action = "begin_input" // Jump to input to start typing (deprecated in favor of tab cycling)
)

// Selection actions
const (
	ActionContinue           Action = "continue"
	ActionSelect             Action = "select"
	ActionSelectConversation Action = "select_conversation" // Select a conversation in sidebar
	ActionToggle             Action = "toggle"
	ActionCancel             Action = "cancel"
)

// Message actions - for chat view
const (
	ActionSendMessage     Action = "send_message"
	ActionEditMessage     Action = "edit_message"
	ActionDeleteMessage   Action = "delete_message"
	ActionReplyInThread   Action = "reply_in_thread"
	ActionOpenThread      Action = "open_thread" // Open a thread from a message
	ActionReact           Action = "react"
	ActionSaveMessage     Action = "save_message"
	ActionPinMessage      Action = "pin_message"
	ActionCopyLink        Action = "copy_link"
	ActionMarkUnread      Action = "mark_unread"
	ActionForwardMessage  Action = "forward_message"
	ActionScheduleMessage Action = "schedule_message"
)

// Text formatting actions
const (
	ActionBold          Action = "bold"
	ActionItalic        Action = "italic"
	ActionStrikethrough Action = "strikethrough"
	ActionLink          Action = "link"
	ActionOrderedList   Action = "ordered_list"
	ActionBulletList    Action = "bullet_list"
	ActionBlockquote    Action = "blockquote"
	ActionCode          Action = "code"
	ActionCodeBlock     Action = "code_block"
)

// Channel/DM actions
const (
	ActionStarChannel     Action = "star_channel"
	ActionOpenChannel     Action = "open_channel"
	ActionCopyChannelName Action = "copy_channel_name"
	ActionCopyChannelLink Action = "copy_channel_link"
)

// User actions
const (
	ActionSetStatus    Action = "set_status"
	ActionOpenUserInfo Action = "open_user_info"
	ActionOpenDM       Action = "open_dm"
)

// File actions
const (
	ActionAttachFile   Action = "attach_file"
	ActionOpenFile     Action = "open_file"
	ActionDownloadFile Action = "download_file"
)

// ActionScope represents where an action is available
type ActionScope string

const (
	ScopeGlobal   ActionScope = "global"
	ScopeChat     ActionScope = "chat"
	ScopeActivity ActionScope = "activity"
	ScopeUser     ActionScope = "user"
	ScopeInit     ActionScope = "init"
	ScopeMessage  ActionScope = "message" // When focused on message input
)

// ActionInfo contains metadata about an action
type ActionInfo struct {
	Action      Action
	Scopes      []ActionScope
	Description string // i18n key for description
}

// Registry maps action names to their metadata
var Registry = map[Action]ActionInfo{
	// Global actions
	ActionQuit:       {Action: ActionQuit, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.quit"},
	ActionHelp:       {Action: ActionHelp, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.help"},
	ActionToggleHelp: {Action: ActionToggleHelp, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.toggle_help"},
	ActionRefresh:    {Action: ActionRefresh, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.refresh"},
	ActionSearch:     {Action: ActionSearch, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.search"},
	ActionCommandBar: {Action: ActionCommandBar, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.command_bar"},

	// Navigation actions
	ActionNextTab:      {Action: ActionNextTab, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.next_tab"},
	ActionPrevTab:      {Action: ActionPrevTab, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.prev_tab"},
	ActionEnterView:    {Action: ActionEnterView, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.enter_view"},
	ActionExitView:     {Action: ActionExitView, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.exit_view"},
	ActionUp:           {Action: ActionUp, Scopes: []ActionScope{ScopeGlobal, ScopeInit}, Description: "keys.up"},
	ActionDown:         {Action: ActionDown, Scopes: []ActionScope{ScopeGlobal, ScopeInit}, Description: "keys.down"},
	ActionLeft:         {Action: ActionLeft, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.left"},
	ActionRight:        {Action: ActionRight, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.right"},
	ActionPageUp:       {Action: ActionPageUp, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.page_up"},
	ActionPageDown:     {Action: ActionPageDown, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.page_down"},
	ActionFirstLine:    {Action: ActionFirstLine, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.first_line"},
	ActionLastLine:     {Action: ActionLastLine, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.last_line"},
	ActionGoToChat:     {Action: ActionGoToChat, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.go_to_chat"},
	ActionGoToActivity: {Action: ActionGoToActivity, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.go_to_activity"},
	ActionGoToUser:     {Action: ActionGoToUser, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.go_to_user"},
	ActionBeginInput:   {Action: ActionBeginInput, Scopes: []ActionScope{ScopeChat}, Description: "keys.begin_input"},

	// Selection actions
	ActionContinue:           {Action: ActionContinue, Scopes: []ActionScope{ScopeGlobal, ScopeInit}, Description: "keys.continue"},
	ActionSelect:             {Action: ActionSelect, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.select"},
	ActionSelectConversation: {Action: ActionSelectConversation, Scopes: []ActionScope{ScopeChat}, Description: "keys.select_conversation"},
	ActionToggle:             {Action: ActionToggle, Scopes: []ActionScope{ScopeInit}, Description: "keys.toggle"},
	ActionCancel:             {Action: ActionCancel, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.cancel"},

	// Message actions
	ActionSendMessage:     {Action: ActionSendMessage, Scopes: []ActionScope{ScopeChat, ScopeMessage}, Description: "keys.send_message"},
	ActionEditMessage:     {Action: ActionEditMessage, Scopes: []ActionScope{ScopeChat}, Description: "keys.edit_message"},
	ActionDeleteMessage:   {Action: ActionDeleteMessage, Scopes: []ActionScope{ScopeChat}, Description: "keys.delete_message"},
	ActionReplyInThread:   {Action: ActionReplyInThread, Scopes: []ActionScope{ScopeChat}, Description: "keys.reply_in_thread"},
	ActionOpenThread:      {Action: ActionOpenThread, Scopes: []ActionScope{ScopeChat}, Description: "keys.open_thread"},
	ActionReact:           {Action: ActionReact, Scopes: []ActionScope{ScopeChat}, Description: "keys.react"},
	ActionSaveMessage:     {Action: ActionSaveMessage, Scopes: []ActionScope{ScopeChat}, Description: "keys.save_message"},
	ActionPinMessage:      {Action: ActionPinMessage, Scopes: []ActionScope{ScopeChat}, Description: "keys.pin_message"},
	ActionCopyLink:        {Action: ActionCopyLink, Scopes: []ActionScope{ScopeChat}, Description: "keys.copy_link"},
	ActionMarkUnread:      {Action: ActionMarkUnread, Scopes: []ActionScope{ScopeChat}, Description: "keys.mark_unread"},
	ActionForwardMessage:  {Action: ActionForwardMessage, Scopes: []ActionScope{ScopeChat}, Description: "keys.forward_message"},
	ActionScheduleMessage: {Action: ActionScheduleMessage, Scopes: []ActionScope{ScopeChat, ScopeMessage}, Description: "keys.schedule_message"},

	// Text formatting actions
	ActionBold:          {Action: ActionBold, Scopes: []ActionScope{ScopeMessage}, Description: "keys.bold"},
	ActionItalic:        {Action: ActionItalic, Scopes: []ActionScope{ScopeMessage}, Description: "keys.italic"},
	ActionStrikethrough: {Action: ActionStrikethrough, Scopes: []ActionScope{ScopeMessage}, Description: "keys.strikethrough"},
	ActionLink:          {Action: ActionLink, Scopes: []ActionScope{ScopeMessage}, Description: "keys.link"},
	ActionOrderedList:   {Action: ActionOrderedList, Scopes: []ActionScope{ScopeMessage}, Description: "keys.ordered_list"},
	ActionBulletList:    {Action: ActionBulletList, Scopes: []ActionScope{ScopeMessage}, Description: "keys.bullet_list"},
	ActionBlockquote:    {Action: ActionBlockquote, Scopes: []ActionScope{ScopeMessage}, Description: "keys.blockquote"},
	ActionCode:          {Action: ActionCode, Scopes: []ActionScope{ScopeMessage}, Description: "keys.code"},
	ActionCodeBlock:     {Action: ActionCodeBlock, Scopes: []ActionScope{ScopeMessage}, Description: "keys.code_block"},

	// Channel/DM actions
	ActionStarChannel:     {Action: ActionStarChannel, Scopes: []ActionScope{ScopeChat}, Description: "keys.star_channel"},
	ActionOpenChannel:     {Action: ActionOpenChannel, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.open_channel"},
	ActionCopyChannelName: {Action: ActionCopyChannelName, Scopes: []ActionScope{ScopeChat}, Description: "keys.copy_channel_name"},
	ActionCopyChannelLink: {Action: ActionCopyChannelLink, Scopes: []ActionScope{ScopeChat}, Description: "keys.copy_channel_link"},

	// User actions
	ActionSetStatus:    {Action: ActionSetStatus, Scopes: []ActionScope{ScopeUser}, Description: "keys.set_status"},
	ActionOpenUserInfo: {Action: ActionOpenUserInfo, Scopes: []ActionScope{ScopeGlobal}, Description: "keys.open_user_info"},
	ActionOpenDM:       {Action: ActionOpenDM, Scopes: []ActionScope{ScopeUser}, Description: "keys.open_dm"},

	// File actions
	ActionAttachFile:   {Action: ActionAttachFile, Scopes: []ActionScope{ScopeMessage}, Description: "keys.attach_file"},
	ActionOpenFile:     {Action: ActionOpenFile, Scopes: []ActionScope{ScopeChat}, Description: "keys.open_file"},
	ActionDownloadFile: {Action: ActionDownloadFile, Scopes: []ActionScope{ScopeChat}, Description: "keys.download_file"},
}

// IsValidAction checks if an action exists in the registry
func IsValidAction(action Action) bool {
	_, exists := Registry[action]
	return exists
}

// IsActionInScope checks if an action is available in a given scope
func IsActionInScope(action Action, scope ActionScope) bool {
	info, exists := Registry[action]
	if !exists {
		return false
	}

	for _, s := range info.Scopes {
		if s == scope || s == ScopeGlobal {
			return true
		}
	}
	return false
}
