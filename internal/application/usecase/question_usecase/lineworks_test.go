package question_usecase_test

import (
	"context"
	"strings"
	"testing"
	"time"

	quc "solvi/internal/application/usecase/question_usecase"
	"solvi/internal/domain/entity"
	"solvi/internal/domain/entity/lineworks"
	repomock "solvi/internal/domain/interface/repository/mock"
	"solvi/internal/domain/valueobject"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func enableLineWorks(uc *quc.QuestionUsecase, channelIDs ...string) {
	uc.ConfigureLineWorks(quc.LineWorksOptions{
		Enabled:    true,
		AppBaseURL: "https://solvi.example",
		ChannelIDs: channelIDs,
		Debounce:   time.Minute,
	})
}

func TestCreate_EnqueuesReceivedNotice(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)
	qid := uuid.New()

	uRepo.EXPECT().GetByID(gomock.Any(), uint(5)).Return(&entity.User{Name: "田中"}, nil)
	qRepo.EXPECT().CreateWithNotification(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, q *entity.Question, notices []*lineworks.Notification) error {
			q.ID = 9
			if len(notices) != 1 {
				t.Fatalf("notices=%+v", notices)
			}
			n := notices[0]
			if n.Event != valueobject.LineWorksEventReceived || n.Destination != valueobject.LineWorksDestinationChannel {
				t.Fatalf("notice=%+v", n)
			}
			if n.ChannelID != "room-a" {
				t.Fatalf("channel=%s", n.ChannelID)
			}
			if !strings.Contains(n.Body, "【新規受付 #"+q.UUID.String()+"】") {
				t.Fatalf("body=%s", n.Body)
			}
			if !strings.HasSuffix(n.Body, "https://solvi.example/questions/"+q.UUID.String()) {
				t.Fatalf("body=%s", n.Body)
			}
			if strings.Contains(n.Body, "#9") {
				t.Fatalf("numeric id in body: %s", n.Body)
			}
			return nil
		},
	)
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(&entity.Question{
		Model: gorm.Model{ID: 9}, UUID: qid, Title: "title", QuestionUserID: 5,
		SupportStatus: valueobject.SupportStatusPending,
		Contents:      []entity.QuestionContent{{Content: "body"}},
	}, nil).AnyTimes()
	uRepo.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound).AnyTimes()

	uc := quc.NewQuestionUsecase(qRepo, uRepo, stubBedrock{}, nil)
	enableLineWorks(uc, "room-a")
	if _, err := uc.Create(context.Background(), 5, "title", "body", nil, nil, false); err != nil {
		t.Fatal(err)
	}
}

func TestCreate_FansOutToMultipleChannels(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	uRepo.EXPECT().GetByID(gomock.Any(), uint(5)).Return(&entity.User{Name: "田中"}, nil)
	qRepo.EXPECT().CreateWithNotification(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, _ *entity.Question, notices []*lineworks.Notification) error {
			if len(notices) != 2 {
				t.Fatalf("notices=%+v", notices)
			}
			if notices[0].ChannelID != "room-a" || notices[1].ChannelID != "room-b" {
				t.Fatalf("channels=%s %s", notices[0].ChannelID, notices[1].ChannelID)
			}
			if notices[0].Body != notices[1].Body {
				t.Fatal("body should match across channels")
			}
			return nil
		},
	)
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(&entity.Question{
		UUID: uuid.New(), Title: "title", QuestionUserID: 5,
		SupportStatus: valueobject.SupportStatusPending,
		Contents:      []entity.QuestionContent{{Content: "body"}},
	}, nil).AnyTimes()
	uRepo.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound).AnyTimes()

	uc := quc.NewQuestionUsecase(qRepo, uRepo, stubBedrock{}, nil)
	enableLineWorks(uc, "room-a", "room-b")
	if _, err := uc.Create(context.Background(), 5, "title", "body", nil, nil, false); err != nil {
		t.Fatal(err)
	}
}

func TestCreate_WithoutChannelIDsDoesNotEnqueueChannelNotice(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	uRepo.EXPECT().GetByID(gomock.Any(), uint(5)).Return(&entity.User{Name: "田中"}, nil)
	qRepo.EXPECT().CreateWithNotification(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, _ *entity.Question, notices []*lineworks.Notification) error {
			if len(notices) != 0 {
				t.Fatalf("notices=%+v", notices)
			}
			return nil
		},
	)
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(&entity.Question{
		UUID: uuid.New(), Title: "title", QuestionUserID: 5,
		SupportStatus: valueobject.SupportStatusPending,
		Contents:      []entity.QuestionContent{{Content: "body"}},
	}, nil).AnyTimes()
	uRepo.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound).AnyTimes()

	uc := quc.NewQuestionUsecase(qRepo, uRepo, stubBedrock{}, nil)
	enableLineWorks(uc)
	if _, err := uc.Create(context.Background(), 5, "title", "body", nil, nil, false); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_AnsweredNoticeGoesToQuestioner(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)
	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(&entity.Question{
		UUID: qid, Title: "title", QuestionUserID: 5,
		SupportStatus: valueobject.SupportStatusSupporting,
		QuestionUser:  entity.User{LineWorksLoginID: "tanaka"},
	}, nil)
	qRepo.EXPECT().CompleteWithNotification(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, q *entity.Question, _, _, _ string, _ []entity.QuestionSummaryReference, n *lineworks.Notification) error {
			if q.SupportStatus != valueobject.SupportStatusDone {
				t.Fatalf("status=%v", q.SupportStatus)
			}
			if n.Destination != valueobject.LineWorksDestinationUser || n.RecipientLoginID != "tanaka" {
				t.Fatalf("notice=%+v", n)
			}
			if !strings.Contains(n.Body, "【回答完了 #"+qid.String()+"】") || !strings.Contains(n.Body, "/questions/"+qid.String()) {
				t.Fatalf("body=%s", n.Body)
			}
			return nil
		},
	)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	enableLineWorks(uc)
	status := "done"
	summary := &quc.QuestionSummaryInput{Content: "質問要約", Answer: "対応要約"}
	if err := uc.Update(context.Background(), 1, true, true, qid.String(), nil, &status, nil, nil, nil, false, summary); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_AnsweredWithoutLoginIDIsRecordedFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)
	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(&entity.Question{
		UUID: qid, Title: "title", QuestionUserID: 5,
		SupportStatus:         valueobject.SupportStatusSupporting,
		IsRequireHumanSupport: false,
		QuestionUser:          entity.User{Email: "tanaka@example.com"},
	}, nil)
	qRepo.EXPECT().CompleteWithNotification(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, _ *entity.Question, _, _, _ string, _ []entity.QuestionSummaryReference, n *lineworks.Notification) error {
			if n.Status != valueobject.LineWorksJobFailed || n.ErrorKind != valueobject.LineWorksErrorRecipientUnresolved {
				t.Fatalf("notice=%+v", n)
			}
			if n.RecipientLoginID != "" || strings.Contains(n.Body, "tanaka@example.com") {
				t.Fatalf("email must not be the recipient: %+v", n)
			}
			return nil
		},
	)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	enableLineWorks(uc)
	if err := uc.Update(context.Background(), 5, false, false, qid.String(), nil, nil, nil, nil, nil, true, nil); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_QuestionerReopenNotifiesChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)
	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(&entity.Question{
		UUID: qid, QuestionUserID: 5, SupportStatus: valueobject.SupportStatusDone,
		Contents: []entity.QuestionContent{{Content: "まだ不明です"}},
		Answers: []entity.QuestionAnswer{{
			Content:    "前回回答",
			AnswerUser: entity.User{Name: "佐藤", Email: "sato@example.com"},
		}},
	}, nil)
	qRepo.EXPECT().UpdateWithNotification(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, q *entity.Question, notices []*lineworks.Notification) error {
			if q.SupportStatus != valueobject.SupportStatusSupporting {
				t.Fatalf("status=%v", q.SupportStatus)
			}
			if len(notices) != 1 {
				t.Fatalf("notices=%+v", notices)
			}
			n := notices[0]
			if n.Event != valueobject.LineWorksEventReopened || n.Destination != valueobject.LineWorksDestinationChannel {
				t.Fatalf("notice=%+v", n)
			}
			if !strings.Contains(n.Body, "【再対応依頼 #"+qid.String()+"】") || !strings.Contains(n.Body, "佐藤") || !strings.Contains(n.Body, "まだ不明です") {
				t.Fatalf("body=%s", n.Body)
			}
			if !strings.HasSuffix(n.Body, "https://solvi.example/questions/"+qid.String()) {
				t.Fatalf("body=%s", n.Body)
			}
			return nil
		},
	)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	enableLineWorks(uc, "room-a")
	status := "supporting"
	if err := uc.Update(context.Background(), 5, false, false, qid.String(), nil, &status, nil, nil, nil, false, nil); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_SupporterReopenDoesNotNotify(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)
	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(&entity.Question{
		UUID: qid, QuestionUserID: 5, SupportStatus: valueobject.SupportStatusDone,
	}, nil)
	qRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	enableLineWorks(uc, "room-a")
	status := "pending"
	if err := uc.Update(context.Background(), 1, true, true, qid.String(), nil, &status, nil, nil, nil, false, nil); err != nil {
		t.Fatal(err)
	}
}

func TestAppendContent_AggregatesWhileSupporting(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)
	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(&entity.Question{
		UUID: qid, QuestionUserID: 5, SupportStatus: valueobject.SupportStatusSupporting,
	}, nil)
	qRepo.EXPECT().AddContentWithReopenAggregate(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c *entity.QuestionContent, follow *lineworks.ReopenFollowUp) error {
			if c.Content != "続けて確認です" || follow.Comment != "続けて確認です" {
				t.Fatalf("content=%+v follow=%+v", c, follow)
			}
			if follow.Debounce != time.Minute || follow.Template.Event != valueobject.LineWorksEventReopened {
				t.Fatalf("follow=%+v", follow)
			}
			if !reflectDeepEqual(follow.ChannelIDs, []string{"room-a", "room-b"}) {
				t.Fatalf("channel ids=%v", follow.ChannelIDs)
			}
			if !strings.Contains(follow.Template.Body, "/questions/"+qid.String()) {
				t.Fatalf("body=%s", follow.Template.Body)
			}
			return nil
		},
	)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	enableLineWorks(uc, "room-a", "room-b")
	if err := uc.AppendContent(context.Background(), 5, qid.String(), "続けて確認です"); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_SecondDoneDoesNotEnqueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)
	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(&entity.Question{
		UUID: qid, Title: "title", QuestionUserID: 1,
		SupportStatus: valueobject.SupportStatusDone,
		Summary:       &entity.QuestionSummary{},
	}, nil).AnyTimes()
	qRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	qRepo.EXPECT().UpsertSummary(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	enableLineWorks(uc, "room-a")
	status := "done"
	summary := &quc.QuestionSummaryInput{Content: "質問要約", Answer: "対応要約"}
	if err := uc.Update(context.Background(), 1, true, true, qid.String(), nil, &status, nil, nil, nil, false, summary); err != nil {
		t.Fatal(err)
	}
}

func reflectDeepEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
