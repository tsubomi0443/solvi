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
	if err := repo.Create(user); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByEmail("test-repo@solvi.local")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Test" {
		t.Fatalf("unexpected user: %+v", got)
	}
}
