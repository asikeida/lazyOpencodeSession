package tui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Base     lipgloss.Style
	Muted    lipgloss.Style
	Accent   lipgloss.Style
	Border   lipgloss.Style
	Selected lipgloss.Style
	Match    lipgloss.Style
	Panel    lipgloss.Style
	Title    lipgloss.Style
	Status   lipgloss.Style
	Warning  lipgloss.Style
	Error    lipgloss.Style
	Help     lipgloss.Style
}

func NewStyles() Styles {
	return Styles{
		Base:     lipgloss.NewStyle(),
		Muted:    lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		Accent:   lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true),
		Border:   lipgloss.NewStyle().BorderForeground(lipgloss.Color("2")),
		Panel:    lipgloss.NewStyle(),
		Selected: lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4")).Bold(true),
		Match:    lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true).Underline(true),
		Title:    lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true),
		Status:   lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		Warning:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		Error:    lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),
		Help:     lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
	}
}
