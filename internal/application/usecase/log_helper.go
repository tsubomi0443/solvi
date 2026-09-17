package usecase

import (
	"context"
	"errors"
	"log/slog"

	logutils "solvi/internal/shared/logUtils"

	"gorm.io/gorm"
)

func LogBusinessWarn(ctx context.Context, op, msg string, err error, attrs ...any) {
	attrs = append(attrs, slog.String("err", err.Error()))
	logutils.Warn(ctx, logutils.LayerUsecase, op, msg, attrs...)
}

func LogRepoPropagation(ctx context.Context, op, msg string, err error, attrs ...any) {
	attrs = append(attrs, slog.String("err", err.Error()))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		logutils.Warn(ctx, logutils.LayerUsecase, op, msg, attrs...)
		return
	}
	logutils.Debug(ctx, logutils.LayerUsecase, op, msg, attrs...)
}
