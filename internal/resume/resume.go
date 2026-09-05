package resume

import (
	"fmt"
	"strconv"
	"strings"
)

const SessionPlaceholder = "{session_id}"

type Config struct {
	Command     string
	Args        []string
	SessionArgs []string
}

func DefaultConfig() Config {
	return Config{Command: "opencode", SessionArgs: []string{"--session", SessionPlaceholder}}
}

func Normalize(cfg Config) (Config, error) {
	if strings.TrimSpace(cfg.Command) == "" {
		return Config{}, fmt.Errorf("resume.command cannot be empty")
	}
	if cfg.SessionArgs == nil {
		cfg.SessionArgs = []string{"--session", SessionPlaceholder}
	}
	if len(cfg.SessionArgs) == 0 {
		return Config{}, fmt.Errorf("resume.session_args cannot be empty")
	}
	found := false
	for _, arg := range cfg.SessionArgs {
		if strings.Contains(arg, SessionPlaceholder) {
			found = true
			break
		}
	}
	if !found {
		return Config{}, fmt.Errorf("resume.session_args must include %s", SessionPlaceholder)
	}
	return cfg, nil
}

func (cfg Config) CommandArgs(sessionID string) (string, []string) {
	args := append([]string{}, cfg.Args...)
	for _, arg := range cfg.SessionArgs {
		args = append(args, strings.ReplaceAll(arg, SessionPlaceholder, sessionID))
	}
	return cfg.Command, args
}

func (cfg Config) CommandLine(sessionID string) string {
	command, args := cfg.CommandArgs(sessionID)
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, quote(command))
	for _, arg := range args {
		parts = append(parts, quote(arg))
	}
	return strings.Join(parts, " ")
}

func quote(value string) string {
	if value == "" || strings.ContainsAny(value, " \t\n\r\"'\\") {
		return strconv.Quote(value)
	}
	return value
}
