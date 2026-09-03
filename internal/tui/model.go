package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/asikeida/lazyOpencodeSession/internal/opencode"
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
const searchDebounceDelay = 100 * time.Millisecond

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
	searchVersion uint64
	searchPending bool
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

type sessionsLoadFailedMsg struct {
	query string
	err   error
}

type previewLoadFailedMsg struct {
	sessionID string
	err       error
}

type statsLoadFailedMsg struct {
	sessionID string
	err       error
}

type titleUpdateFailedMsg struct {
	sessionID string
	err       error
}

type sessionDeleteFailedMsg struct {
	sessionID string
	err       error
}

type clipboardCopiedMsg struct{}

type clipboardCopyFailedMsg struct {
	sessionID string
	err       error
}

type searchDebounceMsg struct {
	version uint64
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
