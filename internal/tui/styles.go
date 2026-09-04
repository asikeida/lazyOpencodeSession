package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type ThemeConfig struct {
	ActiveBorderColor             []string
	InactiveBorderColor           []string
	SearchingActiveBorderColor    []string
	TitleColor                    []string
	AccentColor                   []string
	OptionsTextColor              []string
	DefaultFgColor                []string
	MutedFgColor                  []string
	MatchColor                    []string
	FocusedSelectedMatchColor     []string
	InactiveSelectedMatchColor    []string
	ErrorColor                    []string
	WarningColor                  []string
	SelectedLineBgColor           []string
	SelectedLineFgColor           []string
	InactiveViewSelectedLineColor []string
	StatusModeColor               []string
	StatusTextColor               []string
	StatusKeyColor                []string
	DetailsHintColor              []string
	PreviewTimestampColor         []string
	MemorySnippetColor            []string
	Modal                         ThemeModalConfig
}

type UIConfig struct {
	BorderStyle     string
	SplitRatio      float64
	TwoPaneMinWidth int
}

type ThemeModalConfig struct {
	BackgroundColor string
	BorderColor     string
	KeyColor        string
	TextColor       string
	MutedColor      string
	IconColor       string
	WarningColor    string
}

type Styles struct {
	Base                  lipgloss.Style
	Muted                 lipgloss.Style
	Accent                lipgloss.Style
	ActiveBorder          lipgloss.Style
	InactiveBorder        lipgloss.Style
	SearchingActiveBorder lipgloss.Style
	Selected              lipgloss.Style
	InactiveSelected      lipgloss.Style
	Match                 lipgloss.Style
	SelectedMatch         lipgloss.Style
	InactiveSelectedMatch lipgloss.Style
	OptionsText           lipgloss.Style
	Panel                 lipgloss.Style
	Title                 lipgloss.Style
	Status                lipgloss.Style
	StatusMode            lipgloss.Style
	StatusText            lipgloss.Style
	StatusKey             lipgloss.Style
	DetailsHint           lipgloss.Style
	PreviewTimestamp      lipgloss.Style
	MemorySnippet         lipgloss.Style
	Warning               lipgloss.Style
	Error                 lipgloss.Style
	Help                  lipgloss.Style
	ModalBG               lipgloss.Style
	ModalBorder           lipgloss.Style
	ModalKey              lipgloss.Style
	ModalText             lipgloss.Style
	ModalMuted            lipgloss.Style
	ModalIcon             lipgloss.Style
	ModalWarn             lipgloss.Style
}

func NewStyles() Styles {
	styles, err := BuildStyles(DefaultThemeConfig())
	if err != nil {
		panic(err)
	}
	return styles
}

func DefaultThemeConfig() ThemeConfig {
	return ThemeConfig{
		ActiveBorderColor:             []string{"green", "bold"},
		InactiveBorderColor:           []string{"green"},
		SearchingActiveBorderColor:    []string{"cyan", "bold"},
		TitleColor:                    []string{"green", "bold"},
		AccentColor:                   []string{"green", "bold"},
		OptionsTextColor:              []string{"blue"},
		DefaultFgColor:                []string{"default"},
		MutedFgColor:                  []string{"250"},
		MatchColor:                    []string{"cyan", "bold", "underline"},
		FocusedSelectedMatchColor:     []string{"229", "underline", "bold"},
		InactiveSelectedMatchColor:    []string{"cyan", "underline"},
		ErrorColor:                    []string{"red", "bold"},
		WarningColor:                  []string{"yellow"},
		SelectedLineBgColor:           []string{"blue"},
		SelectedLineFgColor:           []string{"white", "bold"},
		InactiveViewSelectedLineColor: []string{"bold"},
		StatusModeColor:               []string{"252"},
		StatusTextColor:               []string{"252"},
		StatusKeyColor:                []string{"blue"},
		DetailsHintColor:              []string{"blue"},
		PreviewTimestampColor:         []string{"250"},
		MemorySnippetColor:            []string{"778195"},
		Modal: ThemeModalConfig{
			BackgroundColor: "#202330",
			BorderColor:     "#A8B47A",
			KeyColor:        "#69AFC1",
			TextColor:       "#B9C2D0",
			MutedColor:      "#778195",
			IconColor:       "#6096A3",
			WarningColor:    "#B5A06D",
		},
	}
}

func MergeTheme(base ThemeConfig, override ThemeConfig) ThemeConfig {
	merged := base
	mergeTokens := func(dst *[]string, src []string) {
		if len(src) > 0 {
			*dst = append([]string(nil), src...)
		}
	}
	mergeTokens(&merged.ActiveBorderColor, override.ActiveBorderColor)
	mergeTokens(&merged.InactiveBorderColor, override.InactiveBorderColor)
	mergeTokens(&merged.SearchingActiveBorderColor, override.SearchingActiveBorderColor)
	mergeTokens(&merged.TitleColor, override.TitleColor)
	mergeTokens(&merged.AccentColor, override.AccentColor)
	mergeTokens(&merged.OptionsTextColor, override.OptionsTextColor)
	mergeTokens(&merged.DefaultFgColor, override.DefaultFgColor)
	mergeTokens(&merged.MutedFgColor, override.MutedFgColor)
	mergeTokens(&merged.MatchColor, override.MatchColor)
	mergeTokens(&merged.FocusedSelectedMatchColor, override.FocusedSelectedMatchColor)
	mergeTokens(&merged.InactiveSelectedMatchColor, override.InactiveSelectedMatchColor)
	mergeTokens(&merged.ErrorColor, override.ErrorColor)
	mergeTokens(&merged.WarningColor, override.WarningColor)
	mergeTokens(&merged.SelectedLineBgColor, override.SelectedLineBgColor)
	mergeTokens(&merged.SelectedLineFgColor, override.SelectedLineFgColor)
	mergeTokens(&merged.InactiveViewSelectedLineColor, override.InactiveViewSelectedLineColor)
	mergeTokens(&merged.StatusModeColor, override.StatusModeColor)
	mergeTokens(&merged.StatusTextColor, override.StatusTextColor)
	mergeTokens(&merged.StatusKeyColor, override.StatusKeyColor)
	mergeTokens(&merged.DetailsHintColor, override.DetailsHintColor)
	mergeTokens(&merged.PreviewTimestampColor, override.PreviewTimestampColor)
	mergeTokens(&merged.MemorySnippetColor, override.MemorySnippetColor)
	if override.Modal.BackgroundColor != "" {
		merged.Modal.BackgroundColor = override.Modal.BackgroundColor
	}
	if override.Modal.BorderColor != "" {
		merged.Modal.BorderColor = override.Modal.BorderColor
	}
	if override.Modal.KeyColor != "" {
		merged.Modal.KeyColor = override.Modal.KeyColor
	}
	if override.Modal.TextColor != "" {
		merged.Modal.TextColor = override.Modal.TextColor
	}
	if override.Modal.MutedColor != "" {
		merged.Modal.MutedColor = override.Modal.MutedColor
	}
	if override.Modal.IconColor != "" {
		merged.Modal.IconColor = override.Modal.IconColor
	}
	if override.Modal.WarningColor != "" {
		merged.Modal.WarningColor = override.Modal.WarningColor
	}
	return merged
}

func BuildStyles(theme ThemeConfig) (Styles, error) {
	theme = MergeTheme(DefaultThemeConfig(), theme)
	base, err := applyForegroundTokens(lipgloss.NewStyle(), theme.DefaultFgColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.default_fg_color: %w", err)
	}
	muted, err := applyForegroundTokens(lipgloss.NewStyle(), theme.MutedFgColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.muted_fg_color: %w", err)
	}
	accent, err := applyForegroundTokens(lipgloss.NewStyle(), theme.AccentColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.accent_color: %w", err)
	}
	activeBorder, err := applyBorderTokens(lipgloss.NewStyle(), theme.ActiveBorderColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.active_border_color: %w", err)
	}
	inactiveBorder, err := applyBorderTokens(lipgloss.NewStyle(), theme.InactiveBorderColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.inactive_border_color: %w", err)
	}
	searchingBorder, err := applyBorderTokens(lipgloss.NewStyle(), theme.SearchingActiveBorderColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.searching_active_border_color: %w", err)
	}
	selected, err := applyForegroundTokens(lipgloss.NewStyle(), theme.SelectedLineFgColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.selected_line_fg_color: %w", err)
	}
	selected, err = applyBackgroundTokens(selected, theme.SelectedLineBgColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.selected_line_bg_color: %w", err)
	}
	inactiveSelected, err := applyForegroundTokens(base.Copy(), theme.InactiveViewSelectedLineColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.inactive_view_selected_line_color: %w", err)
	}
	match, err := applyForegroundTokens(lipgloss.NewStyle(), theme.MatchColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.match_color: %w", err)
	}
	title, err := applyForegroundTokens(lipgloss.NewStyle(), theme.TitleColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.title_color: %w", err)
	}
	optionsText, err := applyForegroundTokens(lipgloss.NewStyle(), theme.OptionsTextColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.options_text_color: %w", err)
	}
	statusMode, err := applyForegroundTokens(lipgloss.NewStyle(), theme.StatusModeColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.status_mode_color: %w", err)
	}
	statusText, err := applyForegroundTokens(lipgloss.NewStyle(), theme.StatusTextColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.status_text_color: %w", err)
	}
	statusKey, err := applyForegroundTokens(lipgloss.NewStyle(), theme.StatusKeyColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.status_key_color: %w", err)
	}
	detailsHint, err := applyForegroundTokens(lipgloss.NewStyle(), theme.DetailsHintColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.details_hint_color: %w", err)
	}
	previewTimestamp, err := applyForegroundTokens(lipgloss.NewStyle(), theme.PreviewTimestampColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.preview_timestamp_color: %w", err)
	}
	memorySnippet, err := applyForegroundTokens(lipgloss.NewStyle(), theme.MemorySnippetColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.memory_snippet_color: %w", err)
	}
	warning, err := applyForegroundTokens(lipgloss.NewStyle(), theme.WarningColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.warning_color: %w", err)
	}
	errorStyle, err := applyForegroundTokens(lipgloss.NewStyle(), theme.ErrorColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.error_color: %w", err)
	}
	modalBG := lipgloss.NewStyle()
	if err := validateSingleColorToken(theme.Modal.BackgroundColor); err != nil {
		return Styles{}, fmt.Errorf("theme.modal.background_color: %w", err)
	}
	if err := validateSingleColorToken(theme.Modal.BorderColor); err != nil {
		return Styles{}, fmt.Errorf("theme.modal.border_color: %w", err)
	}
	if err := validateSingleColorToken(theme.Modal.KeyColor); err != nil {
		return Styles{}, fmt.Errorf("theme.modal.key_color: %w", err)
	}
	if err := validateSingleColorToken(theme.Modal.TextColor); err != nil {
		return Styles{}, fmt.Errorf("theme.modal.text_color: %w", err)
	}
	if err := validateSingleColorToken(theme.Modal.MutedColor); err != nil {
		return Styles{}, fmt.Errorf("theme.modal.muted_color: %w", err)
	}
	if err := validateSingleColorToken(theme.Modal.IconColor); err != nil {
		return Styles{}, fmt.Errorf("theme.modal.icon_color: %w", err)
	}
	if err := validateSingleColorToken(theme.Modal.WarningColor); err != nil {
		return Styles{}, fmt.Errorf("theme.modal.warning_color: %w", err)
	}
	if theme.Modal.BackgroundColor != "" {
		modalBG = modalBG.Background(lipgloss.Color(theme.Modal.BackgroundColor))
	}
	modalBorder := lipgloss.NewStyle()
	if theme.Modal.BorderColor != "" {
		modalBorder = modalBorder.Foreground(lipgloss.Color(theme.Modal.BorderColor))
	}
	if theme.Modal.BackgroundColor != "" {
		modalBorder = modalBorder.Background(lipgloss.Color(theme.Modal.BackgroundColor))
	}
	selectedMatch := selected.Copy()
	if color := firstColorToken(theme.FocusedSelectedMatchColor); color != "" {
		selectedMatch = selectedMatch.Foreground(lipgloss.Color(color))
	}
	selectedMatch, err = applyStyleOnlyTokens(selectedMatch, theme.FocusedSelectedMatchColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.focused_selected_match_color: %w", err)
	}
	inactiveSelectedMatch := inactiveSelected.Copy()
	if color := firstColorToken(theme.InactiveSelectedMatchColor); color != "" {
		inactiveSelectedMatch = inactiveSelectedMatch.Foreground(lipgloss.Color(color))
	}
	inactiveSelectedMatch, err = applyStyleOnlyTokens(inactiveSelectedMatch, theme.InactiveSelectedMatchColor)
	if err != nil {
		return Styles{}, fmt.Errorf("theme.inactive_selected_match_color: %w", err)
	}
	styles := Styles{
		Base:                  base,
		Muted:                 muted,
		Accent:                accent,
		ActiveBorder:          activeBorder,
		InactiveBorder:        inactiveBorder,
		SearchingActiveBorder: searchingBorder,
		Selected:              selected,
		InactiveSelected:      inactiveSelected,
		Match:                 match,
		SelectedMatch:         selectedMatch,
		InactiveSelectedMatch: inactiveSelectedMatch,
		OptionsText:           optionsText,
		Panel:                 lipgloss.NewStyle(),
		Title:                 title,
		Status:                lipgloss.NewStyle(),
		StatusMode:            statusMode,
		StatusText:            statusText,
		StatusKey:             statusKey,
		DetailsHint:           detailsHint,
		PreviewTimestamp:      previewTimestamp,
		MemorySnippet:         memorySnippet,
		Warning:               warning,
		Error:                 errorStyle,
		Help:                  lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		ModalBG:               modalBG,
		ModalBorder:           modalBorder,
		ModalKey:              lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Modal.KeyColor)),
		ModalText:             lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Modal.TextColor)),
		ModalMuted:            lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Modal.MutedColor)),
		ModalIcon:             lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Modal.IconColor)).Faint(true),
		ModalWarn:             lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Modal.WarningColor)),
	}
	return styles, nil
}

func DefaultUIConfig() UIConfig {
	return UIConfig{BorderStyle: "rounded", SplitRatio: 0.45, TwoPaneMinWidth: 110}
}

func NormalizeUIConfig(ui UIConfig) UIConfig {
	if ui.BorderStyle == "" {
		ui.BorderStyle = DefaultUIConfig().BorderStyle
	}
	if ui.SplitRatio <= 0 {
		ui.SplitRatio = DefaultUIConfig().SplitRatio
	}
	if ui.TwoPaneMinWidth <= 0 {
		ui.TwoPaneMinWidth = DefaultUIConfig().TwoPaneMinWidth
	}
	return ui
}

func ValidateUIConfig(ui UIConfig) error {
	ui = NormalizeUIConfig(ui)
	switch ui.BorderStyle {
	case "rounded", "single", "double", "hidden", "bold":
	default:
		return fmt.Errorf("ui.border_style must be one of rounded, single, double, hidden, bold")
	}
	if ui.SplitRatio < 0.2 || ui.SplitRatio > 0.8 {
		return fmt.Errorf("ui.split_ratio must be between 0.2 and 0.8")
	}
	if ui.TwoPaneMinWidth < 60 || ui.TwoPaneMinWidth > 400 {
		return fmt.Errorf("ui.two_pane_min_width must be between 60 and 400")
	}
	return nil
}

func applyForegroundTokens(style lipgloss.Style, tokens []string) (lipgloss.Style, error) {
	return applyTokens(style, tokens, func(s lipgloss.Style, color string) lipgloss.Style {
		return s.Foreground(lipgloss.Color(color))
	})
}

func applyBackgroundTokens(style lipgloss.Style, tokens []string) (lipgloss.Style, error) {
	return applyTokens(style, tokens, func(s lipgloss.Style, color string) lipgloss.Style {
		return s.Background(lipgloss.Color(color))
	})
}

func applyBorderTokens(style lipgloss.Style, tokens []string) (lipgloss.Style, error) {
	return applyTokens(style, tokens, func(s lipgloss.Style, color string) lipgloss.Style {
		return s.BorderForeground(lipgloss.Color(color))
	})
}

func applyTokens(style lipgloss.Style, tokens []string, applyColor func(lipgloss.Style, string) lipgloss.Style) (lipgloss.Style, error) {
	color := ""
	for _, token := range tokens {
		token = strings.TrimSpace(strings.ToLower(token))
		if token == "" {
			continue
		}
		if isStyleAttribute(token) {
			style = applyStyleAttribute(style, token)
			continue
		}
		if !isColorToken(token) {
			return lipgloss.NewStyle(), fmt.Errorf("unsupported token %q", token)
		}
		if color != "" {
			return lipgloss.NewStyle(), fmt.Errorf("multiple color tokens in %v", tokens)
		}
		color = token
	}
	if color != "" {
		style = applyColor(style, color)
	}
	return style, nil
}

func applyStyleOnlyTokens(style lipgloss.Style, tokens []string) (lipgloss.Style, error) {
	for _, token := range tokens {
		token = strings.TrimSpace(strings.ToLower(token))
		if token == "" || isColorToken(token) {
			continue
		}
		if !isStyleAttribute(token) {
			return lipgloss.NewStyle(), fmt.Errorf("unsupported token %q", token)
		}
		style = applyStyleAttribute(style, token)
	}
	return style, nil
}

func isStyleAttribute(token string) bool {
	switch token {
	case "bold", "underline", "faint", "reverse", "italic":
		return true
	default:
		return false
	}
}

func applyStyleAttribute(style lipgloss.Style, token string) lipgloss.Style {
	switch token {
	case "bold":
		return style.Bold(true)
	case "underline":
		return style.Underline(true)
	case "faint":
		return style.Faint(true)
	case "reverse":
		return style.Reverse(true)
	case "italic":
		return style.Italic(true)
	default:
		return style
	}
}

func isColorToken(token string) bool {
	if token == "default" || strings.HasPrefix(token, "#") {
		return true
	}
	switch token {
	case "black", "red", "green", "yellow", "blue", "magenta", "cyan", "white":
		return true
	}
	for _, r := range token {
		if r < '0' || r > '9' {
			return false
		}
	}
	return token != ""
}

func firstColorToken(tokens []string) string {
	for _, token := range tokens {
		token = strings.TrimSpace(strings.ToLower(token))
		if isColorToken(token) {
			return token
		}
	}
	return ""
}

func hasToken(tokens []string, target string) bool {
	for _, token := range tokens {
		if strings.EqualFold(strings.TrimSpace(token), target) {
			return true
		}
	}
	return false
}

func validateSingleColorToken(token string) error {
	token = strings.TrimSpace(strings.ToLower(token))
	if token == "" {
		return nil
	}
	if !isColorToken(token) {
		return fmt.Errorf("unsupported color %q", token)
	}
	return nil
}
