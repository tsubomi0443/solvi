package repository

import (
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

	dates, err := repo.ListDates()
	if err != nil {
		t.Fatal(err)
	}
	if len(dates) != 2 || dates[0] != "2026-09-17" || dates[1] != "2026-09-16" {
		t.Fatalf("dates = %v", dates)
	}

	day := time.Date(2026, 9, 17, 0, 0, 0, 0, time.Local)
	got, err := repo.ListByDate(day)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("ListByDate = %v", got)
	}
}
