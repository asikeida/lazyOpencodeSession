package opencode

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
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

func createCompatibleDatabase(t *testing.T, path string) {
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
