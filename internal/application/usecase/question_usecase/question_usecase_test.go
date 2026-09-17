package question_usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	outputmodel "solvi/internal/application/model/output_model"
	quc "solvi/internal/application/usecase/question_usecase"
	"solvi/internal/domain/entity"
	bedrockext "solvi/internal/domain/interface/external"
	repomock "solvi/internal/domain/interface/repository/mock"
	"solvi/internal/domain/valueobject"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

type stubBedrock struct{}

func (stubBedrock) AnswerQuestion(string, string) (*bedrockext.FAQAnswerResult, error) {
	return &bedrockext.FAQAnswerResult{Content: "AI回答"}, nil
}

func TestList_ScopedToQuestionUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qRepo.EXPECT().ListByQuestionUserID(gomock.Any(), uint(10)).Return([]entity.Question{
		{UUID: uuid.New(), Title: "mine", SupportStatus: valueobject.SupportStatusPending},
	}, nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	items, err := uc.List(context.Background(), 10, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Title != "mine" {
		t.Fatalf("unexpected list: %+v", items)
	}
}

func TestGet_DeniesOtherUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(&entity.Question{
		UUID: qid, QuestionUserID: 99, Title: "secret",
	}, nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	_, err := uc.Get(context.Background(), 10, false, false, qid.String())
	if err == nil {
		t.Fatal("expected permission error")
	}
}

func TestCreate_PersistsQuestion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	qRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, q *entity.Question) error {
		q.ID = 1
		q.UUID = qid
		return nil
	})
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(&entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		SupportStatus: valueobject.SupportStatusPending,
		Contents:      []entity.QuestionContent{{Content: "body"}},
	}, nil)
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(&entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
	}, nil).AnyTimes()
	qRepo.EXPECT().AddAnswer(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	uRepo.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound).AnyTimes()

	uc := quc.NewQuestionUsecase(qRepo, uRepo, stubBedrock{}, nil)
	tt := time.Date(2006, 1, 2, 3, 4, 5, 0, time.Local)
	out, err := uc.Create(context.Background(), 5, "title", "body", []string{"給与"}, &tt, false)
	if err != nil {
		t.Fatal(err)
	}
	if out.UUID != qid.String() || out.Title != "title" {
		t.Fatalf("unexpected detail: %+v", out)
	}
}

func TestUpdate_SupporterUpdatesTitle(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "old", QuestionUserID: 5,
		SupportStatus: valueobject.SupportStatusPending,
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)
	qRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, updated *entity.Question) error {
		if updated.Title != "new title" {
			t.Fatalf("title=%q", updated.Title)
		}
		return nil
	})

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	title := "new title"
	if err := uc.Update(context.Background(), 1, true, true, qid.String(), &title, nil, nil, nil, nil, false, nil); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_SupporterUpdatesAnswerDue(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		SupportStatus: valueobject.SupportStatusPending,
	}
	due := time.Date(2026, 9, 14, 23, 59, 59, 0, time.FixedZone("Asia/Tokyo", 9*60*60))
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)
	qRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, updated *entity.Question) error {
		if updated.AnswerDue == nil || !updated.AnswerDue.Equal(due) {
			t.Fatalf("due=%v", updated.AnswerDue)
		}
		return nil
	})

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.Update(context.Background(), 1, true, true, qid.String(), nil, nil, &due, nil, nil, false, nil); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_SupporterReopensDoneQuestion(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		SupportStatus: valueobject.SupportStatusDone,
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)
	qRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, updated *entity.Question) error {
		if updated.SupportStatus != valueobject.SupportStatusPending {
			t.Fatalf("status=%v", updated.SupportStatus)
		}
		return nil
	})

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	status := "pending"
	if err := uc.Update(context.Background(), 1, true, true, qid.String(), nil, &status, nil, nil, nil, false, nil); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_SupporterCompletesViaStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	refUUID := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		SupportStatus: valueobject.SupportStatusSupporting,
		Refers: []entity.QuestionRefer{
			{UUID: refUUID, Name: "Doc", URL: "https://example.com/doc"},
		},
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil).Times(2)
	qRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	qRepo.EXPECT().UpsertSummary(gomock.Any(), uint(1), "title", "q-summary", "a-summary", gomock.Any()).Return(nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	status := "done"
	summaryInput := &quc.QuestionSummaryInput{
		Content:    "q-summary",
		Answer:     "a-summary",
		ReferUUIDs: []string{refUUID.String()},
	}
	if err := uc.Update(context.Background(), 1, true, true, qid.String(), nil, &status, nil, nil, nil, false, summaryInput); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_DoneRequiresSummary(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		SupportStatus: valueobject.SupportStatusSupporting,
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	status := "done"
	if err := uc.Update(context.Background(), 1, true, true, qid.String(), nil, &status, nil, nil, nil, false, nil); err == nil {
		t.Fatal("expected error when summary is nil")
	}
}

func TestUpdate_ReDoneUpdatesExistingSummary(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		SupportStatus: valueobject.SupportStatusDone,
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil).Times(2)
	qRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	qRepo.EXPECT().UpsertSummary(gomock.Any(), uint(1), "title", "updated-q", "updated-a", gomock.Any()).Return(nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	status := "done"
	summaryInput := &quc.QuestionSummaryInput{
		Content: "updated-q",
		Answer:  "updated-a",
	}
	if err := uc.Update(context.Background(), 1, true, true, qid.String(), nil, &status, nil, nil, nil, false, summaryInput); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteAnswer_OwnerSupporter(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	answerUUID := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		Answers: []entity.QuestionAnswer{
			{UUID: answerUUID, AnswerUserID: 1, Content: "answer"},
		},
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)
	qRepo.EXPECT().SoftDeleteAnswerByUUID(gomock.Any(), answerUUID.String()).Return(nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.DeleteAnswer(context.Background(), 1, true, true, qid.String(), answerUUID.String()); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteAnswer_DeniesNonSupporter(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.DeleteAnswer(context.Background(), 1, false, false, uuid.NewString(), uuid.NewString()); err == nil {
		t.Fatal("expected permission error")
	}
}

func TestDeleteAnswer_DeniesOtherSupporter(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	answerUUID := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		Answers: []entity.QuestionAnswer{
			{UUID: answerUUID, AnswerUserID: 99, Content: "answer"},
		},
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.DeleteAnswer(context.Background(), 1, true, false, qid.String(), answerUUID.String()); err == nil {
		t.Fatal("expected permission error")
	}
}

func TestDeleteMemo_OwnerSupporter(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	memoUUID := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		Memos: []entity.QuestionMemo{
			{UUID: memoUUID, MemoUserID: 1, Content: "memo"},
		},
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)
	qRepo.EXPECT().SoftDeleteMemoByUUID(gomock.Any(), memoUUID.String()).Return(nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.DeleteMemo(context.Background(), 1, true, true, qid.String(), memoUUID.String()); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteMemo_DeniesOtherSupporter(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	memoUUID := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		Memos: []entity.QuestionMemo{
			{UUID: memoUUID, MemoUserID: 99, Content: "memo"},
		},
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.DeleteMemo(context.Background(), 1, true, false, qid.String(), memoUUID.String()); err == nil {
		t.Fatal("expected permission error")
	}
}

func TestDeleteRefer_OwnerSupporter(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	referUUID := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		Refers: []entity.QuestionRefer{
			{UUID: referUUID, UserID: 1, Name: "ref", URL: "https://example.com"},
		},
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)
	qRepo.EXPECT().SoftDeleteReferByUUID(gomock.Any(), referUUID.String()).Return(nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.DeleteRefer(context.Background(), 1, true, true, qid.String(), referUUID.String()); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteRefer_DeniesOtherSupporter(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	referUUID := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		Refers: []entity.QuestionRefer{
			{UUID: referUUID, UserID: 99, Name: "ref", URL: "https://example.com"},
		},
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.DeleteRefer(context.Background(), 1, true, false, qid.String(), referUUID.String()); err == nil {
		t.Fatal("expected permission error")
	}
}

func TestDeleteRefer_AdminDeletesOther(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	referUUID := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		Refers: []entity.QuestionRefer{
			{UUID: referUUID, UserID: 99, Name: "ref", URL: "https://example.com"},
		},
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)
	qRepo.EXPECT().SoftDeleteReferByUUID(gomock.Any(), referUUID.String()).Return(nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.DeleteRefer(context.Background(), 1, false, true, qid.String(), referUUID.String()); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_AskerUpdatesRequireHuman(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		IsRequireHumanSupport: false,
	}
	requireHuman := true
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)
	qRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, updated *entity.Question) error {
		if !updated.IsRequireHumanSupport {
			t.Fatal("expected require human support")
		}
		return nil
	})

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.Update(context.Background(), 5, false, false, qid.String(), nil, nil, nil, nil, &requireHuman, false, nil); err != nil {
		t.Fatal(err)
	}
}

func TestList_AdminSeesAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qRepo.EXPECT().ListAll(gomock.Any()).Return([]entity.Question{
		{UUID: uuid.New(), Title: "all", SupportStatus: valueobject.SupportStatusPending},
	}, nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	items, err := uc.List(context.Background(), 10, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Title != "all" {
		t.Fatalf("unexpected list: %+v", items)
	}
}

func TestGet_AdminCanViewOthers(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(&entity.Question{
		UUID: qid, QuestionUserID: 99, Title: "secret",
	}, nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	out, err := uc.Get(context.Background(), 10, false, true, qid.String())
	if err != nil {
		t.Fatal(err)
	}
	if out.Title != "secret" {
		t.Fatalf("unexpected detail: %+v", out)
	}
}

func TestDeleteAnswer_AdminDeletesOther(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	answerUUID := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		Answers: []entity.QuestionAnswer{
			{UUID: answerUUID, AnswerUserID: 99, Content: "answer"},
		},
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)
	qRepo.EXPECT().SoftDeleteAnswerByUUID(gomock.Any(), answerUUID.String()).Return(nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.DeleteAnswer(context.Background(), 1, false, true, qid.String(), answerUUID.String()); err != nil {
		t.Fatal(err)
	}
}

func TestDelete_AdminDeletesQuestion(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(&entity.Question{UUID: qid}, nil)
	qRepo.EXPECT().SoftDeleteByUUID(gomock.Any(), qid.String()).Return(nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.Delete(context.Background(), true, qid.String()); err != nil {
		t.Fatal(err)
	}
}

func TestDelete_DeniesNonAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.Delete(context.Background(), false, uuid.NewString()); err == nil {
		t.Fatal("expected permission error")
	}
}

func TestAddRefers_PersistsMultiple(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(&entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, QuestionUserID: 5,
	}, nil)
	qRepo.EXPECT().AddRefer(gomock.Any(), gomock.Any()).Return(nil).Times(2)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	err := uc.AddRefers(context.Background(), 10, qid.String(), []outputmodel.ReferOutput{
		{Name: "Doc A", URL: "https://example.com/a"},
		{Name: "Doc B", URL: "https://example.com/b"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddRefers_SkipsEmptyRows(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(&entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, QuestionUserID: 5,
	}, nil)
	qRepo.EXPECT().AddRefer(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	err := uc.AddRefers(context.Background(), 10, qid.String(), []outputmodel.ReferOutput{
		{Name: "Doc A", URL: "https://example.com/a"},
		{Name: " ", URL: " "},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddRefers_RejectsPartialRow(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(&entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, QuestionUserID: 5,
	}, nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	err := uc.AddRefers(context.Background(), 10, qid.String(), []outputmodel.ReferOutput{
		{Name: "Doc A", URL: ""},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestAddRefers_RejectsEmptyPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(&entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, QuestionUserID: 5,
	}, nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	err := uc.AddRefers(context.Background(), 10, qid.String(), []outputmodel.ReferOutput{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestUpdate_AdminCannotChangeRequireHumanForOthers(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		IsRequireHumanSupport: false,
	}
	requireHuman := true
	qRepo.EXPECT().GetByUUID(gomock.Any(), gomock.Any()).Return(q, nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.Update(context.Background(), 1, false, false, qid.String(), nil, nil, nil, nil, &requireHuman, false, nil); err == nil {
		t.Fatal("expected permission error")
	}
}

func TestGet_OwnerSuccessAndIncludesMemosForSupporter(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	q := &entity.Question{
		Model:          gorm.Model{ID: 1},
		UUID:           qid,
		Title:          "details",
		QuestionUserID: 5,
		Memos: []entity.QuestionMemo{
			{Content: "secret memo"},
		},
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(q, nil).Times(2)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	outOwner, err := uc.Get(context.Background(), 5, false, false, qid.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(outOwner.Memos) != 0 {
		t.Fatalf("regular owner must not see memos, got %+v", outOwner.Memos)
	}

	outSupporter, err := uc.Get(context.Background(), 99, true, false, qid.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(outSupporter.Memos) != 1 {
		t.Fatalf("supporter must see memos, got %+v", outSupporter.Memos)
	}
}

func TestAppendContent_OwnerSuccessAndDeniesOther(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	q := &entity.Question{
		Model:          gorm.Model{ID: 2},
		UUID:           qid,
		QuestionUserID: 10,
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(q, nil).Times(2)
	qRepo.EXPECT().AddContent(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c *entity.QuestionContent) error {
		if c.QuestionID != 2 || c.QuestionUserID != 10 || c.Content != "more info" {
			t.Fatalf("unexpected content: %+v", c)
		}
		return nil
	})

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.AppendContent(context.Background(), 10, qid.String(), "more info"); err != nil {
		t.Fatal(err)
	}
	if err := uc.AppendContent(context.Background(), 99, qid.String(), "hack"); err == nil {
		t.Fatal("expected permission error for other user")
	}
}

func TestAddAnswer_UpdatesStatusFromPending(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	q := &entity.Question{
		Model:         gorm.Model{ID: 3},
		UUID:          qid,
		SupportStatus: valueobject.SupportStatusPending,
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(q, nil)
	qRepo.EXPECT().AddAnswer(gomock.Any(), gomock.Any()).Return(nil)
	qRepo.EXPECT().AddRefer(gomock.Any(), gomock.Any()).Return(nil)
	qRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, updated *entity.Question) error {
		if updated.SupportStatus != valueobject.SupportStatusSupporting {
			t.Fatalf("expected supporting status, got %v", updated.SupportStatus)
		}
		return nil
	})

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.AddAnswer(context.Background(), 1, qid.String(), "ans", []outputmodel.ReferOutput{{Name: "ref", URL: "http://example.com"}}); err != nil {
		t.Fatal(err)
	}
}

func TestAddAnswer_RetainsExistingNonPendingStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	q := &entity.Question{
		Model:         gorm.Model{ID: 3},
		UUID:          qid,
		SupportStatus: valueobject.SupportStatusSupporting,
	}
	qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(q, nil)
	qRepo.EXPECT().AddAnswer(gomock.Any(), gomock.Any()).Return(nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.AddAnswer(context.Background(), 1, qid.String(), "ans", nil); err != nil {
		t.Fatal(err)
	}
}

func TestAddMemo_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	q := &entity.Question{Model: gorm.Model{ID: 7}, UUID: qid}
	qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(q, nil)
	qRepo.EXPECT().AddMemo(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, m *entity.QuestionMemo) error {
		if m.QuestionID != 7 || m.MemoUserID != 3 || m.Content != "internal note" {
			t.Fatalf("unexpected memo: %+v", m)
		}
		return nil
	})

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.AddMemo(context.Background(), 3, qid.String(), "internal note"); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_CompleteBranches(t *testing.T) {
	t.Run("supporter completes question and creates summary when nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		qRepo := repomock.NewMockQuestionRepository(ctrl)
		uRepo := repomock.NewMockUserRepository(ctrl)

		qid := uuid.New()
		q := &entity.Question{
			Model:                 gorm.Model{ID: 11},
			UUID:                  qid,
			Title:                 "Done Title",
			QuestionUserID:        5,
			IsRequireHumanSupport: true,
			Contents:              []entity.QuestionContent{{Content: "c1"}},
			Answers:               []entity.QuestionAnswer{{Content: "a1"}},
			Refers:                []entity.QuestionRefer{{Name: "r1", URL: "u1"}},
		}
		qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(q, nil).Times(2)
		qRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		qRepo.EXPECT().CreateSummary(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
		if err := uc.Update(context.Background(), 99, false, true, qid.String(), nil, nil, nil, nil, nil, true, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("owner completes question without human support", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		qRepo := repomock.NewMockQuestionRepository(ctrl)
		uRepo := repomock.NewMockUserRepository(ctrl)

		qid := uuid.New()
		q := &entity.Question{
			Model:                 gorm.Model{ID: 12},
			UUID:                  qid,
			QuestionUserID:        5,
			IsRequireHumanSupport: false,
		}
		qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(q, nil).Times(2)
		qRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		qRepo.EXPECT().CreateSummary(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
		if err := uc.Update(context.Background(), 5, false, false, qid.String(), nil, nil, nil, nil, nil, true, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("owner cannot complete when require human support", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		qRepo := repomock.NewMockQuestionRepository(ctrl)
		uRepo := repomock.NewMockUserRepository(ctrl)

		qid := uuid.New()
		q := &entity.Question{
			Model:                 gorm.Model{ID: 13},
			UUID:                  qid,
			QuestionUserID:        5,
			IsRequireHumanSupport: true,
		}
		qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(q, nil)

		uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
		if err := uc.Update(context.Background(), 5, false, false, qid.String(), nil, nil, nil, nil, nil, true, nil); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestUpdate_ValidationAndTags(t *testing.T) {
	t.Run("empty title rejected", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		qRepo := repomock.NewMockQuestionRepository(ctrl)
		uRepo := repomock.NewMockUserRepository(ctrl)

		qid := uuid.New()
		qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(&entity.Question{UUID: qid}, nil)

		uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
		emptyTitle := "   "
		if err := uc.Update(context.Background(), 1, true, false, qid.String(), &emptyTitle, nil, nil, nil, nil, false, nil); err == nil {
			t.Fatal("expected error for empty title")
		}
	})

	t.Run("status done requires summary and refers selection", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		qRepo := repomock.NewMockQuestionRepository(ctrl)
		uRepo := repomock.NewMockUserRepository(ctrl)

		qid := uuid.New()
		q := &entity.Question{
			UUID:   qid,
			Refers: []entity.QuestionRefer{{UUID: uuid.New(), Name: "r"}},
		}
		qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(q, nil).Times(2)

		uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
		doneStatus := "done"
		summaryMissingRef := &quc.QuestionSummaryInput{Content: "c", Answer: "a", ReferUUIDs: nil}
		if err := uc.Update(context.Background(), 1, true, false, qid.String(), nil, &doneStatus, nil, nil, nil, false, summaryMissingRef); err == nil {
			t.Fatal("expected error for missing refer selection")
		}

		summaryEmptyText := &quc.QuestionSummaryInput{Content: " ", Answer: "a"}
		if err := uc.Update(context.Background(), 1, true, false, qid.String(), nil, &doneStatus, nil, nil, nil, false, summaryEmptyText); err == nil {
			t.Fatal("expected error for empty summary body")
		}
	})

	t.Run("replaces tags and updates fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		qRepo := repomock.NewMockQuestionRepository(ctrl)
		uRepo := repomock.NewMockUserRepository(ctrl)

		qid := uuid.New()
		q := &entity.Question{Model: gorm.Model{ID: 20}, UUID: qid}
		qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(q, nil)
		qRepo.EXPECT().ReplaceTags(gomock.Any(), uint(20), gomock.Len(2)).Return(nil)
		qRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
		tags := []string{"tag1", "  tag2  ", " "}
		requireHuman := true
		if err := uc.Update(context.Background(), 1, true, false, qid.String(), nil, nil, nil, &tags, &requireHuman, false, nil); err != nil {
			t.Fatal(err)
		}
	})
}

func TestDeleteTargetsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(&entity.Question{UUID: qid}, nil).Times(3)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.DeleteAnswer(context.Background(), 1, true, true, qid.String(), "missing"); err == nil {
		t.Fatal("expected error when answer not found")
	}
	if err := uc.DeleteMemo(context.Background(), 1, true, true, qid.String(), "missing"); err == nil {
		t.Fatal("expected error when memo not found")
	}
	if err := uc.DeleteRefer(context.Background(), 1, true, true, qid.String(), "missing"); err == nil {
		t.Fatal("expected error when refer not found")
	}
}

type mockBedrockClient struct {
	answerFn func(title, content string) (*bedrockext.FAQAnswerResult, error)
}

func (m mockBedrockClient) AnswerQuestion(title, content string) (*bedrockext.FAQAnswerResult, error) {
	return m.answerFn(title, content)
}

func TestRunAI_SuccessAndFailure(t *testing.T) {
	t.Run("success adds answer and refer and invokes callback", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		qRepo := repomock.NewMockQuestionRepository(ctrl)
		uRepo := repomock.NewMockUserRepository(ctrl)

		qid := uuid.New()
		bedrock := mockBedrockClient{
			answerFn: func(title, content string) (*bedrockext.FAQAnswerResult, error) {
				return &bedrockext.FAQAnswerResult{
					Content: "AI Ans",
					References: []bedrockext.FAQReference{
						{Name: "Ref1", URL: "http://ref.local"},
					},
				}, nil
			},
		}

		uRepo.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(&entity.User{Model: gorm.Model{ID: 999}}, nil)
		qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(&entity.Question{Model: gorm.Model{ID: 100}, UUID: qid}, nil)
		qRepo.EXPECT().AddAnswer(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, a *entity.QuestionAnswer) error {
			if a.Content != "AI Ans" || a.AnswerUserID != 999 || a.QuestionID != 100 {
				t.Fatalf("unexpected answer: %+v", a)
			}
			return nil
		})
		qRepo.EXPECT().AddRefer(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, r *entity.QuestionRefer) error {
			if r.Name != "Ref1" || r.UserID != 999 || r.QuestionID != 100 {
				t.Fatalf("unexpected refer: %+v", r)
			}
			return nil
		})

		callbackInvoked := false
		uc := quc.NewQuestionUsecase(qRepo, uRepo, bedrock, func(uuid string) {
			if uuid == qid.String() {
				callbackInvoked = true
			}
		})

		quc.RunAIForTest(uc, qid.String(), "title", "content")
		if !callbackInvoked {
			t.Fatal("expected callback invocation")
		}
	})

	t.Run("bedrock failure falls back to require human support", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		qRepo := repomock.NewMockQuestionRepository(ctrl)
		uRepo := repomock.NewMockUserRepository(ctrl)

		qid := uuid.New()
		bedrock := mockBedrockClient{
			answerFn: func(title, content string) (*bedrockext.FAQAnswerResult, error) {
				return nil, errors.New("bedrock down")
			},
		}

		qRepo.EXPECT().GetByUUID(gomock.Any(), qid.String()).Return(&entity.Question{
			UUID:                  qid,
			IsRequireHumanSupport: false,
		}, nil)
		qRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, q *entity.Question) error {
			if !q.IsRequireHumanSupport {
				t.Fatal("expected IsRequireHumanSupport to become true")
			}
			return nil
		})

		uc := quc.NewQuestionUsecase(qRepo, uRepo, bedrock, nil)
		quc.RunAIForTest(uc, qid.String(), "title", "content")
	})
}
