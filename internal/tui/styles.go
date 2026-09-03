package tui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Base        lipgloss.Style
	Muted       lipgloss.Style
	Accent      lipgloss.Style
	Border      lipgloss.Style
	Selected    lipgloss.Style
	Match       lipgloss.Style
	Panel       lipgloss.Style
	Title       lipgloss.Style
	Status      lipgloss.Style
	Warning     lipgloss.Style
	Error       lipgloss.Style
	Help        lipgloss.Style
	ModalBG     lipgloss.Style
	ModalBorder lipgloss.Style
	ModalKey    lipgloss.Style
	ModalText   lipgloss.Style
	ModalMuted  lipgloss.Style
	ModalIcon   lipgloss.Style
	ModalWarn   lipgloss.Style
}

func NewStyles() Styles {
	return Styles{
		Base:        lipgloss.NewStyle(),
		Muted:       lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		Accent:      lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true),
		Border:      lipgloss.NewStyle().BorderForeground(lipgloss.Color("2")),
		Panel:       lipgloss.NewStyle(),
		Selected:    lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4")).Bold(true),
		Match:       lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true).Underline(true),
		Title:       lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true),
		Status:      lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		Warning:     lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		Error:       lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),
		Help:        lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		ModalBG:     lipgloss.NewStyle().Background(lipgloss.Color("#202330")),
		ModalBorder: lipgloss.NewStyle().Foreground(lipgloss.Color("#A8B47A")).Background(lipgloss.Color("#202330")),
		ModalKey:    lipgloss.NewStyle().Foreground(lipgloss.Color("#69AFC1")),
		ModalText:   lipgloss.NewStyle().Foreground(lipgloss.Color("#B9C2D0")),
		ModalMuted:  lipgloss.NewStyle().Foreground(lipgloss.Color("#778195")),
		ModalIcon:   lipgloss.NewStyle().Foreground(lipgloss.Color("#6096A3")).Faint(true),
		ModalWarn:   lipgloss.NewStyle().Foreground(lipgloss.Color("#B5A06D")),
	}
}
