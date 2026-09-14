package containers

import (
	"context"
	"fmt"
	"sync"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresImage = "postgres:18.4-alpine"
	dbName        = "solvi_test"
	dbUser        = "postgres"
	dbPassword    = "password"
)

var (
	once       sync.Once
	initErr    error
	container  *postgres.PostgresContainer
	connString string
)

func ConnectionString(ctx context.Context) (string, error) {
	if err := ensureStarted(ctx); err != nil {
		return "", err
	}
	return connString, nil
}

func Terminate(ctx context.Context) error {
	if container == nil {
		return nil
	}
	return container.Terminate(ctx)
}

func ensureStarted(ctx context.Context) error {
	once.Do(func() {
		container, initErr = postgres.Run(ctx,
			postgresImage,
			postgres.WithDatabase(dbName),
			postgres.WithUsername(dbUser),
			postgres.WithPassword(dbPassword),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			),
		)
		if initErr != nil {
			initErr = fmt.Errorf("failed to start postgres container: %w", initErr)
			return
		}
		connString, initErr = container.ConnectionString(ctx, "sslmode=disable", "TimeZone=Asia/Tokyo")
		if initErr != nil {
			initErr = fmt.Errorf("failed to get postgres connection string: %w", initErr)
		}
	})
	return initErr
}
