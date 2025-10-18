package keys

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/jcserv/slacky/internal/config"
	"github.com/jcserv/slacky/internal/tui/actions"
)

// ActionKeyMap maps actions to their key bindings
type ActionKeyMap map[actions.Action]key.Binding

// ScopedKeyMap contains keybindings organized by scope
type ScopedKeyMap struct {
	Global   ActionKeyMap
	Chat     ActionKeyMap
	Message  ActionKeyMap
	Init     ActionKeyMap
	User     ActionKeyMap
	Activity ActionKeyMap
}

// LoadKeybindings loads keybindings from config and returns a ScopedKeyMap
// If config is nil or has no keybindings, returns default keybindings
// localizer is optional - if provided, help text will be translated
func LoadKeybindings(cfg *config.Config, localizer *i18n.Localizer) (*ScopedKeyMap, error) {
	scopedMap := &ScopedKeyMap{
		Global:   make(ActionKeyMap),
		Chat:     make(ActionKeyMap),
		Message:  make(ActionKeyMap),
		Init:     make(ActionKeyMap),
		User:     make(ActionKeyMap),
		Activity: make(ActionKeyMap),
	}

	// Start with defaults
	defaults := DefaultKeybindingsByScope()

	// Load global keybindings
	if err := loadScopeKeybindings(scopedMap.Global, defaults[actions.ScopeGlobal], localizer); err != nil {
		return nil, fmt.Errorf("failed to load global keybindings: %w", err)
	}

	// Add number key navigation if enabled in config
	if cfg != nil && cfg.Navigation != nil && cfg.Navigation.NumberKeysGlobal {
		numberKeyBindings := []KeybindingDef{
			{Action: actions.ActionGoToChat, Keys: []string{"1"}},
			{Action: actions.ActionGoToActivity, Keys: []string{"2"}},
			{Action: actions.ActionGoToUser, Keys: []string{"3"}},
		}
		if err := loadScopeKeybindings(scopedMap.Global, numberKeyBindings, localizer); err != nil {
			return nil, fmt.Errorf("failed to load number key bindings: %w", err)
		}
	}

	// Load chat keybindings
	if err := loadScopeKeybindings(scopedMap.Chat, defaults[actions.ScopeChat], localizer); err != nil {
		return nil, fmt.Errorf("failed to load chat keybindings: %w", err)
	}

	// Load message keybindings
	if err := loadScopeKeybindings(scopedMap.Message, defaults[actions.ScopeMessage], localizer); err != nil {
		return nil, fmt.Errorf("failed to load message keybindings: %w", err)
	}

	// Load init keybindings
	if err := loadScopeKeybindings(scopedMap.Init, defaults[actions.ScopeInit], localizer); err != nil {
		return nil, fmt.Errorf("failed to load init keybindings: %w", err)
	}

	// Load user keybindings
	if err := loadScopeKeybindings(scopedMap.User, defaults[actions.ScopeUser], localizer); err != nil {
		return nil, fmt.Errorf("failed to load user keybindings: %w", err)
	}

	// Load activity keybindings
	if err := loadScopeKeybindings(scopedMap.Activity, defaults[actions.ScopeActivity], localizer); err != nil {
		return nil, fmt.Errorf("failed to load activity keybindings: %w", err)
	}

	// Override with user config if present
	if cfg != nil && cfg.Keybindings != nil {
		if err := applyUserKeybindings(scopedMap, cfg.Keybindings); err != nil {
			return nil, fmt.Errorf("failed to apply user keybindings: %w", err)
		}
	}

	return scopedMap, nil
}

// loadScopeKeybindings loads keybindings for a specific scope from defaults
func loadScopeKeybindings(actionMap ActionKeyMap, defs []KeybindingDef, localizer *i18n.Localizer) error {
	for _, def := range defs {
		// Replace OS-specific modifiers in keys
		keys := ReplaceModifiersInSlice(def.Keys)

		// Get action info from registry
		info, exists := actions.Registry[def.Action]
		if !exists {
			return fmt.Errorf("action %s not found in registry", def.Action)
		}

		// Translate description if localizer is provided
		description := info.Description
		if localizer != nil {
			translated, err := localizer.Localize(&i18n.LocalizeConfig{
				MessageID: info.Description,
			})
			if err == nil {
				description = translated
			}
		}

		// Create key binding
		binding := key.NewBinding(
			key.WithKeys(keys...),
			key.WithHelp(FormatKeyForDisplay(keys[0]), description),
		)

		actionMap[def.Action] = binding
	}

	return nil
}

// applyUserKeybindings overlays user-configured keybindings on top of defaults
func applyUserKeybindings(scopedMap *ScopedKeyMap, userBindings *config.Keybindings) error {
	// Apply global keybindings
	if err := applyUserScopeKeybindings(scopedMap.Global, userBindings.Global, actions.ScopeGlobal); err != nil {
		return err
	}

	// Apply chat keybindings
	if err := applyUserScopeKeybindings(scopedMap.Chat, userBindings.Chat, actions.ScopeChat); err != nil {
		return err
	}

	// Apply message keybindings
	if err := applyUserScopeKeybindings(scopedMap.Message, userBindings.Message, actions.ScopeMessage); err != nil {
		return err
	}

	// Apply init keybindings
	if err := applyUserScopeKeybindings(scopedMap.Init, userBindings.Init, actions.ScopeInit); err != nil {
		return err
	}

	// Apply user keybindings
	if err := applyUserScopeKeybindings(scopedMap.User, userBindings.User, actions.ScopeUser); err != nil {
		return err
	}

	// Apply activity keybindings
	if err := applyUserScopeKeybindings(scopedMap.Activity, userBindings.Activity, actions.ScopeActivity); err != nil {
		return err
	}

	return nil
}

// applyUserScopeKeybindings applies user keybindings for a specific scope
func applyUserScopeKeybindings(actionMap ActionKeyMap, userBindings []config.Keybinding, scope actions.ActionScope) error {
	for _, binding := range userBindings {
		action := actions.Action(binding.Action)

		// Validate action exists
		if !actions.IsValidAction(action) {
			slog.Warn("Invalid action in user keybindings, skipping", "action", binding.Action, "scope", scope)
			continue
		}

		// Validate action is valid for this scope
		if !actions.IsActionInScope(action, scope) {
			slog.Warn("Action not valid for scope, skipping", "action", binding.Action, "scope", scope)
			continue
		}

		// Replace OS-specific modifiers
		keyStr := ReplaceModifiers(binding.Key)

		// Get action info for help text
		info := actions.Registry[action]

		// Create key binding
		keyBinding := key.NewBinding(
			key.WithKeys(keyStr),
			key.WithHelp(FormatKeyForDisplay(keyStr), info.Description),
		)

		// Override the default binding
		actionMap[action] = keyBinding

		slog.Debug("Applied user keybinding", "action", action, "key", keyStr, "scope", scope)
	}

	return nil
}

// GetBinding returns the key binding for an action in a specific scope
// Falls back to global scope if not found in the specified scope
func (s *ScopedKeyMap) GetBinding(action actions.Action, scope actions.ActionScope) (key.Binding, bool) {
	var actionMap ActionKeyMap

	switch scope {
	case actions.ScopeChat:
		actionMap = s.Chat
	case actions.ScopeMessage:
		actionMap = s.Message
	case actions.ScopeInit:
		actionMap = s.Init
	case actions.ScopeUser:
		actionMap = s.User
	case actions.ScopeActivity:
		actionMap = s.Activity
	default:
		actionMap = s.Global
	}

	// Try scope-specific binding first
	if binding, ok := actionMap[action]; ok {
		return binding, true
	}

	// Fall back to global
	if binding, ok := s.Global[action]; ok {
		return binding, true
	}

	return key.Binding{}, false
}

// GetAllBindingsForScope returns all keybindings for a scope (including global)
func (s *ScopedKeyMap) GetAllBindingsForScope(scope actions.ActionScope) []key.Binding {
	bindings := []key.Binding{}

	for _, binding := range s.Global {
		bindings = append(bindings, binding)
	}

	var scopeMap ActionKeyMap
	switch scope {
	case actions.ScopeChat:
		scopeMap = s.Chat
	case actions.ScopeMessage:
		scopeMap = s.Message
	case actions.ScopeInit:
		scopeMap = s.Init
	case actions.ScopeUser:
		scopeMap = s.User
	case actions.ScopeActivity:
		scopeMap = s.Activity
	}

	for _, binding := range scopeMap {
		bindings = append(bindings, binding)
	}

	return bindings
}

// GetEssentialBindings returns only the most important keybindings to show in help
// This prevents help text from being overwhelming with too many keybindings
func (s *ScopedKeyMap) GetEssentialBindings(scope actions.ActionScope) []key.Binding {
	// Essential actions that should always be shown
	essentialGlobal := []actions.Action{
		actions.ActionQuit,
		actions.ActionToggleHelp,
		actions.ActionNextTab,
		actions.ActionPrevTab,
		actions.ActionEnterView,
		actions.ActionExitView,
	}

	// Add scope-specific essential actions
	var essentialActions []actions.Action
	switch scope {
	case actions.ScopeChat:
		essentialActions = []actions.Action{
			actions.ActionQuit,
			actions.ActionToggleHelp,
			actions.ActionExitView,
			actions.ActionNextTab,
		}
	case actions.ScopeActivity:
		essentialActions = []actions.Action{
			actions.ActionQuit,
			actions.ActionToggleHelp,
			actions.ActionExitView,
			actions.ActionNextTab,
			actions.ActionPrevTab,
		}
	case actions.ScopeUser:
		essentialActions = []actions.Action{
			actions.ActionQuit,
			actions.ActionToggleHelp,
			actions.ActionExitView,
		}
	default:
		essentialActions = essentialGlobal
	}

	// Collect bindings for essential actions
	bindings := []key.Binding{}
	for _, action := range essentialActions {
		if binding, ok := s.GetBinding(action, scope); ok {
			bindings = append(bindings, binding)
		}
	}

	return bindings
}

// MatchesAction checks if a key message matches an action in a scope
func (s *ScopedKeyMap) MatchesAction(msg interface{}, action actions.Action, scope actions.ActionScope) bool {
	binding, ok := s.GetBinding(action, scope)
	if !ok {
		return false
	}

	// msg should be a tea.KeyMsg, which implements fmt.Stringer
	return key.Matches(msg.(fmt.Stringer), binding)
}
