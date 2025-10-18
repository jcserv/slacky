package messages

import (
	"strings"
	"testing"
)

func TestWrapMessageText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected []string // lines that should be present
	}{
		{
			name:  "simple text wrapping",
			text:  "This is a long message that should be wrapped when it exceeds the width",
			width: 30,
			expected: []string{
				"This is a long message that",
				"should be wrapped when it",
				"exceeds the width",
			},
		},
		{
			name:  "URL preservation",
			text:  "Check this out\n   🔗 https://files.slack.com/files-pri/T028GPCMEHL-F09MBKN6XBN/screenshot_2025-10-17_at_9.39.41___pm.png",
			width: 50,
			expected: []string{
				"Check this out",
				"   🔗 https://files.slack.com/files-pri/T028GPCMEHL-F09MBKN6XBN/screenshot_2025-10-17_at_9.39.41___pm.png",
			},
		},
		{
			name:  "multiple lines with URLs",
			text:  "First line\n   🔗 https://example.com/very/long/url/that/would/normally/wrap\nSecond line",
			width: 30,
			expected: []string{
				"First line",
				"   🔗 https://example.com/very/long/url/that/would/normally/wrap",
				"Second line",
			},
		},
		{
			name:  "URL without emoji marker",
			text:  "https://example.com/test",
			width: 10,
			expected: []string{
				"https://example.com/test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapMessageText(tt.text, tt.width)
			resultLines := strings.Split(result, "\n")

			if len(resultLines) != len(tt.expected) {
				t.Errorf("wrapMessageText() returned %d lines, expected %d\nGot: %v\nExpected: %v",
					len(resultLines), len(tt.expected), resultLines, tt.expected)
				return
			}

			for i, expectedLine := range tt.expected {
				if resultLines[i] != expectedLine {
					t.Errorf("Line %d:\nGot:      %q\nExpected: %q", i, resultLines[i], expectedLine)
				}
			}
		})
	}
}
