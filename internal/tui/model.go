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
	RecentDays      int
	PreviewLimit    int
	Theme           ThemeConfig
	UI              UIConfig
}

type Mode int
type PanelFocus int

const (
	ModeBrowse Mode = iota
	ModeSearch
)

const (
	FocusSessions PanelFocus = iota
	FocusDetails
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
	titleBusy     bool
	titleInput    string
	deleteConfirm bool
	deleteBusy    bool
	deleteImpact  opencode.DeleteImpact
	deleteFor     string
	deleteLoading bool
	deleteErr     error
	preview       []opencode.MessagePreview
	previewFor    string
	detailsOffset int
	focus         PanelFocus
	previewLimit  int
	resumeID      string
	catalog       []opencode.Session
	memories      map[string][]opencode.UserMemory
	memorySearch  map[string]string
	memoryMatches map[string]string
	recentDays    int
	memoryLoading bool
	memoryErr     error
	ui            UIConfig
}

type sessionsLoadedMsg struct {
	query    string
	sessions []opencode.Session
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

type deleteImpactLoadedMsg struct {
	sessionID string
	impact    opencode.DeleteImpact
}

type deleteImpactFailedMsg struct {
	sessionID string
	err       error
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

type userMemoryLoadedMsg struct {
	memories []opencode.UserMemory
}

type userMemoryFailedMsg struct {
	err error
}

func New(opts Options) Model {
	if opts.Limit <= 0 {
		opts.Limit = 500
	}
	opts.UI = NormalizeUIConfig(opts.UI)
	styles, err := BuildStyles(opts.Theme)
	if err != nil {
		styles = NewStyles()
	}
	return Model{
		repo:          opts.Repo,
		limit:         opts.Limit,
		opencode:      defaultString(opts.OpenCodeCommand, "opencode"),
		readOnly:      opts.ReadOnly,
		fields:        normalizeDetailFields(opts.DetailFields),
		stats:         map[string]opencode.SessionStats{},
		statsBusy:     map[string]bool{},
		styles:        styles,
		texts:         NewTexts(opts.Language),
		loading:       true,
		status:        NewTexts(opts.Language).LoadingSessions,
		memories:      map[string][]opencode.UserMemory{},
		memorySearch:  map[string]string{},
		memoryMatches: map[string]string{},
		recentDays:    opts.RecentDays,
		memoryLoading: opts.RecentDays > 0,
		previewLimit:  max(1, opts.PreviewLimit),
		ui:            opts.UI,
	}
}

func (m Model) Init() tea.Cmd {
	if m.recentDays <= 0 {
		return m.loadSessions()
	}
	return tea.Batch(m.loadSessions(), m.loadUserMemory())
}

func (m Model) ResumeSessionID() string {
	return m.resumeID
}
