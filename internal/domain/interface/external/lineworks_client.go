package external

import (
	"context"
	"fmt"
	"time"

	"solvi/internal/domain/valueobject"
)

//go:generate go tool mockgen -typed -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

// LineWorksSendError は送信失敗。Error 文字列に秘密情報や応答本文を含めない。
type LineWorksSendError struct {
	Kind       valueobject.LineWorksErrorKind
	HTTPStatus int
	RetryAfter time.Duration
}

func (e *LineWorksSendError) Error() string {
	return fmt.Sprintf("line works send failed: kind=%d status=%d", e.Kind, e.HTTPStatus)
}

type LineWorksClient interface {
	SendChannelMessage(ctx context.Context, channelID, text string) (int, error)
	SendUserMessage(ctx context.Context, userID, text string) (int, error)
}
