package database

import (
	"solvi/internal/domain/entity"

	"gorm.io/gorm"
)

type AutoMigrator struct{}

func (AutoMigrator) Apply(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.User{},
		&entity.Question{},
		&entity.QuestionContent{},
		&entity.QuestionTag{},
		&entity.QuestionAnswer{},
		&entity.QuestionMemo{},
		&entity.QuestionRefer{},
		&entity.QuestionSummary{},
		&entity.QuestionSummaryReference{},
	)
}
