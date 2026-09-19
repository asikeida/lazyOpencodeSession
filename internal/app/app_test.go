package app

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestOpenCodeDBOverrideResolvesRelativeToDataDirectory(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)
	t.Setenv("OPENCODE_DB", "channels/beta.db")
	want := filepath.Join(dataHome, "opencode", "channels", "beta.db")
	if got := openCodeDBOverride(); got != want {
		t.Fatalf("override = %q, want %q", got, want)
	}
}

func TestResolveDBPathPrefersExplicitPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := resolveDBPath(context.Background(), path, "missing-opencode-command")
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Fatalf("resolved path = %q, want %q", got, path)
	}
}

func TestDiscoverOpenCodeDBUsesDebugPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable shell fixture is Unix-specific")
	}
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "opencode.db")
	command := filepath.Join(dir, "opencode")
	script := "#!/bin/sh\nprintf '%s\\n' '" + dbPath + "'\n"
	if err := os.WriteFile(command, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	if got := discoverOpenCodeDB(context.Background(), command); got != dbPath {
		t.Fatalf("discovered path = %q, want %q", got, dbPath)
	}
}
