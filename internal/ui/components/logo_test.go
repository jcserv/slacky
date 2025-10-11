package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRender(t *testing.T) {
	opts := Opts{
		Version:      "v0.1.0-dev",
		Width:        100,
		FillColor:    lipgloss.Color("#7D56F4"),
		VersionColor: lipgloss.Color("#7D56F4"),
	}

	result := Render(opts)

	if result == "" {
		t.Error("Render() returned empty string")
	}

	if !strings.Contains(result, "v0.1.0-dev") {
		t.Error("Render() does not contain version string")
	}

	if !strings.Contains(result, "███") {
		t.Error("Render() does not contain ASCII art")
	}

	if !strings.Contains(result, fillCharacter) {
		t.Error("Render() does not contain diagonal slashes")
	}

	lines := strings.Split(result, "\n")
	expectedLines := 7 // 1 version + 6 logo lines
	if len(lines) != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, len(lines))
	}
}

func TestSmallRender(t *testing.T) {
	result := SmallRender("v0.1.0", 50)

	// Check that result is not empty
	if result == "" {
		t.Error("SmallRender() returned empty string")
	}

	// Check that it contains the version
	if !strings.Contains(result, "v0.1.0") {
		t.Error("SmallRender() does not contain version string")
	}

	// Check that it contains "Slacky"
	if !strings.Contains(result, "Slacky") {
		t.Error("SmallRender() does not contain 'Slacky'")
	}
}

func TestHeight(t *testing.T) {
	h := Height()
	// Should be 7 (1 version line + 6 logo lines)
	if h != 7 {
		t.Errorf("Expected height 7, got %d", h)
	}
}

func TestLogoWidth(t *testing.T) {
	w := LogoWidth()
	// The ASCII art is 50 characters wide
	if w != 50 {
		t.Errorf("Expected logo width 50, got %d", w)
	}
}

func TestMinWidth(t *testing.T) {
	mw := MinWidth()
	// MinWidth should be LogoWidth + 6 (4 left slashes + 2 spaces)
	expectedMinWidth := LogoWidth() + 6
	if mw != expectedMinWidth {
		t.Errorf("Expected min width %d, got %d", expectedMinWidth, mw)
	}
}
