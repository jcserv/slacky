package testutil

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ExecCmd executes a Bubble Tea command synchronously, running it until
// all messages have been processed. This is essential for testing TEA
// applications where commands are normally async.
//
// Usage:
//
//	m := NewModel()
//	ExecCmd(m, m.Init())
//	assert.Equal(t, expectedState, m.state)
func ExecCmd(m tea.Model, cmd tea.Cmd) tea.Model {
	for cmd != nil {
		msg := cmd()
		var newCmd tea.Cmd
		m, newCmd = m.Update(msg)
		cmd = newCmd
	}
	return m
}

// TempConfigDir creates a temporary directory for config testing.
// It returns the temp directory path and a cleanup function.
// The cleanup function should be called with defer.
//
// Usage:
//
//	tempDir, cleanup := testutil.TempConfigDir(t)
//	defer cleanup()
func TempConfigDir(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "slacky-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// WriteTestConfig writes a test config file to a temporary location.
// Returns the path to the config file and a cleanup function.
//
// Usage:
//
//	configPath, cleanup := testutil.WriteTestConfig(t, "bot_token: xoxb-test\nsocket_token: xapp-test")
//	defer cleanup()
func WriteTestConfig(t *testing.T, content string) (string, func()) {
	t.Helper()

	tempDir, cleanup := TempConfigDir(t)

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		cleanup()
		t.Fatalf("failed to write test config: %v", err)
	}

	return configPath, cleanup
}

// SetEnv sets an environment variable for the duration of a test.
// Returns a cleanup function that should be called with defer.
//
// Usage:
//
//	cleanup := testutil.SetEnv(t, "HOME", tempDir)
//	defer cleanup()
func SetEnv(t *testing.T, key, value string) func() {
	t.Helper()

	oldValue, existed := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set env var %s: %v", key, err)
	}

	return func() {
		if existed {
			os.Setenv(key, oldValue)
		} else {
			os.Unsetenv(key)
		}
	}
}
