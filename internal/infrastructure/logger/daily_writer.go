package logger

import (
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	PrefixApplication = "application"
	PrefixAccess      = "access"
	PrefixAudit       = "audit"

	dateLayout       = "20060102"
	defaultMaxSizeMB = 100
)

// dailyWriter は日付付きファイル名で lumberjack に書き出す。
// 日付が変わると現在のファイルを閉じ、新しい日付のファイルへ切り替える。
// 同一日内のサイズ超過は lumberjack がバックアップへローテーションする。
// https://github.com/natefinch/lumberjack
type dailyWriter struct {
	mu        sync.Mutex
	dir       string
	prefix    string
	now       func() time.Time
	maxSizeMB int
	date      string
	inner     *lumberjack.Logger
}

func newDailyWriter(dir, prefix string) *dailyWriter {
	return &dailyWriter{
		dir:       dir,
		prefix:    prefix,
		now:       time.Now,
		maxSizeMB: defaultMaxSizeMB,
	}
}

func (w *dailyWriter) Write(p []byte) (int, error) {
	date := w.now().Format(dateLayout)

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.inner == nil || w.date != date {
		if err := w.rotateLocked(date); err != nil {
			return 0, err
		}
	}
	return w.inner.Write(p)
}

func (w *dailyWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.inner == nil {
		return nil
	}
	err := w.inner.Close()
	w.inner = nil
	w.date = ""
	return err
}

func (w *dailyWriter) rotateLocked(date string) error {
	if w.inner != nil {
		if err := w.inner.Close(); err != nil {
			return err
		}
	}

	maxSize := w.maxSizeMB
	if maxSize <= 0 {
		maxSize = defaultMaxSizeMB
	}

	w.date = date
	w.inner = &lumberjack.Logger{
		Filename:   filepath.Join(w.dir, w.prefix+"-"+date+".log"),
		MaxSize:    maxSize,
		MaxAge:     0,
		MaxBackups: 0,
		LocalTime:  true,
		Compress:   false,
	}
	return nil
}
