package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveOptionsReadsConfigAndBoolishFields(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.toml")
	if err := os.WriteFile(configPath, []byte(`
db = "/tmp/opencode.db"
language = "zh-CN"
limit = 123
opencode = "custom-opencode"

[details.fields]
title = 1
project = 0
resume_command = true
tokens = false
`), 0o600); err != nil {
		t.Fatal(err)
	}

	opts, err := ResolveOptions(Options{ConfigPath: configPath}, map[string]bool{"config": true})
	if err != nil {
		t.Fatal(err)
	}

	if opts.DBPath != "/tmp/opencode.db" || opts.Language != "zh-CN" || opts.Limit != 123 || opts.OpenCodeCommand != "custom-opencode" {
		t.Fatalf("unexpected options: %+v", opts)
	}
	if !opts.DetailFields["title"] || opts.DetailFields["project"] || !opts.DetailFields["resume_command"] || opts.DetailFields["tokens"] {
		t.Fatalf("unexpected detail fields: %+v", opts.DetailFields)
	}
}

func TestResolveOptionsCLIOverridesConfig(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.toml")
	if err := os.WriteFile(configPath, []byte(`
language = "zh-CN"
limit = 123
opencode = "custom-opencode"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	opts, err := ResolveOptions(Options{
		ConfigPath:      configPath,
		Language:        "en",
		Limit:           42,
		OpenCodeCommand: "opencode-next",
	}, map[string]bool{"config": true, "language": true, "limit": true, "opencode": true})
	if err != nil {
		t.Fatal(err)
	}

	if opts.Language != "en" || opts.Limit != 42 || opts.OpenCodeCommand != "opencode-next" {
		t.Fatalf("CLI did not override config: %+v", opts)
	}
}

func TestResolveOptionsCreatesMissingDefaultConfig(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configPath := filepath.Join(configHome, "lazyocs", "config.toml")

	_, err := ResolveOptions(Options{}, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "[details.fields]") {
		t.Fatalf("default config was not written: %s", string(data))
	}
}

func TestResolveOptionsErrorsForMissingExplicitConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "missing.toml")

	_, err := ResolveOptions(Options{ConfigPath: configPath}, map[string]bool{"config": true})
	if err == nil {
		t.Fatal("expected missing explicit config error")
	}
	if !strings.Contains(err.Error(), "config file not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveOptionsDefaultsToReadOnly(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	opts, err := ResolveOptions(Options{}, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.ReadOnly {
		t.Fatal("expected read-only mode by default")
	}
}

func TestResolveOptionsCanDisableReadOnly(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.toml")
	if err := os.WriteFile(configPath, []byte("read_only = false\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	opts, err := ResolveOptions(Options{ConfigPath: configPath}, map[string]bool{"config": true})
	if err != nil {
		t.Fatal(err)
	}
	if opts.ReadOnly {
		t.Fatal("expected read-only mode to be disabled")
	}
}
