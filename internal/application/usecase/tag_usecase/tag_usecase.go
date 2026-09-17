package tag_usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"solvi/internal/application/usecase"
	outputmodel "solvi/internal/application/model/output_model"
	repo "solvi/internal/domain/interface/repository"
	logutils "solvi/internal/shared/logUtils"
)

const opTag = "tag_usecase"

type TagUsecase struct {
	tagRepo repo.TagRepository
}

func NewTagUsecase(tagRepo repo.TagRepository) *TagUsecase {
	return &TagUsecase{tagRepo: tagRepo}
}

func (uc *TagUsecase) List(ctx context.Context) ([]outputmodel.TagStatOutput, error) {
	const op = opTag + ".List"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始")
	stats, err := uc.tagRepo.ListTagStats(ctx)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "タグ一覧取得失敗", err)
		return nil, err
	}
	out := make([]outputmodel.TagStatOutput, 0, len(stats))
	for _, s := range stats {
		out = append(out, outputmodel.TagStatOutput{Name: s.Name, Count: s.Count})
	}
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理完了", slog.Int("count", len(out)))
	return out, nil
}

func (uc *TagUsecase) Rename(ctx context.Context, from, to string) error {
	const op = opTag + ".Rename"
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("from", from), slog.String("to", to))
	if from == "" || to == "" {
		usecase.LogBusinessWarn(ctx, op, "タグ名未入力", fmt.Errorf("タグ名を入力してください"))
		return fmt.Errorf("タグ名を入力してください")
	}
	if from == to {
		logutils.Debug(ctx, logutils.LayerUsecase, op, "変更なし", slog.String("from", from))
		return nil
	}
	if err := uc.tagRepo.RenameTag(ctx, from, to); err != nil {
		usecase.LogRepoPropagation(ctx, op, "タグ名変更失敗", err, slog.String("from", from), slog.String("to", to))
		return err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "タグ名変更成功", slog.String("from", from), slog.String("to", to))
	return nil
}

func (uc *TagUsecase) Delete(ctx context.Context, name string) error {
	const op = opTag + ".Delete"
	name = strings.TrimSpace(name)
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("name", name))
	if name == "" {
		usecase.LogBusinessWarn(ctx, op, "タグ名未入力", fmt.Errorf("タグ名を入力してください"))
		return fmt.Errorf("タグ名を入力してください")
	}
	if err := uc.tagRepo.DeleteTagByName(ctx, name); err != nil {
		usecase.LogRepoPropagation(ctx, op, "タグ削除失敗", err, slog.String("name", name))
		return err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "タグ削除成功", slog.String("name", name))
	return nil
}
