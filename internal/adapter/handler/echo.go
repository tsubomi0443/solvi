package handler

import (
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func NewEcho(accessLogger *slog.Logger) *echo.Echo {
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

	ec.Use(RequestID(), middleware.Recover(), middleware.RequestLogger())
	return ec
}
