package database

import (
	"solvi/internal/domain/entity"
	"solvi/internal/domain/entity/lineworks"

	"gorm.io/gorm"
)

type AutoMigrator struct{}

func (AutoMigrator) Apply(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&entity.User{},
		&entity.Question{},
		&entity.QuestionContent{},
		&entity.QuestionTag{},
		&entity.QuestionAnswer{},
		&entity.QuestionMemo{},
		&entity.QuestionRefer{},
		&entity.QuestionSummary{},
		&entity.QuestionSummaryReference{},
		&lineworks.Notification{},
		&lineworks.NotificationAttempt{},
	); err != nil {
		return err
	}
	notice := &lineworks.Notification{}
	if db.Migrator().HasIndex(notice, "ux_lineworks_notice") {
		if err := db.Migrator().DropIndex(notice, "ux_lineworks_notice"); err != nil {
			return err
		}
	}
	return db.AutoMigrate(notice)
}
