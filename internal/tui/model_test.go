package tui

import (
	"strings"
	"testing"
	"time"

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
		`        "id": "deepseek-v4-flash-vision-exp"`,
		`"providerID": "opencode-go"`,
		`   "variant": "default"`,
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected model formatting:\n%s", strings.Join(lines, "\n"))
	}
}

func TestAppendModelFieldUsesDetailValueColumn(t *testing.T) {
	model := Model{
		fields: map[string]bool{"model": true},
		texts:  NewTexts("zh-CN"),
	}
	lines := []string{}
	model.appendModelField(&lines, `{"id":"mimo-v2.5","providerID":"opencode-go"}`, 60)
	if len(lines) != 3 {
		t.Fatalf("unexpected model line count: %d", len(lines))
	}
	propertyIndex := strings.Index(lines[1], `"id"`)
	if propertyIndex < 0 {
		t.Fatalf("model property not found: %q", lines[1])
	}
	want := lipgloss.Width(fieldPrefix("模型:"))
	if got := lipgloss.Width(lines[1][:propertyIndex]); got != want {
		t.Fatalf("model property starts at %d, want %d: %q", got, want, lines[1])
	}
}

func TestFormatFullTimeIncludesWeekday(t *testing.T) {
	date := time.Date(2026, time.September, 2, 8, 39, 49, 0, time.Local)
	got := formatFullTime(date, [7]string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"})
	if got != "2026-09-02 周三 08:39:49" {
		t.Fatalf("unexpected full time: %q", got)
	}
}
