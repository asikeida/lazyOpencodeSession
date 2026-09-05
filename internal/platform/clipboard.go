package platform

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func Copy(ctx context.Context, text string) error {
	commands := clipboardCommands(runtime.GOOS)
	if len(commands) == 0 {
		return fmt.Errorf("clipboard copy is not supported on %s", runtime.GOOS)
	}
	for _, spec := range commands {
		if _, err := exec.LookPath(spec[0]); err != nil {
			continue
		}
		cmd := exec.CommandContext(ctx, spec[0], spec[1:]...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("clipboard copy timed out: %w", ctx.Err())
			}
			return fmt.Errorf("clipboard command %s failed: %w", spec[0], err)
		}
		return nil
	}
	return fmt.Errorf("clipboard tool not found; expected one of: %s", clipboardToolNames(commands))
}

func clipboardCommands(goos string) [][]string {
	switch goos {
	case "darwin":
		return [][]string{{"pbcopy"}}
	case "linux":
		return [][]string{
			{"wl-copy"},
			{"xclip", "-selection", "clipboard"},
			{"xsel", "--clipboard", "--input"},
		}
	default:
		return nil
	}
}

func clipboardToolNames(commands [][]string) string {
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		names = append(names, command[0])
	}
	return strings.Join(names, ", ")
}
