package tui

import (
	"context"
	"os"
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
		sessions, err := m.repo.ListSessions(ctx, opencode.SessionFilter{Limit: limit})
		if err != nil {
			return sessionsLoadFailedMsg{query: query, err: opencode.ActionableError(err)}
		}
		annotateDirectoryExists(sessions)
		total, err := m.repo.CountSessions(ctx, opencode.SessionFilter{})
		if err != nil {
			return sessionsLoadFailedMsg{query: query, err: opencode.ActionableError(err)}
		}
		return sessionsLoadedMsg{query: query, sessions: sessions, total: total}
	}
}

func (m Model) loadUserMemory() tea.Cmd {
	days := m.recentDays
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		memories, err := m.repo.RecentUserMemory(ctx, time.Now().AddDate(0, 0, -days), 5000, 4000)
		if err != nil {
			return userMemoryFailedMsg{err: opencode.ActionableError(err)}
		}
		return userMemoryLoadedMsg{memories: memories}
	}
}

func (m Model) loadPreview(sessionID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		messages, err := m.repo.RecentUserMessages(ctx, sessionID, m.previewLimit, 500)
		if err != nil {
			return previewLoadFailedMsg{sessionID: sessionID, err: opencode.ActionableError(err)}
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
			return statsLoadFailedMsg{sessionID: sessionID, err: opencode.ActionableError(err)}
		}
		return statsLoadedMsg{sessionID: sessionID, stats: stats}
	}
}

func (m Model) saveTitle(sessionID string, title string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := m.repo.UpdateSessionTitle(ctx, sessionID, title); err != nil {
			return titleUpdateFailedMsg{sessionID: sessionID, err: opencode.ActionableError(err)}
		}
		return titleUpdatedMsg{sessionID: sessionID, title: title}
	}
}

func (m Model) loadDeleteImpact(sessionID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		impact, err := m.repo.DeleteImpact(ctx, sessionID)
		if err != nil {
			return deleteImpactFailedMsg{sessionID: sessionID, err: opencode.ActionableError(err)}
		}
		return deleteImpactLoadedMsg{sessionID: sessionID, impact: impact}
	}
}

func (m Model) deleteSession(sessionID string, expected opencode.DeleteImpact) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := m.repo.DeleteSessionIfUnchanged(ctx, sessionID, expected); err != nil {
			return sessionDeleteFailedMsg{sessionID: sessionID, err: opencode.ActionableError(err)}
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
