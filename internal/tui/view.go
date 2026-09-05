package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/asikeida/lazyOpencodeSession/internal/opencode"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return m.texts.Loading
	}
	bodyHeight := max(1, m.height-1)
	content := ""
	ui := m.normalizedUI()
	if m.width < ui.TwoPaneMinWidth {
		content = m.renderSessions(m.width, bodyHeight)
	} else {
		leftW := int(float64(m.width) * ui.SplitRatio)
		rightW := m.width - leftW
		left := m.renderSessions(leftW, bodyHeight)
		right := m.renderDetails(rightW, bodyHeight)
		content = lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	}
	view := lipgloss.JoinVertical(lipgloss.Left, content, m.renderStatus())
	if m.help {
		return m.renderOverlay(view, m.renderHelp())
	}
	if m.titleEdit {
		return m.renderOverlay(view, m.renderTitleDialog())
	}
	if m.deleteConfirm {
		return m.renderOverlay(view, m.renderDeleteDialog())
	}
	return view
}

func (m Model) renderSessions(width int, height int) string {
	box := m.panelBox(width, height, m.focus == FocusSessions)
	lines := []string{m.styles.Title.Render(m.texts.AppTitle + " - " + m.texts.SessionsTitle)}
	if m.query != "" || m.mode == ModeSearch {
		prompt := "/ " + m.query
		if m.mode == ModeSearch {
			prompt += "_"
		}
		line := m.styles.Accent.Render(prompt)
		if m.recentDays > 0 {
			memory := fmt.Sprintf(m.texts.MemoryWindow, m.recentDays)
			if m.memoryLoading {
				memory += " " + m.texts.MemoryLoading
			} else if m.memoryErr != nil {
				memory += " " + m.texts.MemoryUnavailable
			}
			line += "  " + m.styles.MemorySnippet.Render("["+memory+"]")
		}
		lines = append(lines, line)
	}
	if m.loading {
		lines = append(lines, m.styles.Muted.Render(m.texts.Loading))
	} else if m.err != nil {
		lines = append(lines, m.styles.Error.Render(m.err.Error()))
	} else if len(m.sessions) == 0 {
		lines = append(lines, m.styles.Muted.Render(m.texts.NoSessions))
	} else {
		for i := m.offset; i < min(len(m.sessions), m.offset+m.visibleItems()); i++ {
			s := m.sessions[i]
			prefix := "  "
			style := m.styles.Base
			matchStyle := m.styles.Match
			if i == m.selected {
				prefix = "> "
				if m.focus == FocusSessions {
					style = m.styles.Selected
					matchStyle = m.styles.SelectedMatch
				} else {
					style = m.styles.InactiveSelected
					matchStyle = m.styles.InactiveSelectedMatch
				}
			}
			title := truncateWidth(s.Title, width-10)
			lines = append(lines, style.Render(prefix)+highlightText(title, m.searchTerms(), style, matchStyle))
			meta := fmt.Sprintf("  %s  %s", truncateWidth(s.Directory, width-18), formatShortTime(s.UpdatedAt))
			lines = append(lines, highlightText(meta, m.searchTerms(), m.styles.Muted, m.styles.Match))
			if snippet := m.memoryMatches[s.ID]; snippet != "" {
				snippet = truncateWidth(snippet, max(1, width-8))
				lines = append(lines, m.styles.MemorySnippet.Render("  │ ")+highlightText(snippet, m.searchTerms(), m.styles.MemorySnippet, m.styles.Match))
			}
		}
	}
	return box.Render(strings.Join(lines, "\n"))
}

func (m Model) renderDetails(width int, height int) string {
	box := m.panelBox(width, height, m.focus == FocusDetails)
	contentWidth := max(20, width-4)
	allLines := m.detailLines(contentWidth)
	viewportHeight := max(1, height-4)
	offset := min(max(0, m.detailsOffset), max(0, len(allLines)-viewportHeight))
	visible := allLines
	if len(visible) > viewportHeight {
		visible = visible[offset : offset+viewportHeight]
	}
	body := make([]string, 0, height-2)
	body = append(body, m.styles.Title.Render(m.texts.DetailsTitle))
	body = append(body, visible...)
	for len(body) < height-3 {
		body = append(body, "")
	}
	foot := m.texts.DetailsScrollHint + "  " + m.detailsProgress(len(allLines), viewportHeight, offset)
	body = append(body, m.styles.DetailsHint.Render(truncateWidth(foot, contentWidth)))
	return box.Render(strings.Join(body, "\n"))
}

func (m Model) renderStatus() string {
	mode := m.texts.FooterBrowse
	if m.mode == ModeSearch {
		mode = m.texts.FooterSearch
	}
	text := m.styles.StatusMode.Render(" "+mode) + m.styles.StatusText.Render(" | "+m.status+" | ") + m.styles.StatusKey.Render(m.texts.FooterHelp+" ")
	return lipgloss.NewStyle().Width(m.width).Render(ansi.Truncate(text, m.width, ""))
}

func (m Model) detailLines(contentWidth int) []string {
	lines := []string{}
	if len(m.sessions) == 0 || m.selected >= len(m.sessions) {
		return append(lines, m.styles.Muted.Render(m.texts.SelectSession))
	}
	s := m.sessions[m.selected]
	m.appendDetailField(&lines, "title", m.texts.FieldTitle, s.Title, contentWidth)
	m.appendDetailField(&lines, "session", m.texts.FieldSession, s.ID, contentWidth)
	m.appendDetailField(&lines, "project", m.texts.FieldProject, s.ProjectID, contentWidth)
	m.appendDetailField(&lines, "directory", m.texts.FieldDirectory, s.Directory, contentWidth)
	m.appendDetailField(&lines, "path_status", m.texts.FieldPathStatus, m.pathStatusText(s.DirectoryExists), contentWidth)
	m.appendDetailField(&lines, "message_count", m.texts.FieldMessageCount, m.statsField(s.ID, func(stats opencode.SessionStats) string {
		return strconv.FormatInt(stats.MessageCount, 10)
	}), contentWidth)
	m.appendDetailField(&lines, "part_count", m.texts.FieldPartCount, m.statsField(s.ID, func(stats opencode.SessionStats) string {
		return strconv.FormatInt(stats.PartCount, 10)
	}), contentWidth)
	m.appendDetailField(&lines, "size", m.texts.FieldSize, m.statsField(s.ID, func(stats opencode.SessionStats) string {
		return formatBytes(stats.SizeBytes)
	}), contentWidth)
	m.appendDetailField(&lines, "large_session", m.texts.FieldLargeSession, m.statsField(s.ID, func(stats opencode.SessionStats) string {
		if stats.SizeBytes >= largeSessionBytes {
			return m.texts.Yes
		}
		return m.texts.No
	}), contentWidth)
	m.appendDetailField(&lines, "updated", m.texts.FieldUpdated, formatFullTime(s.UpdatedAt, m.texts.Weekdays), contentWidth)
	m.appendDetailField(&lines, "created", m.texts.FieldCreated, formatFullTime(s.CreatedAt, m.texts.Weekdays), contentWidth)
	m.appendModelField(&lines, s.Model, contentWidth)
	m.appendDetailField(&lines, "agent", m.texts.FieldAgent, emptyDash(s.Agent), contentWidth)
	m.appendDetailField(&lines, "cost", m.texts.FieldCost, fmt.Sprintf("$%.4f", s.Cost), contentWidth)
	m.appendDetailField(&lines, "tokens", m.texts.FieldTokens, fmt.Sprintf(m.texts.TokensFormat, s.TokensInput, s.TokensOutput, s.TokensReasoning, s.TokensCacheRead), contentWidth)
	m.appendDetailField(&lines, "resume_command", m.texts.FieldResumeCommand, m.resume.CommandLine(s.ID), contentWidth)
	lines = append(lines, "", m.styles.DetailsHint.Render(m.texts.ActionHint))
	if m.previewFor == s.ID {
		lines = append(lines, "", m.styles.Accent.Render(m.texts.RecentUserMessages))
		if len(m.preview) == 0 {
			lines = append(lines, m.styles.Muted.Render(m.texts.NoTextPreview))
		}
		for i, msg := range m.preview {
			lines = append(lines, m.styles.PreviewTimestamp.Render(fmt.Sprintf("%d. %s", i+1, formatShortTime(msg.CreatedAt))))
			for _, line := range wrapAll(msg.Text, max(20, contentWidth-2)) {
				lines = append(lines, "  "+line)
			}
		}
	}
	return lines
}

func (m Model) detailsProgress(total int, visible int, offset int) string {
	if total <= 0 {
		return "0/0"
	}
	if total <= visible {
		return fmt.Sprintf("%d/%d", total, total)
	}
	return fmt.Sprintf("%d/%d", min(total, offset+visible), total)
}

func (m Model) panelBox(width int, height int, focused bool) lipgloss.Style {
	border := m.styles.InactiveBorder
	if focused {
		border = m.styles.ActiveBorder
		if m.mode == ModeSearch && m.focus == FocusSessions {
			border = m.styles.SearchingActiveBorder
		}
	}
	return lipgloss.NewStyle().Width(max(1, width-2)).Height(max(1, height-2)).Border(m.borderStyle()).Inherit(m.styles.Panel).Inherit(border)
}

func (m Model) borderStyle() lipgloss.Border {
	switch m.normalizedUI().BorderStyle {
	case "single":
		return lipgloss.NormalBorder()
	case "double":
		return lipgloss.DoubleBorder()
	case "hidden":
		return lipgloss.HiddenBorder()
	case "bold":
		return lipgloss.ThickBorder()
	default:
		return lipgloss.RoundedBorder()
	}
}

func (m Model) normalizedUI() UIConfig {
	return NormalizeUIConfig(m.ui)
}

func (m Model) appendDetailField(lines *[]string, id string, label string, value string, width int) {
	if !m.fieldEnabled(id) {
		return
	}
	*lines = append(*lines, renderField(label, value, width)...)
}

func (m Model) appendModelField(lines *[]string, value string, width int) {
	if !m.fieldEnabled("model") {
		return
	}
	modelLines := formatModel(value)
	if len(modelLines) <= 1 {
		m.appendDetailField(lines, "model", m.texts.FieldModel, emptyDash(value), width)
		return
	}
	label := m.texts.FieldModel + ":"
	valueIndent := fieldValueIndent(label)
	*lines = append(*lines, fieldPrefix(label))
	for _, line := range modelLines {
		for _, wrapped := range wrapAll(line, max(8, width-lipgloss.Width(valueIndent))) {
			*lines = append(*lines, valueIndent+wrapped)
		}
	}
}
