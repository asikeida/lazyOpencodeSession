package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/asikeida/lazyOpencodeSession/internal/opencode"
	"github.com/asikeida/lazyOpencodeSession/internal/resume"
	lazytui "github.com/asikeida/lazyOpencodeSession/internal/tui"
)

type Options struct {
	DBPath          string
	ConfigPath      string
	Limit           int
	Language        string
	OpenCodeCommand string
	Resume          resume.Config
	DetailFields    map[string]bool
	ReadOnly        bool
	RecentDays      int
	PreviewLimit    int
	Theme           lazytui.ThemeConfig
	UI              lazytui.UIConfig
}

func Run(ctx context.Context, opts Options) error {
	if opts.Limit <= 0 {
		opts.Limit = 500
	}
	if opts.Resume.Command == "" {
		opts.Resume.Command = opts.OpenCodeCommand
	}
	if opts.Resume.Command == "" {
		opts.Resume.Command = "opencode"
	}
	resumeCfg, err := resume.Normalize(opts.Resume)
	if err != nil {
		return err
	}

	dbPath, err := resolveDBPath(ctx, opts.DBPath, resumeCfg.Command)
	if err != nil {
		return err
	}

	repo, err := opencode.Open(ctx, dbPath, opts.ReadOnly)
	if err != nil {
		return err
	}
	defer repo.Close()
	repo.SetAPICommand(resumeCfg.Command)

	model := lazytui.New(lazytui.Options{
		Repo:         repo,
		Limit:        opts.Limit,
		Language:     opts.Language,
		Resume:       resumeCfg,
		DetailFields: opts.DetailFields,
		ReadOnly:     opts.ReadOnly,
		RecentDays:   opts.RecentDays,
		PreviewLimit: opts.PreviewLimit,
		Theme:        opts.Theme,
		UI:           opts.UI,
	})

	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return err
	}

	m, ok := finalModel.(lazytui.Model)
	if !ok || m.ResumeSessionID() == "" {
		return nil
	}

	command, args := resumeCfg.CommandArgs(m.ResumeSessionID())
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = resumeCfg.CommandEnv(os.Environ())
	return cmd.Run()
}

func Check(ctx context.Context, opts Options) error {
	if opts.Limit <= 0 {
		opts.Limit = 5
	}
	command := opts.Resume.Command
	if command == "" {
		command = opts.OpenCodeCommand
	}
	if command == "" {
		command = "opencode"
	}
	dbPath, err := resolveDBPath(ctx, opts.DBPath, command)
	if err != nil {
		return err
	}
	repo, err := opencode.Open(ctx, dbPath, true)
	if err != nil {
		return err
	}
	defer repo.Close()

	sessions, err := repo.ListSessions(ctx, opencode.SessionFilter{Limit: opts.Limit})
	if err != nil {
		return opencode.ActionableError(err)
	}
	fmt.Printf("database: %s\n", dbPath)
	compat := repo.Compatibility()
	fmt.Printf("schema: version=%s browse=%t stats=%t preview=%t rename=%t delete=%t\n", repo.SchemaVersion(), compat.Browse, compat.Stats, compat.Preview, compat.Rename, compat.Delete)
	if opts.RecentDays > 0 {
		memories, err := repo.RecentUserMemory(ctx, time.Now().AddDate(0, 0, -opts.RecentDays), 5000, 4000)
		if err != nil {
			return opencode.ActionableError(err)
		}
		fmt.Printf("memory: %d user messages (%dd)\n", len(memories), opts.RecentDays)
	}
	fmt.Printf("loaded sessions: %d\n", len(sessions))
	for _, s := range sessions {
		fmt.Printf("%s\t%s\t%s\n", s.ID, s.UpdatedAt.Format("2006-01-02 15:04"), s.Title)
	}
	return nil
}

func resolveDBPath(ctx context.Context, path string, command string) (string, error) {
	if env := os.Getenv("LAZYOCS_DB"); path == "" && env != "" {
		path = env
	}
	if path == "" {
		if discovered := discoverOpenCodeDB(ctx, command); discovered != "" {
			path = discovered
		}
	}
	if path == "" {
		path = openCodeDBOverride()
	}
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, ".local", "share", "opencode", "opencode.db")
	}

	expanded, err := expandHome(path)
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("OpenCode database not found: %s", abs)
	}
	if info.IsDir() {
		return "", fmt.Errorf("database path is a directory: %s", abs)
	}
	return abs, nil
}

func discoverOpenCodeDB(ctx context.Context, command string) string {
	if command == "" {
		command = "opencode"
	}
	commandCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	output, err := exec.CommandContext(commandCtx, command, "debug", "paths", "db").Output()
	if err != nil {
		return ""
	}
	path := strings.TrimSpace(string(output))
	if path == "" || strings.ContainsAny(path, "\r\n") {
		return ""
	}
	return path
}

func openCodeDBOverride() string {
	value := strings.TrimSpace(os.Getenv("OPENCODE_DB"))
	if value == "" || filepath.IsAbs(value) {
		return value
	}
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "opencode", value)
}

func expandHome(path string) (string, error) {
	if path == "~" {
		return os.UserHomeDir()
	}
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}
