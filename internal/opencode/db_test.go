package opencode

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSearchTerms(t *testing.T) {
	terms := SearchTerms("  Windows   ISO windows  ")
	if len(terms) != 2 || terms[0] != "windows" || terms[1] != "iso" {
		t.Fatalf("unexpected terms: %#v", terms)
	}
}

func TestListSessionsRequiresAllSearchTerms(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
create table session (
  id text primary key,
  project_id text not null,
  parent_id text,
  title text not null,
  directory text not null,
  time_created integer not null,
  time_updated integer not null,
  model text,
  agent text,
  cost real not null default 0,
  tokens_input integer not null default 0,
  tokens_output integer not null default 0,
  tokens_reasoning integer not null default 0,
  tokens_cache_read integer not null default 0,
  tokens_cache_write integer not null default 0
);
insert into session (id, project_id, title, directory, time_created, time_updated)
values
  ('ses_1', 'global', 'Windows ISO download', '/tmp/a', 1, 3),
  ('ses_2', 'global', 'Windows reinstall', '/tmp/b', 1, 2),
  ('ses_3', 'global', 'ISO checksum', '/tmp/c', 1, 1);
`)
	if err != nil {
		t.Fatal(err)
	}

	repo := &SQLiteRepository{db: db}
	sessions, err := repo.ListSessions(context.Background(), SessionFilter{Query: "windows iso", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].ID != "ses_1" {
		t.Fatalf("unexpected sessions: %#v", sessions)
	}

	count, err := repo.CountSessions(context.Background(), SessionFilter{Query: "windows iso"})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("unexpected count: %d", count)
	}
}

func TestSessionStats(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
create table message (
  id text primary key,
  session_id text not null,
  data text not null
);
create table part (
  id text primary key,
  message_id text not null,
  session_id text not null,
  data text not null
);
insert into message (id, session_id, data) values
  ('msg_1', 'ses_1', 'hello'),
  ('msg_2', 'ses_1', 'world!'),
  ('msg_3', 'ses_2', 'ignored');
insert into part (id, message_id, session_id, data) values
  ('part_1', 'msg_1', 'ses_1', 'abc'),
  ('part_2', 'msg_2', 'ses_1', 'defg'),
  ('part_3', 'msg_3', 'ses_2', 'ignored');
`)
	if err != nil {
		t.Fatal(err)
	}

	repo := &SQLiteRepository{db: db}
	stats, err := repo.SessionStats(context.Background(), "ses_1")
	if err != nil {
		t.Fatal(err)
	}
	if stats.MessageCount != 2 || stats.PartCount != 2 || stats.SizeBytes != 18 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
}
