package i18n

import (
	"embed"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.toml
var localeFS embed.FS

var bundle *i18n.Bundle

// Init initializes the i18n bundle with embedded locale files
func Init() error {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// Load all locale files from embedded FS
	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := "locales/" + entry.Name()
		if _, err := bundle.LoadMessageFileFS(localeFS, filePath); err != nil {
			return fmt.Errorf("failed to load message file %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// NewLocalizer creates a new localizer for the given language preferences.
// It accepts language tags in order of preference (e.g., "es", "en").
// If no languages are provided, it defaults to English.
//
// This function panics if the i18n bundle is not initialized. Since i18n.Init()
// is called at application startup (in main.go), this panic indicates a programming
// error rather than a runtime condition. Use this in model initialization where
// Init() has already been called.
func NewLocalizer(langs ...string) *i18n.Localizer {
	if bundle == nil {
		panic("i18n bundle not initialized - call i18n.Init() first")
	}

	if len(langs) == 0 {
		langs = []string{"en"}
	}

	return i18n.NewLocalizer(bundle, langs...)
}

// DetectLocale attempts to detect the user's locale from the environment
// Returns the detected language tag (e.g., "en", "es") or "en" as fallback
func DetectLocale() string {
	// Try to detect from environment variables
	// This works with the existing mattn/go-localereader dependency
	if locale := detectFromEnv(); locale != "" {
		return locale
	}

	return "en"
}

// detectFromEnv detects locale from environment variables
func detectFromEnv() string {
	// Common environment variables for locale
	for _, envVar := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if locale := getEnvLocale(envVar); locale != "" {
			return locale
		}
	}
	return ""
}

// getEnvLocale parses a locale from environment variable value
// Converts values like "en_US.UTF-8" to "en"
func getEnvLocale(envVar string) string {
	value := os.Getenv(envVar)
	if value == "" {
		return ""
	}

	// Parse language tag from locale string
	// e.g., "en_US.UTF-8" -> "en"
	tag, err := language.Parse(value)
	if err != nil {
		return ""
	}

	base, _ := tag.Base()
	return base.String()
}
