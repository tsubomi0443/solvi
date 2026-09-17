package log_usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"solvi/internal/domain/entity"
)

type mockLogRepo struct {
	listDatesFn  func(ctx context.Context) ([]string, error)
	listNamesFn  func(ctx context.Context) ([]string, error)
	listByDateFn func(ctx context.Context, date time.Time) ([]string, error)
	openFn       func(ctx context.Context, name string) (io.ReadCloser, error)
}

func (m *mockLogRepo) ListDates(ctx context.Context) ([]string, error) {
	if m.listDatesFn != nil {
		return m.listDatesFn(ctx)
	}
	return nil, nil
}

func (m *mockLogRepo) ListNames(ctx context.Context) ([]string, error) {
	if m.listNamesFn != nil {
		return m.listNamesFn(ctx)
	}
	return nil, nil
}

func (m *mockLogRepo) ListByDate(ctx context.Context, date time.Time) ([]string, error) {
	if m.listByDateFn != nil {
		return m.listByDateFn(ctx, date)
	}
	return nil, nil
}

func (m *mockLogRepo) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	if m.openFn != nil {
		return m.openFn(ctx, name)
	}
	return io.NopCloser(strings.NewReader("dummy log")), nil
}

func TestListAvailableDates(t *testing.T) {
	t.Run("returns dates on success", func(t *testing.T) {
		repo := &mockLogRepo{
			listDatesFn: func(ctx context.Context) ([]string, error) {
				return []string{"2026-09-18", "2026-09-17"}, nil
			},
		}
		uc := NewLogUsecase(repo)
		dates := uc.ListAvailableDates(context.Background())
		if len(dates) != 2 || dates[0] != "2026-09-18" {
			t.Fatalf("unexpected dates: %v", dates)
		}
	})

	t.Run("returns empty slice on repo error", func(t *testing.T) {
		repo := &mockLogRepo{
			listDatesFn: func(ctx context.Context) ([]string, error) {
				return nil, errors.New("io error")
			},
		}
		uc := NewLogUsecase(repo)
		dates := uc.ListAvailableDates(context.Background())
		if dates == nil || len(dates) != 0 {
			t.Fatalf("expected empty non-nil slice, got: %v", dates)
		}
	})
}

func TestIssueAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockLogRepo{
			listNamesFn: func(ctx context.Context) ([]string, error) {
				return []string{"app.log"}, nil
			},
		}
		uc := NewLogUsecase(repo)
		res, err := uc.IssueAll(context.Background(), "user-1")
		if err != nil {
			t.Fatal(err)
		}
		if res.DownloadKey == "" || len(res.Password) == 0 || res.Filename != "solvi-logs-all.zip" {
			t.Fatalf("unexpected issue result: %+v", res)
		}
	})

	t.Run("repo error and empty files", func(t *testing.T) {
		repoErr := &mockLogRepo{
			listNamesFn: func(ctx context.Context) ([]string, error) {
				return nil, errors.New("fail")
			},
		}
		ucErr := NewLogUsecase(repoErr)
		if _, err := ucErr.IssueAll(context.Background(), "user-1"); err == nil {
			t.Fatal("expected error on repo failure")
		}

		repoEmpty := &mockLogRepo{
			listNamesFn: func(ctx context.Context) ([]string, error) {
				return []string{}, nil
			},
		}
		ucEmpty := NewLogUsecase(repoEmpty)
		if _, err := ucEmpty.IssueAll(context.Background(), "user-1"); err == nil {
			t.Fatal("expected error on empty log list")
		}
	})
}

func TestIssueDate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockLogRepo{
			listByDateFn: func(ctx context.Context, date time.Time) ([]string, error) {
				return []string{"access-20260918.log"}, nil
			},
		}
		uc := NewLogUsecase(repo)
		res, err := uc.IssueDate(context.Background(), "user-1", "2026-09-18")
		if err != nil {
			t.Fatal(err)
		}
		if res.Filename != "solvi-logs-20260918.zip" {
			t.Fatalf("unexpected filename: %s", res.Filename)
		}
	})

	t.Run("invalid date and empty list", func(t *testing.T) {
		repo := &mockLogRepo{}
		uc := NewLogUsecase(repo)
		if _, err := uc.IssueDate(context.Background(), "user-1", "invalid-date"); err == nil {
			t.Fatal("expected date parse error")
		}

		repoEmpty := &mockLogRepo{
			listByDateFn: func(ctx context.Context, date time.Time) ([]string, error) {
				return []string{}, nil
			},
		}
		ucEmpty := NewLogUsecase(repoEmpty)
		if _, err := ucEmpty.IssueDate(context.Background(), "user-1", "2026-09-18"); err == nil {
			t.Fatal("expected error on empty date logs")
		}
	})
}

func TestLookupTicket(t *testing.T) {
	uc := NewLogUsecase(&mockLogRepo{})
	key := "valid-key"
	uc.tickets[key] = entity.DownloadTicket{
		Key:       key,
		UserUUID:  "user-1",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	t.Run("success", func(t *testing.T) {
		ticket, err := uc.LookupTicket(context.Background(), key, "user-1")
		if err != nil {
			t.Fatal(err)
		}
		if ticket.Key != key {
			t.Fatalf("unexpected ticket: %+v", ticket)
		}
	})

	t.Run("missing key", func(t *testing.T) {
		if _, err := uc.LookupTicket(context.Background(), "missing", "user-1"); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("wrong user", func(t *testing.T) {
		if _, err := uc.LookupTicket(context.Background(), key, "user-2"); err == nil {
			t.Fatal("expected error for wrong user")
		}
	})

	t.Run("expired ticket", func(t *testing.T) {
		expiredKey := "exp-key"
		uc.tickets[expiredKey] = entity.DownloadTicket{
			Key:       expiredKey,
			UserUUID:  "user-1",
			ExpiresAt: time.Now().Add(-1 * time.Minute),
		}
		if _, err := uc.LookupTicket(context.Background(), expiredKey, "user-1"); err == nil {
			t.Fatal("expected expired ticket error")
		}
	})
}

func TestStream_SuccessAndDeletesTicket(t *testing.T) {
	repo := &mockLogRepo{
		listNamesFn: func(ctx context.Context) ([]string, error) {
			return []string{"app.log"}, nil
		},
		openFn: func(ctx context.Context, name string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("log line content\n")), nil
		},
	}
	uc := NewLogUsecase(repo)
	res, err := uc.IssueAll(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := uc.Stream(context.Background(), &buf, res.DownloadKey, "user-1"); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected zip output")
	}

	// 完了後はチケットが削除されていること
	if _, err := uc.LookupTicket(context.Background(), res.DownloadKey, "user-1"); err == nil {
		t.Fatal("ticket should have been deleted after successful stream")
	}
}

func TestStream_UnknownKind(t *testing.T) {
	uc := NewLogUsecase(&mockLogRepo{})
	key := "unknown-kind"
	uc.tickets[key] = entity.DownloadTicket{
		Key:       key,
		UserUUID:  "user-1",
		Kind:      entity.DownloadKind(999),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	var buf bytes.Buffer
	if err := uc.Stream(context.Background(), &buf, key, "user-1"); err == nil {
		t.Fatal("expected error for unknown ticket kind")
	}
}
