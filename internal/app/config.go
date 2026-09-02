package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
# Built-in color theme: default or dark. Default: default.
# 内置颜色主题：default 或 dark。默认值：default。
theme = "default"
# Maximum number of sessions loaded into the list. Default: 500.
# 会话列表最多加载的数量。默认值：500。
limit = 500
# OpenCode executable or command path used by Enter. Default: opencode.
# 按 Enter 恢复会话时使用的 OpenCode 命令或路径。默认值：opencode。
opencode = "opencode"

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
	DB       string     `toml:"db"`
	Language string     `toml:"language"`
	Theme    string     `toml:"theme"`
	Limit    int        `toml:"limit"`
	OpenCode string     `toml:"opencode"`
	Details  rawDetails `toml:"details"`
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
	mergeConfig(&opts, cfg)

	if cliSet["db"] {
		opts.DBPath = cli.DBPath
	}
	if cliSet["limit"] {
		opts.Limit = cli.Limit
	}
	if cliSet["theme"] {
		opts.Theme = cli.Theme
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
	if opts.Theme == "" {
		opts.Theme = "default"
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

	return opts, nil
}

func DefaultOptions() Options {
	return Options{
		Limit:           500,
		Theme:           "default",
		Language:        "auto",
		OpenCodeCommand: "opencode",
		DetailFields:    DefaultDetailFields(),
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
	if cfg.Theme != "" {
		opts.Theme = cfg.Theme
	}
	if cfg.Limit > 0 {
		opts.Limit = cfg.Limit
	}
	if cfg.OpenCode != "" {
		opts.OpenCodeCommand = cfg.OpenCode
	}
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
