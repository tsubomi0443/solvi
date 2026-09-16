package repository

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"
	"time"
)

type LogRepository struct {
	logDirectory *os.Root
}

func NewLogRepository(logDirectoryRootPath string) (*LogRepository, error) {
	root, err := os.OpenRoot(logDirectoryRootPath)
	if err != nil {
		return nil, fmt.Errorf("ログディレクトリのルートが参照できませんでした: %w", err)
	}
	return &LogRepository{
		logDirectory: root,
	}, nil
}

const (
	logNamePattern       = "log_20060102.log"
	accessLogNamePattern = "log_20060102_access.log"
)

func (repo *LogRepository) ListNames() ([]string, error) {
	names := []string{}
	err := fs.WalkDir(repo.logDirectory.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		names = append(names, d.Name())
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("ログファイルの走査中にエラーが発生しました: %w", err)
	}
	sort.Strings(names)
	return names, nil
}

func (repo *LogRepository) Open(name string) (io.ReadCloser, error) {
	f, err := repo.logDirectory.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("ログファイルを開けませんでした: %w", err)
	}
	return f, nil
}

func (repo *LogRepository) ListDates() ([]string, error) {
	names, err := repo.ListNames()
	if err != nil {
		return nil, err
	}

	dates := make(map[string]struct{})
	for _, name := range names {
		if !strings.HasPrefix(name, "log_") || !strings.HasSuffix(name, ".log") || strings.Contains(name, "_access") {
			continue
		}
		compact := strings.TrimSuffix(strings.TrimPrefix(name, "log_"), ".log")
		if len(compact) != 8 {
			continue
		}
		date, err := time.ParseInLocation("20060102", compact, time.Local)
		if err != nil {
			continue
		}
		if !repo.HasDatePair(date) {
			continue
		}
		dates[formatLogDate(compact)] = struct{}{}
	}

	result := make([]string, 0, len(dates))
	for date := range dates {
		result = append(result, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(result)))
	return result, nil
}

func (repo *LogRepository) HasDatePair(date time.Time) bool {
	logName := date.Format(logNamePattern)
	accessLogName := date.Format(accessLogNamePattern)
	if _, err := repo.logDirectory.OpenFile(logName, os.O_RDONLY, 0); err != nil {
		return false
	}
	if _, err := repo.logDirectory.OpenFile(accessLogName, os.O_RDONLY, 0); err != nil {
		return false
	}
	return true
}

func formatLogDate(compact string) string {
	return compact[:4] + "-" + compact[4:6] + "-" + compact[6:8]
}
