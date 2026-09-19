package opencode

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOpenEnablesForeignKeysAndDeleteCascades(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	createCompatibleDatabase(t, path)
	repo, err := Open(context.Background(), path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	compat := repo.Compatibility()
	if !compat.Browse || !compat.Stats || !compat.Preview || !compat.Rename || !compat.Delete {
		t.Fatalf("unexpected compatibility: %#v", compat)
	}
	var foreignKeys int
	if err := repo.db.QueryRow("pragma foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatal(err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign_keys = %d, want 1", foreignKeys)
	}
	if err := repo.DeleteSession(context.Background(), "ses_root"); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"session", "message", "part"} {
		var count int
		if err := repo.db.QueryRow("select count(*) from " + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%s rows remain after production-path delete: %d", table, count)
		}
	}
}

func TestOpenReadOnlyRejectsWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	createCompatibleDatabase(t, path)
	repo, err := Open(context.Background(), path, true)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	if err := repo.UpdateSessionTitle(context.Background(), "ses_root", "changed"); err == nil {
		t.Fatal("read-only repository allowed title update")
	}
	if err := repo.DeleteSession(context.Background(), "ses_root"); err == nil {
		t.Fatal("read-only repository allowed deletion")
	}
}

func TestDeleteImpactIncludesDescendantsAndPayloadRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	createCompatibleDatabase(t, path)
	repo, err := Open(context.Background(), path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	_, err = repo.db.Exec(`
insert into session (id, project_id, parent_id, title, directory, time_created, time_updated)
values ('ses_child', 'global', 'ses_root', 'child', '/tmp', 1, 1);
insert into message values ('msg_child', 'ses_child', '{}', 1);
insert into part values ('part_child', 'msg_child', 'ses_child', '{}', 1);
`)
	if err != nil {
		t.Fatal(err)
	}
	impact, err := repo.DeleteImpact(context.Background(), "ses_root")
	if err != nil {
		t.Fatal(err)
	}
	want := (DeleteImpact{SessionCount: 2, MessageCount: 2, PartCount: 2})
	if impact != want {
		t.Fatalf("impact = %#v, want %#v", impact, want)
	}
}

func TestDeleteSessionStopsWhenImpactChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	createCompatibleDatabase(t, path)
	repo, err := Open(context.Background(), path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	expected, err := repo.DeleteImpact(context.Background(), "ses_root")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.db.Exec("insert into message values ('msg_new', 'ses_root', '{}', 2)"); err != nil {
		t.Fatal(err)
	}
	err = repo.DeleteSessionIfUnchanged(context.Background(), "ses_root", expected)
	if !errors.Is(err, ErrDeleteImpactChanged) {
		t.Fatalf("delete error = %v, want impact changed", err)
	}
	var sessions int
	if err := repo.db.QueryRow("select count(*) from session").Scan(&sessions); err != nil {
		t.Fatal(err)
	}
	if sessions != 1 {
		t.Fatalf("delete was partially applied: %d sessions remain", sessions)
	}
}

func TestRecentUserMemoryMapsChildMessagesToRoot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	createCompatibleDatabase(t, path)
	repo, err := Open(context.Background(), path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	now := time.Now().UnixMilli()
	_, err = repo.db.Exec(`
insert into session (id, project_id, parent_id, title, directory, time_created, time_updated)
values ('ses_child', 'global', 'ses_root', 'child', '/tmp', ?, ?);
insert into message values ('msg_user', 'ses_child', '{"role":"user"}', ?);
insert into part values ('part_user', 'msg_user', 'ses_child', '{"type":"text","text":"remember the cobalt migration"}', ?);
`, now, now, now, now)
	if err != nil {
		t.Fatal(err)
	}
	memories, err := repo.RecentUserMemory(context.Background(), time.Now().Add(-time.Hour), 100, 4000)
	if err != nil {
		t.Fatal(err)
	}
	if len(memories) != 1 || memories[0].SessionID != "ses_root" || memories[0].Text != "remember the cobalt migration" {
		t.Fatalf("unexpected memories: %#v", memories)
	}
}

func TestOpenPrefersV2SchemaAndReadsV2Data(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	createV2Database(t, path)
	repo, err := Open(context.Background(), path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	if repo.SchemaVersion() != SchemaV2 {
		t.Fatalf("schema version = %s, want v2", repo.SchemaVersion())
	}
	sessions, err := repo.ListSessions(context.Background(), SessionFilter{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].ID != "ses_v2_root" {
		t.Fatalf("unexpected v2 sessions: %#v", sessions)
	}
	if sessions[0].Title != "" || sessions[0].Model != "gpt-5.5" {
		t.Fatalf("unexpected nullable title or JSON model handling: %#v", sessions[0])
	}
	count, err := repo.CountSessions(context.Background(), SessionFilter{})
	if err != nil || count != 1 {
		t.Fatalf("v2 count = %d, err = %v", count, err)
	}
	stats, err := repo.SessionStats(context.Background(), "ses_v2_root")
	if err != nil {
		t.Fatal(err)
	}
	if stats.MessageCount != 2 || stats.PartCount != 3 || stats.SizeBytes == 0 {
		t.Fatalf("unexpected v2 stats: %#v", stats)
	}
	previews, err := repo.RecentUserMessages(context.Background(), "ses_v2_root", 5, 500)
	if err != nil {
		t.Fatal(err)
	}
	if len(previews) != 1 || previews[0].Text != "root v2 message" {
		t.Fatalf("unexpected v2 previews: %#v", previews)
	}
	memories, err := repo.RecentUserMemory(context.Background(), time.Now().Add(-time.Hour), 100, 4000)
	if err != nil {
		t.Fatal(err)
	}
	if len(memories) != 2 || memories[0].SessionID != "ses_v2_root" || memories[1].SessionID != "ses_v2_root" {
		t.Fatalf("unexpected v2 memories: %#v", memories)
	}
}

func TestV2WritesUseOpenCodeAPI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	createV2Database(t, path)
	repo, err := Open(context.Background(), path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	type call struct {
		method string
		path   string
		body   any
	}
	calls := []call{}
	repo.callAPI = func(_ context.Context, method string, path string, body any) error {
		calls = append(calls, call{method: method, path: path, body: body})
		if method == "patch" {
			return errors.New("HTTP 404 Not Found")
		}
		return nil
	}
	if err := repo.UpdateSessionTitle(context.Background(), "ses_v2_root", "new title"); err != nil {
		t.Fatal(err)
	}
	impact, err := repo.DeleteImpact(context.Background(), "ses_v2_root")
	if err != nil {
		t.Fatal(err)
	}
	if impact != (DeleteImpact{SessionCount: 2, MessageCount: 3, PartCount: 4}) {
		t.Fatalf("unexpected v2 impact: %#v", impact)
	}
	if err := repo.DeleteSessionIfUnchanged(context.Background(), "ses_v2_root", impact); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 3 || calls[0].method != "patch" || calls[1].method != "post" || calls[2].method != "delete" {
		t.Fatalf("unexpected API calls: %#v", calls)
	}
	if calls[0].path != "/api/session/ses_v2_root" || calls[1].path != "/api/session/ses_v2_root/rename" || calls[2].path != "/api/session/ses_v2_root" {
		t.Fatalf("unexpected API paths: %#v", calls)
	}
	body, ok := calls[0].body.(map[string]string)
	if !ok || body["title"] != "new title" {
		t.Fatalf("unexpected update body: %#v", calls[0].body)
	}
}

func TestOpenRejectsMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.db")
	if _, err := Open(context.Background(), path, false); err == nil {
		t.Fatal("read-write Open created a missing database")
	}
}

func TestOpenRejectsIncompatibleBrowseSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("create table session (id text primary key, title text not null)"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = Open(context.Background(), path, false)
	if err == nil || !strings.Contains(err.Error(), "missing session columns") {
		t.Fatalf("Open error = %v, want readable schema incompatibility", err)
	}
	if !strings.Contains(err.Error(), "update lazyocs or select a compatible OpenCode database") {
		t.Fatalf("Open error is not actionable: %v", err)
	}
}

func TestOpenDoesNotFallBackToV1WhenV2SchemaIsIncomplete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	createCompatibleDatabase(t, path)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("create table session_v2 (id text primary key)"); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = Open(context.Background(), path, false)
	if err == nil || !strings.Contains(err.Error(), "missing session_v2 columns") {
		t.Fatalf("Open error = %v, want V2 incompatibility instead of V1 fallback", err)
	}
}

func TestWriteOperationsFailClosedWithoutDeleteCascades(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	createCompatibleDatabase(t, path)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("drop table part; create table part (id text primary key, message_id text, session_id text, data text, time_created integer)"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	repo, err := Open(context.Background(), path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	if repo.Compatibility().Delete {
		t.Fatal("schema without part cascade was marked delete-compatible")
	}
	if err := repo.DeleteSession(context.Background(), "ses_root"); err == nil || !strings.Contains(err.Error(), "delete cascades") {
		t.Fatalf("DeleteSession error = %v, want fail-closed error", err)
	}
}

func createCompatibleDatabase(t testing.TB, path string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
pragma foreign_keys = on;
create table session (
  id text primary key, project_id text not null, parent_id text,
  title text not null, directory text not null,
  time_created integer not null, time_updated integer not null,
  model text, agent text, cost real not null default 0,
  tokens_input integer not null default 0, tokens_output integer not null default 0,
  tokens_reasoning integer not null default 0, tokens_cache_read integer not null default 0,
  tokens_cache_write integer not null default 0
);
create table message (
  id text primary key, session_id text not null references session(id) on delete cascade,
  data text not null, time_created integer not null
);
create table part (
  id text primary key, message_id text not null references message(id) on delete cascade,
  session_id text not null, data text not null, time_created integer not null
);
insert into session (id, project_id, title, directory, time_created, time_updated)
values ('ses_root', 'global', 'root', '/tmp', 1, 1);
insert into message values ('msg_root', 'ses_root', '{}', 1);
insert into part values ('part_root', 'msg_root', 'ses_root', '{}', 1);
`)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func createV2Database(t testing.TB, path string) {
	t.Helper()
	createCompatibleDatabase(t, path)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UnixMilli()
	_, err = db.Exec(`
create table session_v2 (
  id text primary key, project_id text not null, workspace_id text, parent_id text,
  title text, directory text not null, time_created integer not null, time_updated integer not null,
  model text, agent text, cost real not null default 0,
  tokens_input integer not null default 0, tokens_output integer not null default 0,
  tokens_reasoning integer not null default 0, tokens_cache_read integer not null default 0,
  tokens_cache_write integer not null default 0
);
create table session_message (
  id text primary key, session_id text not null, type text not null, seq integer not null,
  time_created integer not null, time_updated integer not null, data text not null
);
insert into session_v2 (id, project_id, title, directory, time_created, time_updated, model, agent)
values ('ses_v2_root', 'global', null, '/tmp/v2', ?, ?, '{"id":"gpt-5.5","providerID":"openai"}', 'build');
insert into session_v2 (id, project_id, parent_id, title, directory, time_created, time_updated)
values ('ses_v2_child', 'global', 'ses_v2_root', 'v2 child', '/tmp/v2', ?, ?);
insert into session_message values
  ('msg_v2_user', 'ses_v2_root', 'user', 1, ?, ?, '{"text":"root v2 message","time":{"created":1}}'),
  ('msg_v2_assistant', 'ses_v2_root', 'assistant', 2, ?, ?, '{"content":[{"type":"text","text":"answer"},{"type":"tool"}]}'),
  ('msg_v2_child_user', 'ses_v2_child', 'user', 1, ?, ?, '{"text":"child v2 memory","time":{"created":1}}');
`, now, now, now, now, now, now, now, now, now, now)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}
