package models

import (
	"strings"
	"testing"

	"github.com/slack-go/slack"
)

func TestGetFileIcon(t *testing.T) {
	tests := []struct {
		filetype string
		expected string
	}{
		{"png", "📷"},
		{"jpg", "📷"},
		{"mp4", "📹"},
		{"mp3", "🎵"},
		{"pdf", "📄"},
		{"xlsx", "📊"},
		{"zip", "🗜️"},
		{"go", "💻"},
		{"unknown", "📎"},
	}

	for _, tt := range tests {
		t.Run(tt.filetype, func(t *testing.T) {
			result := getFileIcon(tt.filetype)
			if result != tt.expected {
				t.Errorf("getFileIcon(%q) = %q, expected %q", tt.filetype, result, tt.expected)
			}
		})
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    int
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{2621440, "2.5 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatFileSize(%d) = %q, expected %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestMakeClickableURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		text     string
		expected string
	}{
		{
			name:     "valid URL",
			url:      "https://example.com",
			text:     "Click here",
			expected: "\x1b]8;;https://example.com\x1b\\Click here\x1b]8;;\x1b\\",
		},
		{
			name:     "empty URL",
			url:      "",
			text:     "No link",
			expected: "No link",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeClickableURL(tt.url, tt.text)
			if result != tt.expected {
				t.Errorf("makeClickableURL(%q, %q) = %q, expected %q", tt.url, tt.text, result, tt.expected)
			}
		})
	}
}

func TestFormatFileInfo(t *testing.T) {
	tests := []struct {
		name     string
		file     slack.File
		expected []string // parts that should be in the output
	}{
		{
			name: "image with dimensions",
			file: slack.File{
				Name:      "vacation.jpg",
				Filetype:  "jpg",
				Size:      2500000,
				OriginalW: 1920,
				OriginalH: 1080,
				Permalink: "https://files.slack.com/vacation.jpg",
			},
			expected: []string{"📷", "vacation.jpg", "2.4 MB", "1920x1080", "🔗", "https://files.slack.com/vacation.jpg"},
		},
		{
			name: "document without dimensions",
			file: slack.File{
				Name:      "report.pdf",
				Filetype:  "pdf",
				Size:      512000,
				Permalink: "https://files.slack.com/report.pdf",
			},
			expected: []string{"📄", "report.pdf", "500.0 KB", "🔗", "https://files.slack.com/report.pdf"},
		},
		{
			name: "file with no name uses title",
			file: slack.File{
				Title:     "My File",
				Filetype:  "txt",
				Size:      1024,
				Permalink: "https://files.slack.com/file.txt",
			},
			expected: []string{"📄", "My File", "1.0 KB", "🔗"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileInfo(tt.file)
			for _, part := range tt.expected {
				if !strings.Contains(result, part) {
					t.Errorf("formatFileInfo() result %q does not contain expected part %q", result, part)
				}
			}
		})
	}
}

func TestGetDisplayText(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		expected []string // parts that should be in the output
	}{
		{
			name: "text only",
			message: Message{
				Text: "Hello world",
			},
			expected: []string{"Hello world"},
		},
		{
			name: "file only",
			message: Message{
				Files: []slack.File{
					{
						Name:      "image.png",
						Filetype:  "png",
						Size:      1024,
						Permalink: "https://files.slack.com/image.png",
					},
				},
			},
			expected: []string{"📷", "image.png", "1.0 KB"},
		},
		{
			name: "text and file",
			message: Message{
				Text: "Check this out",
				Files: []slack.File{
					{
						Name:      "doc.pdf",
						Filetype:  "pdf",
						Size:      2048,
						Permalink: "https://files.slack.com/doc.pdf",
					},
				},
			},
			expected: []string{"Check this out", "📄", "doc.pdf"},
		},
		{
			name: "multiple files",
			message: Message{
				Files: []slack.File{
					{
						Name:      "img1.jpg",
						Filetype:  "jpg",
						Size:      1024,
						Permalink: "https://files.slack.com/img1.jpg",
					},
					{
						Name:      "img2.png",
						Filetype:  "png",
						Size:      2048,
						Permalink: "https://files.slack.com/img2.png",
					},
				},
			},
			expected: []string{"📷", "img1.jpg", "img2.png"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.message.GetDisplayText()
			for _, part := range tt.expected {
				if !strings.Contains(result, part) {
					t.Errorf("GetDisplayText() result %q does not contain expected part %q", result, part)
				}
			}
		})
	}
}
