package platform

import "testing"

func TestClipboardCommandsByPlatform(t *testing.T) {
	tests := []struct {
		goos string
		want string
	}{
		{goos: "linux", want: "wl-copy, xclip, xsel"},
		{goos: "darwin", want: "pbcopy"},
		{goos: "windows", want: "clip.exe"},
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
