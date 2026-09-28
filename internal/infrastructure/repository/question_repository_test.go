package repository_test

import (
	"context"
	"testing"
	"time"

	"solvi/internal/domain/entity"
	"solvi/internal/domain/valueobject"
	"solvi/internal/infrastructure/repository"
	"solvi/internal/shared/testutils/database"
)

func TestQuestionRepository_Integration(t *testing.T) {
	ctx := context.Background()
	db, err := database.DB(ctx)
	if err != nil {
		t.Skipf("postgres testcontainer unavailable: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	user := &entity.User{Name: "Q Author", Email: "author@solvi.local"}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatal(err)
	}

	qRepo := repository.NewQuestionRepository(db)

	// 1. Create
	due := time.Now().Add(24 * time.Hour)
	question := &entity.Question{
		Title:                 "Integration Question",
		QuestionUserID:        user.ID,
		AnswerDue:             &due,
		SupportStatus:         valueobject.SupportStatusPending,
		IsRequireHumanSupport: false,
	}
	if err := qRepo.Create(ctx, question); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// 2. GetByUUID
	got, err := qRepo.GetByUUID(ctx, question.UUID.String())
	if err != nil || got.Title != "Integration Question" {
		t.Fatalf("GetByUUID failed: %v, got: %+v", err, got)
	}

	// 3. ListByQuestionUserID & ListAll
	byUser, err := qRepo.ListByQuestionUserID(ctx, user.ID)
	if err != nil || len(byUser) == 0 {
		t.Fatalf("ListByQuestionUserID failed: %v, len: %d", err, len(byUser))
	}
	all, err := qRepo.ListAll(ctx)
	if err != nil || len(all) == 0 {
		t.Fatalf("ListAll failed: %v, len: %d", err, len(all))
	}

	// 4. Update
	question.Title = "Updated Question Title"
	if err := qRepo.Update(ctx, question); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 5. AddContent, AddAnswer, AddMemo, AddRefer
	content := &entity.QuestionContent{QuestionID: question.ID, QuestionUserID: user.ID, Content: "detail 1"}
	if err := qRepo.AddContent(ctx, content); err != nil {
		t.Fatalf("AddContent failed: %v", err)
	}

	answer := &entity.QuestionAnswer{QuestionID: question.ID, AnswerUserID: user.ID, Content: "solution 1"}
	if err := qRepo.AddAnswer(ctx, answer); err != nil {
		t.Fatalf("AddAnswer failed: %v", err)
	}

	memo := &entity.QuestionMemo{QuestionID: question.ID, MemoUserID: user.ID, Content: "internal 1"}
	if err := qRepo.AddMemo(ctx, memo); err != nil {
		t.Fatalf("AddMemo failed: %v", err)
	}

	refer := &entity.QuestionRefer{QuestionID: question.ID, UserID: user.ID, Name: "ref manual", URL: "http://ref.local"}
	if err := qRepo.AddRefer(ctx, refer); err != nil {
		t.Fatalf("AddRefer failed: %v", err)
	}

	// 6. ReplaceTags
	tags := []entity.QuestionTag{{Name: "Tag1", QuestionID: question.ID}, {Name: "Tag2", QuestionID: question.ID}}
	if err := qRepo.ReplaceTags(ctx, question.ID, tags); err != nil {
		t.Fatalf("ReplaceTags failed: %v", err)
	}
	// empty tags replace
	if err := qRepo.ReplaceTags(ctx, question.ID, nil); err != nil {
		t.Fatalf("ReplaceTags with empty slice failed: %v", err)
	}

	// 7. CreateSummary
	summary := &entity.QuestionSummary{
		Title:      "Summary",
		Content:    "Q Summary",
		Answer:     "A Summary",
		QuestionID: question.ID,
	}
	summaryRefs := []entity.QuestionSummaryReference{
		{Name: "Ref A", URL: "http://ref-a"},
	}
	if err := qRepo.CreateSummary(ctx, summary, summaryRefs); err != nil {
		t.Fatalf("CreateSummary failed: %v", err)
	}

	// 8. UpsertSummary (update branch)
	upsertRefs := []entity.QuestionSummaryReference{
		{Name: "Ref A", URL: "http://ref-a"},
	}
	if err := qRepo.UpsertSummary(ctx, question.ID, "New Summary Title", "New Content", "New Answer", upsertRefs); err != nil {
		t.Fatalf("UpsertSummary update failed: %v", err)
	}

	// 8b. ReplaceTags and ListSummaries / ListTagsByQuestionIDs
	tagsWithNames := []entity.QuestionTag{{Name: "FAQTag", QuestionID: question.ID}}
	if err := qRepo.ReplaceTags(ctx, question.ID, tagsWithNames); err != nil {
		t.Fatalf("ReplaceTags for summary test failed: %v", err)
	}
	summaries, err := qRepo.ListSummaries(ctx)
	if err != nil || len(summaries) == 0 {
		t.Fatalf("ListSummaries failed: %v, len: %d", err, len(summaries))
	}
	tagMap, err := qRepo.ListTagsByQuestionIDs(ctx, []uint{question.ID})
	if err != nil || len(tagMap[question.ID]) != 1 || tagMap[question.ID][0] != "FAQTag" {
		t.Fatalf("ListTagsByQuestionIDs failed: %v, map: %+v", err, tagMap)
	}
	summaryUUID := summaries[0].UUID.String()
	if err := qRepo.SoftDeleteSummaryByUUID(ctx, summaryUUID); err != nil {
		t.Fatalf("SoftDeleteSummaryByUUID failed: %v", err)
	}
	afterDelete, err := qRepo.ListSummaries(ctx)
	if err != nil {
		t.Fatalf("ListSummaries after delete failed: %v", err)
	}
	for _, s := range afterDelete {
		if s.UUID.String() == summaryUUID {
			t.Fatal("deleted summary still listed")
		}
	}

	// 9. Soft deletes
	if err := qRepo.SoftDeleteAnswerByUUID(ctx, answer.UUID.String()); err != nil {
		t.Fatalf("SoftDeleteAnswerByUUID failed: %v", err)
	}
	if err := qRepo.SoftDeleteMemoByUUID(ctx, memo.UUID.String()); err != nil {
		t.Fatalf("SoftDeleteMemoByUUID failed: %v", err)
	}
	if err := qRepo.SoftDeleteReferByUUID(ctx, refer.UUID.String()); err != nil {
		t.Fatalf("SoftDeleteReferByUUID failed: %v", err)
	}
	if err := qRepo.SoftDeleteByUUID(ctx, question.UUID.String()); err != nil {
		t.Fatalf("SoftDeleteByUUID failed: %v", err)
	}

	// Not found
	if _, err := qRepo.GetByUUID(ctx, question.UUID.String()); err == nil {
		t.Fatal("expected deleted question to be not found")
	}
}
