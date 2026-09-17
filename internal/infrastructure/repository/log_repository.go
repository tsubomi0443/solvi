package repository

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"regexp"
	"sort"
	"time"
)

type LogRepository struct {
	logDirectory *os.Root
}

func NewLogRepository(logDirectoryRootPath string) (*LogRepository, error) {
	root, err := os.OpenRoot(logDirectoryRootPath)
	if err != nil {
		slog.Error("root directory could not be found", slog.String("error", err.Error()))
		return nil, fmt.Errorf("ログディレクトリのルートが参照できませんでした: %w", err)
	}
	return &LogRepository{
		logDirectory: root,
	}, nil
}

func (repo *LogRepository) Close() error {
	if repo == nil || repo.logDirectory == nil {
		return nil
	}
	return repo.logDirectory.Close()
}

var (
	newLogNameRE = regexp.MustCompile(`^(?:application|access|audit)-(\d{8})(?:-.+)?\.log$`)
	oldAppLogRE  = regexp.MustCompile(`^log_(\d{8})\.log$`)
	oldAccLogRE  = regexp.MustCompile(`^log_(\d{8})_access\.log$`)
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
		slog.Error("")
		return nil, fmt.Errorf("ログファイルを開けませんでした: %w", err)
	}
	return f, nil
}

func (repo *LogRepository) ListDates() ([]string, error) {
	names, err := repo.ListNames()
	if err != nil {
		slog.Error("")
		return nil, err
	}

	dates := make(map[string]struct{})
	for _, name := range names {
		compact, ok := parseLogDate(name)
		if !ok {
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

func (repo *LogRepository) ListByDate(date time.Time) ([]string, error) {
	names, err := repo.ListNames()
	if err != nil {
		slog.Error("")
		return nil, err
	}

	want := date.Format("20060102")
	matched := make([]string, 0)
	for _, name := range names {
		compact, ok := parseLogDate(name)
		if !ok || compact != want {
			continue
		}
		matched = append(matched, name)
	}
	return matched, nil
}

func parseLogDate(name string) (string, bool) {
	if m := newLogNameRE.FindStringSubmatch(name); len(m) == 2 {
		return m[1], true
	}
	if m := oldAccLogRE.FindStringSubmatch(name); len(m) == 2 {
		return m[1], true
	}
	if m := oldAppLogRE.FindStringSubmatch(name); len(m) == 2 {
		return m[1], true
	}
	return "", false
}

func formatLogDate(compact string) string {
	return compact[:4] + "-" + compact[4:6] + "-" + compact[6:8]
}
