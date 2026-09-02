package opencode

import "context"

type Repository interface {
	ListSessions(ctx context.Context, filter SessionFilter) ([]Session, error)
	CountSessions(ctx context.Context, filter SessionFilter) (int, error)
	SessionStats(ctx context.Context, sessionID string) (SessionStats, error)
	UpdateSessionTitle(ctx context.Context, sessionID string, title string) error
	RecentUserMessages(ctx context.Context, sessionID string, limit int, maxChars int) ([]MessagePreview, error)
	Close() error
}
