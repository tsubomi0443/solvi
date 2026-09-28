package repository

import (
	"context"

	"solvi/internal/domain/entity"
)

//go:generate go tool mockgen -typed -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

type QuestionRepository interface {
	Create(ctx context.Context, question *entity.Question) error
	GetByUUID(ctx context.Context, uuid string) (*entity.Question, error)
	ListByQuestionUserID(ctx context.Context, userID uint) ([]entity.Question, error)
	ListAll(ctx context.Context) ([]entity.Question, error)
	Update(ctx context.Context, question *entity.Question) error
	AddContent(ctx context.Context, content *entity.QuestionContent) error
	AddAnswer(ctx context.Context, answer *entity.QuestionAnswer) error
	AddMemo(ctx context.Context, memo *entity.QuestionMemo) error
	SoftDeleteAnswerByUUID(ctx context.Context, uuid string) error
	SoftDeleteMemoByUUID(ctx context.Context, uuid string) error
	SoftDeleteReferByUUID(ctx context.Context, uuid string) error
	SoftDeleteByUUID(ctx context.Context, uuid string) error
	AddRefer(ctx context.Context, refer *entity.QuestionRefer) error
	ReplaceTags(ctx context.Context, questionID uint, tags []entity.QuestionTag) error
	CreateSummary(ctx context.Context, summary *entity.QuestionSummary, refs []entity.QuestionSummaryReference) error
	UpsertSummary(ctx context.Context, questionID uint, title, content, answer string, refs []entity.QuestionSummaryReference) error
	ListSummaries(ctx context.Context) ([]entity.QuestionSummary, error)
	ListTagsByQuestionIDs(ctx context.Context, questionIDs []uint) (map[uint][]string, error)
	SoftDeleteSummaryByUUID(ctx context.Context, uuid string) error
}
