package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

func (m Model) renderHelp() string {
	width := m.dialogWidth()
	contentWidth := max(1, width-4)
	keyWidth := 0
	for _, line := range m.texts.HelpLines {
		keyWidth = max(keyWidth, lipgloss.Width(line.Key))
	}
	keyWidth = min(keyWidth, max(1, contentWidth/2))
	descriptionWidth := max(1, contentWidth-keyWidth-2)
	body := make([]string, 0, len(m.texts.HelpLines)+1)
	for _, line := range m.texts.HelpLines {
		key := truncateWidth(line.Key, keyWidth)
		key += strings.Repeat(" ", max(0, keyWidth-lipgloss.Width(key)))
		body = append(body, m.styles.ModalKey.Render(key)+"  "+m.styles.ModalText.Render(truncateWidth(line.Description, descriptionWidth)))
	}
	body = append(body, m.styles.ModalMuted.Render(truncateWidth(m.texts.ReadOnlyNotice, contentWidth)))
	return m.renderDialog(m.texts.HelpTitle, body, width)
}

func (m Model) renderTitleDialog() string {
	width := m.dialogWidth()
	contentWidth := max(1, width-4)
	inputWidth := max(1, contentWidth-3)
	input := tailWidth(m.titleInput, max(1, inputWidth-1)) + m.styles.ModalKey.Render("▏")
	inputLine := m.styles.ModalIcon.Render("✎") + " " + m.styles.ModalText.Render(input)
	hint := m.styles.ModalKey.Render("Enter") + " " + m.styles.ModalMuted.Render(m.texts.SaveAction) +
		"  " + m.styles.ModalKey.Render("Esc") + " " + m.styles.ModalMuted.Render(m.texts.CancelAction)
	return m.renderDialog(m.texts.SaveAsTitle, []string{
		inputLine,
		hint,
	}, width)
}

func (m Model) renderDeleteDialog() string {
	width := m.dialogWidth()
	contentWidth := max(1, width-4)
	title := ""
	id := m.currentID()
	if m.selected >= 0 && m.selected < len(m.sessions) {
		title = m.sessions[m.selected].Title
	}
	target := fmt.Sprintf("%s: %s", m.texts.DeleteTarget, truncateWidth(title, max(1, contentWidth-lipgloss.Width(m.texts.DeleteTarget)-2)))
	hint := m.styles.ModalKey.Render("y / Enter") + " " + m.styles.ModalMuted.Render(m.texts.DeleteAction) +
		"  " + m.styles.ModalKey.Render("n / Esc") + " " + m.styles.ModalMuted.Render(m.texts.CancelAction)
	impact := []string{}
	if m.deleteLoading {
		impact = append(impact, m.styles.ModalMuted.Render(m.texts.LoadingDeleteImpact))
		hint = m.styles.ModalKey.Render("Esc") + " " + m.styles.ModalMuted.Render(m.texts.CancelAction)
	} else if m.deleteErr != nil {
		impact = append(impact, m.styles.Error.Render(truncateWidth(m.deleteErr.Error(), contentWidth)))
		hint = m.styles.ModalKey.Render("Esc") + " " + m.styles.ModalMuted.Render(m.texts.CancelAction)
	} else {
		impact = append(impact,
			m.styles.ModalText.Render(fmt.Sprintf("%s: %d", m.texts.DeleteSessions, m.deleteImpact.SessionCount)),
			m.styles.ModalText.Render(fmt.Sprintf("%s: %d", m.texts.DeleteMessages, m.deleteImpact.MessageCount)),
			m.styles.ModalText.Render(fmt.Sprintf("%s: %d", m.texts.DeleteParts, m.deleteImpact.PartCount)),
		)
	}
	if m.deleteBusy {
		hint = m.styles.ModalMuted.Render(m.texts.DeletingSession)
	}
	body := []string{
		m.styles.ModalWarn.Render(truncateWidth(m.texts.DeleteWarning, contentWidth)),
		m.styles.ModalText.Render(target),
		m.styles.ModalText.Faint(true).Render(truncateWidth(id, contentWidth)),
	}
	body = append(body, impact...)
	body = append(body, ansi.Truncate(hint, contentWidth, ""))
	return m.renderDialog(m.texts.DeleteDialogTitle, body, width)
}

func (m Model) dialogWidth() int {
	if m.width <= 4 {
		return max(0, m.width)
	}
	return min(max(44, m.width*3/4), m.width-2)
}

func (m Model) renderDialog(title string, body []string, width int) string {
	if width <= 0 {
		return ""
	}
	borderStyle := m.styles.ModalBorder
	if width == 1 {
		return borderStyle.Render("│")
	}

	innerWidth := width - 2
	topContent := strings.Repeat("─", innerWidth)
	if innerWidth >= 5 {
		title = truncateWidth(title, innerWidth-4)
		legend := " " + title + " "
		left := max(1, (innerWidth-lipgloss.Width(legend))/2)
		right := max(0, innerWidth-lipgloss.Width(legend)-left)
		topContent = strings.Repeat("─", left) + legend + strings.Repeat("─", right)
	}
	top := borderStyle.Render("┌" + topContent + "┐")
	lines := []string{top}
	for _, line := range body {
		content := strings.Repeat(" ", innerWidth)
		if innerWidth >= 2 {
			contentWidth := innerWidth - 2
			line = fitANSIWidth(line, contentWidth)
			content = " " + line + " "
		}
		cell := m.styles.ModalBG.Inline(true).Render(content)
		lines = append(lines, borderStyle.Render("│")+cell+borderStyle.Render("│"))
	}
	lines = append(lines, borderStyle.Render("└"+strings.Repeat("─", innerWidth)+"┘"))
	return strings.Join(lines, "\n")
}

func (m Model) renderOverlay(base string, modal string) string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	screen := cellbuf.NewBuffer(m.width, m.height)
	cellbuf.SetContent(screen, base)
	dimCells(screen)

	modalWidth := min(lipgloss.Width(modal), m.width)
	modalHeight := min(len(strings.Split(modal, "\n")), m.height)
	if modalWidth > 0 && modalHeight > 0 {
		x := max(0, (m.width-modalWidth)/2)
		y := max(0, (m.height-modalHeight)/2)
		rect := cellbuf.Rect(x, y, modalWidth, modalHeight)
		cellbuf.SetContentRect(screen, modal, rect)
		applyBackground(screen, rect, modalBackground(m.styles.ModalBG))
	}

	lines := make([]string, m.height)
	for row := range lines {
		width, line := cellbuf.RenderLine(screen, row)
		lines[row] = line + strings.Repeat(" ", max(0, m.width-width))
	}
	return strings.Join(lines, "\n")
}

func fitANSIWidth(value string, width int) string {
	if width <= 0 {
		return ""
	}
	value = strings.NewReplacer("\r", " ", "\n", " ", "\t", "    ").Replace(value)
	value = ansi.Truncate(value, width, "")
	return value + strings.Repeat(" ", max(0, width-ansi.StringWidth(value)))
}

func dimCells(buffer *cellbuf.Buffer) {
	for y, line := range buffer.Lines {
		for x, cell := range line {
			if cell == nil || cell.Width == 0 {
				continue
			}
			cell = cell.Clone()
			cell.Style.Attrs |= cellbuf.FaintAttr
			buffer.SetCell(x, y, cell)
		}
	}
}

func modalBackground(style lipgloss.Style) ansi.Color {
	sample := cellbuf.NewBuffer(1, 1)
	cellbuf.SetContent(sample, style.Inline(true).Render(" "))
	return sample.Cell(0, 0).Style.Bg
}

func applyBackground(buffer *cellbuf.Buffer, rect cellbuf.Rectangle, background ansi.Color) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			cell := buffer.Cell(x, y)
			if cell == nil || cell.Width == 0 {
				continue
			}
			cell = cell.Clone()
			cell.Style.Bg = background
			buffer.SetCell(x, y, cell)
		}
	}
}
