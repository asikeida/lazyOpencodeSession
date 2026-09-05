package app

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/asikeida/lazyOpencodeSession/internal/resume"
	lazytui "github.com/asikeida/lazyOpencodeSession/internal/tui"
)

func TestResolveOptionsReadsConfigAndBoolishFields(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.toml")
	themePath := filepath.Join(tmp, "theme.toml")
	themesDir := filepath.Join(tmp, "themes")
	if err := os.MkdirAll(themesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themesDir, "moss.toml"), []byte(`
[theme]
default_fg_color = ["252"]
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(themePath, []byte(`
[theme]
inactive_border_color = ["blue"]

[theme.modal]
text_color = "#eeeeee"
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte(`
db = "/tmp/opencode.db"
language = "zh-CN"
limit = 123
opencode = "custom-opencode"
theme_name = "moss"
theme_file = "theme.toml"

[resume]
command = "opencode"
args = ["--proxy"]
session_args = ["--session", "{session_id}"]

[search]
recent_days = 14

[preview]
recent_messages_limit = 9

[ui]
border_style = "double"
split_ratio = 0.4
two_pane_min_width = 120

[theme]
active_border_color = ["yellow", "bold"]
match_color = ["magenta", "underline"]
status_mode_color = ["cyan", "bold"]

[theme.modal]
border_color = "#ffffff"

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

	if opts.DBPath != "/tmp/opencode.db" || opts.Language != "zh-CN" || opts.Limit != 123 || opts.OpenCodeCommand != "opencode" {
		t.Fatalf("unexpected options: %+v", opts)
	}
	if opts.RecentDays != 14 {
		t.Fatalf("recent days = %d, want 14", opts.RecentDays)
	}
	if opts.Resume.Command != "opencode" || !reflect.DeepEqual(opts.Resume.Args, []string{"--proxy"}) {
		t.Fatalf("unexpected resume config: %+v", opts.Resume)
	}
	if opts.PreviewLimit != 9 {
		t.Fatalf("preview limit = %d, want 9", opts.PreviewLimit)
	}
	if opts.UI.BorderStyle != "double" || opts.UI.SplitRatio != 0.4 || opts.UI.TwoPaneMinWidth != 120 {
		t.Fatalf("unexpected ui config: %+v", opts.UI)
	}
	if len(opts.Theme.ActiveBorderColor) != 2 || opts.Theme.ActiveBorderColor[0] != "yellow" {
		t.Fatalf("unexpected active border theme: %#v", opts.Theme.ActiveBorderColor)
	}
	if opts.Theme.Modal.BorderColor != "#ffffff" {
		t.Fatalf("unexpected modal border color: %q", opts.Theme.Modal.BorderColor)
	}
	if len(opts.Theme.InactiveBorderColor) != 1 || opts.Theme.InactiveBorderColor[0] != "blue" {
		t.Fatalf("unexpected inherited inactive border theme: %#v", opts.Theme.InactiveBorderColor)
	}
	if opts.Theme.Modal.TextColor != "#eeeeee" {
		t.Fatalf("unexpected inherited modal text color: %q", opts.Theme.Modal.TextColor)
	}
	if len(opts.Theme.StatusModeColor) != 2 || opts.Theme.StatusModeColor[0] != "cyan" {
		t.Fatalf("unexpected status mode theme: %#v", opts.Theme.StatusModeColor)
	}
	if len(opts.Theme.DefaultFgColor) != 1 || opts.Theme.DefaultFgColor[0] != "252" {
		t.Fatalf("unexpected inherited theme preset fg: %#v", opts.Theme.DefaultFgColor)
	}
	if !opts.DetailFields["title"] || opts.DetailFields["project"] || !opts.DetailFields["resume_command"] || opts.DetailFields["tokens"] {
		t.Fatalf("unexpected detail fields: %+v", opts.DetailFields)
	}
}

func TestResolveOptionsRejectsMissingThemePreset(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("theme_name = 'missing'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ResolveOptions(Options{ConfigPath: configPath}, map[string]bool{"config": true})
	if err == nil || !strings.Contains(err.Error(), "theme preset not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestThemePresetCandidatesIncludeSystemPackageDirectory(t *testing.T) {
	candidates := themePresetCandidates("/home/user/.config/lazyocs/config.toml", "/usr/bin/lazyocs", "moss.toml")
	want := filepath.FromSlash("/usr/share/lazyocs/themes/moss.toml")
	if candidates[len(candidates)-1] != want {
		t.Fatalf("system theme candidate = %q, want %q", candidates[len(candidates)-1], want)
	}
}

func TestResolveOptionsRejectsMissingThemeFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("theme_file = 'missing-theme.toml'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ResolveOptions(Options{ConfigPath: configPath}, map[string]bool{"config": true})
	if err == nil || !strings.Contains(err.Error(), "theme file not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveOptionsRejectsInvalidThemeToken(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("[theme]\nactive_border_color = ['bogus']\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ResolveOptions(Options{ConfigPath: configPath}, map[string]bool{"config": true})
	if err == nil || !strings.Contains(err.Error(), "theme.active_border_color") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveOptionsRejectsInvalidRecentDays(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("[search]\nrecent_days = 366\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ResolveOptions(Options{ConfigPath: configPath}, map[string]bool{"config": true})
	if err == nil || !strings.Contains(err.Error(), "between 0 and 365") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveOptionsCanDisableRecentMemory(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("[search]\nrecent_days = 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	opts, err := ResolveOptions(Options{ConfigPath: configPath}, map[string]bool{"config": true})
	if err != nil {
		t.Fatal(err)
	}
	if opts.RecentDays != 0 {
		t.Fatalf("recent days = %d, want disabled", opts.RecentDays)
	}
}

func TestResolveOptionsRejectsInvalidPreviewLimit(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("[preview]\nrecent_messages_limit = 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ResolveOptions(Options{ConfigPath: configPath}, map[string]bool{"config": true})
	if err == nil || !strings.Contains(err.Error(), "preview.recent_messages_limit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveOptionsRejectsInvalidUIConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("[ui]\nborder_style = 'weird'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ResolveOptions(Options{ConfigPath: configPath}, map[string]bool{"config": true})
	if err == nil || !strings.Contains(err.Error(), "ui.border_style") {
		t.Fatalf("unexpected error: %v", err)
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
	if opts.Resume.Command != "opencode-next" {
		t.Fatalf("CLI did not override resume command: %+v", opts.Resume)
	}
}

func TestResolveOptionsKeepsLegacyOpenCodeFallback(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte(`opencode = "opencode-beta"
`), 0o600); err != nil {
		t.Fatal(err)
	}
	opts, err := ResolveOptions(Options{ConfigPath: configPath}, map[string]bool{"config": true})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Resume.Command != "opencode-beta" {
		t.Fatalf("resume command = %q", opts.Resume.Command)
	}
}

func TestResolveOptionsRejectsInvalidResumeConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte(`[resume]
session_args = ["--session"]
`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ResolveOptions(Options{ConfigPath: configPath}, map[string]bool{"config": true})
	if err == nil || !strings.Contains(err.Error(), resume.SessionPlaceholder) {
		t.Fatalf("unexpected error: %v", err)
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
	if !strings.Contains(string(data), "[theme]") {
		t.Fatalf("default config is missing theme section: %s", string(data))
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

func TestResolveOptionsDefaultsToReadWrite(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	opts, err := ResolveOptions(Options{}, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if opts.ReadOnly {
		t.Fatal("expected read-write mode by default")
	}
	if opts.RecentDays != 7 {
		t.Fatalf("recent days = %d, want default 7", opts.RecentDays)
	}
	if opts.PreviewLimit != 5 {
		t.Fatalf("preview limit = %d, want default 5", opts.PreviewLimit)
	}
	if opts.UI != lazytui.DefaultUIConfig() {
		t.Fatalf("unexpected ui defaults: %+v", opts.UI)
	}
	if opts.Resume.Command != "opencode" {
		t.Fatalf("unexpected resume default: %+v", opts.Resume)
	}
}

func TestResolveOptionsCanEnableReadOnly(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.toml")
	if err := os.WriteFile(configPath, []byte("read_only = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	opts, err := ResolveOptions(Options{ConfigPath: configPath}, map[string]bool{"config": true})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.ReadOnly {
		t.Fatal("expected read-only mode to be enabled")
	}
}
