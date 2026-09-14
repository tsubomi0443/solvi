package tag_usecase

import (
	"fmt"
	"strings"

	outputmodel "solvi/internal/application/model/output_model"
	repo "solvi/internal/domain/interface/repository"
)

type TagUsecase struct {
	tagRepo repo.TagRepository
}

func NewTagUsecase(tagRepo repo.TagRepository) *TagUsecase {
	return &TagUsecase{tagRepo: tagRepo}
}

func (uc *TagUsecase) List() ([]outputmodel.TagStatOutput, error) {
	stats, err := uc.tagRepo.ListTagStats()
	if err != nil {
		return nil, err
	}
	out := make([]outputmodel.TagStatOutput, 0, len(stats))
	for _, s := range stats {
		out = append(out, outputmodel.TagStatOutput{Name: s.Name, Count: s.Count})
	}
	return out, nil
}

func (uc *TagUsecase) Rename(from, to string) error {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if from == "" || to == "" {
		return fmt.Errorf("タグ名を入力してください")
	}
	if from == to {
		return nil
	}
	return uc.tagRepo.RenameTag(from, to)
}

func (uc *TagUsecase) Delete(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("タグ名を入力してください")
	}
	return uc.tagRepo.DeleteTagByName(name)
}
