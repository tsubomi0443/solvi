package repository

import (
	"io"
	"time"
)

type LogRepository interface {
	ListNames() ([]string, error)
	Open(name string) (io.ReadCloser, error)
	ListDates() ([]string, error)
	ListByDate(date time.Time) ([]string, error)
}
