package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"solvi/internal/shared/testutils/containers"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbOnce sync.Once
	db     *gorm.DB
	dbErr  error
)

func DB(ctx context.Context) (*gorm.DB, error) {
	dbOnce.Do(func() {
		dsn, err := containers.ConnectionString(ctx)
		if err != nil {
			dbErr = err
			return
		}
		jst, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			dbErr = fmt.Errorf("failed to load timezone: %w", err)
			return
		}
		db, dbErr = gorm.Open(postgres.Open(dsn), &gorm.Config{
			NowFunc: func() time.Time { return time.Now().In(jst) },
			Logger: logger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags),
				logger.Config{
					SlowThreshold:             time.Second,
					LogLevel:                  logger.Warn,
					IgnoreRecordNotFoundError: true,
					Colorful:                  true,
				},
			),
		})
		if dbErr != nil {
			dbErr = fmt.Errorf("failed to open gorm db: %w", dbErr)
			return
		}
		if err := (AutoMigrator{}).Apply(db); err != nil {
			dbErr = fmt.Errorf("failed to migrate database: %w", err)
		}
	})
	return db, dbErr
}
