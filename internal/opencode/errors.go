package opencode

import (
	"context"
	"errors"
	"fmt"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

func ActionableError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("OpenCode database operation timed out; wait for OpenCode to finish writing, then retry: %w", err)
	}
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		switch sqliteErr.Code() & 0xff {
		case sqlite3.SQLITE_BUSY, sqlite3.SQLITE_LOCKED:
			return fmt.Errorf("OpenCode database is busy or locked; wait for OpenCode to finish writing, then retry: %w", err)
		}
	}
	return err
}
