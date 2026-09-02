package opencode

import "context"

type Repository interface {
	ListSessions(ctx context.Context, filter SessionFilter) ([]Session, error)
	RecentUserMessages(ctx context.Context, sessionID string, limit int, maxChars int) ([]MessagePreview, error)
	Close() error
}
