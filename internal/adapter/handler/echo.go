package handler

import (
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func NewEcho(accessLogger, panicLogger *slog.Logger) *echo.Echo {
	ec := echo.New()

	tmpl := template.New("")
	tmpl.Funcs(template.FuncMap{
		"dict":   MakeMapFunc,
		"themes": ThemeOptions,
	})

	err := filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			if _, err := tmpl.ParseFiles(path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		slog.Error("failed to load templates", "err", err)
	}

	ec.Renderer = &echo.TemplateRenderer{
		Template: tmpl,
	}

	if accessLogger != nil {
		ec.Logger = accessLogger
	}

	recover := newRecoverMiddleware(panicLogger)
	ec.Use(RequestID(), middleware.RequestLogger(), recover.middleware)
	return ec
}

type recoverMiddleware struct {
	panicLogger *slog.Logger
}

func newRecoverMiddleware(panicLogger *slog.Logger) *recoverMiddleware {
	return &recoverMiddleware{
		panicLogger: panicLogger,
	}
}

func (rm *recoverMiddleware) middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				rm.panicLogger.Error(
					"Unhandled panic occurred",
					slog.Any("panic", r),
					slog.String("stacktrace", string(stack)),
				)
			}
		}()

		return next(c)
	}
}
