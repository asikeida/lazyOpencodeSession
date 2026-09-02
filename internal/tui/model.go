package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/asikeida/lazyOpencodeSession/internal/opencode"
	"github.com/asikeida/lazyOpencodeSession/internal/platform"
)

type Options struct {
	Repo     opencode.Repository
	Limit    int
	Theme    string
	Language string
}

type Mode int

const (
	ModeBrowse Mode = iota
	ModeSearch
)

type Model struct {
	repo       opencode.Repository
	limit      int
	styles     Styles
	texts      Texts
	width      int
	height     int
	sessions   []opencode.Session
	selected   int
	offset     int
	query      string
	mode       Mode
	loading    bool
	status     string
	err        error
	help       bool
	preview    []opencode.MessagePreview
	previewFor string
	resumeID   string
}

type sessionsLoadedMsg struct {
	query    string
	sessions []opencode.Session
}

type previewLoadedMsg struct {
	sessionID string
	messages  []opencode.MessagePreview
}

type errMsg struct {
	err error
}

func New(opts Options) Model {
	if opts.Limit <= 0 {
		opts.Limit = 500
	}
	return Model{
		repo:    opts.Repo,
		limit:   opts.Limit,
		styles:  NewStyles(opts.Theme),
		texts:   NewTexts(opts.Language),
		loading: true,
		status:  NewTexts(opts.Language).LoadingSessions,
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
		m.loading = false
		m.err = nil
		if m.selected >= len(m.sessions) {
			m.selected = max(0, len(m.sessions)-1)
		}
		m.ensureVisible()
		m.status = fmt.Sprintf("%d sessions", len(m.sessions))
		return m, nil
	case previewLoadedMsg:
		if msg.sessionID == m.currentID() {
			m.preview = msg.messages
			m.previewFor = msg.sessionID
			m.status = fmt.Sprintf(m.texts.LoadedPreviewMessages, len(msg.messages))
		}
		return m, nil
	case errMsg:
		m.loading = false
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
		case "up":
			m.moveUp(1)
			return m, nil
		case "down":
			m.moveDown(1)
			return m, nil
		case "pgup":
			m.moveUp(m.visibleItems())
			return m, nil
		case "pgdown":
			m.moveDown(m.visibleItems())
			return m, nil
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
	case "r":
		m.loading = true
		m.status = m.texts.ReloadingSessions
		return m, m.loadSessions()
	case "up", "k":
		m.moveUp(1)
		return m, nil
	case "down", "j":
		m.moveDown(1)
		return m, nil
	case "pgup":
		m.moveUp(m.visibleItems())
		return m, nil
	case "pgdown":
		m.moveDown(m.visibleItems())
		return m, nil
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
	if m.help {
		return m.renderHelp()
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
	return lipgloss.JoinVertical(lipgloss.Left, content, m.renderStatus())
}

func (m Model) renderSessions(width int, height int) string {
	box := lipgloss.NewStyle().Width(max(1, width-2)).Height(max(1, height-2)).Border(lipgloss.RoundedBorder()).Inherit(m.styles.Border)
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
			if i == m.selected {
				prefix = "> "
				style = m.styles.Selected
			}
			title := truncateWidth(s.Title, width-10)
			lines = append(lines, style.Render(prefix+title))
			meta := fmt.Sprintf("  %s  %s", truncateWidth(s.Directory, width-18), formatShortTime(s.UpdatedAt))
			lines = append(lines, m.styles.Muted.Render(meta))
		}
	}
	return box.Render(strings.Join(lines, "\n"))
}

func (m Model) renderDetails(width int, height int) string {
	box := lipgloss.NewStyle().Width(max(1, width-2)).Height(max(1, height-2)).Border(lipgloss.RoundedBorder()).Inherit(m.styles.Border)
	lines := []string{m.styles.Title.Render(m.texts.DetailsTitle)}
	if len(m.sessions) == 0 || m.selected >= len(m.sessions) {
		lines = append(lines, m.styles.Muted.Render(m.texts.SelectSession))
		return box.Render(strings.Join(lines, "\n"))
	}

	s := m.sessions[m.selected]
	lines = append(lines,
		field(m.texts.FieldTitle, s.Title),
		field(m.texts.FieldSession, s.ID),
		field(m.texts.FieldProject, s.ProjectID),
		field(m.texts.FieldDirectory, s.Directory),
		field(m.texts.FieldUpdated, formatFullTime(s.UpdatedAt)),
		field(m.texts.FieldCreated, formatFullTime(s.CreatedAt)),
		field(m.texts.FieldModel, emptyDash(s.Model)),
		field(m.texts.FieldAgent, emptyDash(s.Agent)),
		field(m.texts.FieldCost, fmt.Sprintf("$%.4f", s.Cost)),
		field(m.texts.FieldTokens, fmt.Sprintf(m.texts.TokensFormat, s.TokensInput, s.TokensOutput, s.TokensReasoning, s.TokensCacheRead)),
		"",
		m.styles.Muted.Render(m.texts.ActionHint),
	)

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
	text := strings.Join([]string{
		m.texts.HelpTitle,
		"",
		strings.Join(m.texts.HelpLines, "\n"),
		"",
		m.texts.ReadOnlyNotice,
	}, "\n")
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Inherit(m.styles.Border)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box.Render(text))
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
		return sessionsLoadedMsg{query: query, sessions: sessions}
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

func (m Model) currentID() string {
	if m.selected < 0 || m.selected >= len(m.sessions) {
		return ""
	}
	return m.sessions[m.selected].ID
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

func field(name string, value string) string {
	label := name + ":"
	padding := strings.Repeat(" ", max(1, 12-lipgloss.Width(label)))
	return label + padding + value
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

func formatFullTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
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

func wrap(value string, width int, maxLines int) []string {
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
			if len(lines) >= maxLines {
				return lines
			}
			continue
		}
		next := line + string(r)
		if lipgloss.Width(next) > width && line != "" {
			lines = append(lines, truncateWidth(line, width))
			line = string(r)
			if len(lines) >= maxLines {
				return lines
			}
			continue
		}
		line = next
	}
	if line != "" && len(lines) < maxLines {
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
