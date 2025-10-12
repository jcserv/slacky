package keys

import (
	"runtime"
	"strings"
)

const (
	// ModifierPlaceholder is the placeholder used in key definitions
	// that will be replaced with the OS-specific modifier
	ModifierPlaceholder = "{mod}"
)

// PrimaryModifier returns the primary modifier key for the current OS
// - macOS: "cmd"
// - Windows/Linux: "ctrl"
func PrimaryModifier() string {
	if runtime.GOOS == "darwin" {
		return "cmd"
	}
	return "ctrl"
}

// SecondaryModifier returns the secondary modifier key for the current OS
// - macOS: "option" (alt)
// - Windows/Linux: "alt"
func SecondaryModifier() string {
	if runtime.GOOS == "darwin" {
		return "option"
	}
	return "alt"
}

// ReplaceModifiers replaces modifier placeholders with OS-specific modifiers
// Supports:
// - {mod} → cmd (macOS) or ctrl (Windows/Linux)
// - {alt} → option (macOS) or alt (Windows/Linux)
func ReplaceModifiers(key string) string {
	result := key
	result = strings.ReplaceAll(result, ModifierPlaceholder, PrimaryModifier())
	result = strings.ReplaceAll(result, "{alt}", SecondaryModifier())
	return result
}

// ReplaceModifiersInSlice replaces modifiers in a slice of key strings
func ReplaceModifiersInSlice(keys []string) []string {
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = ReplaceModifiers(key)
	}
	return result
}

// FormatKeyForDisplay formats a key string for display in help text
// Replaces modifier placeholders and formats nicely
func FormatKeyForDisplay(key string) string {
	display := ReplaceModifiers(key)

	// Capitalize modifier keys for display
	if runtime.GOOS == "darwin" {
		display = strings.ReplaceAll(display, "cmd", "Cmd")
		display = strings.ReplaceAll(display, "option", "Option")
	} else {
		display = strings.ReplaceAll(display, "ctrl", "Ctrl")
		display = strings.ReplaceAll(display, "alt", "Alt")
	}

	// Capitalize shift
	display = strings.ReplaceAll(display, "shift", "Shift")

	return display
}
