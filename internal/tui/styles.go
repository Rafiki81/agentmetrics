package tui

import (
	"github.com/Rafiki81/libagentmetrics/config"
	"github.com/charmbracelet/lipgloss"
)

// Theme holds resolved lipgloss colors from config
type Theme struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Success   lipgloss.Color
	Warning   lipgloss.Color
	Danger    lipgloss.Color
	Muted     lipgloss.Color
	Bg        lipgloss.Color
	BgAlt     lipgloss.Color
	Fg        lipgloss.Color
	Border    lipgloss.Color
}

// Styles holds all pre-built lipgloss styles derived from the theme
type Styles struct {
	Theme             Theme
	Title             lipgloss.Style
	Header            lipgloss.Style
	AgentCard         lipgloss.Style
	AgentCardSelected lipgloss.Style
	AgentName         lipgloss.Style
	StatusRunning     lipgloss.Style
	StatusIdle        lipgloss.Style
	StatusStopped     lipgloss.Style
	MetricLabel       lipgloss.Style
	MetricValue       lipgloss.Style
	DetailPanel       lipgloss.Style
	FileCreate        lipgloss.Style
	FileModify        lipgloss.Style
	FileDelete        lipgloss.Style
	Help              lipgloss.Style
	BarFull           lipgloss.Style
	BarEmpty          lipgloss.Style
	Empty             lipgloss.Style
	Logo              lipgloss.Style
	// Token & cost
	TokenLabel  lipgloss.Style
	TokenValue  lipgloss.Style
	TokenSource lipgloss.Style
	Cost        lipgloss.Style
	// Alert
	AlertWarn lipgloss.Style
	AlertCrit lipgloss.Style
	AlertInfo lipgloss.Style
	// Git & session
	Git     lipgloss.Style
	Session lipgloss.Style
	// Security
	SecurityCritical lipgloss.Style
	SecurityHigh     lipgloss.Style
	SecurityMedium   lipgloss.Style
	SecurityLow      lipgloss.Style
	SecurityBanner   lipgloss.Style
}

// NewStyles creates styles from a theme config
func NewStyles(tc config.ThemeConfig) *Styles {
	t := Theme{
		Primary:   lipgloss.Color(tc.Primary),
		Secondary: lipgloss.Color(tc.Secondary),
		Success:   lipgloss.Color(tc.Success),
		Warning:   lipgloss.Color(tc.Warning),
		Danger:    lipgloss.Color(tc.Danger),
		Muted:     lipgloss.Color(tc.Muted),
		Bg:        lipgloss.Color(tc.Background),
		BgAlt:     lipgloss.Color(tc.BackgroundAlt),
		Fg:        lipgloss.Color(tc.Foreground),
		Border:    lipgloss.Color(tc.Border),
	}

	return &Styles{
		Theme: t,

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Primary).
			Background(t.Bg).
			Padding(0, 1),

		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Secondary).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(t.Border).
			MarginBottom(1),

		AgentCard: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Border).
			Padding(0, 1).
			MarginBottom(1),

		AgentCardSelected: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary).
			Padding(0, 1).
			MarginBottom(1),

		AgentName: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Fg),

		StatusRunning: lipgloss.NewStyle().
			Foreground(t.Success).
			Bold(true),

		StatusIdle: lipgloss.NewStyle().
			Foreground(t.Warning).
			Bold(true),

		StatusStopped: lipgloss.NewStyle().
			Foreground(t.Danger).
			Bold(true),

		MetricLabel: lipgloss.NewStyle().
			Foreground(t.Muted),

		MetricValue: lipgloss.NewStyle().
			Foreground(t.Fg).
			Bold(true),

		DetailPanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Border).
			Padding(1, 2),

		FileCreate: lipgloss.NewStyle().Foreground(t.Success),
		FileModify: lipgloss.NewStyle().Foreground(t.Warning),
		FileDelete: lipgloss.NewStyle().Foreground(t.Danger),

		Help: lipgloss.NewStyle().
			Foreground(t.Muted).
			MarginTop(1),

		BarFull:  lipgloss.NewStyle().Foreground(t.Success),
		BarEmpty: lipgloss.NewStyle().Foreground(t.Border),

		Empty: lipgloss.NewStyle().
			Foreground(t.Muted).
			Italic(true).
			Padding(2, 0).
			Align(lipgloss.Center),

		Logo: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true),

		// Token & cost styles
		TokenLabel:  lipgloss.NewStyle().Foreground(t.Warning),
		TokenValue:  lipgloss.NewStyle().Foreground(t.Secondary).Bold(true),
		TokenSource: lipgloss.NewStyle().Foreground(t.Muted).Italic(true),
		Cost:        lipgloss.NewStyle().Foreground(t.Success).Bold(true),

		// Alert styles
		AlertWarn: lipgloss.NewStyle().Foreground(t.Warning).Bold(true),
		AlertCrit: lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
		AlertInfo: lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")),

		// Git & session
		Git:     lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA")),
		Session: lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399")),

		// Security styles
		SecurityCritical: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Blink(true),
		SecurityHigh: lipgloss.NewStyle().
			Foreground(t.Danger).
			Bold(true),
		SecurityMedium: lipgloss.NewStyle().
			Foreground(t.Warning).
			Bold(true),
		SecurityLow: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")),
		SecurityBanner: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000")).
			Background(lipgloss.Color("#2D0000")).
			Padding(0, 1),
	}
}

// DefaultStyles returns styles with the default Tokyo Night theme
func DefaultStyles() *Styles {
	return NewStyles(config.DefaultConfig().Theme)
}

// StatusStyle returns the appropriate style for a status string
func (s *Styles) StatusStyle(status string) lipgloss.Style {
	switch status {
	case "RUNNING":
		return s.StatusRunning
	case "IDLE":
		return s.StatusIdle
	case "STOPPED":
		return s.StatusStopped
	default:
		return lipgloss.NewStyle().Foreground(s.Theme.Muted)
	}
}

// FileOpStyle returns the style for a file operation type
func (s *Styles) FileOpStyle(op string) lipgloss.Style {
	switch op {
	case "CREATE":
		return s.FileCreate
	case "MODIFY":
		return s.FileModify
	case "DELETE":
		return s.FileDelete
	default:
		return lipgloss.NewStyle().Foreground(s.Theme.Muted)
	}
}

// RenderBar renders a progress bar
func (s *Styles) RenderBar(value, max float64, width int) string {
	if max <= 0 {
		max = 100
	}
	filled := int(value / max * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += s.BarFull.Render("#")
		} else {
			bar += s.BarEmpty.Render("-")
		}
	}
	return bar
}
