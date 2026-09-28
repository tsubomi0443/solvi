package repository_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"solvi/internal/domain/entity"
	"solvi/internal/domain/entity/lineworks"
	"solvi/internal/domain/valueobject"
	"solvi/internal/infrastructure/repository"
	"solvi/internal/shared/testutils/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestLineWorksNotificationRepository_Integration(t *testing.T) {
	ctx := context.Background()
	db, err := database.DB(ctx)
	if err != nil {
		t.Skipf("postgres testcontainer unavailable: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	asker := &entity.User{Name: "Asker", Email: "asker-" + uuid.NewString() + "@solvi.local", LineWorksLoginID: "asker"}
	if err := userRepo.Create(ctx, asker); err != nil {
		t.Fatal(err)
	}
	qRepo := repository.NewQuestionRepository(db)
	question := &entity.Question{
		Title:          "通知テスト",
		QuestionUserID: asker.ID,
		SupportStatus:  valueobject.SupportStatusPending,
		Contents:       []entity.QuestionContent{{Content: "本文", QuestionUserID: asker.ID}},
	}
	notice := &lineworks.Notification{
		Event:         valueobject.LineWorksEventReceived,
		Destination:   valueobject.LineWorksDestinationChannel,
		Body:          "【新規受付】",
		Status:        valueobject.LineWorksJobPending,
		NextAttemptAt: time.Now().Add(-time.Second),
	}
	if err := qRepo.CreateWithNotification(ctx, question, notice); err != nil {
		t.Fatal(err)
	}

	count, err := gorm.G[lineworks.Notification](db).
		Where("question_uuid = ? AND event = ?", question.UUID, valueobject.LineWorksEventReceived).
		Count(ctx, "*")
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || notice.StatusRevision != 1 {
		t.Fatalf("count=%d revision=%d", count, notice.StatusRevision)
	}

	dup := &lineworks.Notification{
		QuestionUUID:   question.UUID,
		Event:          valueobject.LineWorksEventReceived,
		StatusRevision: 1,
		Burst:          0,
		Destination:    valueobject.LineWorksDestinationChannel,
		Body:           "duplicate",
		Status:         valueobject.LineWorksJobPending,
		NextAttemptAt:  time.Now(),
	}
	if err := gorm.G[lineworks.Notification](db).Create(ctx, dup); err == nil {
		t.Fatal("expected unique violation")
	}

	question.SupportStatus = valueobject.SupportStatusSupporting
	reopen := &lineworks.Notification{
		Event:         valueobject.LineWorksEventReopened,
		Destination:   valueobject.LineWorksDestinationChannel,
		Body:          "【再対応依頼】\n追加コメント: 初回\nhttps://solvi.example/questions/" + question.UUID.String(),
		Status:        valueobject.LineWorksJobPending,
		NextAttemptAt: time.Now(),
	}
	if err := qRepo.UpdateWithNotification(ctx, question, reopen); err != nil {
		t.Fatal(err)
	}
	if reopen.StatusRevision != 2 {
		t.Fatalf("reopen revision=%d", reopen.StatusRevision)
	}

	follow := &lineworks.ReopenFollowUp{
		QuestionUUID: question.UUID,
		Comment:      "続き",
		Debounce:     time.Minute,
		Template: lineworks.Notification{
			Destination: valueobject.LineWorksDestinationChannel,
			Body:        "fresh",
			Status:      valueobject.LineWorksJobPending,
		},
	}
	if err := qRepo.AddContentWithReopenAggregate(ctx, &entity.QuestionContent{
		Content: "続き", QuestionUserID: asker.ID, QuestionID: question.ID,
	}, follow); err != nil {
		t.Fatal(err)
	}
	pending, err := gorm.G[lineworks.Notification](db).
		Where("question_uuid = ? AND event = ? AND status = ?", question.UUID, valueobject.LineWorksEventReopened, valueobject.LineWorksJobPending).
		First(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(pending.Body, "追加コメント: 続き") {
		t.Fatalf("body=%s", pending.Body)
	}
	count, err = gorm.G[lineworks.Notification](db).
		Where("question_uuid = ? AND event = ?", question.UUID, valueobject.LineWorksEventReopened).
		Count(ctx, "*")
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("reopen count=%d", count)
	}

	noticeRepo := repository.NewLineWorksNotificationRepository(db)
	claimed, err := noticeRepo.ClaimDue(ctx, time.Now().Add(time.Hour), 20)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, c := range claimed {
		if c.QuestionUUID == question.UUID && c.Event == valueobject.LineWorksEventReceived {
			found = true
			if c.Status != valueobject.LineWorksJobSending {
				t.Fatalf("status=%v", c.Status)
			}
		}
	}
	if !found {
		t.Fatal("received notice was not claimed")
	}
}
