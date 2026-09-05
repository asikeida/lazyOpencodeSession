package opencode

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestActionableErrorExplainsDatabaseTimeout(t *testing.T) {
	err := ActionableError(context.DeadlineExceeded)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("wrapped error lost deadline: %v", err)
	}
	for _, want := range []string{"timed out", "retry"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error does not contain %q: %v", want, err)
		}
	}
}

func TestActionableErrorLeavesOtherErrorsUnchanged(t *testing.T) {
	want := errors.New("query failed")
	if got := ActionableError(want); got != want {
		t.Fatalf("ActionableError returned %v, want original error", got)
	}
}

func TestActionableErrorExplainsSQLiteWriteLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	createCompatibleDatabase(t, path)
	locker, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer locker.Close()
	contender, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer contender.Close()

	tx, err := locker.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec("update session set title = 'locked' where id = 'ses_root'"); err != nil {
		t.Fatal(err)
	}
	_, err = contender.Exec("update session set title = 'blocked' where id = 'ses_root'")
	if err == nil {
		t.Fatal("second writer unexpectedly acquired the database lock")
	}
	err = ActionableError(err)
	for _, want := range []string{"busy or locked", "retry"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error does not contain %q: %v", want, err)
		}
	}
}
