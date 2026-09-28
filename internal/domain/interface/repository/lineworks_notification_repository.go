package repository

import (
	"context"
	"time"

	"solvi/internal/domain/entity/lineworks"
)

//go:generate go tool mockgen -typed -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

type LineWorksNotificationRepository interface {
	ClaimDue(ctx context.Context, now time.Time, limit int) ([]lineworks.Notification, error)
	RecoverStale(ctx context.Context, staleBefore time.Time, maxAttempts int) (int, error)
	SaveResult(ctx context.Context, notice *lineworks.Notification, attempt *lineworks.NotificationAttempt) error
}
