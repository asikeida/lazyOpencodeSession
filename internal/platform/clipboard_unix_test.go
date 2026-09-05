//go:build !windows

package platform

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCopyWritesTextToAvailableTool(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "clipboard.txt")
	writeClipboardTool(t, dir, preferredClipboardTool(t), "#!/bin/sh\n/bin/cat > \"$CLIPBOARD_TEST_OUTPUT\"\n")
	t.Setenv("PATH", dir)
	t.Setenv("CLIPBOARD_TEST_OUTPUT", output)

	if err := Copy(context.Background(), "ses_example"); err != nil {
		t.Fatalf("Copy returned an error: %v", err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read copied text: %v", err)
	}
	if string(got) != "ses_example" {
		t.Fatalf("copied text = %q, want %q", got, "ses_example")
	}
}

func TestCopyStopsToolWhenContextExpires(t *testing.T) {
	dir := t.TempDir()
	writeClipboardTool(t, dir, preferredClipboardTool(t), "#!/bin/sh\nexec /bin/sleep 10\n")
	t.Setenv("PATH", dir)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	started := time.Now()
	err := Copy(ctx, "ses_example")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Copy error = %v, want context deadline exceeded", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("Copy did not stop promptly after timeout: %v", elapsed)
	}
}

func TestCopyReportsMissingTool(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	err := Copy(context.Background(), "ses_example")
	want := clipboardToolNames(clipboardCommands(runtime.GOOS))
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("Copy error = %v, want missing tool error", err)
	}
}

func TestClipboardCommandsByPlatform(t *testing.T) {
	tests := []struct {
		goos string
		want string
	}{
		{goos: "linux", want: "wl-copy, xclip, xsel"},
		{goos: "darwin", want: "pbcopy"},
		{goos: "windows", want: ""},
	}
	for _, test := range tests {
		t.Run(test.goos, func(t *testing.T) {
			got := clipboardToolNames(clipboardCommands(test.goos))
			if got != test.want {
				t.Fatalf("clipboard commands = %q, want %q", got, test.want)
			}
		})
	}
}

func preferredClipboardTool(t *testing.T) string {
	t.Helper()
	commands := clipboardCommands(runtime.GOOS)
	if len(commands) == 0 {
		t.Skipf("clipboard is not supported on %s", runtime.GOOS)
	}
	return commands[0][0]
}

func writeClipboardTool(t *testing.T, dir string, name string, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write fake clipboard tool: %v", err)
	}
}
