package tui

import (
	"context"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/asikeida/lazyOpencodeSession/internal/opencode"
	"github.com/asikeida/lazyOpencodeSession/internal/platform"
)

func (m Model) loadSessions() tea.Cmd {
	query := m.query
	limit := m.limit
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		sessions, err := m.repo.ListSessions(ctx, opencode.SessionFilter{Query: query, Limit: limit})
		if err != nil {
			return sessionsLoadFailedMsg{query: query, err: err}
		}
		annotateDirectoryExists(sessions)
		matched, err := m.repo.CountSessions(ctx, opencode.SessionFilter{Query: query})
		if err != nil {
			return sessionsLoadFailedMsg{query: query, err: err}
		}
		total := matched
		if strings.TrimSpace(query) != "" {
			total, err = m.repo.CountSessions(ctx, opencode.SessionFilter{})
			if err != nil {
				return sessionsLoadFailedMsg{query: query, err: err}
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
			return previewLoadFailedMsg{sessionID: sessionID, err: err}
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
			return statsLoadFailedMsg{sessionID: sessionID, err: err}
		}
		return statsLoadedMsg{sessionID: sessionID, stats: stats}
	}
}

func (m Model) saveTitle(sessionID string, title string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := m.repo.UpdateSessionTitle(ctx, sessionID, title); err != nil {
			return titleUpdateFailedMsg{sessionID: sessionID, err: err}
		}
		return titleUpdatedMsg{sessionID: sessionID, title: title}
	}
}

func (m Model) deleteSession(sessionID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := m.repo.DeleteSession(ctx, sessionID); err != nil {
			return sessionDeleteFailedMsg{sessionID: sessionID, err: err}
		}
		return sessionDeletedMsg{sessionID: sessionID}
	}
}

func (m Model) copySessionID(sessionID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := platform.Copy(ctx, sessionID); err != nil {
			return clipboardCopyFailedMsg{sessionID: sessionID, err: err}
		}
		return clipboardCopiedMsg{}
	}
}

func (m *Model) searchChanged() tea.Cmd {
	m.searchVersion++
	m.searchPending = true
	m.selected = 0
	m.offset = 0
	m.preview = nil
	m.previewFor = ""
	m.loading = true
	version := m.searchVersion
	return tea.Tick(searchDebounceDelay, func(time.Time) tea.Msg {
		return searchDebounceMsg{version: version}
	})
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
