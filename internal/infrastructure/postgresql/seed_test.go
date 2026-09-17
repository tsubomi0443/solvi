package postgresql_test

import (
	"context"
	"os"
	"testing"

	"solvi/internal/domain/entity"
	"solvi/internal/infrastructure/postgresql"
	"solvi/internal/shared/testutils/database"
)

func TestSeed_CreatesAndIdempotent(t *testing.T) {
	ctx := context.Background()
	db, err := database.DB(ctx)
	if err != nil {
		t.Skipf("postgres testcontainer unavailable: %v", err)
	}

	os.Setenv("PEPPER", "test-pepper")

	// 1回目の実行
	if err := postgresql.Seed(db); err != nil {
		t.Fatalf("Seed failed: %v", err)
	}

	var count int64
	if err := db.Model(&entity.User{}).Where("email = ?", "admin@solvi.local").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("expected admin user to be created")
	}

	// 2回目の実行（冪等性確認）
	if err := postgresql.Seed(db); err != nil {
		t.Fatalf("Second Seed run failed: %v", err)
	}
}
