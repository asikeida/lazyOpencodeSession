package tui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Base     lipgloss.Style
	Muted    lipgloss.Style
	Accent   lipgloss.Style
	Border   lipgloss.Style
	Selected lipgloss.Style
	Match    lipgloss.Style
	Title    lipgloss.Style
	Status   lipgloss.Style
	Warning  lipgloss.Style
	Error    lipgloss.Style
	Help     lipgloss.Style
}

func NewStyles(theme string) Styles {
	if theme == "dark" {
		return Styles{
			Base:     lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
			Muted:    lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
			Accent:   lipgloss.NewStyle().Foreground(lipgloss.Color("110")).Bold(true),
			Border:   lipgloss.NewStyle().BorderForeground(lipgloss.Color("238")),
			Selected: lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("24")).Bold(true),
			Match:    lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true).Underline(true),
			Title:    lipgloss.NewStyle().Foreground(lipgloss.Color("110")).Bold(true),
			Status:   lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Background(lipgloss.Color("236")),
			Warning:  lipgloss.NewStyle().Foreground(lipgloss.Color("179")),
			Error:    lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true),
			Help:     lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Background(lipgloss.Color("236")),
		}
	}
	return Styles{
		Base:     lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		Muted:    lipgloss.NewStyle().Foreground(lipgloss.Color("246")),
		Accent:   lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true),
		Border:   lipgloss.NewStyle().BorderForeground(lipgloss.Color("240")),
		Selected: lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("25")).Bold(true),
		Match:    lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true).Underline(true),
		Title:    lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true),
		Status:   lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Background(lipgloss.Color("238")),
		Warning:  lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		Error:    lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true),
		Help:     lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Background(lipgloss.Color("238")),
	}
}
