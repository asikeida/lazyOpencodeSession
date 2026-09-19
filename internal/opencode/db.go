package opencode

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db       *sql.DB
	compat   SchemaCompatibility
	version  SchemaVersion
	readOnly bool
	callAPI  apiCaller
}

type apiCaller func(ctx context.Context, method string, path string, body any) error

type SchemaVersion int

const (
	SchemaV1 SchemaVersion = iota + 1
	SchemaV2
)

func (v SchemaVersion) String() string {
	if v == SchemaV2 {
		return "v2"
	}
	return "v1"
}

type SchemaCompatibility struct {
	Browse  bool
	Stats   bool
	Preview bool
	Rename  bool
	Delete  bool
	Issues  []string
}

var ErrDeleteImpactChanged = errors.New("delete impact changed; review the updated scope and confirm again")

func Open(ctx context.Context, path string, readOnly bool) (*SQLiteRepository, error) {
	u := url.URL{Scheme: "file", Path: path}
	q := u.Query()
	if readOnly {
		q.Set("mode", "ro")
	} else {
		q.Set("mode", "rw")
	}
	q.Set("cache", "shared")
	q.Add("_pragma", "busy_timeout(1000)")
	q.Add("_pragma", "foreign_keys(ON)")
	u.RawQuery = q.Encode()

	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to open OpenCode database: %w", ActionableError(err))
	}
	compat, version, err := inspectSchema(ctx, db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to inspect OpenCode database schema: %w", ActionableError(err))
	}
	if !compat.Browse {
		db.Close()
		return nil, fmt.Errorf("incompatible OpenCode database schema: %s; update lazyocs or select a compatible OpenCode database", strings.Join(compat.Issues, "; "))
	}
	repo := &SQLiteRepository{db: db, compat: compat, version: version, readOnly: readOnly}
	repo.SetAPICommand("opencode")
	return repo, nil
}

func (r *SQLiteRepository) SetAPICommand(command string) {
	if strings.TrimSpace(command) == "" {
		command = "opencode"
	}
	r.callAPI = func(ctx context.Context, method string, path string, body any) error {
		return callOpenCodeAPI(ctx, command, method, path, body)
	}
}

func (r *SQLiteRepository) UpdateSessionTitle(ctx context.Context, sessionID string, title string) error {
	if !r.compat.Rename {
		return errors.New("OpenCode database schema does not support safe title updates")
	}
	if r.readOnly {
		return errors.New("read-only repository does not allow title updates")
	}
	if r.version == SchemaV2 {
		path := "/api/session/" + url.PathEscape(sessionID)
		body := map[string]string{"title": title}
		err := r.callAPI(ctx, "patch", path, body)
		if err != nil && strings.Contains(err.Error(), "HTTP 404 Not Found") {
			return r.callAPI(ctx, "post", path+"/rename", body)
		}
		return err
	}
	result, err := r.db.ExecContext(ctx, `update session set title = ? where id = ?`, title, sessionID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	return nil
}

func (r *SQLiteRepository) DeleteSession(ctx context.Context, sessionID string) error {
	impact, err := r.DeleteImpact(ctx, sessionID)
	if err != nil {
		return err
	}
	return r.DeleteSessionIfUnchanged(ctx, sessionID, impact)
}

func (r *SQLiteRepository) DeleteImpact(ctx context.Context, sessionID string) (DeleteImpact, error) {
	if !r.compat.Delete {
		return DeleteImpact{}, errors.New("OpenCode database schema does not have the required delete cascades")
	}
	if r.version == SchemaV2 {
		return queryDeleteImpactV2(ctx, r.db, sessionID)
	}
	return queryDeleteImpact(ctx, r.db, sessionID)
}

func (r *SQLiteRepository) DeleteSessionIfUnchanged(ctx context.Context, sessionID string, expected DeleteImpact) error {
	if !r.compat.Delete {
		return errors.New("OpenCode database schema does not have the required delete cascades")
	}
	if r.readOnly {
		return errors.New("read-only repository does not allow deletion")
	}
	if r.version == SchemaV2 {
		current, err := queryDeleteImpactV2(ctx, r.db, sessionID)
		if err != nil {
			return err
		}
		if current != expected {
			return fmt.Errorf("%w: expected %+v, current %+v", ErrDeleteImpactChanged, expected, current)
		}
		return r.callAPI(ctx, "delete", "/api/session/"+url.PathEscape(sessionID), nil)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	current, err := queryDeleteImpact(ctx, tx, sessionID)
	if err != nil {
		return err
	}
	if current != expected {
		return fmt.Errorf("%w: expected %+v, current %+v", ErrDeleteImpactChanged, expected, current)
	}
	result, err := tx.ExecContext(ctx, `
with recursive descendants(id) as (
  select id from session where id = ?
  union
  select s.id from session s join descendants d on s.parent_id = d.id
)
delete from session where id in (select id from descendants)`, sessionID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	if count != expected.SessionCount {
		return fmt.Errorf("%w: expected %d sessions, deleted %d", ErrDeleteImpactChanged, expected.SessionCount, count)
	}
	return tx.Commit()
}

func callOpenCodeAPI(ctx context.Context, command string, method string, path string, body any) error {
	args := []string{"api", strings.ToLower(method), path}
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		args = append(args, "--data", string(data))
	}
	output, err := exec.CommandContext(ctx, command, args...).CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail == "" {
			return fmt.Errorf("opencode api %s %s failed: %w", method, path, err)
		}
		return fmt.Errorf("opencode api %s %s failed: %s", method, path, detail)
	}
	return nil
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func queryDeleteImpact(ctx context.Context, db queryRower, sessionID string) (DeleteImpact, error) {
	var impact DeleteImpact
	err := db.QueryRowContext(ctx, `
with recursive descendants(id) as (
  select id from session where id = ?
  union
  select s.id from session s join descendants d on s.parent_id = d.id
)
select
  (select count(*) from descendants),
  (select count(*) from message where session_id in (select id from descendants)),
  (select count(*) from part p join message m on p.message_id = m.id where m.session_id in (select id from descendants))`, sessionID).Scan(
		&impact.SessionCount,
		&impact.MessageCount,
		&impact.PartCount,
	)
	if err != nil {
		return DeleteImpact{}, err
	}
	if impact.SessionCount == 0 {
		return DeleteImpact{}, fmt.Errorf("session not found: %s", sessionID)
	}
	return impact, nil
}

func queryDeleteImpactV2(ctx context.Context, db queryRower, sessionID string) (DeleteImpact, error) {
	var impact DeleteImpact
	err := db.QueryRowContext(ctx, `
with recursive descendants(id) as (
  select id from session_v2 where id = ?
  union
  select s.id from session_v2 s join descendants d on s.parent_id = d.id
)
select
  (select count(*) from descendants),
  (select count(*) from session_message where session_id in (select id from descendants)),
  coalesce((select sum(
    case
      when type = 'assistant' and json_type(data, '$.content') = 'array' then json_array_length(data, '$.content')
      when type in ('user', 'synthetic', 'system') and trim(coalesce(json_extract(data, '$.text'), '')) != '' then 1
      else 0
    end
  ) from session_message where session_id in (select id from descendants)), 0)`, sessionID).Scan(
		&impact.SessionCount,
		&impact.MessageCount,
		&impact.PartCount,
	)
	if err != nil {
		return DeleteImpact{}, err
	}
	if impact.SessionCount == 0 {
		return DeleteImpact{}, fmt.Errorf("session not found: %s", sessionID)
	}
	return impact, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

func (r *SQLiteRepository) Compatibility() SchemaCompatibility {
	return r.compat
}

func (r *SQLiteRepository) SchemaVersion() SchemaVersion {
	return r.version
}

func inspectSchema(ctx context.Context, db *sql.DB) (SchemaCompatibility, SchemaVersion, error) {
	sessionV2, err := tableColumns(ctx, db, "session_v2")
	if err != nil {
		return SchemaCompatibility{}, 0, err
	}
	messageV2, err := tableColumns(ctx, db, "session_message")
	if err != nil {
		return SchemaCompatibility{}, 0, err
	}
	if len(sessionV2) > 0 || len(messageV2) > 0 {
		v2Browse := []string{"id", "project_id", "parent_id", "title", "directory", "time_created", "time_updated", "model", "agent", "cost", "tokens_input", "tokens_output", "tokens_reasoning", "tokens_cache_read", "tokens_cache_write"}
		missingBrowse := missingColumns(sessionV2, v2Browse)
		compat := SchemaCompatibility{
			Browse:  len(missingBrowse) == 0,
			Stats:   hasColumns(messageV2, "session_id", "type", "data"),
			Preview: hasColumns(messageV2, "id", "session_id", "type", "data", "time_created"),
			Rename:  hasColumns(sessionV2, "id", "title"),
			Delete:  hasColumns(sessionV2, "id", "parent_id") && hasColumns(messageV2, "session_id", "type", "data"),
		}
		if len(missingBrowse) > 0 {
			compat.Issues = append(compat.Issues, "missing session_v2 columns: "+strings.Join(missingBrowse, ", "))
		}
		if !compat.Stats || !compat.Preview {
			compat.Issues = append(compat.Issues, "required session_message columns are missing")
		}
		return compat, SchemaV2, nil
	}

	session, err := tableColumns(ctx, db, "session")
	if err != nil {
		return SchemaCompatibility{}, 0, err
	}
	message, err := tableColumns(ctx, db, "message")
	if err != nil {
		return SchemaCompatibility{}, 0, err
	}
	part, err := tableColumns(ctx, db, "part")
	if err != nil {
		return SchemaCompatibility{}, 0, err
	}

	browseColumns := []string{"id", "project_id", "parent_id", "title", "directory", "time_created", "time_updated", "model", "agent", "cost", "tokens_input", "tokens_output", "tokens_reasoning", "tokens_cache_read", "tokens_cache_write"}
	missingBrowse := missingColumns(session, browseColumns)
	compat := SchemaCompatibility{
		Browse:  len(missingBrowse) == 0,
		Stats:   hasColumns(message, "session_id", "data") && hasColumns(part, "session_id", "data"),
		Preview: hasColumns(message, "id", "session_id", "data", "time_created") && hasColumns(part, "id", "message_id", "data", "time_created"),
		Rename:  hasColumns(session, "id", "title"),
	}
	if len(missingBrowse) > 0 {
		compat.Issues = append(compat.Issues, "missing session columns: "+strings.Join(missingBrowse, ", "))
	}

	messageCascade, err := hasDeleteCascade(ctx, db, "message", "session_id", "session", "id")
	if err != nil {
		return SchemaCompatibility{}, 0, err
	}
	partCascade, err := hasDeleteCascade(ctx, db, "part", "message_id", "message", "id")
	if err != nil {
		return SchemaCompatibility{}, 0, err
	}
	compat.Delete = hasColumns(session, "id", "parent_id") && messageCascade && partCascade
	if !compat.Delete {
		compat.Issues = append(compat.Issues, "required session/message/part delete cascades are missing")
	}
	return compat, SchemaV1, nil
}

func tableColumns(ctx context.Context, db *sql.DB, table string) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "pragma table_info("+table+")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return nil, err
		}
		columns[name] = true
	}
	return columns, rows.Err()
}

func hasColumns(columns map[string]bool, required ...string) bool {
	return len(missingColumns(columns, required)) == 0
}

func missingColumns(columns map[string]bool, required []string) []string {
	missing := make([]string, 0)
	for _, column := range required {
		if !columns[column] {
			missing = append(missing, column)
		}
	}
	return missing
}

func hasDeleteCascade(ctx context.Context, db *sql.DB, table, from, targetTable, targetColumn string) (bool, error) {
	rows, err := db.QueryContext(ctx, "pragma foreign_key_list("+table+")")
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, seq int
		var referencedTable, sourceColumn, referencedColumn, onUpdate, onDelete, match string
		if err := rows.Scan(&id, &seq, &referencedTable, &sourceColumn, &referencedColumn, &onUpdate, &onDelete, &match); err != nil {
			return false, err
		}
		if sourceColumn == from && referencedTable == targetTable && referencedColumn == targetColumn && strings.EqualFold(onDelete, "cascade") {
			return true, nil
		}
	}
	return false, rows.Err()
}

func (r *SQLiteRepository) ListSessions(ctx context.Context, filter SessionFilter) ([]Session, error) {
	if filter.Limit <= 0 {
		filter.Limit = 500
	}

	table := "session"
	titleExpr := "title"
	modelExpr := "coalesce(model, '')"
	if r.version == SchemaV2 {
		table = "session_v2"
		titleExpr = "coalesce(title, '')"
		modelExpr = "coalesce(case when json_valid(model) then json_extract(model, '$.id') else model end, '')"
	}
	base := `
select
  id,
  project_id,
  coalesce(parent_id, ''),
  ` + titleExpr + `,
  directory,
  time_created,
  time_updated,
  ` + modelExpr + `,
  coalesce(agent, ''),
  cost,
  tokens_input,
  tokens_output,
  tokens_reasoning,
  tokens_cache_read,
  tokens_cache_write
from ` + table + `
where (parent_id is null or parent_id = '')`

	where, args := metadataSearchWhere(filter.Query)
	base += where
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

func (r *SQLiteRepository) CountSessions(ctx context.Context, filter SessionFilter) (int, error) {
	table := "session"
	if r.version == SchemaV2 {
		table = "session_v2"
	}
	base := `
select count(*)
from ` + table + `
where (parent_id is null or parent_id = '')`
	where, args := metadataSearchWhere(filter.Query)
	base += where

	var count int
	if err := r.db.QueryRowContext(ctx, base, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *SQLiteRepository) SessionStats(ctx context.Context, sessionID string) (SessionStats, error) {
	if !r.compat.Stats {
		return SessionStats{}, errors.New("OpenCode database schema does not support session statistics")
	}
	var stats SessionStats
	if r.version == SchemaV2 {
		err := r.db.QueryRowContext(ctx, `
select
  count(*),
  coalesce(sum(
    case
      when type = 'assistant' and json_type(data, '$.content') = 'array' then json_array_length(data, '$.content')
      when type in ('user', 'synthetic', 'system') and trim(coalesce(json_extract(data, '$.text'), '')) != '' then 1
      else 0
    end
  ), 0),
  coalesce(sum(length(data)), 0)
from session_message
where session_id = ?`, sessionID).Scan(&stats.MessageCount, &stats.PartCount, &stats.SizeBytes)
		return stats, err
	}
	err := r.db.QueryRowContext(ctx, `
select
  (select count(*) from message where session_id = ?),
  (select count(*) from part where session_id = ?),
  coalesce((select sum(length(data)) from message where session_id = ?), 0) +
  coalesce((select sum(length(data)) from part where session_id = ?), 0)`, sessionID, sessionID, sessionID, sessionID).Scan(
		&stats.MessageCount,
		&stats.PartCount,
		&stats.SizeBytes,
	)
	if err != nil {
		return SessionStats{}, err
	}
	return stats, nil
}

func (r *SQLiteRepository) RecentUserMessages(ctx context.Context, sessionID string, limit int, maxChars int) ([]MessagePreview, error) {
	if !r.compat.Preview {
		return nil, errors.New("OpenCode database schema does not support message previews")
	}
	if limit <= 0 {
		limit = 5
	}
	if maxChars <= 0 {
		maxChars = 500
	}

	if r.version == SchemaV2 {
		return r.recentUserMessagesV2(ctx, sessionID, limit, maxChars)
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

func (r *SQLiteRepository) recentUserMessagesV2(ctx context.Context, sessionID string, limit int, maxChars int) ([]MessagePreview, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, json_extract(data, '$.text'), time_created
from session_message
where session_id = ?
  and type = 'user'
  and trim(coalesce(json_extract(data, '$.text'), '')) != ''
order by time_created desc
limit ?`, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	previews := make([]MessagePreview, 0, limit)
	for rows.Next() {
		var preview MessagePreview
		var created int64
		if err := rows.Scan(&preview.ID, &preview.Text, &created); err != nil {
			return nil, err
		}
		preview.Text = truncateRunes(preview.Text, maxChars)
		preview.CreatedAt = millis(created)
		previews = append(previews, preview)
	}
	return previews, rows.Err()
}

func (r *SQLiteRepository) RecentUserMemory(ctx context.Context, since time.Time, limit int, maxChars int) ([]UserMemory, error) {
	if !r.compat.Preview {
		return nil, errors.New("OpenCode database schema does not support user memory search")
	}
	if limit <= 0 {
		limit = 5000
	}
	if maxChars <= 0 {
		maxChars = 4000
	}
	if r.version == SchemaV2 {
		return r.recentUserMemoryV2(ctx, since, limit, maxChars)
	}

	rows, err := r.db.QueryContext(ctx, `
with recursive roots(root_id, id) as (
  select id, id from session where parent_id is null or parent_id = ''
  union
  select roots.root_id, s.id from session s join roots on s.parent_id = roots.id
)
select roots.root_id, p.data, p.time_created
from roots
cross join message m on m.session_id = roots.id
cross join part p on p.message_id = m.id
where m.time_created >= ?
  and m.data like '%"role":"user"%'
  and p.data like '%"type":"text"%'
order by m.time_created desc, p.time_created asc
limit ?`, since.UnixMilli(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	memories := make([]UserMemory, 0)
	for rows.Next() {
		var sessionID, raw string
		var created int64
		if err := rows.Scan(&sessionID, &raw, &created); err != nil {
			return nil, err
		}
		text := textFromPart(raw)
		if strings.TrimSpace(text) == "" {
			continue
		}
		memories = append(memories, UserMemory{SessionID: sessionID, Text: truncateRunes(text, maxChars), CreatedAt: millis(created)})
	}
	return memories, rows.Err()
}

func (r *SQLiteRepository) recentUserMemoryV2(ctx context.Context, since time.Time, limit int, maxChars int) ([]UserMemory, error) {
	rows, err := r.db.QueryContext(ctx, `
with recursive roots(root_id, id) as (
  select id, id from session_v2 where parent_id is null or parent_id = ''
  union
  select roots.root_id, s.id from session_v2 s join roots on s.parent_id = roots.id
)
select roots.root_id, json_extract(sm.data, '$.text'), sm.time_created
from roots
join session_message sm on sm.session_id = roots.id
where sm.time_created >= ?
  and sm.type = 'user'
  and trim(coalesce(json_extract(sm.data, '$.text'), '')) != ''
order by sm.time_created desc
limit ?`, since.UnixMilli(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	memories := make([]UserMemory, 0)
	for rows.Next() {
		var memory UserMemory
		var created int64
		if err := rows.Scan(&memory.SessionID, &memory.Text, &created); err != nil {
			return nil, err
		}
		memory.Text = truncateRunes(memory.Text, maxChars)
		memory.CreatedAt = millis(created)
		memories = append(memories, memory)
	}
	return memories, rows.Err()
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

func metadataSearchWhere(query string) (string, []any) {
	terms := SearchTerms(query)
	if len(terms) == 0 {
		return "", nil
	}
	field := `lower(
    coalesce(id, '') || ' ' ||
    coalesce(project_id, '') || ' ' ||
    coalesce(title, '') || ' ' ||
    coalesce(directory, '') || ' ' ||
    coalesce(model, '') || ' ' ||
    coalesce(agent, '')
  )`
	parts := make([]string, 0, len(terms))
	args := make([]any, 0, len(terms))
	for _, term := range terms {
		parts = append(parts, field+" like ?")
		args = append(args, "%"+strings.ToLower(term)+"%")
	}
	return "\n  and " + strings.Join(parts, "\n  and "), args
}

func SearchTerms(query string) []string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	terms := make([]string, 0, len(fields))
	seen := map[string]bool{}
	for _, field := range fields {
		if field == "" || seen[field] {
			continue
		}
		seen[field] = true
		terms = append(terms, field)
	}
	return terms
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
