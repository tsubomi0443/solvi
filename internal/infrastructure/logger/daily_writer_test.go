package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDailyWriter_rollsOnDateChange(t *testing.T) {
	dir := t.TempDir()
	current := time.Date(2026, 9, 17, 23, 59, 0, 0, time.Local)
	w := &dailyWriter{
		dir:       dir,
		prefix:    PrefixApplication,
		now:       func() time.Time { return current },
		maxSizeMB: 100,
	}
	t.Cleanup(func() { _ = w.Close() })

	if _, err := w.Write([]byte("day1\n")); err != nil {
		t.Fatalf("write day1: %v", err)
	}

	current = time.Date(2026, 9, 18, 0, 0, 1, 0, time.Local)
	if _, err := w.Write([]byte("day2\n")); err != nil {
		t.Fatalf("write day2: %v", err)
	}

	day1 := filepath.Join(dir, "application-20260917.log")
	day2 := filepath.Join(dir, "application-20260918.log")
	if _, err := os.Stat(day1); err != nil {
		t.Fatalf("expected %s: %v", day1, err)
	}
	if _, err := os.Stat(day2); err != nil {
		t.Fatalf("expected %s: %v", day2, err)
	}
}

func TestDailyWriter_CloseWhenInnerNil(t *testing.T) {
	w := newDailyWriter(t.TempDir(), PrefixAccess)
	if err := w.Close(); err != nil {
		t.Fatalf("expected nil error on uninitialized close, got %v", err)
	}
}

func TestLoggers_NewAndClose(t *testing.T) {
	t.Run("empty directory fails", func(t *testing.T) {
		_, err := New(Options{Dir: ""})
		if err == nil {
			t.Fatal("expected error for empty directory")
		}
	})

	t.Run("success with access stdout and close", func(t *testing.T) {
		dir := t.TempDir()
		loggers, err := New(Options{
			Dir:          dir,
			AccessStdout: true,
		})
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}
		if loggers.App == nil || loggers.Access == nil || loggers.Audit == nil {
			t.Fatal("expected all slog loggers to be initialized")
		}

		loggers.App.Info("test app log")
		loggers.Access.Info("test access log")
		loggers.Audit.Info("test audit log")

		if err := loggers.Close(); err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		var nilLoggers *Loggers
		if err := nilLoggers.Close(); err != nil {
			t.Fatalf("nil Loggers close must return nil, got %v", err)
		}
	})
}
