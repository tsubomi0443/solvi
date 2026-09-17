package repository

import (
	"context"
	"log/slog"
	repo "solvi/internal/domain/interface/repository"
	"solvi/internal/domain/entity"
	logutils "solvi/internal/shared/logUtils"

	"gorm.io/gorm"
)

const opTagRepo = "TagRepository"

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) ListTagStats(ctx context.Context) ([]repo.TagStat, error) {
	const op = opTagRepo + ".ListTagStats"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索")
	var rows []repo.TagStat
	err := r.db.WithContext(ctx).Model(&entity.QuestionTag{}).
		Select("name, count(*) as count").
		Group("name").
		Order("name ASC").
		Scan(&rows).Error
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB検索失敗", slog.String("err", err.Error()))
		return nil, err
	}
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索完了", slog.Int("count", len(rows)))
	return rows, err
}

func (r *TagRepository) RenameTag(ctx context.Context, from, to string) error {
	const op = opTagRepo + ".RenameTag"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB更新", slog.String("from", from), slog.String("to", to))
	if err := r.db.WithContext(ctx).Model(&entity.QuestionTag{}).Where("name = ?", from).Update("name", to).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", slog.String("from", from), slog.String("to", to), slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *TagRepository) DeleteTagByName(ctx context.Context, name string) error {
	const op = opTagRepo + ".DeleteTagByName"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB削除", slog.String("name", name))
	if err := r.db.WithContext(ctx).Where("name = ?", name).Delete(&entity.QuestionTag{}).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB削除失敗", slog.String("name", name), slog.String("err", err.Error()))
		return err
	}
	return nil
}
