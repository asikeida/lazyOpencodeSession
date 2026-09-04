package opencode

import (
	"context"
	"time"
)

type Repository interface {
	ListSessions(ctx context.Context, filter SessionFilter) ([]Session, error)
	CountSessions(ctx context.Context, filter SessionFilter) (int, error)
	SessionStats(ctx context.Context, sessionID string) (SessionStats, error)
	UpdateSessionTitle(ctx context.Context, sessionID string, title string) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteImpact(ctx context.Context, sessionID string) (DeleteImpact, error)
	DeleteSessionIfUnchanged(ctx context.Context, sessionID string, expected DeleteImpact) error
	RecentUserMessages(ctx context.Context, sessionID string, limit int, maxChars int) ([]MessagePreview, error)
	RecentUserMemory(ctx context.Context, since time.Time, limit int, maxChars int) ([]UserMemory, error)
	Close() error
}
