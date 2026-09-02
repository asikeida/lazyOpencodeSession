package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderHelpLinesAlignsDescriptions(t *testing.T) {
	lines := strings.Split(renderHelpLines([]HelpLine{
		{Key: "q / Ctrl+C", Description: "退出"},
		{Key: "搜索中 Ctrl-K/J", Description: "上下移动"},
		{Key: "Esc", Description: "清空搜索"},
	}), "\n")
	if len(lines) != 3 {
		t.Fatalf("unexpected line count: %d", len(lines))
	}
	firstIndex := strings.Index(lines[0], "退出")
	if firstIndex < 0 {
		t.Fatalf("description not found: %#v", lines)
	}
	first := lipgloss.Width(lines[0][:firstIndex])
	for i, description := range []string{"退出", "上下移动", "清空搜索"} {
		index := strings.Index(lines[i], description)
		if index < 0 || lipgloss.Width(lines[i][:index]) != first {
			t.Fatalf("descriptions are not aligned: %#v", lines)
		}
	}
}

func TestFormatModel(t *testing.T) {
	lines := formatModel(`{"id":"deepseek-v4-flash-vision-exp","providerID":"opencode-go","variant":"default"}`)
	want := []string{
		`"id":         "deepseek-v4-flash-vision-exp"`,
		`"providerID": "opencode-go"`,
		`"variant":    "default"`,
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected model formatting:\n%s", strings.Join(lines, "\n"))
	}
}
