package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"

	"github.com/asikeida/lazyOpencodeSession/internal/opencode"
	"github.com/asikeida/lazyOpencodeSession/internal/platform"
)

type Options struct {
	Repo            opencode.Repository
	Limit           int
	Language        string
	OpenCodeCommand string
	DetailFields    map[string]bool
	ReadOnly        bool
}

type Mode int

const (
	ModeBrowse Mode = iota
	ModeSearch
)

const largeSessionBytes = 10 * 1024 * 1024

type Model struct {
	repo          opencode.Repository
	limit         int
	opencode      string
	readOnly      bool
	fields        map[string]bool
	styles        Styles
	texts         Texts
	width         int
	height        int
	sessions      []opencode.Session
	stats         map[string]opencode.SessionStats
	statsBusy     map[string]bool
	matched       int
	total         int
	selected      int
	offset        int
	query         string
	mode          Mode
	loading       bool
	status        string
	err           error
	help          bool
	titleEdit     bool
	titleInput    string
	deleteConfirm bool
	deleteBusy    bool
	preview       []opencode.MessagePreview
	previewFor    string
	resumeID      string
}

type sessionsLoadedMsg struct {
	query    string
	sessions []opencode.Session
	matched  int
	total    int
}

type previewLoadedMsg struct {
	sessionID string
	messages  []opencode.MessagePreview
}

type statsLoadedMsg struct {
	sessionID string
	stats     opencode.SessionStats
}

type titleUpdatedMsg struct {
	sessionID string
	title     string
}

type sessionDeletedMsg struct {
	sessionID string
}

type errMsg struct {
	err error
}

func New(opts Options) Model {
	if opts.Limit <= 0 {
		opts.Limit = 500
	}
	return Model{
		repo:      opts.Repo,
		limit:     opts.Limit,
		opencode:  defaultString(opts.OpenCodeCommand, "opencode"),
		readOnly:  opts.ReadOnly,
		fields:    normalizeDetailFields(opts.DetailFields),
		stats:     map[string]opencode.SessionStats{},
		statsBusy: map[string]bool{},
		styles:    NewStyles(),
		texts:     NewTexts(opts.Language),
		loading:   true,
		status:    NewTexts(opts.Language).LoadingSessions,
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadSessions()
}

func (m Model) ResumeSessionID() string {
	return m.resumeID
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureVisible()
		return m, nil
	case sessionsLoadedMsg:
		if msg.query != m.query {
			return m, nil
		}
		m.sessions = msg.sessions
		m.matched = msg.matched
		m.total = msg.total
		m.loading = false
		m.err = nil
		if m.selected >= len(m.sessions) {
			m.selected = max(0, len(m.sessions)-1)
		}
		m.ensureVisible()
		m.status = m.resultStatus()
		return m, m.maybeLoadStats()
	case statsLoadedMsg:
		m.stats[msg.sessionID] = msg.stats
		delete(m.statsBusy, msg.sessionID)
		return m, nil
	case titleUpdatedMsg:
		for i := range m.sessions {
			if m.sessions[i].ID == msg.sessionID {
				m.sessions[i].Title = msg.title
				break
			}
		}
		m.titleEdit = false
		m.titleInput = ""
		m.status = m.texts.TitleSaved
		return m, nil
	case sessionDeletedMsg:
		for i := range m.sessions {
			if m.sessions[i].ID == msg.sessionID {
				m.sessions = append(m.sessions[:i], m.sessions[i+1:]...)
				break
			}
		}
		delete(m.stats, msg.sessionID)
		delete(m.statsBusy, msg.sessionID)
		m.deleteConfirm = false
		m.deleteBusy = false
		m.preview = nil
		m.previewFor = ""
		m.selected = min(m.selected, max(0, len(m.sessions)-1))
		if m.total > 0 {
			m.total--
		}
		if m.matched > 0 {
			m.matched--
		}
		m.status = m.texts.SessionDeleted
		return m, m.maybeLoadStats()
	case previewLoadedMsg:
		if msg.sessionID == m.currentID() {
			m.preview = msg.messages
			m.previewFor = msg.sessionID
			m.status = fmt.Sprintf(m.texts.LoadedPreviewMessages, len(msg.messages))
		}
		return m, nil
	case errMsg:
		m.loading = false
		m.deleteBusy = false
		m.err = msg.err
		m.status = msg.err.Error()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.help {
		switch key.String() {
		case "esc", "?", "q", "enter":
			m.help = false
		}
		return m, nil
	}

	if m.titleEdit {
		switch key.String() {
		case "esc":
			m.titleEdit = false
			m.titleInput = ""
			m.status = m.texts.TitleCancelled
			return m, nil
		case "enter":
			if strings.TrimSpace(m.titleInput) == "" {
				m.status = m.texts.TitleEmpty
				return m, nil
			}
			id := m.currentID()
			m.status = m.texts.SavingTitle
			return m, m.saveTitle(id, strings.TrimSpace(m.titleInput))
		case "ctrl+c":
			return m, tea.Quit
		}
		switch key.Type {
		case tea.KeyBackspace:
			if len([]rune(m.titleInput)) > 0 {
				runes := []rune(m.titleInput)
				m.titleInput = string(runes[:len(runes)-1])
			}
		case tea.KeyRunes:
			m.titleInput += string(key.Runes)
		case tea.KeySpace:
			m.titleInput += " "
		}
		return m, nil
	}

	if m.deleteConfirm {
		if m.deleteBusy {
			return m, nil
		}
		switch key.String() {
		case "y", "enter":
			id := m.currentID()
			m.deleteBusy = true
			m.status = m.texts.DeletingSession
			return m, m.deleteSession(id)
		case "n", "esc":
			m.deleteConfirm = false
			m.status = m.texts.DeleteCancelled
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
		return m, nil
	}

	if m.mode == ModeSearch {
		switch key.String() {
		case "esc":
			m.mode = ModeBrowse
			m.query = ""
			m.selected = 0
			m.offset = 0
			m.preview = nil
			m.previewFor = ""
			m.loading = true
			return m, m.loadSessions()
		case "enter":
			m.mode = ModeBrowse
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		case "up", "ctrl+k":
			m.moveUp(1)
			return m, m.maybeLoadStats()
		case "down", "ctrl+j":
			m.moveDown(1)
			return m, m.maybeLoadStats()
		case "ctrl+p":
			if id := m.currentID(); id != "" {
				m.status = m.texts.LoadingPreview
				return m, m.loadPreview(id)
			}
			return m, nil
		case "pgup":
			m.moveUp(m.visibleItems())
			return m, m.maybeLoadStats()
		case "pgdown":
			m.moveDown(m.visibleItems())
			return m, m.maybeLoadStats()
		}
		switch key.Type {
		case tea.KeyBackspace:
			if len([]rune(m.query)) > 0 {
				r := []rune(m.query)
				m.query = string(r[:len(r)-1])
				m.selected = 0
				m.offset = 0
				m.preview = nil
				m.previewFor = ""
				m.loading = true
				return m, m.loadSessions()
			}
		case tea.KeyRunes:
			m.query += string(key.Runes)
			m.selected = 0
			m.offset = 0
			m.preview = nil
			m.previewFor = ""
			m.loading = true
			return m, m.loadSessions()
		case tea.KeySpace:
			m.query += " "
			m.selected = 0
			m.offset = 0
			m.preview = nil
			m.previewFor = ""
			m.loading = true
			return m, m.loadSessions()
		}
		return m, nil
	}

	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?":
		m.help = true
		return m, nil
	case "/":
		m.mode = ModeSearch
		m.status = m.texts.SearchSessions
		return m, nil
	case "e":
		if m.currentID() == "" {
			return m, nil
		}
		if m.readOnly {
			m.status = m.texts.TitleReadOnly
			return m, nil
		}
		m.titleEdit = true
		m.titleInput = m.sessions[m.selected].Title
		m.status = m.texts.EditingTitle
		return m, nil
	case "d":
		if m.currentID() == "" {
			return m, nil
		}
		if m.readOnly {
			m.status = m.texts.TitleReadOnly
			return m, nil
		}
		m.deleteConfirm = true
		m.status = m.texts.ConfirmDelete
		return m, nil
	case "r":
		m.loading = true
		m.status = m.texts.ReloadingSessions
		return m, m.loadSessions()
	case "up", "k":
		m.moveUp(1)
		return m, m.maybeLoadStats()
	case "down", "j":
		m.moveDown(1)
		return m, m.maybeLoadStats()
	case "pgup":
		m.moveUp(m.visibleItems())
		return m, m.maybeLoadStats()
	case "pgdown":
		m.moveDown(m.visibleItems())
		return m, m.maybeLoadStats()
	case "enter":
		if id := m.currentID(); id != "" {
			m.resumeID = id
			return m, tea.Quit
		}
	case "p":
		if id := m.currentID(); id != "" {
			m.status = m.texts.LoadingPreview
			return m, m.loadPreview(id)
		}
	case "y":
		if id := m.currentID(); id != "" {
			if err := platform.Copy(id); err != nil {
				m.status = fmt.Sprintf(m.texts.ClipboardUnavailable, id)
			} else {
				m.status = m.texts.CopiedSessionID
			}
		}
	}
	return m, nil
}

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
	if m.deleteBusy {
		hint = m.styles.ModalMuted.Render(m.texts.DeletingSession)
	}
	return m.renderDialog(m.texts.DeleteDialogTitle, []string{
		m.styles.ModalWarn.Render(truncateWidth(m.texts.DeleteWarning, contentWidth)),
		m.styles.ModalText.Render(target),
		m.styles.ModalText.Faint(true).Render(truncateWidth(id, contentWidth)),
		ansi.Truncate(hint, contentWidth, ""),
	}, width)
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

func (m Model) loadSessions() tea.Cmd {
	query := m.query
	limit := m.limit
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		sessions, err := m.repo.ListSessions(ctx, opencode.SessionFilter{Query: query, Limit: limit})
		if err != nil {
			return errMsg{err: err}
		}
		annotateDirectoryExists(sessions)
		matched, err := m.repo.CountSessions(ctx, opencode.SessionFilter{Query: query})
		if err != nil {
			return errMsg{err: err}
		}
		total := matched
		if strings.TrimSpace(query) != "" {
			total, err = m.repo.CountSessions(ctx, opencode.SessionFilter{})
			if err != nil {
				return errMsg{err: err}
			}
		}
		return sessionsLoadedMsg{query: query, sessions: sessions, matched: matched, total: total}
	}
}

func (m Model) loadPreview(sessionID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		messages, err := m.repo.RecentUserMessages(ctx, sessionID, 5, 500)
		if err != nil {
			return errMsg{err: err}
		}
		return previewLoadedMsg{sessionID: sessionID, messages: messages}
	}
}

func (m Model) loadStats(sessionID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		stats, err := m.repo.SessionStats(ctx, sessionID)
		if err != nil {
			return errMsg{err: err}
		}
		return statsLoadedMsg{sessionID: sessionID, stats: stats}
	}
}

func (m Model) saveTitle(sessionID string, title string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := m.repo.UpdateSessionTitle(ctx, sessionID, title); err != nil {
			return errMsg{err: err}
		}
		return titleUpdatedMsg{sessionID: sessionID, title: title}
	}
}

func (m Model) deleteSession(sessionID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := m.repo.DeleteSession(ctx, sessionID); err != nil {
			return errMsg{err: err}
		}
		return sessionDeletedMsg{sessionID: sessionID}
	}
}

func (m Model) currentID() string {
	if m.selected < 0 || m.selected >= len(m.sessions) {
		return ""
	}
	return m.sessions[m.selected].ID
}

func (m Model) resultStatus() string {
	if strings.TrimSpace(m.query) == "" {
		return fmt.Sprintf(m.texts.StatusSessions, m.total)
	}
	return fmt.Sprintf(m.texts.StatusMatched, m.matched, m.total)
}

func (m Model) searchTerms() []string {
	return opencode.SearchTerms(m.query)
}

func (m Model) pathStatusText(exists bool) string {
	if exists {
		return m.texts.PathExists
	}
	return m.texts.PathMissing
}

func (m Model) statsField(sessionID string, format func(opencode.SessionStats) string) string {
	stats, ok := m.stats[sessionID]
	if !ok {
		return m.texts.Loading
	}
	return format(stats)
}

func (m *Model) maybeLoadStats() tea.Cmd {
	if !m.usesStats() {
		return nil
	}
	id := m.currentID()
	if id == "" || m.statsBusy[id] {
		return nil
	}
	if _, ok := m.stats[id]; ok {
		return nil
	}
	m.statsBusy[id] = true
	return m.loadStats(id)
}

func (m Model) usesStats() bool {
	return m.fieldEnabled("message_count") || m.fieldEnabled("part_count") || m.fieldEnabled("size") || m.fieldEnabled("large_session")
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

func (m Model) fieldEnabled(id string) bool {
	enabled, ok := m.fields[id]
	return ok && enabled
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

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func annotateDirectoryExists(sessions []opencode.Session) {
	cache := make(map[string]bool)
	for i := range sessions {
		dir := sessions[i].Directory
		if exists, ok := cache[dir]; ok {
			sessions[i].DirectoryExists = exists
			continue
		}
		exists := false
		if dir != "" {
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				exists = true
			}
		}
		cache[dir] = exists
		sessions[i].DirectoryExists = exists
	}
}

func (m *Model) moveUp(count int) {
	if m.selected > 0 {
		m.selected = max(0, m.selected-count)
		m.ensureVisible()
	}
}

func (m *Model) moveDown(count int) {
	if m.selected < len(m.sessions)-1 {
		m.selected = min(len(m.sessions)-1, m.selected+count)
		m.ensureVisible()
	}
}

func (m *Model) ensureVisible() {
	visible := m.visibleItems()
	if m.selected < m.offset {
		m.offset = m.selected
	}
	if m.selected >= m.offset+visible {
		m.offset = m.selected - visible + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m Model) visibleItems() int {
	reserved := 5
	if m.query != "" || m.mode == ModeSearch {
		reserved = 6
	}
	return max(1, (m.height-reserved)/2)
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
