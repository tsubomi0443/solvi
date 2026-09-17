package repository_test

import (
	"context"
	"testing"

	"solvi/internal/domain/entity"
	"solvi/internal/infrastructure/repository"
	"solvi/internal/shared/testutils/database"
)

func TestUserRepository_CreateAndGetByEmail(t *testing.T) {
	ctx := context.Background()
	db, err := database.DB(ctx)
	if err != nil {
		t.Skipf("postgres testcontainer unavailable: %v", err)
	}

	repo := repository.NewUserRepository(db)
	user := &entity.User{Name: "Test", Email: "test-repo@solvi.local"}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByEmail(ctx, "test-repo@solvi.local")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Test" {
		t.Fatalf("unexpected user: %+v", got)
	}
}

func TestUserRepository_SoftDeleteByUUID(t *testing.T) {
	ctx := context.Background()
	db, err := database.DB(ctx)
	if err != nil {
		t.Skipf("postgres testcontainer unavailable: %v", err)
	}

	repo := repository.NewUserRepository(db)
	user := &entity.User{Name: "DeleteMe", Email: "delete-repo@solvi.local"}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatal(err)
	}
	if err := repo.SoftDeleteByUUID(ctx, user.UUID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByEmail(ctx, "delete-repo@solvi.local"); err == nil {
		t.Fatal("expected deleted user to be hidden")
	}
}
