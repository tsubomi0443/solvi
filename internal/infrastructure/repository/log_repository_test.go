package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogRepository_listNewAndLegacyNames(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		"application-20260917.log",
		"access-20260917.log",
		"audit-20260917.log",
		"application-20260917-2026-09-17T18-30-00.000.log",
		"log_20260916.log",
		"log_20260916_access.log",
		"readme.txt",
	}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	repo, err := NewLogRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	ctx := context.Background()
	dates, err := repo.ListDates(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(dates) != 2 || dates[0] != "2026-09-17" || dates[1] != "2026-09-16" {
		t.Fatalf("dates = %v", dates)
	}

	day := time.Date(2026, 9, 17, 0, 0, 0, 0, time.Local)
	got, err := repo.ListByDate(ctx, day)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("ListByDate = %v", got)
	}

	// Open existing file
	rc, err := repo.Open(ctx, "application-20260917.log")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	_ = rc.Close()

	// Open non-existing file
	if _, err := repo.Open(ctx, "not-found.log"); err == nil {
		t.Fatal("expected error opening missing log")
	}

	// ListByDate with no matching files
	dayNone := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	none, err := repo.ListByDate(ctx, dayNone)
	if err != nil {
		t.Fatal(err)
	}
	if len(none) != 0 {
		t.Fatalf("expected 0 files, got %d", len(none))
	}
}

func TestLogRepository_InvalidRootPath(t *testing.T) {
	_, err := NewLogRepository(filepath.Join(t.TempDir(), "nonexistent-dir"))
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}
