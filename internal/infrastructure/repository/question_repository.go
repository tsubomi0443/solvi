package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"solvi/internal/domain/entity"
	"solvi/internal/domain/valueobject"
	logutils "solvi/internal/shared/logUtils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const opQuestionRepo = "QuestionRepository"

type QuestionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository(db *gorm.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

func noopPreload(_ gorm.PreloadBuilder) error { return nil }

func (r *QuestionRepository) questionQuery(ctx context.Context) gorm.ChainInterface[entity.Question] {
	return gorm.G[entity.Question](r.db.WithContext(ctx)).
		Preload("QuestionUser", noopPreload).
		Preload("Contents", func(pb gorm.PreloadBuilder) error {
			pb.Order("id ASC")
			return nil
		}).
		Preload("Contents.QuestionUser", noopPreload).
		Preload("Tags", noopPreload).
		Preload("Answers", func(pb gorm.PreloadBuilder) error {
			pb.Order("id ASC")
			return nil
		}).
		Preload("Answers.AnswerUser", noopPreload).
		Preload("Memos", func(pb gorm.PreloadBuilder) error {
			pb.Order("id ASC")
			return nil
		}).
		Preload("Memos.MemoUser", noopPreload).
		Preload("Refers", noopPreload).
		Preload("Refers.User", noopPreload).
		Preload("Summary.References", noopPreload)
}

func (r *QuestionRepository) Create(ctx context.Context, question *entity.Question) error {
	const op = opQuestionRepo + ".Create"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB作成", slog.Uint64("question_user_id", uint64(question.QuestionUserID)))
	if err := gorm.G[entity.Question](r.db.WithContext(ctx)).Create(ctx, question); err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB作成失敗", slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) GetByUUID(ctx context.Context, uuid string) (*entity.Question, error) {
	const op = opQuestionRepo + ".GetByUUID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索", slog.String("question_uuid", uuid))
	q, err := r.questionQuery(ctx).Where("uuid = ?", uuid).First(ctx)
	if err != nil {
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
	qs, err := r.questionQuery(ctx).Where("question_user_id = ?", userID).Order("id DESC").Find(ctx)
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
	qs, err := r.questionQuery(ctx).Order("id DESC").Find(ctx)
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
	if err := gorm.G[entity.QuestionContent](r.db.WithContext(ctx)).Create(ctx, content); err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB作成失敗", slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) AddAnswer(ctx context.Context, answer *entity.QuestionAnswer) error {
	const op = opQuestionRepo + ".AddAnswer"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB作成", slog.Uint64("question_id", uint64(answer.QuestionID)))
	if err := gorm.G[entity.QuestionAnswer](r.db.WithContext(ctx)).Create(ctx, answer); err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB作成失敗", slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) AddMemo(ctx context.Context, memo *entity.QuestionMemo) error {
	const op = opQuestionRepo + ".AddMemo"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB作成", slog.Uint64("question_id", uint64(memo.QuestionID)))
	if err := gorm.G[entity.QuestionMemo](r.db.WithContext(ctx)).Create(ctx, memo); err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB作成失敗", slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) SoftDeleteAnswerByUUID(ctx context.Context, uuid string) error {
	const op = opQuestionRepo + ".SoftDeleteAnswerByUUID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB削除", slog.String("answer_uuid", uuid))
	if _, err := gorm.G[entity.QuestionAnswer](r.db.WithContext(ctx)).Where("uuid = ?", uuid).Delete(ctx); err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB削除失敗", slog.String("answer_uuid", uuid), slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) SoftDeleteMemoByUUID(ctx context.Context, uuid string) error {
	const op = opQuestionRepo + ".SoftDeleteMemoByUUID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB削除", slog.String("memo_uuid", uuid))
	if _, err := gorm.G[entity.QuestionMemo](r.db.WithContext(ctx)).Where("uuid = ?", uuid).Delete(ctx); err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB削除失敗", slog.String("memo_uuid", uuid), slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) SoftDeleteReferByUUID(ctx context.Context, uuid string) error {
	const op = opQuestionRepo + ".SoftDeleteReferByUUID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB削除", slog.String("refer_uuid", uuid))
	if _, err := gorm.G[entity.QuestionRefer](r.db.WithContext(ctx)).Where("uuid = ?", uuid).Delete(ctx); err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB削除失敗", slog.String("refer_uuid", uuid), slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) SoftDeleteByUUID(ctx context.Context, uuid string) error {
	const op = opQuestionRepo + ".SoftDeleteByUUID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB削除", slog.String("question_uuid", uuid))
	if _, err := gorm.G[entity.Question](r.db.WithContext(ctx)).Where("uuid = ?", uuid).Delete(ctx); err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB削除失敗", slog.String("question_uuid", uuid), slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) AddRefer(ctx context.Context, refer *entity.QuestionRefer) error {
	const op = opQuestionRepo + ".AddRefer"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB作成", slog.Uint64("question_id", uint64(refer.QuestionID)))
	if err := gorm.G[entity.QuestionRefer](r.db.WithContext(ctx)).Create(ctx, refer); err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB作成失敗", slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *QuestionRepository) ReplaceTags(ctx context.Context, questionID uint, tags []entity.QuestionTag) error {
	const op = opQuestionRepo + ".ReplaceTags"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB更新", slog.Uint64("question_id", uint64(questionID)), slog.Int("tag_count", len(tags)))
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := gorm.G[entity.QuestionTag](tx).Where("question_id = ?", questionID).Delete(ctx); err != nil {
			return err
		}
		if len(tags) == 0 {
			return nil
		}
		return gorm.G[entity.QuestionTag](tx).CreateInBatches(ctx, &tags, len(tags))
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
		if err := gorm.G[entity.QuestionSummary](tx).Create(ctx, summary); err != nil {
			return err
		}
		for i := range refs {
			refs[i].QuestionSummaryID = summary.ID
		}
		if len(refs) > 0 {
			return gorm.G[entity.QuestionSummaryReference](tx).CreateInBatches(ctx, &refs, len(refs))
		}
		return nil
	})
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB作成失敗", slog.Uint64("question_id", uint64(summary.QuestionID)), slog.String("err", err.Error()))
	}
	return err
}

func upsertSummaryTx(ctx context.Context, tx *gorm.DB, questionID uint, title, content, answer string, refs []entity.QuestionSummaryReference) error {
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
		if err := gorm.G[entity.QuestionSummary](tx).Create(ctx, &summary); err != nil {
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

	if _, err := gorm.G[entity.QuestionSummaryReference](tx).Where("question_summary_id = ?", summaryID).Delete(ctx); err != nil {
		return err
	}

	for i := range refs {
		refs[i].ID = 0
		refs[i].UUID = uuid.Nil
		refs[i].QuestionSummaryID = summaryID
	}
	if len(refs) > 0 {
		if err := gorm.G[entity.QuestionSummaryReference](tx).CreateInBatches(ctx, &refs, len(refs)); err != nil {
			return err
		}
	}
	return nil
}

func (r *QuestionRepository) UpsertSummary(ctx context.Context, questionID uint, title, content, answer string, refs []entity.QuestionSummaryReference) error {
	const op = opQuestionRepo + ".UpsertSummary"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DBサマリー更新", slog.Uint64("question_id", uint64(questionID)))
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return upsertSummaryTx(ctx, tx, questionID, title, content, answer, refs)
	})
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DBサマリー更新失敗", slog.Uint64("question_id", uint64(questionID)), slog.String("err", err.Error()))
	}
	return err
}

func (r *QuestionRepository) ListSummaries(ctx context.Context) ([]entity.QuestionSummary, error) {
	const op = opQuestionRepo + ".ListSummaries"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索")
	doneQuestionIDs := gorm.G[entity.Question](r.db.WithContext(ctx)).
		Select("id").
		Where("support_status = ?", valueobject.SupportStatusDone)
	summaries, err := gorm.G[entity.QuestionSummary](r.db.WithContext(ctx)).
		Preload("References", noopPreload).
		Where("question_id IN (?)", doneQuestionIDs).
		Order("created_at DESC").
		Find(ctx)
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
	tags, err := gorm.G[entity.QuestionTag](r.db.WithContext(ctx)).
		Where("question_id IN ?", questionIDs).
		Order("id ASC").
		Find(ctx)
	if err != nil {
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
		summary, err := gorm.G[entity.QuestionSummary](tx).Where("uuid = ?", uuid).First(ctx)
		if err != nil {
			return err
		}
		if _, err := gorm.G[entity.QuestionSummaryReference](tx).Where("question_summary_id = ?", summary.ID).Delete(ctx); err != nil {
			return err
		}
		_, err = gorm.G[entity.QuestionSummary](tx).Where("id = ?", summary.ID).Delete(ctx)
		return err
	})
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB削除失敗", slog.String("summary_uuid", uuid), slog.String("err", err.Error()))
	}
	return err
}

func (r *QuestionRepository) ListIncompleteDueOnDates(ctx context.Context, dates ...time.Time) ([]entity.Question, error) {
	const op = opQuestionRepo + ".ListIncompleteDueOnDates"
	if len(dates) == 0 {
		return nil, nil
	}
	dateStrings := make([]string, 0, len(dates))
	for _, d := range dates {
		y, m, day := d.Date()
		dateStrings = append(dateStrings, fmt.Sprintf("%04d-%02d-%02d", y, int(m), day))
	}
	qs, err := gorm.G[entity.Question](r.db.WithContext(ctx)).
		Preload("QuestionUser", noopPreload).
		Where("support_status IN ?", []valueobject.SupportStatus{
			valueobject.SupportStatusPending,
			valueobject.SupportStatusSupporting,
		}).
		Where("answer_due IS NOT NULL").
		Where("DATE(timezone('Asia/Tokyo', answer_due)) IN ?", dateStrings).
		Order("answer_due ASC, title ASC").
		Find(ctx)
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB検索失敗", slog.String("err", err.Error()))
		return nil, err
	}
	return qs, nil
}
