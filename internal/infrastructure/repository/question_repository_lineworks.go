package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"solvi/internal/domain/entity"
	"solvi/internal/domain/entity/lineworks"
	"solvi/internal/domain/valueobject"
	logutils "solvi/internal/shared/logUtils"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func lockQuestion(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	_, err := gorm.G[entity.Question](tx, clause.Locking{Strength: "UPDATE"}).
		Select("id").
		Where("uuid = ?", id).
		First(ctx)
	return err
}

func nextStatusRevision(ctx context.Context, tx *gorm.DB, questionUUID uuid.UUID) (int, error) {
	latest, err := gorm.G[lineworks.Notification](tx).
		Where("question_uuid = ?", questionUUID).
		Order("status_revision DESC").
		First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return latest.StatusRevision + 1, nil
}

func currentStatusRevision(ctx context.Context, tx *gorm.DB, questionUUID uuid.UUID) (int, error) {
	latest, err := gorm.G[lineworks.Notification](tx).
		Where("question_uuid = ?", questionUUID).
		Order("status_revision DESC").
		First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return latest.StatusRevision, nil
}

func insertNotice(ctx context.Context, tx *gorm.DB, notice *lineworks.Notification) error {
	if notice.StatusRevision == 0 {
		rev, err := nextStatusRevision(ctx, tx, notice.QuestionUUID)
		if err != nil {
			return err
		}
		notice.StatusRevision = rev
	}
	return gorm.G[lineworks.Notification](tx, clause.OnConflict{
		Columns: []clause.Column{
			{Name: "question_uuid"},
			{Name: "event"},
			{Name: "status_revision"},
			{Name: "burst"},
		},
		DoNothing: true,
	}).Create(ctx, notice)
}

func (r *QuestionRepository) CreateWithNotification(ctx context.Context, question *entity.Question, notice *lineworks.Notification) error {
	const op = opQuestionRepo + ".CreateWithNotification"
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := gorm.G[entity.Question](tx).Create(ctx, question); err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", "err", err.Error())
			return err
		}
		if notice == nil {
			return nil
		}
		notice.QuestionUUID = question.UUID
		notice.Burst = 0
		notice.StatusRevision = 0
		if err := insertNotice(ctx, tx, notice); err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", "err", err.Error())
			return err
		}
		return nil
	})
}

func (r *QuestionRepository) UpdateWithNotification(ctx context.Context, question *entity.Question, notice *lineworks.Notification) error {
	const op = opQuestionRepo + ".UpdateWithNotification"
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockQuestion(ctx, tx, question.UUID); err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", "err", err.Error())
			return err
		}
		if err := tx.Omit("QuestionUser", "Contents", "Tags", "Answers", "Memos", "Refers", "Summary").Save(question).Error; err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", "err", err.Error())
			return err
		}
		if notice == nil {
			return nil
		}
		notice.QuestionUUID = question.UUID
		notice.StatusRevision = 0
		notice.Burst = 0
		if err := insertNotice(ctx, tx, notice); err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", "err", err.Error())
			return err
		}
		return nil
	})
}

func (r *QuestionRepository) CompleteWithNotification(ctx context.Context, question *entity.Question, summaryTitle, summaryContent, summaryAnswer string, refs []entity.QuestionSummaryReference, notice *lineworks.Notification) error {
	const op = opQuestionRepo + ".CompleteWithNotification"
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockQuestion(ctx, tx, question.UUID); err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", "err", err.Error())
			return err
		}
		if err := tx.Omit("QuestionUser", "Contents", "Tags", "Answers", "Memos", "Refers", "Summary").Save(question).Error; err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", "err", err.Error())
			return err
		}
		if err := upsertSummaryTx(ctx, tx, question.ID, summaryTitle, summaryContent, summaryAnswer, refs); err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", "err", err.Error())
			return err
		}
		if notice == nil {
			return nil
		}
		notice.QuestionUUID = question.UUID
		notice.StatusRevision = 0
		notice.Burst = 0
		if err := insertNotice(ctx, tx, notice); err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", "err", err.Error())
			return err
		}
		return nil
	})
}

func (r *QuestionRepository) AddContentWithReopenAggregate(ctx context.Context, content *entity.QuestionContent, followUp *lineworks.ReopenFollowUp) error {
	const op = opQuestionRepo + ".AddContentWithReopenAggregate"
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if followUp != nil {
			if err := lockQuestion(ctx, tx, followUp.QuestionUUID); err != nil {
				logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", "err", err.Error())
				return err
			}
		}
		if err := gorm.G[entity.QuestionContent](tx).Create(ctx, content); err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", "err", err.Error())
			return err
		}
		if followUp == nil {
			return nil
		}
		if err := aggregateReopen(ctx, tx, followUp); err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", "err", err.Error())
			return err
		}
		return nil
	})
}

func aggregateReopen(ctx context.Context, tx *gorm.DB, followUp *lineworks.ReopenFollowUp) error {
	pending, err := gorm.G[lineworks.Notification](tx).
		Where("question_uuid = ? AND event = ? AND status = ?", followUp.QuestionUUID, valueobject.LineWorksEventReopened, valueobject.LineWorksJobPending).
		Order("status_revision DESC, burst DESC").
		First(ctx)
	if err == nil {
		pending.Body = appendReopenComment(pending.Body, followUp.Comment)
		pending.NextAttemptAt = time.Now().Add(followUp.Debounce)
		return tx.Save(&pending).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	latest, err := gorm.G[lineworks.Notification](tx).
		Where("question_uuid = ? AND event = ?", followUp.QuestionUUID, valueobject.LineWorksEventReopened).
		Order("status_revision DESC, burst DESC").
		First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		rev, revErr := currentStatusRevision(ctx, tx, followUp.QuestionUUID)
		if revErr != nil {
			return revErr
		}
		notice := followUp.Template
		notice.ID = 0
		notice.QuestionUUID = followUp.QuestionUUID
		notice.Event = valueobject.LineWorksEventReopened
		notice.StatusRevision = rev
		notice.Burst = 0
		notice.Status = valueobject.LineWorksJobPending
		notice.NextAttemptAt = time.Now().Add(followUp.Debounce)
		return insertNotice(ctx, tx, &notice)
	}
	if err != nil {
		return err
	}
	if latest.Status == valueobject.LineWorksJobSending {
		return nil
	}
	if latest.Status == valueobject.LineWorksJobSent {
		sentAt := latest.UpdatedAt
		if latest.SentAt != nil {
			sentAt = *latest.SentAt
		}
		if time.Since(sentAt) < followUp.Debounce {
			return nil
		}
	}
	notice := followUp.Template
	notice.ID = 0
	notice.UUID = uuid.Nil
	notice.QuestionUUID = followUp.QuestionUUID
	notice.Event = valueobject.LineWorksEventReopened
	notice.StatusRevision = latest.StatusRevision
	notice.Burst = latest.Burst + 1
	notice.Status = valueobject.LineWorksJobPending
	notice.NextAttemptAt = time.Now().Add(followUp.Debounce)
	return insertNotice(ctx, tx, &notice)
}

func appendReopenComment(body, comment string) string {
	line := "追加コメント: " + strings.TrimSpace(comment)
	body = strings.TrimRight(body, "\n")
	if i := strings.LastIndex(body, "\n"); i >= 0 {
		last := body[i+1:]
		if strings.Contains(last, "://") {
			return body[:i] + "\n" + line + "\n" + last
		}
	}
	return body + "\n" + line
}
