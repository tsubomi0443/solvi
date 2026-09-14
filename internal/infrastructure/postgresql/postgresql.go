package postgresql

import (
	"time"

	"solvi/internal/domain/entity"

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
