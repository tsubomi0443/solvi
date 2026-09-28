package lineworks_usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"solvi/internal/domain/entity/lineworks"
	ext "solvi/internal/domain/interface/external"
	repo "solvi/internal/domain/interface/repository"
	"solvi/internal/domain/valueobject"
	logutils "solvi/internal/shared/logUtils"
)

const opLineWorks = "lineworks_usecase"

type Usecase struct {
	repo   repo.LineWorksNotificationRepository
	client ext.LineWorksClient
}

func New(noticeRepo repo.LineWorksNotificationRepository, client ext.LineWorksClient) *Usecase {
	return &Usecase{repo: noticeRepo, client: client}
}

func (uc *Usecase) ProcessDue(ctx context.Context) error {
	const op = opLineWorks + ".ProcessDue"
	if _, err := uc.repo.RecoverStale(ctx, time.Now().Add(-2*time.Minute), valueobject.LineWorksMaxSendAttempts); err != nil {
		logutils.Error(ctx, logutils.LayerUsecase, op, "送信中の復旧失敗", slog.String("err", err.Error()))
		return err
	}
	jobs, err := uc.repo.ClaimDue(ctx, time.Now(), 20)
	if err != nil {
		logutils.Error(ctx, logutils.LayerUsecase, op, "通知予定の取得失敗", slog.String("err", err.Error()))
		return err
	}
	for i := range jobs {
		uc.dispatch(ctx, &jobs[i])
	}
	return nil
}

func (uc *Usecase) dispatch(ctx context.Context, n *lineworks.Notification) {
	const op = opLineWorks + ".dispatch"
	var (
		status int
		err    error
	)
	switch n.Destination {
	case valueobject.LineWorksDestinationUser:
		if strings.TrimSpace(n.RecipientLoginID) == "" {
			n.Status = valueobject.LineWorksJobFailed
			n.ErrorKind = valueobject.LineWorksErrorRecipientUnresolved
			n.LastError = safeMessage(valueobject.LineWorksErrorRecipientUnresolved)
			attempt := &lineworks.NotificationAttempt{
				NotificationUUID: n.UUID,
				Outcome:          valueobject.LineWorksAttemptPermanent,
				AttemptedAt:      time.Now(),
			}
			if saveErr := uc.repo.SaveResult(ctx, n, attempt); saveErr != nil {
				logutils.Error(ctx, logutils.LayerUsecase, op, "送信結果の保存失敗", slog.String("question_uuid", n.QuestionUUID.String()), slog.String("err", saveErr.Error()))
			}
			return
		}
		status, err = uc.client.SendUserMessage(ctx, n.RecipientLoginID, n.Body)
	default:
		status, err = uc.client.SendChannelMessage(ctx, n.Body)
	}
	attempt := applyResult(n, status, err)
	if saveErr := uc.repo.SaveResult(ctx, n, attempt); saveErr != nil {
		logutils.Error(ctx, logutils.LayerUsecase, op, "送信結果の保存失敗", slog.String("question_uuid", n.QuestionUUID.String()), slog.String("err", saveErr.Error()))
	}
}

func applyResult(n *lineworks.Notification, httpStatus int, err error) *lineworks.NotificationAttempt {
	now := time.Now()
	attempt := &lineworks.NotificationAttempt{
		NotificationUUID: n.UUID,
		HTTPStatus:       httpStatus,
		AttemptedAt:      now,
	}
	if err == nil {
		n.Status = valueobject.LineWorksJobSent
		n.SentAt = &now
		n.ErrorKind = valueobject.LineWorksErrorNone
		n.LastError = ""
		n.AttemptCount++
		attempt.Outcome = valueobject.LineWorksAttemptSuccess
		if attempt.HTTPStatus == 0 {
			attempt.HTTPStatus = 200
		}
		return attempt
	}

	sendErr := &ext.LineWorksSendError{Kind: valueobject.LineWorksErrorTransient, HTTPStatus: httpStatus}
	var typed *ext.LineWorksSendError
	if errors.As(err, &typed) {
		sendErr = typed
	}
	if sendErr.HTTPStatus != 0 {
		attempt.HTTPStatus = sendErr.HTTPStatus
	}
	n.AttemptCount++
	n.ErrorKind = sendErr.Kind
	n.LastError = safeMessage(sendErr.Kind)

	switch sendErr.Kind {
	case valueobject.LineWorksErrorRateLimited:
		attempt.Outcome = valueobject.LineWorksAttemptRateLimited
		wait := sendErr.RetryAfter
		if wait <= 0 {
			wait = time.Second
		}
		if n.AttemptCount >= valueobject.LineWorksMaxSendAttempts {
			n.Status = valueobject.LineWorksJobFailed
		} else {
			n.Status = valueobject.LineWorksJobPending
			n.NextAttemptAt = now.Add(wait)
		}
	case valueobject.LineWorksErrorUnknownOutcome:
		attempt.Outcome = valueobject.LineWorksAttemptUnknown
		if n.AttemptCount >= valueobject.LineWorksMaxSendAttempts {
			n.Status = valueobject.LineWorksJobFailed
		} else {
			n.Status = valueobject.LineWorksJobPending
			n.NextAttemptAt = now.Add(backoff(n.AttemptCount))
		}
	case valueobject.LineWorksErrorBotUnavailable, valueobject.LineWorksErrorInvalidDestination, valueobject.LineWorksErrorRecipientUnresolved:
		attempt.Outcome = valueobject.LineWorksAttemptPermanent
		n.Status = valueobject.LineWorksJobFailed
	default:
		attempt.Outcome = valueobject.LineWorksAttemptTransient
		if n.AttemptCount >= valueobject.LineWorksMaxSendAttempts {
			n.Status = valueobject.LineWorksJobFailed
		} else {
			n.Status = valueobject.LineWorksJobPending
			n.NextAttemptAt = now.Add(backoff(n.AttemptCount))
		}
	}
	return attempt
}

func backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 8 {
		attempt = 8
	}
	return time.Second << (attempt - 1)
}

func safeMessage(kind valueobject.LineWorksErrorKind) string {
	switch kind {
	case valueobject.LineWorksErrorRateLimited:
		return "LINE WORKSのレート制限です"
	case valueobject.LineWorksErrorTransient:
		return "LINE WORKSへの送信が一時的に失敗しました"
	case valueobject.LineWorksErrorUnknownOutcome:
		return "LINE WORKSへの送信成否が不明です"
	case valueobject.LineWorksErrorBotUnavailable:
		return "Botがトークルームにいないか、利用できません"
	case valueobject.LineWorksErrorInvalidDestination:
		return "送信先IDが不正です"
	case valueobject.LineWorksErrorRecipientUnresolved:
		return "質問者のLINE WORKSログインIDが未設定です"
	default:
		return "LINE WORKSへの送信に失敗しました"
	}
}
