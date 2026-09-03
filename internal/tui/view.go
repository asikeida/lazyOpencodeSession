package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/asikeida/lazyOpencodeSession/internal/opencode"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return m.texts.Loading
	}
	bodyHeight := max(1, m.height-1)
	content := ""
	if m.width < 110 {
		content = m.renderSessions(m.width, bodyHeight)
	} else {
		leftW := int(float64(m.width) * 0.45)
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
	box := lipgloss.NewStyle().Width(max(1, width-2)).Height(max(1, height-2)).Border(lipgloss.RoundedBorder()).Inherit(m.styles.Panel).Inherit(m.styles.Border)
	lines := []string{m.styles.Title.Render(m.texts.AppTitle + " - " + m.texts.SessionsTitle)}
	if m.query != "" || m.mode == ModeSearch {
		prompt := "/ " + m.query
		if m.mode == ModeSearch {
			prompt += "_"
		}
		lines = append(lines, m.styles.Accent.Render(prompt))
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
				style = m.styles.Selected
				matchStyle = m.styles.Selected.Copy().Foreground(lipgloss.Color("229")).Underline(true)
			}
			title := truncateWidth(s.Title, width-10)
			lines = append(lines, style.Render(prefix)+highlightText(title, m.searchTerms(), style, matchStyle))
			meta := fmt.Sprintf("  %s  %s", truncateWidth(s.Directory, width-18), formatShortTime(s.UpdatedAt))
			lines = append(lines, highlightText(meta, m.searchTerms(), m.styles.Muted, m.styles.Match))
		}
	}
	return box.Render(strings.Join(lines, "\n"))
}

func (m Model) renderDetails(width int, height int) string {
	box := lipgloss.NewStyle().Width(max(1, width-2)).Height(max(1, height-2)).Border(lipgloss.RoundedBorder()).Inherit(m.styles.Panel).Inherit(m.styles.Border)
	lines := []string{m.styles.Title.Render(m.texts.DetailsTitle)}
	if len(m.sessions) == 0 || m.selected >= len(m.sessions) {
		lines = append(lines, m.styles.Muted.Render(m.texts.SelectSession))
		return box.Render(strings.Join(lines, "\n"))
	}

	s := m.sessions[m.selected]
	contentWidth := max(20, width-4)
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
	m.appendDetailField(&lines, "resume_command", m.texts.FieldResumeCommand, m.opencode+" --session "+s.ID, contentWidth)
	lines = append(lines, "", m.styles.Muted.Render(m.texts.ActionHint))

	if m.previewFor == s.ID {
		lines = append(lines, "", m.styles.Accent.Render(m.texts.RecentUserMessages))
		if len(m.preview) == 0 {
			lines = append(lines, m.styles.Muted.Render(m.texts.NoTextPreview))
		}
		for i, msg := range m.preview {
			lines = append(lines, m.styles.Muted.Render(fmt.Sprintf("%d. %s", i+1, formatShortTime(msg.CreatedAt))))
			for _, line := range wrap(msg.Text, max(20, width-8), 4) {
				lines = append(lines, "  "+line)
			}
		}
	}

	return box.Render(strings.Join(lines, "\n"))
}

func (m Model) renderStatus() string {
	mode := m.texts.FooterBrowse
	if m.mode == ModeSearch {
		mode = m.texts.FooterSearch
	}
	text := fmt.Sprintf(" %s | %s | %s ", mode, m.status, m.texts.FooterHelp)
	return m.styles.Status.Width(m.width).Render(truncateWidth(text, m.width))
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
