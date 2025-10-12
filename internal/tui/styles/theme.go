package styles

import "github.com/charmbracelet/lipgloss"

var (
	Primary     = lipgloss.Color("#82aaff") // Title, primary actions
	Secondary   = lipgloss.Color("#5c8fa6") // Secondary elements
	Tertiary    = lipgloss.Color("#d07e60") // Additional elements
	Quarternary = lipgloss.Color("#c099ff")

	ColourSuccess = lipgloss.Color("#86e1b3")
	ColourError   = lipgloss.Color("#ff757f")
	ColourInfo    = lipgloss.Color("#82aaff")
	ColourWarning = lipgloss.Color("#d07e60")

	// grayscale
	ColourForeground = lipgloss.Color("#ffffff") // #c8d3f5
	ColourBackground = lipgloss.Color("#222436")
	ColourDim        = lipgloss.Color("#828bb8")
	ColourSubtle     = lipgloss.Color("#444a73")
)

var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary)

	Subtitle = lipgloss.NewStyle().
			Foreground(ColourForeground)

	Label = lipgloss.NewStyle().
		Bold(true).
		Foreground(Tertiary)

	LabelActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)

	Success = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColourSuccess)

	Error = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColourError)

	Info = lipgloss.NewStyle().
		Foreground(ColourInfo)

	Warning = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColourWarning)

	Dim = lipgloss.NewStyle().
		Foreground(ColourDim)

	Subtle = lipgloss.NewStyle().
		Foreground(ColourSubtle)

	Highlight = lipgloss.NewStyle().
			Foreground(Tertiary)

	Help = lipgloss.NewStyle().
		Foreground(ColourDim)

	Completed = lipgloss.NewStyle().
			Foreground(ColourSuccess)

	Logo = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary)

	// Tabs
	Tab = lipgloss.NewStyle().
		Foreground(ColourDim).
		Padding(0, 2)

	ActiveTab = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		Background(ColourSubtle).
		Padding(0, 2)

	TabSeparator = lipgloss.NewStyle().
		Foreground(ColourSubtle)

	TabsRow = lipgloss.NewStyle().
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColourSubtle)

	// Status Bar
	StatusBar = lipgloss.NewStyle().
		Foreground(ColourForeground).
		Background(ColourSubtle).
		Padding(0, 1).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColourSubtle)
)

var (
	Border = lipgloss.NewStyle().
		Foreground(ColourSubtle)

	BorderActive = lipgloss.NewStyle().
			Foreground(Primary)
)

// GetTheme returns a Theme struct with all styles
type Theme struct {
	Title        lipgloss.Style
	Subtitle     lipgloss.Style
	Label        lipgloss.Style
	Success      lipgloss.Style
	Error        lipgloss.Style
	Info         lipgloss.Style
	Warning      lipgloss.Style
	Dim          lipgloss.Style
	Help         lipgloss.Style
	Completed    lipgloss.Style
	Highlight    lipgloss.Style
	Border       lipgloss.Style
	Tab          lipgloss.Style
	ActiveTab    lipgloss.Style
	TabSeparator lipgloss.Style
	TabsRow      lipgloss.Style
	StatusBar    lipgloss.Style
}

func DefaultTheme() Theme {
	return Theme{
		Title:        Title,
		Subtitle:     Subtitle,
		Label:        Label,
		Success:      Success,
		Error:        Error,
		Info:         Info,
		Warning:      Warning,
		Dim:          Dim,
		Help:         Help,
		Completed:    Completed,
		Highlight:    Highlight,
		Border:       Border,
		Tab:          Tab,
		ActiveTab:    ActiveTab,
		TabSeparator: TabSeparator,
		TabsRow:      TabsRow,
		StatusBar:    StatusBar,
	}
}
