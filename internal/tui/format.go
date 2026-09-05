package tui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/lipgloss"

	"github.com/asikeida/lazyOpencodeSession/internal/opencode"
)

func (m *Model) applySearch() {
	terms := opencode.SearchTerms(m.query)
	m.memoryMatches = map[string]string{}
	if len(terms) == 0 {
		m.sessions = append(m.sessions[:0], m.catalog...)
		m.matched = len(m.sessions)
		return
	}
	filtered := make([]opencode.Session, 0, len(m.catalog))
	for _, session := range m.catalog {
		metadata := strings.ToLower(strings.Join([]string{session.ID, session.ProjectID, session.Title, session.Directory, session.Model, session.Agent}, " "))
		if !containsAllTermsAcross(metadata, m.memorySearch[session.ID], terms) {
			continue
		}
		filtered = append(filtered, session)
		if !containsAllTerms(metadata, terms) {
			m.memoryMatches[session.ID] = bestMemorySnippet(m.memories[session.ID], terms)
		}
	}
	m.sessions = filtered
	m.matched = len(filtered)
}

func containsAllTermsAcross(metadata string, memory string, terms []string) bool {
	for _, term := range terms {
		if !strings.Contains(metadata, term) && !strings.Contains(memory, term) {
			return false
		}
	}
	return true
}

func containsAllTerms(value string, terms []string) bool {
	for _, term := range terms {
		if !strings.Contains(value, term) {
			return false
		}
	}
	return true
}

func bestMemorySnippet(memories []opencode.UserMemory, terms []string) string {
	best := ""
	bestScore := 0
	for _, memory := range memories {
		text := sanitizeSingleLine(memory.Text)
		lower := strings.ToLower(text)
		score := 0
		for _, term := range terms {
			if strings.Contains(lower, term) {
				score++
			}
		}
		if score > bestScore {
			best = contextualSnippet(text, terms, 180)
			bestScore = score
		}
	}
	return best
}

func sanitizeSingleLine(value string) string {
	return strings.Join(strings.FieldsFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}), " ")
}

func sanitizeTitleInput(value string) string {
	return sanitizeSingleLine(value)
}

func sanitizeTitleFragment(value string) string {
	var out []rune
	for _, r := range value {
		switch {
		case unicode.IsControl(r):
			continue
		case r == '\n' || r == '\r' || r == '\t':
			out = append(out, ' ')
		default:
			out = append(out, r)
		}
	}
	return string(out)
}

func contextualSnippet(text string, terms []string, maxRunes int) string {
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	lower := []rune(strings.ToLower(text))
	position := 0
	found := false
	for _, term := range terms {
		if index := runeIndex(lower, []rune(term)); index >= 0 {
			if !found || index < position {
				position = index
				found = true
			}
		}
	}
	start := max(0, position-24)
	end := min(len(runes), start+maxRunes)
	if end-start < maxRunes {
		start = max(0, end-maxRunes)
	}
	snippet := string(runes[start:end])
	if start > 0 {
		snippet = "…" + snippet
	}
	if end < len(runes) {
		snippet += "…"
	}
	return snippet
}

func runeIndex(value []rune, term []rune) int {
	if len(term) == 0 {
		return 0
	}
	for i := 0; i+len(term) <= len(value); i++ {
		match := true
		for j := range term {
			if value[i+j] != term[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func fieldPrefix(label string) string {
	padding := strings.Repeat(" ", max(1, 12-lipgloss.Width(label)))
	return label + padding
}

func fieldValueIndent(label string) string {
	return strings.Repeat(" ", lipgloss.Width(fieldPrefix(label)))
}

func formatModel(value string) []string {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(value), &fields); err != nil || len(fields) == 0 {
		return []string{value}
	}
	keys := make([]string, 0, len(fields))
	for _, key := range []string{"id", "providerID", "variant"} {
		if _, ok := fields[key]; ok {
			keys = append(keys, key)
		}
	}
	var extra []string
	for key := range fields {
		if key != "id" && key != "providerID" && key != "variant" {
			extra = append(extra, key)
		}
	}
	sort.Strings(extra)
	keys = append(keys, extra...)
	if len(keys) == 0 {
		return []string{value}
	}
	keyWidth := 0
	keyLabels := make(map[string]string, len(keys))
	for _, key := range keys {
		label := `"` + key + `":`
		keyLabels[key] = label
		keyWidth = max(keyWidth, lipgloss.Width(label))
	}
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		value := fields[key]
		var text string
		if err := json.Unmarshal(value, &text); err == nil {
			text = strconv.Quote(text)
		} else {
			text = string(value)
		}
		lines = append(lines, fmt.Sprintf("%*s %s", keyWidth, keyLabels[key], text))
	}
	return lines
}

func renderHelpLines(lines []HelpLine) string {
	keyWidth := 0
	for _, line := range lines {
		keyWidth = max(keyWidth, lipgloss.Width(line.Key))
	}
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		padding := strings.Repeat(" ", keyWidth-lipgloss.Width(line.Key)+4)
		result = append(result, line.Key+padding+line.Description)
	}
	return strings.Join(result, "\n")
}

func normalizeDetailFields(fields map[string]bool) map[string]bool {
	defaults := map[string]bool{
		"title":          true,
		"session":        true,
		"project":        true,
		"directory":      true,
		"path_status":    true,
		"message_count":  false,
		"part_count":     false,
		"size":           false,
		"large_session":  false,
		"updated":        true,
		"created":        true,
		"model":          true,
		"agent":          true,
		"cost":           true,
		"tokens":         true,
		"resume_command": false,
	}
	for key, value := range fields {
		defaults[key] = value
	}
	return defaults
}

func (m Model) visibleItems() int {
	reserved := 5
	if m.query != "" || m.mode == ModeSearch {
		reserved = 6
	}
	linesPerItem := 2
	if m.query != "" {
		linesPerItem = 3
	}
	return max(1, (m.height-reserved)/linesPerItem)
}

func renderField(name string, value string, width int) []string {
	label := name + ":"
	labelWidth := 12
	padding := strings.Repeat(" ", max(1, labelWidth-lipgloss.Width(label)))
	prefix := label + padding
	continuation := strings.Repeat(" ", lipgloss.Width(prefix))
	valueWidth := max(8, width-lipgloss.Width(prefix))
	wrapped := wrapAll(value, valueWidth)
	if len(wrapped) == 0 {
		return []string{prefix}
	}
	lines := make([]string, 0, len(wrapped))
	lines = append(lines, prefix+wrapped[0])
	for _, line := range wrapped[1:] {
		lines = append(lines, continuation+line)
	}
	return lines
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func formatShortTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	now := time.Now()
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return t.Format("15:04")
	}
	return t.Format("01-02 15:04")
}

func formatFullTime(t time.Time, weekdays [7]string) string {
	if t.IsZero() {
		return "-"
	}
	weekday := ""
	if int(t.Weekday()) < len(weekdays) {
		weekday = weekdays[t.Weekday()]
	}
	return fmt.Sprintf("%s %s %s", t.Format("2006-01-02"), weekday, t.Format("15:04:05"))
}

func formatBytes(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	units := []string{"KB", "MB", "GB", "TB"}
	value := float64(size)
	for _, unit := range units {
		value /= 1024
		if value < 1024 {
			return fmt.Sprintf("%.1f %s", value, unit)
		}
	}
	return fmt.Sprintf("%.1f PB", value/1024)
}

func truncateWidth(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= width {
		return value
	}
	runes := []rune(value)
	for len(runes) > 0 && lipgloss.Width(string(runes)+"...") > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "..."
}

func tailWidth(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= width {
		return value
	}
	runes := []rune(value)
	for len(runes) > 0 && lipgloss.Width("…"+string(runes)) > width {
		runes = runes[1:]
	}
	return "…" + string(runes)
}

func highlightText(value string, terms []string, normal lipgloss.Style, match lipgloss.Style) string {
	if len(terms) == 0 || value == "" {
		return normal.Render(value)
	}
	lower := strings.ToLower(value)
	var out strings.Builder
	pos := 0
	for pos < len(value) {
		bestStart := -1
		bestEnd := -1
		for _, term := range terms {
			idx := strings.Index(lower[pos:], term)
			if idx < 0 {
				continue
			}
			start := pos + idx
			end := start + len(term)
			if bestStart == -1 || start < bestStart || start == bestStart && end > bestEnd {
				bestStart = start
				bestEnd = end
			}
		}
		if bestStart < 0 {
			out.WriteString(normal.Render(value[pos:]))
			break
		}
		if bestStart > pos {
			out.WriteString(normal.Render(value[pos:bestStart]))
		}
		out.WriteString(match.Render(value[bestStart:bestEnd]))
		pos = bestEnd
	}
	return out.String()
}

func wrap(value string, width int, maxLines int) []string {
	lines := wrapAll(value, width)
	if len(lines) > maxLines {
		return lines[:maxLines]
	}
	return lines
}

func wrapAll(value string, width int) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	lines := []string{}
	line := ""
	for _, r := range []rune(value) {
		if r == '\n' {
			if line != "" {
				lines = append(lines, line)
				line = ""
			}
			continue
		}
		next := line + string(r)
		if lipgloss.Width(next) > width && line != "" {
			lines = append(lines, truncateWidth(line, width))
			line = string(r)
			continue
		}
		line = next
	}
	if line != "" {
		lines = append(lines, truncateWidth(line, width))
	}
	return lines
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
