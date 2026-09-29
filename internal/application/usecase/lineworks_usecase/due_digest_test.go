package lineworks_usecase

import (
	"context"
	"strings"
	"testing"
	"time"

	"solvi/internal/domain/entity"
	"solvi/internal/domain/entity/lineworks"
	"solvi/internal/domain/valueobject"
	extmock "solvi/internal/domain/interface/external/mock"
	repomock "solvi/internal/domain/interface/repository/mock"
	"solvi/internal/shared/config"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func jst(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func TestEnqueueDueDigestIfScheduledSkipsWrongMinute(t *testing.T) {
	ctrl := gomock.NewController(t)
	noticeRepo := repomock.NewMockLineWorksNotificationRepository(ctrl)
	questionRepo := repomock.NewMockQuestionRepository(ctrl)
	client := extmock.NewMockLineWorksClient(ctrl)

	loc := jst(t)
	now := time.Date(2026, 9, 30, 9, 1, 0, 0, loc)
	uc := New(noticeRepo, questionRepo, client, DueDigestSchedule{
		NoticeTimes: []config.NoticeTime{{Hour: 9, Minute: 0}},
		ChannelIDs:  []string{"room-a"},
		AppBaseURL:  "https://solvi.example",
	})
	uc.jst = loc

	if err := uc.EnqueueDueDigestIfScheduled(context.Background(), now); err != nil {
		t.Fatal(err)
	}
}

func TestEnqueueDueDigestIfScheduledEnqueuesAtMatchedMinute(t *testing.T) {
	ctrl := gomock.NewController(t)
	noticeRepo := repomock.NewMockLineWorksNotificationRepository(ctrl)
	questionRepo := repomock.NewMockQuestionRepository(ctrl)
	client := extmock.NewMockLineWorksClient(ctrl)

	loc := jst(t)
	now := time.Date(2026, 9, 30, 9, 0, 30, 0, loc)
	today := time.Date(2026, 9, 30, 0, 0, 0, 0, loc)
	tomorrow := time.Date(2026, 10, 1, 0, 0, 0, 0, loc)
	dueToday := time.Date(2026, 9, 30, 15, 0, 0, 0, loc)
	dueTomorrow := time.Date(2026, 10, 1, 10, 0, 0, 0, loc)

	questionRepo.EXPECT().ListIncompleteDueOnDates(gomock.Any(), today, tomorrow).Return([]entity.Question{
		{UUID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Title: "当日", AnswerDue: &dueToday, QuestionUser: entity.User{Name: "田中"}},
		{UUID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Title: "明日", AnswerDue: &dueTomorrow, QuestionUser: entity.User{Name: "佐藤"}},
	}, nil)
	noticeRepo.EXPECT().EnqueueDueDigests(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, notices []lineworks.Notification) error {
			if len(notices) != 2 {
				t.Fatalf("notices=%d", len(notices))
			}
			for _, n := range notices {
				if n.Event != valueobject.LineWorksEventDueDigest || n.StatusRevision != 20260930 || n.Burst != 540 {
					t.Fatalf("notice=%+v", n)
				}
				if n.ChannelID != "room-a" && n.ChannelID != "room-b" {
					t.Fatalf("channel=%s", n.ChannelID)
				}
				if !strings.Contains(n.Body, "期日当日の問い合わせが、1件あります。") || !strings.Contains(n.Body, "期日が明日に迫った問い合わせが、1件あります。") {
					t.Fatalf("body=%s", n.Body)
				}
			}
			return nil
		},
	)

	uc := New(noticeRepo, questionRepo, client, DueDigestSchedule{
		NoticeTimes: []config.NoticeTime{{Hour: 9, Minute: 0}},
		ChannelIDs:  []string{"room-a", "room-b"},
		AppBaseURL:  "https://solvi.example",
	})
	uc.jst = loc
	if err := uc.EnqueueDueDigestIfScheduled(context.Background(), now); err != nil {
		t.Fatal(err)
	}
}

func TestEnqueueDueDigestIfScheduledEnqueuesEmptyMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	noticeRepo := repomock.NewMockLineWorksNotificationRepository(ctrl)
	questionRepo := repomock.NewMockQuestionRepository(ctrl)
	client := extmock.NewMockLineWorksClient(ctrl)

	loc := jst(t)
	now := time.Date(2026, 9, 30, 18, 30, 0, 0, loc)
	today := time.Date(2026, 9, 30, 0, 0, 0, 0, loc)
	tomorrow := time.Date(2026, 10, 1, 0, 0, 0, 0, loc)

	questionRepo.EXPECT().ListIncompleteDueOnDates(gomock.Any(), today, tomorrow).Return(nil, nil)
	noticeRepo.EXPECT().EnqueueDueDigests(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, notices []lineworks.Notification) error {
			if len(notices) != 1 {
				t.Fatalf("notices=%d", len(notices))
			}
			if notices[0].Body != "【定期通知】\n期日に近づいている問い合わせはありませんでした。" {
				t.Fatalf("body=%s", notices[0].Body)
			}
			return nil
		},
	)

	uc := New(noticeRepo, questionRepo, client, DueDigestSchedule{
		NoticeTimes: []config.NoticeTime{{Hour: 18, Minute: 30}},
		ChannelIDs:  []string{"room-a"},
		AppBaseURL:  "https://solvi.example",
	})
	uc.jst = loc
	if err := uc.EnqueueDueDigestIfScheduled(context.Background(), now); err != nil {
		t.Fatal(err)
	}
}
