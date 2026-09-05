package opencode

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
)

func BenchmarkStartupOpenAndList(b *testing.B) {
	path := createBenchmarkDatabase(b)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		repo, err := Open(context.Background(), path, true)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := repo.ListSessions(context.Background(), SessionFilter{Limit: 500}); err != nil {
			repo.Close()
			b.Fatal(err)
		}
		if _, err := repo.CountSessions(context.Background(), SessionFilter{}); err != nil {
			repo.Close()
			b.Fatal(err)
		}
		if err := repo.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRecentUserMessages(b *testing.B) {
	repo := openBenchmarkRepository(b)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := repo.RecentUserMessages(context.Background(), "ses_bench_000", 5, 500); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSessionStats(b *testing.B) {
	repo := openBenchmarkRepository(b)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := repo.SessionStats(context.Background(), "ses_bench_000"); err != nil {
			b.Fatal(err)
		}
	}
}

func openBenchmarkRepository(b *testing.B) *SQLiteRepository {
	b.Helper()
	repo, err := Open(context.Background(), createBenchmarkDatabase(b), true)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { repo.Close() })
	return repo
}

func createBenchmarkDatabase(b *testing.B) string {
	b.Helper()
	path := filepath.Join(b.TempDir(), "opencode.db")
	createCompatibleDatabase(b, path)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		b.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`
create index session_parent_idx on session(parent_id);
create index session_time_updated_idx on session(time_updated desc);
create index message_session_time_created_id_idx on message(session_id, time_created desc, id);
create index part_message_id_id_idx on part(message_id, id);
create index part_session_id_idx on part(session_id);`); err != nil {
		b.Fatal(err)
	}
	for sessionIndex := range 500 {
		sessionID := fmt.Sprintf("ses_bench_%03d", sessionIndex)
		title := fmt.Sprintf("Benchmark session %03d", sessionIndex)
		if sessionIndex%10 == 0 {
			title += " Windows ISO"
		}
		if _, err := tx.Exec(`insert into session
(id, project_id, title, directory, time_created, time_updated)
values (?, 'global', ?, '/tmp/benchmark', ?, ?)`, sessionID, title, sessionIndex, sessionIndex); err != nil {
			b.Fatal(err)
		}
		for messageIndex := range 10 {
			messageID := fmt.Sprintf("msg_%03d_%02d", sessionIndex, messageIndex)
			partID := fmt.Sprintf("part_%03d_%02d", sessionIndex, messageIndex)
			created := sessionIndex*10 + messageIndex
			if _, err := tx.Exec(`insert into message (id, session_id, data, time_created)
values (?, ?, '{"role":"user"}', ?)`, messageID, sessionID, created); err != nil {
				b.Fatal(err)
			}
			if _, err := tx.Exec(`insert into part (id, message_id, session_id, data, time_created)
values (?, ?, ?, ?, ?)`, partID, messageID, sessionID, fmt.Sprintf(`{"type":"text","text":"benchmark message %d windows checksum"}`, messageIndex), created); err != nil {
				b.Fatal(err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		b.Fatal(err)
	}
	return path
}
