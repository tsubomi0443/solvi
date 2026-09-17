package repository_test

import (
	"context"
	"testing"

	"solvi/internal/domain/entity"
	"solvi/internal/infrastructure/repository"
	"solvi/internal/shared/testutils/database"
)

func TestTagRepository_Integration(t *testing.T) {
	ctx := context.Background()
	db, err := database.DB(ctx)
	if err != nil {
		t.Skipf("postgres testcontainer unavailable: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	user := &entity.User{Name: "Tag Author", Email: "tag-author@solvi.local"}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatal(err)
	}

	qRepo := repository.NewQuestionRepository(db)
	q := &entity.Question{
		Title:          "Tag Question",
		QuestionUserID: user.ID,
	}
	if err := qRepo.Create(ctx, q); err != nil {
		t.Fatal(err)
	}

	tagRepo := repository.NewTagRepository(db)
	tags := []entity.QuestionTag{
		{Name: "TagA", QuestionID: q.ID},
		{Name: "TagA", QuestionID: q.ID},
		{Name: "TagB", QuestionID: q.ID},
	}
	if err := qRepo.ReplaceTags(ctx, q.ID, tags); err != nil {
		t.Fatal(err)
	}

	// ListTagStats
	stats, err := tagRepo.ListTagStats(ctx)
	if err != nil {
		t.Fatalf("ListTagStats failed: %v", err)
	}
	if len(stats) == 0 {
		t.Fatal("expected at least 1 tag stat")
	}

	// RenameTag
	if err := tagRepo.RenameTag(ctx, "TagA", "TagA-Renamed"); err != nil {
		t.Fatalf("RenameTag failed: %v", err)
	}

	// DeleteTagByName
	if err := tagRepo.DeleteTagByName(ctx, "TagB"); err != nil {
		t.Fatalf("DeleteTagByName failed: %v", err)
	}
}
