package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"

	"github.com/asikeida/lazyOpencodeSession/internal/opencode"
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

func TestTitleDialogUsesLegendAndStableWidth(t *testing.T) {
	model := Model{
		width:      100,
		styles:     NewStyles(),
		texts:      NewTexts("en"),
		titleInput: "untitled",
	}
	dialog := model.renderTitleDialog()
	if !strings.Contains(ansi.Strip(dialog), "Save as") || !strings.Contains(ansi.Strip(dialog), "✎ untitled") {
		t.Fatalf("title dialog is missing its legend or input: %q", ansi.Strip(dialog))
	}
	if got := len(strings.Split(dialog, "\n")); got != 4 {
		t.Fatalf("title dialog has %d lines, want compact 4-line layout", got)
	}
	wantWidth := model.dialogWidth()
	for i, line := range strings.Split(dialog, "\n") {
		if got := lipgloss.Width(line); got != wantWidth {
			t.Fatalf("dialog line %d has width %d, want %d", i, got, wantWidth)
		}
	}
}

func TestOverlayKeepsScreenWidthWhenBoundaryCrossesWideCharacter(t *testing.T) {
	model := Model{
		width:  20,
		height: 5,
		styles: NewStyles(),
	}
	base := strings.Join([]string{
		strings.Repeat("a", 20),
		"1234567中" + strings.Repeat("b", 11),
		"12345678901中" + strings.Repeat("c", 7),
		strings.Repeat("d", 20),
		strings.Repeat("e", 20),
	}, "\n")
	modal := "┌──┐\n│  │\n└──┘"

	for i, line := range strings.Split(model.renderOverlay(base, modal), "\n") {
		if got := lipgloss.Width(line); got != model.width {
			t.Fatalf("overlay line %d has width %d, want %d", i, got, model.width)
		}
	}
}

func TestOverlayPlacesEveryModalBorderAtFixedCells(t *testing.T) {
	model := Model{
		width:  20,
		height: 5,
		styles: NewStyles(),
	}
	base := strings.Join([]string{
		strings.Repeat("界", 10),
		"1234567中" + strings.Repeat("b", 11),
		"12345678901中" + strings.Repeat("c", 7),
		strings.Repeat("文", 10),
		strings.Repeat("e", 20),
	}, "\n")
	modal := "┌──┐\n│  │\n└──┘"
	view := model.renderOverlay(base, modal)

	screen := cellbuf.NewBuffer(model.width, model.height)
	cellbuf.SetContent(screen, view)
	x := (model.width - lipgloss.Width(modal)) / 2
	y := (model.height - len(strings.Split(modal, "\n"))) / 2
	want := [][2]rune{{'┌', '┐'}, {'│', '│'}, {'└', '┘'}}
	for row, borders := range want {
		left := screen.Cell(x, y+row)
		right := screen.Cell(x+lipgloss.Width(modal)-1, y+row)
		if left.Rune != borders[0] || right.Rune != borders[1] {
			t.Fatalf("modal row %d borders are %q/%q, want %q/%q", row, left.Rune, right.Rune, borders[0], borders[1])
		}
	}
}

func TestOverlayRecentersAfterResize(t *testing.T) {
	modal := "┌──┐\n│  │\n└──┘"
	for _, width := range []int{20, 21, 40} {
		model := Model{width: width, height: 5, styles: NewStyles()}
		view := model.renderOverlay(strings.Repeat("界", width/2), modal)
		screen := cellbuf.NewBuffer(width, model.height)
		cellbuf.SetContent(screen, view)
		x := (width - lipgloss.Width(modal)) / 2
		y := (model.height - len(strings.Split(modal, "\n"))) / 2
		if got := screen.Cell(x, y).Rune; got != '┌' {
			t.Fatalf("terminal width %d placed top-left border at the wrong cell: %q", width, got)
		}
	}
}

func TestRenderDialogNeverImplicitlyWrapsBodyLine(t *testing.T) {
	model := Model{styles: NewStyles()}
	for _, body := range []string{strings.Repeat("内容", 20), "line one\nline two", "left\tright"} {
		dialog := model.renderDialog("Title", []string{body}, 20)
		lines := strings.Split(dialog, "\n")
		if len(lines) != 3 {
			t.Fatalf("dialog produced %d physical lines for one body line %q, want 3", len(lines), body)
		}
		for i, line := range lines {
			if got := lipgloss.Width(line); got != 20 {
				t.Fatalf("dialog line %d has width %d, want 20", i, got)
			}
		}
	}
}

func TestDialogWidthNeverExceedsTerminal(t *testing.T) {
	for _, width := range []int{1, 2, 3, 8, 20, 40, 100} {
		model := Model{width: width}
		if got := model.dialogWidth(); got < 0 || got > width {
			t.Fatalf("terminal width %d produced dialog width %d", width, got)
		}
	}
}

func TestHelpUsesSharedDialogFrame(t *testing.T) {
	model := Model{
		width:  100,
		height: 30,
		styles: NewStyles(),
		texts:  NewTexts("en"),
	}
	dialog := model.renderHelp()
	lines := strings.Split(ansi.Strip(dialog), "\n")
	if !strings.Contains(lines[0], model.texts.HelpTitle) || !strings.HasPrefix(lines[0], "┌") {
		t.Fatalf("help does not use a top border legend: %q", lines[0])
	}
	if !strings.HasPrefix(lines[len(lines)-1], "└") {
		t.Fatalf("help does not use the shared dialog footer: %q", lines[len(lines)-1])
	}
	for i, line := range strings.Split(dialog, "\n") {
		if got := lipgloss.Width(line); got != model.dialogWidth() {
			t.Fatalf("help line %d has width %d, want %d", i, got, model.dialogWidth())
		}
	}
}

func TestHelpRemainsAnOverlayOverMainView(t *testing.T) {
	model := Model{
		width:  100,
		height: 30,
		styles: NewStyles(),
		texts:  NewTexts("en"),
		help:   true,
	}
	view := ansi.Strip(model.View())
	if !strings.Contains(view, "lazyOpencodeSession - Sessions") || !strings.Contains(view, model.texts.HelpTitle) {
		t.Fatalf("help should preserve the main view behind its dialog: %q", view)
	}
}

func TestDeleteDialogShowsTargetAndConfirmationKeys(t *testing.T) {
	model := Model{
		width:    100,
		styles:   NewStyles(),
		texts:    NewTexts("en"),
		sessions: []opencode.Session{{ID: "ses_example", Title: "Important session"}},
	}
	dialog := ansi.Strip(model.renderDeleteDialog())
	if !strings.HasPrefix(dialog, "┌") {
		t.Fatalf("delete dialog does not use the shared frame: %q", dialog)
	}
	for _, want := range []string{"Delete session", "Important session", "ses_example", "y / Enter delete"} {
		if !strings.Contains(dialog, want) {
			t.Fatalf("delete dialog does not contain %q: %q", want, dialog)
		}
	}
}

func TestDeleteErrorUnlocksConfirmationDialog(t *testing.T) {
	model := Model{
		deleteConfirm: true,
		deleteBusy:    true,
		sessions:      []opencode.Session{{ID: "ses_example"}},
	}
	updated, _ := model.Update(sessionDeleteFailedMsg{sessionID: "ses_example", err: errors.New("delete failed")})
	got := updated.(Model)
	if got.deleteBusy {
		t.Fatal("delete dialog stayed locked after an error")
	}
	if !got.deleteConfirm {
		t.Fatal("delete dialog should remain open so the user can retry or cancel")
	}
	if got.status != "delete failed" {
		t.Fatalf("status = %q, want operation error", got.status)
	}
}

func TestStatsErrorUnlocksOnlyFailedSession(t *testing.T) {
	model := Model{
		sessions:  []opencode.Session{{ID: "ses_example"}},
		statsBusy: map[string]bool{"ses_example": true, "ses_other": true},
	}
	updated, _ := model.Update(statsLoadFailedMsg{sessionID: "ses_example", err: errors.New("stats failed")})
	got := updated.(Model)
	if got.statsBusy["ses_example"] {
		t.Fatal("failed session stayed locked after stats error")
	}
	if !got.statsBusy["ses_other"] {
		t.Fatal("unrelated session loading state was cleared")
	}
	if got.err != nil {
		t.Fatalf("detail error became a fatal list error: %v", got.err)
	}
	if got.status != "stats failed" {
		t.Fatalf("status = %q, want operation error", got.status)
	}
}

func TestTitleErrorKeepsEditorOpen(t *testing.T) {
	model := Model{
		titleEdit:  true,
		titleInput: "new title",
		sessions:   []opencode.Session{{ID: "ses_example"}},
	}
	updated, _ := model.Update(titleUpdateFailedMsg{sessionID: "ses_example", err: errors.New("save failed")})
	got := updated.(Model)
	if !got.titleEdit || got.titleInput != "new title" {
		t.Fatalf("title editor state was lost after save error: edit=%v input=%q", got.titleEdit, got.titleInput)
	}
	if got.status != "save failed" {
		t.Fatalf("status = %q, want operation error", got.status)
	}
}

func TestStaleSessionsErrorDoesNotReplaceCurrentResults(t *testing.T) {
	model := Model{query: "new", loading: true, status: "loading"}
	updated, _ := model.Update(sessionsLoadFailedMsg{query: "old", err: errors.New("old query failed")})
	got := updated.(Model)
	if !got.loading || got.err != nil || got.status != "loading" {
		t.Fatalf("stale error changed current state: loading=%v err=%v status=%q", got.loading, got.err, got.status)
	}
}

func TestCurrentSessionsErrorStopsLoading(t *testing.T) {
	wantErr := errors.New("list failed")
	model := Model{query: "current", loading: true}
	updated, _ := model.Update(sessionsLoadFailedMsg{query: "current", err: wantErr})
	got := updated.(Model)
	if got.loading || !errors.Is(got.err, wantErr) || got.status != wantErr.Error() {
		t.Fatalf("current list error was not applied: loading=%v err=%v status=%q", got.loading, got.err, got.status)
	}
}

func TestPreviewErrorOnlyAffectsCurrentSession(t *testing.T) {
	model := Model{sessions: []opencode.Session{{ID: "ses_current"}}, status: "unchanged"}
	updated, _ := model.Update(previewLoadFailedMsg{sessionID: "ses_other", err: errors.New("stale preview")})
	got := updated.(Model)
	if got.status != "unchanged" {
		t.Fatalf("stale preview error changed status: %q", got.status)
	}
	updated, _ = got.Update(previewLoadFailedMsg{sessionID: "ses_current", err: errors.New("preview failed")})
	got = updated.(Model)
	if got.status != "preview failed" {
		t.Fatalf("current preview error was not shown: %q", got.status)
	}
}

func TestCopyKeyReturnsAsyncCommand(t *testing.T) {
	model := Model{
		sessions: []opencode.Session{{ID: "ses_example"}},
		texts:    NewTexts("en"),
	}
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	got := updated.(Model)
	if cmd == nil {
		t.Fatal("copy key did not return an asynchronous command")
	}
	if got.status != got.texts.CopyingSessionID {
		t.Fatalf("status = %q, want copying status", got.status)
	}
}

func TestClipboardMessagesUpdateStatus(t *testing.T) {
	model := Model{texts: NewTexts("en")}
	updated, _ := model.Update(clipboardCopiedMsg{})
	got := updated.(Model)
	if got.status != got.texts.CopiedSessionID {
		t.Fatalf("success status = %q", got.status)
	}
	updated, _ = got.Update(clipboardCopyFailedMsg{sessionID: "ses_example", err: errors.New("copy failed")})
	got = updated.(Model)
	if !strings.Contains(got.status, "ses_example") {
		t.Fatalf("failure status does not expose fallback ID: %q", got.status)
	}
}

func TestSearchDebounceOnlyLoadsLatestVersion(t *testing.T) {
	model := Model{mode: ModeSearch, texts: NewTexts("en")}
	updated, firstCmd := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	first := updated.(Model)
	if firstCmd == nil || first.query != "a" || !first.searchPending {
		t.Fatalf("first key did not schedule debounce: query=%q pending=%v", first.query, first.searchPending)
	}

	updated, secondCmd := first.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	second := updated.(Model)
	if secondCmd == nil || second.query != "ab" || second.searchVersion <= first.searchVersion {
		t.Fatalf("second key did not replace debounce: query=%q version=%d", second.query, second.searchVersion)
	}

	updated, staleCmd := second.Update(searchDebounceMsg{version: first.searchVersion})
	afterStale := updated.(Model)
	if staleCmd != nil || !afterStale.searchPending {
		t.Fatal("stale debounce message started a query")
	}

	updated, currentCmd := afterStale.Update(searchDebounceMsg{version: second.searchVersion})
	afterCurrent := updated.(Model)
	if currentCmd == nil || afterCurrent.searchPending {
		t.Fatal("current debounce message did not start the latest query")
	}
}

func TestSearchEnterFlushesPendingQuery(t *testing.T) {
	model := Model{
		mode:          ModeSearch,
		query:         "opencode",
		searchVersion: 3,
		searchPending: true,
	}
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)
	if got.mode != ModeBrowse || got.searchPending || cmd == nil {
		t.Fatalf("enter did not flush pending search: mode=%v pending=%v cmd=%v", got.mode, got.searchPending, cmd != nil)
	}
	if got.searchVersion != 4 {
		t.Fatalf("version = %d, want pending timer invalidated", got.searchVersion)
	}
}

func TestSearchEscapeClearsImmediatelyAndInvalidatesDebounce(t *testing.T) {
	model := Model{
		mode:          ModeSearch,
		query:         "opencode",
		searchVersion: 7,
		searchPending: true,
	}
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(Model)
	if got.mode != ModeBrowse || got.query != "" || got.searchPending || cmd == nil {
		t.Fatalf("escape did not immediately clear search: mode=%v query=%q pending=%v cmd=%v", got.mode, got.query, got.searchPending, cmd != nil)
	}
	if got.searchVersion != 8 {
		t.Fatalf("version = %d, want pending timer invalidated", got.searchVersion)
	}
}
