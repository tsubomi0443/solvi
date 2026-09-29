package repository_test

import (
	"context"
	"testing"
	"time"

	"solvi/internal/domain/entity/lineworks"
	"solvi/internal/domain/valueobject"
	"solvi/internal/infrastructure/repository"
	"solvi/internal/shared/testutils/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestLineWorksNotificationRepository_EnqueueDueDigestsDedupes(t *testing.T) {
	ctx := context.Background()
	db, err := database.DB(ctx)
	if err != nil {
		t.Skipf("postgres testcontainer unavailable: %v", err)
	}

	repo := repository.NewLineWorksNotificationRepository(db)
	notices := []lineworks.Notification{{
		QuestionUUID:   uuid.Nil,
		Event:          valueobject.LineWorksEventDueDigest,
		StatusRevision: 20260930,
		Burst:          540,
		ChannelID:      "room-a",
		Destination:    valueobject.LineWorksDestinationChannel,
		Body:           "【定期通知】\n期日に近づいている問い合わせはありませんでした。",
		Status:         valueobject.LineWorksJobPending,
		NextAttemptAt:  time.Now(),
	}}
	if err := repo.EnqueueDueDigests(ctx, notices); err != nil {
		t.Fatal(err)
	}
	if err := repo.EnqueueDueDigests(ctx, notices); err != nil {
		t.Fatal(err)
	}
	count, err := gorm.G[lineworks.Notification](db).
		Where("event = ? AND status_revision = ? AND burst = ? AND channel_id = ?", valueobject.LineWorksEventDueDigest, 20260930, 540, "room-a").
		Count(ctx, "*")
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count=%d", count)
	}
}
