package platform

import (
	"errors"
	"os/exec"
)

func Copy(text string) error {
	commands := [][]string{
		{"wl-copy"},
		{"xclip", "-selection", "clipboard"},
		{"xsel", "--clipboard", "--input"},
	}
	for _, spec := range commands {
		if _, err := exec.LookPath(spec[0]); err != nil {
			continue
		}
		cmd := exec.Command(spec[0], spec[1:]...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return err
		}
		_, writeErr := stdin.Write([]byte(text))
		closeErr := stdin.Close()
		waitErr := cmd.Wait()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
		return waitErr
	}
	return errors.New("clipboard tool not found")
}
