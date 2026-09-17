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
