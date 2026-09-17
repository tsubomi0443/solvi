package repository

import (
	"context"
	"io"
	"time"
)

type LogRepository interface {
	ListNames(ctx context.Context) ([]string, error)
	Open(ctx context.Context, name string) (io.ReadCloser, error)
	ListDates(ctx context.Context) ([]string, error)
	ListByDate(ctx context.Context, date time.Time) ([]string, error)
}
