package repository

import (
	"context"
	"time"

	"solvi/internal/domain/entity/lineworks"
	"solvi/internal/domain/valueobject"
	logutils "solvi/internal/shared/logUtils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const opLineWorksRepo = "LineWorksNotificationRepository"

type LineWorksNotificationRepository struct {
	db *gorm.DB
}

func NewLineWorksNotificationRepository(db *gorm.DB) *LineWorksNotificationRepository {
	return &LineWorksNotificationRepository{db: db}
}

func (r *LineWorksNotificationRepository) ClaimDue(ctx context.Context, now time.Time, limit int) ([]lineworks.Notification, error) {
	const op = opLineWorksRepo + ".ClaimDue"
	if limit <= 0 {
		limit = 20
	}
	var claimed []lineworks.Notification
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		rows, err := gorm.G[lineworks.Notification](tx, clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND next_attempt_at <= ?", valueobject.LineWorksJobPending, now).
			Order("id").
			Limit(limit).
			Find(ctx)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		ids := make([]uint, len(rows))
		for i := range rows {
			ids[i] = rows[i].ID
		}
		if _, err := gorm.G[lineworks.Notification](tx).
			Where("id IN ?", ids).
			Set(clause.Assignments(map[string]interface{}{
				"status":     int(valueobject.LineWorksJobSending),
				"updated_at": now,
			})).
			Update(ctx); err != nil {
			return err
		}
		claimed, err = gorm.G[lineworks.Notification](tx).Where("id IN ?", ids).Find(ctx)
		return err
	})
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "通知予定の取得失敗", "err", err.Error())
		return nil, err
	}
	return claimed, nil
}

func (r *LineWorksNotificationRepository) RecoverStale(ctx context.Context, staleBefore time.Time, maxAttempts int) (int, error) {
	const op = opLineWorksRepo + ".RecoverStale"
	if maxAttempts <= 0 {
		maxAttempts = valueobject.LineWorksMaxSendAttempts
	}
	var recovered int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		rows, err := gorm.G[lineworks.Notification](tx).
			Where("status = ? AND updated_at < ?", valueobject.LineWorksJobSending, staleBefore).
			Find(ctx)
		if err != nil {
			return err
		}
		now := time.Now()
		for i := range rows {
			n := &rows[i]
			n.AttemptCount++
			n.ErrorKind = valueobject.LineWorksErrorUnknownOutcome
			n.LastError = "LINE WORKSへの送信成否が不明です"
			if n.AttemptCount >= maxAttempts {
				n.Status = valueobject.LineWorksJobFailed
			} else {
				n.Status = valueobject.LineWorksJobPending
				n.NextAttemptAt = now.Add(30 * time.Second)
			}
			if err := tx.Save(n).Error; err != nil {
				return err
			}
			attempt := lineworks.NotificationAttempt{
				NotificationUUID: n.UUID,
				Outcome:          valueobject.LineWorksAttemptUnknown,
				AttemptedAt:      now,
			}
			if err := gorm.G[lineworks.NotificationAttempt](tx).Create(ctx, &attempt); err != nil {
				return err
			}
			recovered++
		}
		return nil
	})
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "送信中の復旧失敗", "err", err.Error())
		return 0, err
	}
	return recovered, nil
}

func (r *LineWorksNotificationRepository) SaveResult(ctx context.Context, notice *lineworks.Notification, attempt *lineworks.NotificationAttempt) error {
	const op = opLineWorksRepo + ".SaveResult"
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(notice).Error; err != nil {
			return err
		}
		if attempt == nil {
			return nil
		}
		attempt.NotificationUUID = notice.UUID
		return gorm.G[lineworks.NotificationAttempt](tx).Create(ctx, attempt)
	})
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "送信結果の保存失敗", "err", err.Error())
	}
	return err
}

func (r *LineWorksNotificationRepository) EnqueueDueDigests(ctx context.Context, notices []lineworks.Notification) error {
	const op = opLineWorksRepo + ".EnqueueDueDigests"
	if len(notices) == 0 {
		return nil
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range notices {
			n := notices[i]
			if err := gorm.G[lineworks.Notification](tx, clause.OnConflict{
				Columns: []clause.Column{
					{Name: "question_uuid"},
					{Name: "event"},
					{Name: "status_revision"},
					{Name: "burst"},
					{Name: "channel_id"},
				},
				DoNothing: true,
			}).Create(ctx, &n); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "期日通知の登録失敗", "err", err.Error())
	}
	return err
}
