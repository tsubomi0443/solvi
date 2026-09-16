package repository

import (
	"solvi/internal/domain/entity"
)

//go:generate go tool mockgen -typed -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

type QuestionRepository interface {
	Create(question *entity.Question) error
	GetByUUID(uuid string) (*entity.Question, error)
	ListByQuestionUserID(userID uint) ([]entity.Question, error)
	ListAll() ([]entity.Question, error)
	Update(question *entity.Question) error
	AddContent(content *entity.QuestionContent) error
	AddAnswer(answer *entity.QuestionAnswer) error
	AddMemo(memo *entity.QuestionMemo) error
	SoftDeleteAnswerByUUID(uuid string) error
	SoftDeleteMemoByUUID(uuid string) error
	SoftDeleteReferByUUID(uuid string) error
	SoftDeleteByUUID(uuid string) error
	AddRefer(refer *entity.QuestionRefer) error
	ReplaceTags(questionID uint, tags []entity.QuestionTag) error
	CreateSummary(summary *entity.QuestionSummary, refs []entity.QuestionSummaryReference) error
}
