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

func TestUserRepository_CRUDAndNotFoundBranches(t *testing.T) {
	ctx := context.Background()
	db, err := database.DB(ctx)
	if err != nil {
		t.Skipf("postgres testcontainer unavailable: %v", err)
	}

	repo := repository.NewUserRepository(db)
	user := &entity.User{Name: "User1", Email: "u1@solvi.local", IsAdmin: true}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatal(err)
	}

	byID, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if byID.Email != "u1@solvi.local" {
		t.Fatalf("unexpected user: %+v", byID)
	}

	byUUID, err := repo.GetByUUID(ctx, user.UUID.String())
	if err != nil {
		t.Fatalf("GetByUUID failed: %v", err)
	}
	if byUUID.ID != user.ID {
		t.Fatalf("unexpected id: %d", byUUID.ID)
	}

	// Update
	byID.Name = "User1-Renamed"
	if err := repo.Update(ctx, byID); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	updated, err := repo.GetByID(ctx, user.ID)
	if err != nil || updated.Name != "User1-Renamed" {
		t.Fatalf("expected updated name, got %+v, err=%v", updated, err)
	}

	// Count admins
	count, err := repo.CountAdmins(ctx)
	if err != nil {
		t.Fatalf("CountAdmins failed: %v", err)
	}
	if count < 1 {
		t.Fatalf("expected at least 1 admin, got %d", count)
	}

	// ListAll
	all, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(all) == 0 {
		t.Fatal("expected at least one user in list")
	}

	// Not found branches
	if _, err := repo.GetByID(ctx, 999999); err == nil {
		t.Fatal("expected error for non-existent ID")
	}
	if _, err := repo.GetByUUID(ctx, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Fatal("expected error for non-existent UUID")
	}
	if _, err := repo.GetByEmail(ctx, "not-exist@solvi.local"); err == nil {
		t.Fatal("expected error for non-existent Email")
	}
}
