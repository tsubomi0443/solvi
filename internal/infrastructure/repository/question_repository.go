package repository

import (
	"fmt"
	"solvi/internal/domain/entity"

	"gorm.io/gorm"
)

type QuestionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository(db *gorm.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

func (r *QuestionRepository) preload(q *gorm.DB) *gorm.DB {
	return q.Preload("QuestionUser").
		Preload("Contents", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		Preload("Contents.QuestionUser").
		Preload("Tags").
		Preload("Answers", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		Preload("Answers.AnswerUser").
		Preload("Memos", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		Preload("Memos.MemoUser").
		Preload("Refers").
		Preload("Refers.User").
		Preload("Summary.References")
}

func (r *QuestionRepository) Create(question *entity.Question) error {
	return r.db.Create(question).Error
}

func (r *QuestionRepository) GetByUUID(uuid string) (*entity.Question, error) {
	var q entity.Question
	if err := r.preload(r.db).Where("uuid = ?", uuid).First(&q).Error; err != nil {
		return nil, fmt.Errorf("質問が見つかりません: %w", err)
	}
	return &q, nil
}

func (r *QuestionRepository) ListByQuestionUserID(userID uint) ([]entity.Question, error) {
	var qs []entity.Question
	err := r.preload(r.db).Where("question_user_id = ?", userID).Order("id DESC").Find(&qs).Error
	return qs, err
}

func (r *QuestionRepository) ListAll() ([]entity.Question, error) {
	var qs []entity.Question
	err := r.preload(r.db).Order("id DESC").Find(&qs).Error
	return qs, err
}

func (r *QuestionRepository) Update(question *entity.Question) error {
	return r.db.Save(question).Error
}

func (r *QuestionRepository) AddContent(content *entity.QuestionContent) error {
	return r.db.Create(content).Error
}

func (r *QuestionRepository) AddAnswer(answer *entity.QuestionAnswer) error {
	return r.db.Create(answer).Error
}

func (r *QuestionRepository) AddMemo(memo *entity.QuestionMemo) error {
	return r.db.Create(memo).Error
}

func (r *QuestionRepository) AddRefer(refer *entity.QuestionRefer) error {
	return r.db.Create(refer).Error
}

func (r *QuestionRepository) ReplaceTags(questionID uint, tags []entity.QuestionTag) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("question_id = ?", questionID).Delete(&entity.QuestionTag{}).Error; err != nil {
			return err
		}
		if len(tags) == 0 {
			return nil
		}
		return tx.Create(&tags).Error
	})
}

func (r *QuestionRepository) CreateSummary(summary *entity.QuestionSummary, refs []entity.QuestionSummaryReference) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(summary).Error; err != nil {
			return err
		}
		for i := range refs {
			refs[i].QuestionSummaryID = summary.ID
		}
		if len(refs) > 0 {
			return tx.Create(&refs).Error
		}
		return nil
	})
}
