package question_usecase_test

import (
	"testing"
	"time"

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

	qRepo.EXPECT().ListByQuestionUserID(uint(10)).Return([]entity.Question{
		{UUID: uuid.New(), Title: "mine", SupportStatus: valueobject.SupportStatusPending},
	}, nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	items, err := uc.List(10, false)
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
	qRepo.EXPECT().GetByUUID(qid.String()).Return(&entity.Question{
		UUID: qid, QuestionUserID: 99, Title: "secret",
	}, nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	_, err := uc.Get(10, false, qid.String())
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
	qRepo.EXPECT().Create(gomock.Any()).DoAndReturn(func(q *entity.Question) error {
		q.ID = 1
		q.UUID = qid
		return nil
	})
	qRepo.EXPECT().GetByUUID(qid.String()).Return(&entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		SupportStatus: valueobject.SupportStatusPending,
		Contents:      []entity.QuestionContent{{Content: "body"}},
	}, nil)
	qRepo.EXPECT().GetByUUID(qid.String()).Return(&entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
	}, nil).AnyTimes()
	qRepo.EXPECT().AddAnswer(gomock.Any()).Return(nil).AnyTimes()
	uRepo.EXPECT().GetByEmail(gomock.Any()).Return(nil, gorm.ErrRecordNotFound).AnyTimes()

	uc := quc.NewQuestionUsecase(qRepo, uRepo, stubBedrock{}, nil)
	tt := time.Date(2006, 1, 2, 3, 4, 5, 0, time.Local)
	out, err := uc.Create(5, "title", "body", []string{"給与"}, &tt, false)
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
	qRepo.EXPECT().GetByUUID(qid.String()).Return(q, nil)
	qRepo.EXPECT().Update(gomock.Any()).DoAndReturn(func(updated *entity.Question) error {
		if updated.Title != "new title" {
			t.Fatalf("title=%q", updated.Title)
		}
		return nil
	})

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	title := "new title"
	if err := uc.Update(1, true, qid.String(), &title, nil, nil, nil, nil, false); err != nil {
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
	qRepo.EXPECT().GetByUUID(qid.String()).Return(q, nil)
	qRepo.EXPECT().Update(gomock.Any()).DoAndReturn(func(updated *entity.Question) error {
		if updated.AnswerDue == nil || !updated.AnswerDue.Equal(due) {
			t.Fatalf("due=%v", updated.AnswerDue)
		}
		return nil
	})

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	if err := uc.Update(1, true, qid.String(), nil, nil, &due, nil, nil, false); err != nil {
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
	qRepo.EXPECT().GetByUUID(qid.String()).Return(q, nil)
	qRepo.EXPECT().Update(gomock.Any()).DoAndReturn(func(updated *entity.Question) error {
		if updated.SupportStatus != valueobject.SupportStatusPending {
			t.Fatalf("status=%v", updated.SupportStatus)
		}
		return nil
	})

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	status := "pending"
	if err := uc.Update(1, true, qid.String(), nil, &status, nil, nil, nil, false); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_SupporterCompletesViaStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	qRepo := repomock.NewMockQuestionRepository(ctrl)
	uRepo := repomock.NewMockUserRepository(ctrl)

	qid := uuid.New()
	q := &entity.Question{
		Model: gorm.Model{ID: 1}, UUID: qid, Title: "title", QuestionUserID: 5,
		SupportStatus: valueobject.SupportStatusSupporting,
	}
	qRepo.EXPECT().GetByUUID(qid.String()).Return(q, nil).Times(2)
	qRepo.EXPECT().Update(gomock.Any()).Return(nil)
	qRepo.EXPECT().CreateSummary(gomock.Any(), gomock.Any()).Return(nil)

	uc := quc.NewQuestionUsecase(qRepo, uRepo, nil, nil)
	status := "done"
	if err := uc.Update(1, true, qid.String(), nil, &status, nil, nil, nil, false); err != nil {
		t.Fatal(err)
	}
}
