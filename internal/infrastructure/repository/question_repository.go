package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"solvi/internal/domain/entity"
	logutils "solvi/internal/shared/logUtils"

	"gorm.io/gorm"
)

const opQuestionRepo = "QuestionRepository"

type QuestionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository(db *gorm.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

func (r *QuestionRepository) preload(q *gorm.DB) *gorm.DB {
	return q.Preload("QuestionUser").
		Preload("Contents", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		Preload("Contents.QuestionUser").
		Preload("Tags").
		Preload("Answers", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		Preload("Answers.AnswerUser").
		Preload("Memos", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		Preload("Memos.MemoUser").
		Preload("Refers").
		Preload("Refers.User").
		Preload("Summary.References")
}

func (r *QuestionRepository) Create(ctx context.Context, question *entity.Question) error {
	const op = opQuestionRepo + ".Create"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB作成", slog.Uint64("question_user_id", uint64(question.QuestionUserID)))
	if err := r.db.WithContext(ctx).Create(question).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB作成失敗", slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) GetByUUID(ctx context.Context, uuid string) (*entity.Question, error) {
	const op = opQuestionRepo + ".GetByUUID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索", slog.String("question_uuid", uuid))
	var q entity.Question
	if err := r.preload(r.db.WithContext(ctx)).Where("uuid = ?", uuid).First(&q).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logutils.Debug(ctx, logutils.LayerRepository, op, "レコードなし", slog.String("question_uuid", uuid))
			return nil, fmt.Errorf("質問が見つかりません: %w", err)
		}
		logutils.Error(ctx, logutils.LayerRepository, op, "DB検索失敗", slog.String("question_uuid", uuid), slog.String("err", err.Error()))
		return nil, fmt.Errorf("質問が見つかりません: %w", err)
	}
	return &q, nil
}

func (r *QuestionRepository) ListByQuestionUserID(ctx context.Context, userID uint) ([]entity.Question, error) {
	const op = opQuestionRepo + ".ListByQuestionUserID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索", slog.Uint64("user_id", uint64(userID)))
	var qs []entity.Question
	err := r.preload(r.db.WithContext(ctx)).Where("question_user_id = ?", userID).Order("id DESC").Find(&qs).Error
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB検索失敗", slog.Uint64("user_id", uint64(userID)), slog.String("err", err.Error()))
		return nil, err
	}
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索完了", slog.Int("count", len(qs)))
	return qs, err
}

func (r *QuestionRepository) ListAll(ctx context.Context) ([]entity.Question, error) {
	const op = opQuestionRepo + ".ListAll"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索")
	var qs []entity.Question
	err := r.preload(r.db.WithContext(ctx)).Order("id DESC").Find(&qs).Error
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB検索失敗", slog.String("err", err.Error()))
		return nil, err
	}
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索完了", slog.Int("count", len(qs)))
	return qs, err
}

func (r *QuestionRepository) Update(ctx context.Context, question *entity.Question) error {
	const op = opQuestionRepo + ".Update"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB更新", slog.String("question_uuid", question.UUID.String()))
	if err := r.db.WithContext(ctx).Save(question).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", slog.String("question_uuid", question.UUID.String()), slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) AddContent(ctx context.Context, content *entity.QuestionContent) error {
	const op = opQuestionRepo + ".AddContent"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB作成", slog.Uint64("question_id", uint64(content.QuestionID)))
	if err := r.db.WithContext(ctx).Create(content).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB作成失敗", slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) AddAnswer(ctx context.Context, answer *entity.QuestionAnswer) error {
	const op = opQuestionRepo + ".AddAnswer"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB作成", slog.Uint64("question_id", uint64(answer.QuestionID)))
	if err := r.db.WithContext(ctx).Create(answer).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB作成失敗", slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) AddMemo(ctx context.Context, memo *entity.QuestionMemo) error {
	const op = opQuestionRepo + ".AddMemo"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB作成", slog.Uint64("question_id", uint64(memo.QuestionID)))
	if err := r.db.WithContext(ctx).Create(memo).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB作成失敗", slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) SoftDeleteAnswerByUUID(ctx context.Context, uuid string) error {
	const op = opQuestionRepo + ".SoftDeleteAnswerByUUID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB削除", slog.String("answer_uuid", uuid))
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&entity.QuestionAnswer{}).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB削除失敗", slog.String("answer_uuid", uuid), slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) SoftDeleteMemoByUUID(ctx context.Context, uuid string) error {
	const op = opQuestionRepo + ".SoftDeleteMemoByUUID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB削除", slog.String("memo_uuid", uuid))
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&entity.QuestionMemo{}).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB削除失敗", slog.String("memo_uuid", uuid), slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) SoftDeleteReferByUUID(ctx context.Context, uuid string) error {
	const op = opQuestionRepo + ".SoftDeleteReferByUUID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB削除", slog.String("refer_uuid", uuid))
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&entity.QuestionRefer{}).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB削除失敗", slog.String("refer_uuid", uuid), slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) SoftDeleteByUUID(ctx context.Context, uuid string) error {
	const op = opQuestionRepo + ".SoftDeleteByUUID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB削除", slog.String("question_uuid", uuid))
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&entity.Question{}).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB削除失敗", slog.String("question_uuid", uuid), slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) AddRefer(ctx context.Context, refer *entity.QuestionRefer) error {
	const op = opQuestionRepo + ".AddRefer"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB作成", slog.Uint64("question_id", uint64(refer.QuestionID)))
	if err := r.db.WithContext(ctx).Create(refer).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB作成失敗", slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) ReplaceTags(ctx context.Context, questionID uint, tags []entity.QuestionTag) error {
	const op = opQuestionRepo + ".ReplaceTags"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB更新", slog.Uint64("question_id", uint64(questionID)), slog.Int("tag_count", len(tags)))
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("question_id = ?", questionID).Delete(&entity.QuestionTag{}).Error; err != nil {
			return err
		}
		if len(tags) == 0 {
			return nil
		}
		return tx.Create(&tags).Error
	})
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", slog.Uint64("question_id", uint64(questionID)), slog.String("err", err.Error()))
	}
	return err
}

func (r *QuestionRepository) CreateSummary(ctx context.Context, summary *entity.QuestionSummary, refs []entity.QuestionSummaryReference) error {
	const op = opQuestionRepo + ".CreateSummary"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB作成", slog.Uint64("question_id", uint64(summary.QuestionID)))
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(summary).Error; err != nil {
			return err
		}
		for i := range refs {
			refs[i].QuestionSummaryID = summary.ID
		}
		if len(refs) > 0 {
			return tx.Create(&refs).Error
		}
		return nil
	})
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB作成失敗", slog.Uint64("question_id", uint64(summary.QuestionID)), slog.String("err", err.Error()))
	}
	return err
}

func (r *QuestionRepository) UpsertSummary(ctx context.Context, questionID uint, title, content, answer string, refs []entity.QuestionSummaryReference) error {
	const op = opQuestionRepo + ".UpsertSummary"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DBサマリー更新", slog.Uint64("question_id", uint64(questionID)))
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing entity.QuestionSummary
		findErr := tx.Unscoped().Where("question_id = ?", questionID).First(&existing).Error
		if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}

		var summaryID uint
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			summary := entity.QuestionSummary{
				Title:      title,
				Content:    content,
				Answer:     answer,
				QuestionID: questionID,
			}
			if err := tx.Create(&summary).Error; err != nil {
				return err
			}
			summaryID = summary.ID
		} else {
			existing.Title = title
			existing.Content = content
			existing.Answer = answer
			existing.DeletedAt = gorm.DeletedAt{}
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
			summaryID = existing.ID
		}

		if err := tx.Where("question_summary_id = ?", summaryID).Delete(&entity.QuestionSummaryReference{}).Error; err != nil {
			return err
		}

		for i := range refs {
			refs[i].ID = 0
			refs[i].QuestionSummaryID = summaryID
		}
		if len(refs) > 0 {
			if err := tx.Create(&refs).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DBサマリー更新失敗", slog.Uint64("question_id", uint64(questionID)), slog.String("err", err.Error()))
	}
	return err
}

func (r *QuestionRepository) ListSummaries(ctx context.Context) ([]entity.QuestionSummary, error) {
	const op = opQuestionRepo + ".ListSummaries"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索")
	var summaries []entity.QuestionSummary
	err := r.db.WithContext(ctx).
		Preload("References").
		Order("created_at DESC").
		Find(&summaries).Error
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB検索失敗", slog.String("err", err.Error()))
		return nil, err
	}
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索完了", slog.Int("count", len(summaries)))
	return summaries, nil
}

func (r *QuestionRepository) ListTagsByQuestionIDs(ctx context.Context, questionIDs []uint) (map[uint][]string, error) {
	const op = opQuestionRepo + ".ListTagsByQuestionIDs"
	if len(questionIDs) == 0 {
		return map[uint][]string{}, nil
	}
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索", slog.Int("question_count", len(questionIDs)))
	var tags []entity.QuestionTag
	if err := r.db.WithContext(ctx).Where("question_id IN ?", questionIDs).Order("id ASC").Find(&tags).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB検索失敗", slog.String("err", err.Error()))
		return nil, err
	}
	out := make(map[uint][]string, len(questionIDs))
	for _, t := range tags {
		out[t.QuestionID] = append(out[t.QuestionID], t.Name)
	}
	return out, nil
}

func (r *QuestionRepository) SoftDeleteSummaryByUUID(ctx context.Context, uuid string) error {
	const op = opQuestionRepo + ".SoftDeleteSummaryByUUID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB削除", slog.String("summary_uuid", uuid))
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var summary entity.QuestionSummary
		if err := tx.Where("uuid = ?", uuid).First(&summary).Error; err != nil {
			return err
		}
		if err := tx.Where("question_summary_id = ?", summary.ID).Delete(&entity.QuestionSummaryReference{}).Error; err != nil {
			return err
		}
		return tx.Delete(&summary).Error
	})
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB削除失敗", slog.String("summary_uuid", uuid), slog.String("err", err.Error()))
	}
	return err
}
