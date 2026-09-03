package platform

import (
	"context"
	"errors"
	"os/exec"
	"strings"
)

func Copy(ctx context.Context, text string) error {
	commands := [][]string{
		{"wl-copy"},
		{"xclip", "-selection", "clipboard"},
		{"xsel", "--clipboard", "--input"},
	}
	for _, spec := range commands {
		if _, err := exec.LookPath(spec[0]); err != nil {
			continue
		}
		cmd := exec.CommandContext(ctx, spec[0], spec[1:]...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		return nil
	}
	return errors.New("clipboard tool not found")
}
