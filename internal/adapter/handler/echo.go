package handler

import (
	"html/template"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func NewEcho(outputs io.Writer) *echo.Echo {
	ec := echo.New()

	ec.Use(middleware.RequestLogger())
	ec.Use(middleware.Recover())

	tmpl := template.New("")
	tmpl.Funcs(template.FuncMap{
		"dict":   MakeMapFunc,
		"themes": ThemeOptions,
	})

	err := filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// .html ファイルのみを対象とする
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

	// テンプレートの設定
	ec.Renderer = &echo.TemplateRenderer{
		Template: tmpl,
	}

	handler := slog.NewJSONHandler(outputs, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	ec.Logger = slog.New(handler)

	// ミドルウェア
	ec.Use(middleware.Recover(), middleware.RequestLogger())
	return ec
}
