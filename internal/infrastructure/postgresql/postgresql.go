package postgresql

import (
	"time"

	"solvi/internal/domain/entity"
	"solvi/internal/domain/entity/lineworks"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresqlDB(dsn string) (*gorm.DB, error) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().In(jst) },
	})
	if err != nil {
		return nil, err
	}

	if err := autoMigrateOnce(db); err != nil {
		return nil, err
	}
	if err := Seed(db); err != nil {
		return nil, err
	}
	return db, nil
}

func autoMigrateOnce(db *gorm.DB) error {
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
	return remigrateLineWorksNoticeIndex(db)
}

func remigrateLineWorksNoticeIndex(db *gorm.DB) error {
	notice := &lineworks.Notification{}
	if db.Migrator().HasIndex(notice, "ux_lineworks_notice") {
		if err := db.Migrator().DropIndex(notice, "ux_lineworks_notice"); err != nil {
			return err
		}
	}
	return db.AutoMigrate(notice)
}
