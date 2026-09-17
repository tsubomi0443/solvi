package logger

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type Options struct {
	Dir          string
	Level        slog.Level
	AccessStdout bool
}

type Loggers struct {
	App    *slog.Logger
	Access *slog.Logger
	Audit  *slog.Logger

	closers []io.Closer
}

func New(opts Options) (*Loggers, error) {
	if opts.Dir == "" {
		return nil, fmt.Errorf("log directory is empty")
	}
	if err := os.MkdirAll(opts.Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	appW := newDailyWriter(opts.Dir, PrefixApplication)
	accessW := newDailyWriter(opts.Dir, PrefixAccess)
	auditW := newDailyWriter(opts.Dir, PrefixAudit)

	handlerOpts := &slog.HandlerOptions{Level: opts.Level}

	var accessOut io.Writer = accessW
	if opts.AccessStdout {
		accessOut = io.MultiWriter(os.Stdout, accessW)
	}

	return &Loggers{
		App:     slog.New(slog.NewJSONHandler(appW, handlerOpts)),
		Access:  slog.New(slog.NewJSONHandler(accessOut, &slog.HandlerOptions{Level: slog.LevelInfo})),
		Audit:   slog.New(slog.NewJSONHandler(auditW, &slog.HandlerOptions{Level: slog.LevelInfo})),
		closers: []io.Closer{appW, accessW, auditW},
	}, nil
}

func (l *Loggers) Close() error {
	if l == nil {
		return nil
	}
	var errs []error
	for _, c := range l.closers {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
