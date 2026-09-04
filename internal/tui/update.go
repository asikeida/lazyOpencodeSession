package tui

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/asikeida/lazyOpencodeSession/internal/opencode"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureVisible()
		m.clampDetailsOffset()
		return m, nil
	case sessionsLoadedMsg:
		m.catalog = msg.sessions
		m.total = msg.total
		m.loading = false
		m.err = nil
		m.applySearch()
		if m.selected >= len(m.sessions) {
			m.selected = max(0, len(m.sessions)-1)
		}
		m.ensureVisible()
		m.status = m.resultStatus()
		return m, m.maybeLoadStats()
	case userMemoryLoadedMsg:
		m.memories = map[string][]opencode.UserMemory{}
		m.memorySearch = map[string]string{}
		for _, memory := range msg.memories {
			m.memories[memory.SessionID] = append(m.memories[memory.SessionID], memory)
			m.memorySearch[memory.SessionID] += " " + strings.ToLower(memory.Text)
		}
		m.memoryLoading = false
		m.memoryErr = nil
		m.applySearch()
		if m.selected >= len(m.sessions) {
			m.selected = max(0, len(m.sessions)-1)
		}
		m.ensureVisible()
		m.status = m.resultStatus()
		return m, m.maybeLoadStats()
	case userMemoryFailedMsg:
		m.memoryLoading = false
		m.memoryErr = msg.err
		return m, nil
	case statsLoadedMsg:
		m.stats[msg.sessionID] = msg.stats
		delete(m.statsBusy, msg.sessionID)
		return m, nil
	case titleUpdatedMsg:
		for i := range m.catalog {
			if m.catalog[i].ID == msg.sessionID {
				m.catalog[i].Title = msg.title
				break
			}
		}
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
		for i := range m.catalog {
			if m.catalog[i].ID == msg.sessionID {
				m.catalog = append(m.catalog[:i], m.catalog[i+1:]...)
				break
			}
		}
		for i := range m.sessions {
			if m.sessions[i].ID == msg.sessionID {
				m.sessions = append(m.sessions[:i], m.sessions[i+1:]...)
				break
			}
		}
		delete(m.stats, msg.sessionID)
		delete(m.statsBusy, msg.sessionID)
		delete(m.memories, msg.sessionID)
		delete(m.memorySearch, msg.sessionID)
		delete(m.memoryMatches, msg.sessionID)
		m.deleteConfirm = false
		m.deleteBusy = false
		m.deleteImpact = opencode.DeleteImpact{}
		m.deleteFor = ""
		m.deleteLoading = false
		m.deleteErr = nil
		m.preview = nil
		m.previewFor = ""
		m.detailsOffset = 0
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
			m.detailsOffset = 0
			m.status = fmt.Sprintf(m.texts.LoadedPreviewMessages, len(msg.messages))
		}
		return m, nil
	case deleteImpactLoadedMsg:
		if m.deleteConfirm && msg.sessionID == m.currentID() {
			m.deleteImpact = msg.impact
			m.deleteFor = msg.sessionID
			m.deleteLoading = false
			m.deleteErr = nil
			m.status = m.texts.ConfirmDelete
		}
		return m, nil
	case deleteImpactFailedMsg:
		if m.deleteConfirm && msg.sessionID == m.currentID() {
			m.deleteLoading = false
			m.deleteErr = msg.err
			m.status = msg.err.Error()
		}
		return m, nil
	case sessionsLoadFailedMsg:
		if msg.query != m.query {
			return m, nil
		}
		m.loading = false
		m.err = msg.err
		m.status = msg.err.Error()
		return m, nil
	case previewLoadFailedMsg:
		if msg.sessionID == m.currentID() {
			m.status = msg.err.Error()
		}
		return m, nil
	case statsLoadFailedMsg:
		delete(m.statsBusy, msg.sessionID)
		if msg.sessionID == m.currentID() {
			m.status = msg.err.Error()
		}
		return m, nil
	case titleUpdateFailedMsg:
		if msg.sessionID == m.currentID() {
			m.status = msg.err.Error()
		}
		return m, nil
	case sessionDeleteFailedMsg:
		m.deleteBusy = false
		if msg.sessionID == m.currentID() {
			m.status = msg.err.Error()
			if errors.Is(msg.err, opencode.ErrDeleteImpactChanged) {
				m.deleteImpact = opencode.DeleteImpact{}
				m.deleteFor = ""
				m.deleteLoading = true
				return m, m.loadDeleteImpact(msg.sessionID)
			}
		}
		return m, nil
	case clipboardCopiedMsg:
		m.status = m.texts.CopiedSessionID
		return m, nil
	case clipboardCopyFailedMsg:
		m.status = fmt.Sprintf(m.texts.ClipboardUnavailable, msg.sessionID)
		return m, nil
	case searchDebounceMsg:
		if m.mode != ModeSearch || !m.searchPending || msg.version != m.searchVersion {
			return m, nil
		}
		m.searchPending = false
		m.applySearch()
		m.status = m.resultStatus()
		return m, m.maybeLoadStats()
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
			if m.deleteLoading || m.deleteErr != nil || m.deleteFor != id {
				return m, nil
			}
			m.deleteBusy = true
			m.status = m.texts.DeletingSession
			return m, m.deleteSession(id, m.deleteImpact)
		case "n", "esc":
			m.deleteConfirm = false
			m.deleteImpact = opencode.DeleteImpact{}
			m.deleteFor = ""
			m.deleteLoading = false
			m.deleteErr = nil
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
			m.searchVersion++
			m.searchPending = false
			m.selected = 0
			m.offset = 0
			m.preview = nil
			m.previewFor = ""
			m.applySearch()
			m.status = m.resultStatus()
			return m, m.maybeLoadStats()
		case "enter":
			m.mode = ModeBrowse
			if m.searchPending {
				m.searchVersion++
				m.searchPending = false
				m.applySearch()
				m.status = m.resultStatus()
				return m, m.maybeLoadStats()
			}
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		case "up", "ctrl+k":
			if m.detailsFocused() {
				m.scrollDetails(-1)
				return m, nil
			}
			m.moveUp(1)
			m.detailsOffset = 0
			return m, m.maybeLoadStats()
		case "down", "ctrl+j":
			if m.detailsFocused() {
				m.scrollDetails(1)
				return m, nil
			}
			m.moveDown(1)
			m.detailsOffset = 0
			return m, m.maybeLoadStats()
		case "ctrl+p":
			if id := m.currentID(); id != "" {
				m.status = m.texts.LoadingPreview
				return m, m.loadPreview(id)
			}
			return m, nil
		case "pgup":
			if m.detailsFocused() {
				m.scrollDetails(-m.detailScrollStep())
				return m, nil
			}
			m.moveUp(m.visibleItems())
			m.detailsOffset = 0
			return m, m.maybeLoadStats()
		case "pgdown":
			if m.detailsFocused() {
				m.scrollDetails(m.detailScrollStep())
				return m, nil
			}
			m.moveDown(m.visibleItems())
			m.detailsOffset = 0
			return m, m.maybeLoadStats()
		}
		switch key.Type {
		case tea.KeyBackspace:
			if len([]rune(m.query)) > 0 {
				r := []rune(m.query)
				m.query = string(r[:len(r)-1])
				return m, m.searchChanged()
			}
		case tea.KeyRunes:
			m.query += string(key.Runes)
			return m, m.searchChanged()
		case tea.KeySpace:
			m.query += " "
			return m, m.searchChanged()
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
		m.focus = FocusSessions
		m.status = m.texts.SearchSessions
		return m, nil
	case "h":
		m.focus = FocusSessions
		return m, nil
	case "l":
		if m.hasDetailsPane() {
			m.focus = FocusDetails
		}
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
		id := m.currentID()
		m.deleteImpact = opencode.DeleteImpact{}
		m.deleteFor = ""
		m.deleteLoading = true
		m.deleteErr = nil
		m.status = m.texts.LoadingDeleteImpact
		return m, m.loadDeleteImpact(id)
	case "r":
		m.loading = true
		m.status = m.texts.ReloadingSessions
		if m.recentDays > 0 {
			m.memoryLoading = true
			m.memoryErr = nil
			m.memories = map[string][]opencode.UserMemory{}
			m.memorySearch = map[string]string{}
			m.memoryMatches = map[string]string{}
			return m, tea.Batch(m.loadSessions(), m.loadUserMemory())
		}
		return m, m.loadSessions()
	case "up", "k":
		if m.detailsFocused() {
			m.scrollDetails(-1)
			return m, nil
		}
		m.moveUp(1)
		m.detailsOffset = 0
		return m, m.maybeLoadStats()
	case "down", "j":
		if m.detailsFocused() {
			m.scrollDetails(1)
			return m, nil
		}
		m.moveDown(1)
		m.detailsOffset = 0
		return m, m.maybeLoadStats()
	case "pgup":
		if m.detailsFocused() {
			m.scrollDetails(-m.detailScrollStep())
			return m, nil
		}
		m.moveUp(m.visibleItems())
		m.detailsOffset = 0
		return m, m.maybeLoadStats()
	case "pgdown":
		if m.detailsFocused() {
			m.scrollDetails(m.detailScrollStep())
			return m, nil
		}
		m.moveDown(m.visibleItems())
		m.detailsOffset = 0
		return m, m.maybeLoadStats()
	case "g":
		if m.detailsFocused() {
			m.detailsOffset = 0
			return m, nil
		}
	case "G":
		if m.detailsFocused() {
			m.detailsOffset = 1 << 30
			m.clampDetailsOffset()
			return m, nil
		}
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
			m.status = m.texts.CopyingSessionID
			return m, m.copySessionID(id)
		}
	}
	return m, nil
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

func (m Model) fieldEnabled(id string) bool {
	enabled, ok := m.fields[id]
	return ok && enabled
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

func (m *Model) detailScrollStep() int {
	return max(1, (m.height-5)/2)
}

func (m *Model) scrollDetails(delta int) {
	m.detailsOffset += delta
	m.clampDetailsOffset()
}

func (m *Model) clampDetailsOffset() {
	if !m.hasDetailsPane() {
		m.detailsOffset = 0
		m.focus = FocusSessions
		return
	}
	rightWidth := m.width - int(float64(m.width)*m.normalizedUI().SplitRatio)
	contentWidth := max(20, rightWidth-4)
	viewportHeight := max(1, m.height-5)
	maxOffset := max(0, len(m.detailLines(contentWidth))-viewportHeight)
	if m.detailsOffset < 0 {
		m.detailsOffset = 0
	}
	if m.detailsOffset > maxOffset {
		m.detailsOffset = maxOffset
	}
}

func (m Model) hasDetailsPane() bool {
	return m.width >= m.normalizedUI().TwoPaneMinWidth
}

func (m Model) detailsFocused() bool {
	return m.focus == FocusDetails && m.hasDetailsPane()
}
