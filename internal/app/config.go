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
# Path to OpenCode SQLite database. Leave empty to use ~/.local/share/opencode/opencode.db.
db = ""
language = "auto"
theme = "default"
limit = 500
opencode = "opencode"

[details.fields]
title = true
session = true
project = true
directory = true
path_status = true
message_count = false
part_count = false
size = false
large_session = false
updated = true
created = true
model = true
agent = true
cost = true
tokens = true
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
