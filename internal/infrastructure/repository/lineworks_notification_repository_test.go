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
	notices := []*lineworks.Notification{
		{
			Event:         valueobject.LineWorksEventReceived,
			Destination:   valueobject.LineWorksDestinationChannel,
			ChannelID:     "room-a",
			Body:          "【新規受付】",
			Status:        valueobject.LineWorksJobPending,
			NextAttemptAt: time.Now().Add(-time.Second),
		},
		{
			Event:         valueobject.LineWorksEventReceived,
			Destination:   valueobject.LineWorksDestinationChannel,
			ChannelID:     "room-b",
			Body:          "【新規受付】",
			Status:        valueobject.LineWorksJobPending,
			NextAttemptAt: time.Now().Add(-time.Second),
		},
	}
	if err := qRepo.CreateWithNotification(ctx, question, notices); err != nil {
		t.Fatal(err)
	}

	rows, err := gorm.G[lineworks.Notification](db).
		Where("question_uuid = ? AND event = ?", question.UUID, valueobject.LineWorksEventReceived).
		Find(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("count=%d", len(rows))
	}
	if rows[0].StatusRevision != rows[1].StatusRevision {
		t.Fatalf("revisions=%d %d", rows[0].StatusRevision, rows[1].StatusRevision)
	}
	channels := map[string]bool{rows[0].ChannelID: true, rows[1].ChannelID: true}
	if !channels["room-a"] || !channels["room-b"] {
		t.Fatalf("channels=%v", channels)
	}

	dup := &lineworks.Notification{
		QuestionUUID:   question.UUID,
		Event:          valueobject.LineWorksEventReceived,
		StatusRevision: rows[0].StatusRevision,
		Burst:          0,
		ChannelID:      "room-a",
		Destination:    valueobject.LineWorksDestinationChannel,
		Body:           "duplicate",
		Status:         valueobject.LineWorksJobPending,
		NextAttemptAt:  time.Now(),
	}
	if err := gorm.G[lineworks.Notification](db).Create(ctx, dup); err == nil {
		t.Fatal("expected unique violation")
	}

	question.SupportStatus = valueobject.SupportStatusSupporting
	reopenNotices := []*lineworks.Notification{{
		Event:         valueobject.LineWorksEventReopened,
		Destination:   valueobject.LineWorksDestinationChannel,
		ChannelID:     "room-a",
		Body:          "【再対応依頼】\n追加コメント: 初回\nhttps://solvi.example/questions/" + question.UUID.String(),
		Status:        valueobject.LineWorksJobPending,
		NextAttemptAt: time.Now(),
	}, {
		Event:         valueobject.LineWorksEventReopened,
		Destination:   valueobject.LineWorksDestinationChannel,
		ChannelID:     "room-b",
		Body:          "【再対応依頼】\n追加コメント: 初回\nhttps://solvi.example/questions/" + question.UUID.String(),
		Status:        valueobject.LineWorksJobPending,
		NextAttemptAt: time.Now(),
	}}
	if err := qRepo.UpdateWithNotification(ctx, question, reopenNotices); err != nil {
		t.Fatal(err)
	}
	if reopenNotices[0].StatusRevision != 2 || reopenNotices[1].StatusRevision != 2 {
		t.Fatalf("reopen revisions=%d %d", reopenNotices[0].StatusRevision, reopenNotices[1].StatusRevision)
	}

	follow := &lineworks.ReopenFollowUp{
		QuestionUUID: question.UUID,
		Comment:      "続き",
		Debounce:     time.Minute,
		ChannelIDs:   []string{"room-a", "room-b"},
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
	pendingRows, err := gorm.G[lineworks.Notification](db).
		Where("question_uuid = ? AND event = ? AND status = ?", question.UUID, valueobject.LineWorksEventReopened, valueobject.LineWorksJobPending).
		Find(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(pendingRows) != 2 {
		t.Fatalf("pending count=%d", len(pendingRows))
	}
	for _, pending := range pendingRows {
		if !strings.Contains(pending.Body, "追加コメント: 続き") {
			t.Fatalf("body=%s", pending.Body)
		}
	}

	reopenCount, err := gorm.G[lineworks.Notification](db).
		Where("question_uuid = ? AND event = ?", question.UUID, valueobject.LineWorksEventReopened).
		Count(ctx, "*")
	if err != nil {
		t.Fatal(err)
	}
	if reopenCount != 2 {
		t.Fatalf("reopen count=%d", reopenCount)
	}

	noticeRepo := repository.NewLineWorksNotificationRepository(db)
	claimed, err := noticeRepo.ClaimDue(ctx, time.Now().Add(time.Hour), 20)
	if err != nil {
		t.Fatal(err)
	}
	var found int
	for _, c := range claimed {
		if c.QuestionUUID == question.UUID && c.Event == valueobject.LineWorksEventReceived {
			found++
			if c.Status != valueobject.LineWorksJobSending {
				t.Fatalf("status=%v", c.Status)
			}
		}
	}
	if found != 2 {
		t.Fatalf("received notices claimed=%d", found)
	}
}
