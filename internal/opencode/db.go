package opencode

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*SQLiteRepository, error) {
	u := url.URL{Scheme: "file", Path: path}
	q := u.Query()
	q.Set("mode", "ro")
	q.Set("cache", "shared")
	q.Add("_pragma", "busy_timeout(1000)")
	u.RawQuery = q.Encode()

	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to open OpenCode database: %w", err)
	}
	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

func (r *SQLiteRepository) ListSessions(ctx context.Context, filter SessionFilter) ([]Session, error) {
	if filter.Limit <= 0 {
		filter.Limit = 500
	}

	base := `
select
  id,
  project_id,
  coalesce(parent_id, ''),
  title,
  directory,
  time_created,
  time_updated,
  coalesce(model, ''),
  coalesce(agent, ''),
  cost,
  tokens_input,
  tokens_output,
  tokens_reasoning,
  tokens_cache_read,
  tokens_cache_write
from session
where (parent_id is null or parent_id = '')`

	args := []any{}
	query := strings.TrimSpace(filter.Query)
	if query != "" {
		like := "%" + strings.ToLower(query) + "%"
		base += `
  and (
    lower(id) like ?
    or lower(project_id) like ?
    or lower(title) like ?
    or lower(directory) like ?
    or lower(coalesce(model, '')) like ?
    or lower(coalesce(agent, '')) like ?
  )`
		args = append(args, like, like, like, like, like, like)
	}
	base += "\norder by time_updated desc\nlimit ?"
	args = append(args, filter.Limit)

	rows, err := r.db.QueryContext(ctx, base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]Session, 0, filter.Limit)
	for rows.Next() {
		var s Session
		var created, updated int64
		if err := rows.Scan(
			&s.ID,
			&s.ProjectID,
			&s.ParentID,
			&s.Title,
			&s.Directory,
			&created,
			&updated,
			&s.Model,
			&s.Agent,
			&s.Cost,
			&s.TokensInput,
			&s.TokensOutput,
			&s.TokensReasoning,
			&s.TokensCacheRead,
			&s.TokensCacheWrite,
		); err != nil {
			return nil, err
		}
		s.CreatedAt = millis(created)
		s.UpdatedAt = millis(updated)
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (r *SQLiteRepository) RecentUserMessages(ctx context.Context, sessionID string, limit int, maxChars int) ([]MessagePreview, error) {
	if limit <= 0 {
		limit = 5
	}
	if maxChars <= 0 {
		maxChars = 500
	}

	rows, err := r.db.QueryContext(ctx, `
select
  p.id,
  p.data,
  p.time_created
from message m
join part p on p.message_id = m.id
where m.session_id = ?
  and m.data like '%"role":"user"%'
  and p.data like '%"type":"text"%'
order by m.time_created desc, p.time_created asc
limit ?`, sessionID, limit*3)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	previews := make([]MessagePreview, 0, limit)
	for rows.Next() {
		var id string
		var raw string
		var created int64
		if err := rows.Scan(&id, &raw, &created); err != nil {
			return nil, err
		}
		text := textFromPart(raw)
		if strings.TrimSpace(text) == "" {
			continue
		}
		previews = append(previews, MessagePreview{ID: id, Text: truncateRunes(text, maxChars), CreatedAt: millis(created)})
		if len(previews) >= limit {
			break
		}
	}
	return previews, rows.Err()
}

func textFromPart(raw string) string {
	var part struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(raw), &part); err != nil {
		return ""
	}
	if part.Type != "text" {
		return ""
	}
	return part.Text
}

func millis(ms int64) time.Time {
	if ms <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms).Local()
}

func truncateRunes(value string, max int) string {
	r := []rune(value)
	if len(r) <= max {
		return value
	}
	return string(r[:max]) + "..."
}
