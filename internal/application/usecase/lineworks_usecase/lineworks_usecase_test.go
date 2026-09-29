package lineworks_usecase

import (
	"context"
	"strings"
	"testing"
	"time"

	"solvi/internal/domain/entity/lineworks"
	ext "solvi/internal/domain/interface/external"
	extmock "solvi/internal/domain/interface/external/mock"
	repomock "solvi/internal/domain/interface/repository/mock"
	"solvi/internal/domain/valueobject"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestApplyResultSuccessAndFailures(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		n := &lineworks.Notification{}
		attempt := applyResult(n, 201, nil)
		if n.Status != valueobject.LineWorksJobSent || n.SentAt == nil || attempt.Outcome != valueobject.LineWorksAttemptSuccess {
			t.Fatalf("notice=%+v attempt=%+v", n, attempt)
		}
	})

	t.Run("rate limit waits for reset", func(t *testing.T) {
		n := &lineworks.Notification{}
		before := time.Now()
		attempt := applyResult(n, 429, &ext.LineWorksSendError{
			Kind:       valueobject.LineWorksErrorRateLimited,
			HTTPStatus: 429,
			RetryAfter: 30 * time.Second,
		})
		if n.Status != valueobject.LineWorksJobPending || attempt.Outcome != valueobject.LineWorksAttemptRateLimited {
			t.Fatalf("notice=%+v attempt=%+v", n, attempt)
		}
		if n.NextAttemptAt.Before(before.Add(29 * time.Second)) {
			t.Fatalf("next=%s", n.NextAttemptAt)
		}
	})

	t.Run("transient backs off until max", func(t *testing.T) {
		n := &lineworks.Notification{AttemptCount: valueobject.LineWorksMaxSendAttempts - 1}
		applyResult(n, 503, &ext.LineWorksSendError{Kind: valueobject.LineWorksErrorTransient, HTTPStatus: 503})
		if n.Status != valueobject.LineWorksJobFailed {
			t.Fatalf("status=%v", n.Status)
		}
	})

	t.Run("bot unavailable is permanent", func(t *testing.T) {
		n := &lineworks.Notification{}
		applyResult(n, 403, &ext.LineWorksSendError{Kind: valueobject.LineWorksErrorBotUnavailable, HTTPStatus: 403})
		if n.Status != valueobject.LineWorksJobFailed || n.ErrorKind != valueobject.LineWorksErrorBotUnavailable {
			t.Fatalf("notice=%+v", n)
		}
		if strings.Contains(n.LastError, "secret") {
			t.Fatal(n.LastError)
		}
	})
}

func TestProcessDueFailsPermanentWhenChannelIDMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	noticeRepo := repomock.NewMockLineWorksNotificationRepository(ctrl)
	client := extmock.NewMockLineWorksClient(ctrl)

	job := lineworks.Notification{
		UUID:         uuid.New(),
		QuestionUUID: uuid.New(),
		Event:        valueobject.LineWorksEventReceived,
		Destination:  valueobject.LineWorksDestinationChannel,
		Body:         "【新規受付】",
		Status:       valueobject.LineWorksJobSending,
	}
	noticeRepo.EXPECT().RecoverStale(gomock.Any(), gomock.Any(), valueobject.LineWorksMaxSendAttempts).Return(0, nil)
	noticeRepo.EXPECT().ClaimDue(gomock.Any(), gomock.Any(), 20).Return([]lineworks.Notification{job}, nil)
	noticeRepo.EXPECT().SaveResult(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, n *lineworks.Notification, attempt *lineworks.NotificationAttempt) error {
			if n.Status != valueobject.LineWorksJobFailed || n.ErrorKind != valueobject.LineWorksErrorInvalidDestination {
				t.Fatalf("notice=%+v", n)
			}
			if attempt.Outcome != valueobject.LineWorksAttemptPermanent {
				t.Fatalf("attempt=%+v", attempt)
			}
			return nil
		},
	)

	if err := New(noticeRepo, repomock.NewMockQuestionRepository(ctrl), client, DueDigestSchedule{}).ProcessDue(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestProcessDueSendsChannelMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	noticeRepo := repomock.NewMockLineWorksNotificationRepository(ctrl)
	client := extmock.NewMockLineWorksClient(ctrl)

	job := lineworks.Notification{
		UUID:         uuid.New(),
		QuestionUUID: uuid.New(),
		Event:        valueobject.LineWorksEventReceived,
		Destination:  valueobject.LineWorksDestinationChannel,
		ChannelID:    "room-a",
		Body:         "【新規受付】",
		Status:       valueobject.LineWorksJobSending,
	}
	noticeRepo.EXPECT().RecoverStale(gomock.Any(), gomock.Any(), valueobject.LineWorksMaxSendAttempts).Return(0, nil)
	noticeRepo.EXPECT().ClaimDue(gomock.Any(), gomock.Any(), 20).Return([]lineworks.Notification{job}, nil)
	client.EXPECT().SendChannelMessage(gomock.Any(), job.ChannelID, job.Body).Return(200, nil)
	noticeRepo.EXPECT().SaveResult(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, n *lineworks.Notification, attempt *lineworks.NotificationAttempt) error {
			if n.Status != valueobject.LineWorksJobSent || attempt.Outcome != valueobject.LineWorksAttemptSuccess {
				t.Fatalf("saved status=%v outcome=%v", n.Status, attempt.Outcome)
			}
			return nil
		},
	)

	if err := New(noticeRepo, repomock.NewMockQuestionRepository(ctrl), client, DueDigestSchedule{}).ProcessDue(context.Background()); err != nil {
		t.Fatal(err)
	}
}
