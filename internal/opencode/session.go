package opencode

import "time"

type Session struct {
	ID               string
	ProjectID        string
	ParentID         string
	Title            string
	Directory        string
	DirectoryExists  bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Model            string
	Agent            string
	Cost             float64
	TokensInput      int64
	TokensOutput     int64
	TokensReasoning  int64
	TokensCacheRead  int64
	TokensCacheWrite int64
}

type MessagePreview struct {
	ID        string
	Text      string
	CreatedAt time.Time
}

type UserMemory struct {
	SessionID string
	Text      string
	CreatedAt time.Time
}

type SessionStats struct {
	MessageCount int64
	PartCount    int64
	SizeBytes    int64
}

type DeleteImpact struct {
	SessionCount int64
	MessageCount int64
	PartCount    int64
}

type SessionFilter struct {
	Query string
	Limit int
}
