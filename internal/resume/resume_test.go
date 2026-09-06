package resume

import (
	"reflect"
	"strings"
	"testing"
)

func TestDefaultCommandArgs(t *testing.T) {
	cfg, err := Normalize(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	command, args := cfg.CommandArgs("ses_example")
	if command != "opencode" {
		t.Fatalf("command = %q", command)
	}
	if want := []string{"--session", "ses_example"}; !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestFixedCommandArgs(t *testing.T) {
	cfg, err := Normalize(Config{Command: "opencode", Args: []string{"--pure"}})
	if err != nil {
		t.Fatal(err)
	}
	_, args := cfg.CommandArgs("ses_fixed")
	if want := []string{"--pure", "--session", "ses_fixed"}; !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestCommandEnvAppendsConfiguredVariables(t *testing.T) {
	cfg, err := Normalize(Config{Command: "opencode", Env: map[string]string{"ALL_PROXY": "http://127.0.0.1:7897"}})
	if err != nil {
		t.Fatal(err)
	}
	env := cfg.CommandEnv([]string{"PATH=/usr/bin"})
	if want := []string{"PATH=/usr/bin", "ALL_PROXY=http://127.0.0.1:7897"}; !reflect.DeepEqual(env, want) {
		t.Fatalf("env = %#v, want %#v", env, want)
	}
}

func TestCommandEnvOverridesConfiguredVariables(t *testing.T) {
	cfg, err := Normalize(Config{Command: "opencode", Env: map[string]string{"ALL_PROXY": "http://127.0.0.1:7897"}})
	if err != nil {
		t.Fatal(err)
	}
	env := cfg.CommandEnv([]string{"PATH=/usr/bin", "ALL_PROXY=http://old"})
	if want := []string{"PATH=/usr/bin", "ALL_PROXY=http://127.0.0.1:7897"}; !reflect.DeepEqual(env, want) {
		t.Fatalf("env = %#v, want %#v", env, want)
	}
}

func TestNormalizeRejectsInvalidEnvKey(t *testing.T) {
	_, err := Normalize(Config{Command: "opencode", Env: map[string]string{"BAD=KEY": "value"}})
	if err == nil || !strings.Contains(err.Error(), "resume.env") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCustomSessionArgs(t *testing.T) {
	cfg, err := Normalize(Config{Command: "opencode", SessionArgs: []string{"run", SessionPlaceholder}})
	if err != nil {
		t.Fatal(err)
	}
	_, args := cfg.CommandArgs("ses_custom")
	if want := []string{"run", "ses_custom"}; !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestNormalizeRejectsMissingPlaceholder(t *testing.T) {
	_, err := Normalize(Config{Command: "opencode", SessionArgs: []string{"--session"}})
	if err == nil || !strings.Contains(err.Error(), SessionPlaceholder) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCommandLineQuotesDisplayOnly(t *testing.T) {
	cfg := Config{Command: "/opt/OpenCode/opencode", Args: []string{"--config", "my config.json"}, SessionArgs: []string{"--session", SessionPlaceholder}}
	line := cfg.CommandLine("ses x")
	for _, want := range []string{"/opt/OpenCode/opencode", "\"my config.json\"", "\"ses x\""} {
		if !strings.Contains(line, want) {
			t.Fatalf("command line missing %q: %q", want, line)
		}
	}
}
