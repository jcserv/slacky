package styles_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/jcserv/slacky/internal/tui/styles"
	"github.com/stretchr/testify/assert"
)

func TestColorsDefined(t *testing.T) {
	t.Parallel()

	// Test that all color constants are defined and not empty
	colors := []lipgloss.Color{
		styles.Primary,
		styles.Secondary,
		styles.Tertiary,
		styles.Quarternary,
		styles.ColourSuccess,
		styles.ColourError,
		styles.ColourInfo,
		styles.ColourWarning,
		styles.ColourForeground,
		styles.ColourBackground,
		styles.ColourDim,
		styles.ColourSubtle,
	}

	for _, color := range colors {
		assert.NotEmpty(t, color, "color should not be empty")
		// Verify it's a valid color string (starts with #)
		assert.Contains(t, string(color), "#", "color should be a hex color")
	}
}

func TestStylesDefined(t *testing.T) {
	t.Parallel()

	// Test that all style variables are defined
	// We can't test exact values, but we can verify they exist and render
	testStyles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"Title", styles.Title},
		{"Subtitle", styles.Subtitle},
		{"Label", styles.Label},
		{"LabelActive", styles.LabelActive},
		{"Success", styles.Success},
		{"Error", styles.Error},
		{"Info", styles.Info},
		{"Warning", styles.Warning},
		{"Dim", styles.Dim},
		{"Subtle", styles.Subtle},
		{"Highlight", styles.Highlight},
		{"Help", styles.Help},
		{"Completed", styles.Completed},
		{"Logo", styles.Logo},
		{"Border", styles.Border},
		{"BorderActive", styles.BorderActive},
	}

	for _, tt := range testStyles {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test that the style can render text
			rendered := tt.style.Render("test")
			assert.NotEmpty(t, rendered)
			assert.Contains(t, rendered, "test")
		})
	}
}

func TestDefaultTheme(t *testing.T) {
	t.Parallel()

	theme := styles.DefaultTheme()

	// Test that theme has all required fields
	assert.NotNil(t, theme.Title)
	assert.NotNil(t, theme.Subtitle)
	assert.NotNil(t, theme.Label)
	assert.NotNil(t, theme.Success)
	assert.NotNil(t, theme.Error)
	assert.NotNil(t, theme.Info)
	assert.NotNil(t, theme.Warning)
	assert.NotNil(t, theme.Dim)
	assert.NotNil(t, theme.Help)
	assert.NotNil(t, theme.Completed)
	assert.NotNil(t, theme.Highlight)
	assert.NotNil(t, theme.Border)
}

func TestStyleConsistency(t *testing.T) {
	t.Parallel()

	// Test that styles render consistently
	text := "Test Text"

	// Title should be bold
	titleRendered := styles.Title.Render(text)
	assert.Contains(t, titleRendered, text)

	// Success should be bold
	successRendered := styles.Success.Render(text)
	assert.Contains(t, successRendered, text)

	// Error should be bold
	errorRendered := styles.Error.Render(text)
	assert.Contains(t, errorRendered, text)

	// All should produce different output due to color differences
	// (though we can't test exact ANSI codes without more complex parsing)
}

func TestBorderStyles(t *testing.T) {
	t.Parallel()

	// Test border styles
	text := "bordered"

	borderRendered := styles.Border.Render(text)
	assert.Contains(t, borderRendered, text)

	borderActiveRendered := styles.BorderActive.Render(text)
	assert.Contains(t, borderActiveRendered, text)

	// Both styles should work (they only differ in color)
	// We can't easily test color differences without parsing ANSI codes
}
