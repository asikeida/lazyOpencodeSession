package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	lazytui "github.com/asikeida/lazyOpencodeSession/internal/tui"
	"github.com/pelletier/go-toml/v2"
)

const DefaultConfigTOML = `# lazyocs configuration
# OpenCode SQLite database path. Default: ~/.local/share/opencode/opencode.db.
# Leave empty to use the default path.
# OpenCode SQLite 数据库路径。默认值：~/.local/share/opencode/opencode.db。
# 留空表示使用默认路径。
db = ""
# UI language: auto, en, or zh-CN. Default: auto.
# 界面语言：auto、en 或 zh-CN。默认值：auto。
language = "auto"
# Maximum number of sessions loaded into the list. Default: 500.
# 会话列表最多加载的数量。默认值：500。
limit = 500
# OpenCode executable or command path used by Enter. Default: opencode.
# 按 Enter 恢复会话时使用的 OpenCode 命令或路径。默认值：opencode。
opencode = "opencode"
# Database write protection. Default: false.
# 数据库写保护。默认值：false。使用 --read-only 或设为 true 可禁止修改标题。
read_only = false

# Optional built-in or local theme preset name. Example: lazygit-classic, moss, iris-night.
# 可选的内置或本地主题预设名称，例如 lazygit-classic、moss、iris-night。
theme_name = ""

# Optional external theme file. Relative paths are resolved from this config file.
# 可选的外部主题文件。相对路径将基于当前配置文件解析。
theme_file = ""

[search]
# User-message memory window in days. Use 0 to disable. Default: 7.
# 用户消息记忆搜索的时间范围（天）。设为 0 可关闭。默认值：7。
recent_days = 7

[preview]
# Number of recent user messages shown in the right-side preview. Default: 5.
# 右侧预览显示的最近用户消息条数。默认值：5。
recent_messages_limit = 5

[ui]
# Panel border style: rounded, single, double, hidden, or bold. Default: rounded.
# 面板边框样式：rounded、single、double、hidden 或 bold。默认值：rounded。
border_style = "rounded"
# Left panel width ratio in two-pane mode. Default: 0.45.
# 双栏模式下左侧面板宽度比例。默认值：0.45。
split_ratio = 0.45
# Minimum terminal width for two-pane mode. Default: 110.
# 进入双栏模式所需的最小终端宽度。默认值：110。
two_pane_min_width = 110

# Theme configuration. Uncomment and adjust any values you want to override.
# 主题配置。只需取消注释并修改你想覆盖的值。
#
# [theme]
# active_border_color = ["green", "bold"]
# inactive_border_color = ["green"]
# searching_active_border_color = ["cyan", "bold"]
# title_color = ["green", "bold"]
# accent_color = ["green", "bold"]
# options_text_color = ["blue"]
# default_fg_color = ["default"]
# muted_fg_color = ["250"]
# match_color = ["cyan", "bold", "underline"]
# focused_selected_match_color = ["229", "underline", "bold"]
# inactive_selected_match_color = ["cyan", "underline"]
# error_color = ["red", "bold"]
# warning_color = ["yellow"]
# selected_line_bg_color = ["blue"]
# selected_line_fg_color = ["white", "bold"]
# inactive_view_selected_line_color = ["bold"]
# status_mode_color = ["252"]
# status_text_color = ["252"]
# status_key_color = ["blue"]
# details_hint_color = ["blue"]
# preview_timestamp_color = ["250"]
# memory_snippet_color = ["250"]
#
# [theme.modal]
# background_color = "#202330"
# border_color = "#A8B47A"
# key_color = "#69AFC1"
# text_color = "#B9C2D0"
# muted_color = "#778195"
# icon_color = "#6096A3"
# warning_color = "#B5A06D"

[details.fields]
# Session title. Default: true.
# 会话标题。默认值：true。
title = true
# Session ID. Default: true.
# 会话 ID。默认值：true。
session = true
# Project ID. Default: true.
# 项目 ID。默认值：true。
project = true
# Session working directory. Default: true.
# 会话工作目录。默认值：true。
directory = true
# Whether the session directory currently exists. Default: true.
# 会话目录当前是否存在。默认值：true。
path_status = true
# Number of messages. Loaded lazily when enabled. Default: false.
# 消息数量。启用后按需加载。默认值：false。
message_count = false
# Number of parts. Loaded lazily when enabled. Default: false.
# 内容片段数量。启用后按需加载。默认值：false。
part_count = false
# Combined message and part payload size. Loaded lazily. Default: false.
# 消息和片段数据总大小。启用后按需加载。默认值：false。
size = false
# Whether payload size is at least 10 MB. Loaded lazily. Default: false.
# 数据总大小是否达到 10 MB。启用后按需加载。默认值：false。
large_session = false
# Last update time. Default: true.
# 最后更新时间。默认值：true。
updated = true
# Creation time. Default: true.
# 创建时间。默认值：true。
created = true
# Model information. Default: true.
# 模型信息。默认值：true。
model = true
# Agent name. Default: true.
# Agent 名称。默认值：true。
agent = true
# Session cost. Default: true.
# 会话费用。默认值：true。
cost = true
# Token usage. Default: true.
# Token 使用情况。默认值：true。
tokens = true
# Command used to resume the session. Default: false.
# 恢复会话时使用的命令。默认值：false。
resume_command = false
`

type rawConfig struct {
	DB        string     `toml:"db"`
	Language  string     `toml:"language"`
	Limit     int        `toml:"limit"`
	OpenCode  string     `toml:"opencode"`
	ReadOnly  *bool      `toml:"read_only"`
	Search    rawSearch  `toml:"search"`
	Preview   rawPreview `toml:"preview"`
	UI        rawUI      `toml:"ui"`
	ThemeName string     `toml:"theme_name"`
	ThemeFile string     `toml:"theme_file"`
	Theme     rawTheme   `toml:"theme"`
	Details   rawDetails `toml:"details"`
}

type rawThemeFile struct {
	Theme rawTheme `toml:"theme"`
}

type rawSearch struct {
	RecentDays *int `toml:"recent_days"`
}

type rawPreview struct {
	RecentMessagesLimit *int `toml:"recent_messages_limit"`
}

type rawUI struct {
	BorderStyle     string   `toml:"border_style"`
	SplitRatio      *float64 `toml:"split_ratio"`
	TwoPaneMinWidth *int     `toml:"two_pane_min_width"`
}

type rawTheme struct {
	ActiveBorderColor             []string      `toml:"active_border_color"`
	InactiveBorderColor           []string      `toml:"inactive_border_color"`
	SearchingActiveBorderColor    []string      `toml:"searching_active_border_color"`
	TitleColor                    []string      `toml:"title_color"`
	AccentColor                   []string      `toml:"accent_color"`
	OptionsTextColor              []string      `toml:"options_text_color"`
	DefaultFgColor                []string      `toml:"default_fg_color"`
	MutedFgColor                  []string      `toml:"muted_fg_color"`
	MatchColor                    []string      `toml:"match_color"`
	FocusedSelectedMatchColor     []string      `toml:"focused_selected_match_color"`
	InactiveSelectedMatchColor    []string      `toml:"inactive_selected_match_color"`
	ErrorColor                    []string      `toml:"error_color"`
	WarningColor                  []string      `toml:"warning_color"`
	SelectedLineBgColor           []string      `toml:"selected_line_bg_color"`
	SelectedLineFgColor           []string      `toml:"selected_line_fg_color"`
	InactiveViewSelectedLineColor []string      `toml:"inactive_view_selected_line_color"`
	StatusModeColor               []string      `toml:"status_mode_color"`
	StatusTextColor               []string      `toml:"status_text_color"`
	StatusKeyColor                []string      `toml:"status_key_color"`
	DetailsHintColor              []string      `toml:"details_hint_color"`
	PreviewTimestampColor         []string      `toml:"preview_timestamp_color"`
	MemorySnippetColor            []string      `toml:"memory_snippet_color"`
	Modal                         rawThemeModal `toml:"modal"`
}

type rawThemeModal struct {
	BackgroundColor string `toml:"background_color"`
	BorderColor     string `toml:"border_color"`
	KeyColor        string `toml:"key_color"`
	TextColor       string `toml:"text_color"`
	MutedColor      string `toml:"muted_color"`
	IconColor       string `toml:"icon_color"`
	WarningColor    string `toml:"warning_color"`
}

type rawDetails struct {
	Fields map[string]any `toml:"fields"`
}

func ResolveOptions(cli Options, cliSet map[string]bool) (Options, error) {
	opts := DefaultOptions()
	explicitConfig := cliSet["config"]
	configPath := cli.ConfigPath
	if configPath == "" {
		configPath = defaultConfigPath()
	}
	opts.ConfigPath = configPath

	cfg, err := loadConfig(configPath, !explicitConfig)
	if err != nil {
		return Options{}, err
	}
	if cfg.ThemeName != "" {
		themePath, err := resolveThemeName(configPath, cfg.ThemeName)
		if err != nil {
			return Options{}, err
		}
		themeCfg, err := loadThemeFile(themePath)
		if err != nil {
			return Options{}, err
		}
		cfg.Theme = mergeRawTheme(themeCfg.Theme, cfg.Theme)
	}
	if cfg.ThemeFile != "" {
		themePath, err := resolveThemePath(configPath, cfg.ThemeFile)
		if err != nil {
			return Options{}, err
		}
		themeCfg, err := loadThemeFile(themePath)
		if err != nil {
			return Options{}, err
		}
		cfg.Theme = mergeRawTheme(themeCfg.Theme, cfg.Theme)
	}
	mergeConfig(&opts, cfg)

	if cliSet["db"] {
		opts.DBPath = cli.DBPath
	}
	if cliSet["limit"] {
		opts.Limit = cli.Limit
	}
	if cliSet["language"] {
		opts.Language = cli.Language
	}
	if cliSet["opencode"] {
		opts.OpenCodeCommand = cli.OpenCodeCommand
	}
	if explicitConfig {
		opts.ConfigPath = cli.ConfigPath
	}

	if opts.Limit <= 0 {
		opts.Limit = 500
	}
	if opts.Language == "" {
		opts.Language = "auto"
	}
	if opts.OpenCodeCommand == "" {
		opts.OpenCodeCommand = "opencode"
	}
	if opts.DetailFields == nil {
		opts.DetailFields = DefaultDetailFields()
	}
	if opts.RecentDays < 0 || opts.RecentDays > 365 {
		return Options{}, fmt.Errorf("search.recent_days must be between 0 and 365")
	}
	if opts.PreviewLimit < 1 || opts.PreviewLimit > 100 {
		return Options{}, fmt.Errorf("preview.recent_messages_limit must be between 1 and 100")
	}
	if err := lazytui.ValidateUIConfig(opts.UI); err != nil {
		return Options{}, err
	}
	if _, err := lazytui.BuildStyles(opts.Theme); err != nil {
		return Options{}, err
	}

	return opts, nil
}

func resolveThemeName(configPath string, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("theme_name cannot be empty")
	}
	fileName := name
	if filepath.Ext(fileName) == "" {
		fileName += ".toml"
	}
	configDir := filepath.Dir(configPath)
	candidates := []string{
		filepath.Join(configDir, "themes", fileName),
	}
	if exe, err := os.Executable(); err == nil && exe != "" {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "themes", fileName))
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("theme preset not found: %s", name)
}

func loadThemeFile(path string) (rawThemeFile, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return rawThemeFile{}, fmt.Errorf("theme file not found: %s", path)
	}
	if err != nil {
		return rawThemeFile{}, err
	}
	var cfg rawThemeFile
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return rawThemeFile{}, err
	}
	return cfg, nil
}

func resolveThemePath(configPath string, themePath string) (string, error) {
	expanded, err := expandHome(themePath)
	if err != nil {
		return "", err
	}
	if filepath.IsAbs(expanded) {
		return expanded, nil
	}
	base := filepath.Dir(configPath)
	if base == "" {
		base = "."
	}
	return filepath.Clean(filepath.Join(base, expanded)), nil
}

func mergeRawTheme(base rawTheme, override rawTheme) rawTheme {
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

func DefaultOptions() Options {
	return Options{
		Limit:           500,
		Language:        "auto",
		OpenCodeCommand: "opencode",
		DetailFields:    DefaultDetailFields(),
		ReadOnly:        false,
		RecentDays:      7,
		PreviewLimit:    5,
		Theme:           lazytui.DefaultThemeConfig(),
		UI:              lazytui.DefaultUIConfig(),
	}
}

func DefaultDetailFields() map[string]bool {
	return map[string]bool{
		"title":          true,
		"session":        true,
		"project":        true,
		"directory":      true,
		"path_status":    true,
		"message_count":  false,
		"part_count":     false,
		"size":           false,
		"large_session":  false,
		"updated":        true,
		"created":        true,
		"model":          true,
		"agent":          true,
		"cost":           true,
		"tokens":         true,
		"resume_command": false,
	}
}

func PrintDefaultConfig(w io.Writer) error {
	_, err := io.WriteString(w, DefaultConfigTOML)
	return err
}

func defaultConfigPath() string {
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return filepath.Join(configHome, "lazyocs", "config.toml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "lazyocs", "config.toml")
}

func loadConfig(path string, createDefault bool) (rawConfig, error) {
	if path == "" {
		return rawConfig{}, nil
	}
	expanded, err := expandHome(path)
	if err != nil {
		return rawConfig{}, err
	}
	data, err := os.ReadFile(expanded)
	if errors.Is(err, os.ErrNotExist) {
		if createDefault {
			_ = createDefaultConfig(expanded)
			return rawConfig{}, nil
		}
		return rawConfig{}, fmt.Errorf("config file not found: %s", expanded)
	}
	if err != nil {
		return rawConfig{}, err
	}
	var cfg rawConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return rawConfig{}, err
	}
	return cfg, nil
}

func createDefaultConfig(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = io.WriteString(file, DefaultConfigTOML)
	if err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func mergeConfig(opts *Options, cfg rawConfig) {
	if cfg.DB != "" {
		opts.DBPath = cfg.DB
	}
	if cfg.Language != "" {
		opts.Language = cfg.Language
	}
	if cfg.Limit > 0 {
		opts.Limit = cfg.Limit
	}
	if cfg.OpenCode != "" {
		opts.OpenCodeCommand = cfg.OpenCode
	}
	if cfg.ReadOnly != nil {
		opts.ReadOnly = *cfg.ReadOnly
	}
	if cfg.Search.RecentDays != nil {
		opts.RecentDays = *cfg.Search.RecentDays
	}
	if cfg.Preview.RecentMessagesLimit != nil {
		opts.PreviewLimit = *cfg.Preview.RecentMessagesLimit
	}
	if cfg.UI.BorderStyle != "" {
		opts.UI.BorderStyle = cfg.UI.BorderStyle
	}
	if cfg.UI.SplitRatio != nil {
		opts.UI.SplitRatio = *cfg.UI.SplitRatio
	}
	if cfg.UI.TwoPaneMinWidth != nil {
		opts.UI.TwoPaneMinWidth = *cfg.UI.TwoPaneMinWidth
	}
	opts.Theme = lazytui.MergeTheme(opts.Theme, toThemeConfig(cfg.Theme))
	if len(cfg.Details.Fields) > 0 {
		fields := DefaultDetailFields()
		for key, value := range cfg.Details.Fields {
			enabled, ok := parseBoolish(value)
			if !ok {
				continue
			}
			fields[key] = enabled
		}
		opts.DetailFields = fields
	}
}

func toThemeConfig(raw rawTheme) lazytui.ThemeConfig {
	return lazytui.ThemeConfig{
		ActiveBorderColor:             raw.ActiveBorderColor,
		InactiveBorderColor:           raw.InactiveBorderColor,
		SearchingActiveBorderColor:    raw.SearchingActiveBorderColor,
		TitleColor:                    raw.TitleColor,
		AccentColor:                   raw.AccentColor,
		OptionsTextColor:              raw.OptionsTextColor,
		DefaultFgColor:                raw.DefaultFgColor,
		MutedFgColor:                  raw.MutedFgColor,
		MatchColor:                    raw.MatchColor,
		FocusedSelectedMatchColor:     raw.FocusedSelectedMatchColor,
		InactiveSelectedMatchColor:    raw.InactiveSelectedMatchColor,
		ErrorColor:                    raw.ErrorColor,
		WarningColor:                  raw.WarningColor,
		SelectedLineBgColor:           raw.SelectedLineBgColor,
		SelectedLineFgColor:           raw.SelectedLineFgColor,
		InactiveViewSelectedLineColor: raw.InactiveViewSelectedLineColor,
		StatusModeColor:               raw.StatusModeColor,
		StatusTextColor:               raw.StatusTextColor,
		StatusKeyColor:                raw.StatusKeyColor,
		DetailsHintColor:              raw.DetailsHintColor,
		PreviewTimestampColor:         raw.PreviewTimestampColor,
		MemorySnippetColor:            raw.MemorySnippetColor,
		Modal: lazytui.ThemeModalConfig{
			BackgroundColor: raw.Modal.BackgroundColor,
			BorderColor:     raw.Modal.BorderColor,
			KeyColor:        raw.Modal.KeyColor,
			TextColor:       raw.Modal.TextColor,
			MutedColor:      raw.Modal.MutedColor,
			IconColor:       raw.Modal.IconColor,
			WarningColor:    raw.Modal.WarningColor,
		},
	}
}

func parseBoolish(value any) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case int64:
		if v == 0 || v == 1 {
			return v == 1, true
		}
	case int:
		if v == 0 || v == 1 {
			return v == 1, true
		}
	}
	return false, false
}
