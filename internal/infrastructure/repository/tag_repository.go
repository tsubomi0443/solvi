package repository

import (
	repo "solvi/internal/domain/interface/repository"
	"solvi/internal/domain/entity"

	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) ListTagStats() ([]repo.TagStat, error) {
	var rows []repo.TagStat
	err := r.db.Model(&entity.QuestionTag{}).
		Select("name, count(*) as count").
		Group("name").
		Order("name ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *TagRepository) RenameTag(from, to string) error {
	return r.db.Model(&entity.QuestionTag{}).Where("name = ?", from).Update("name", to).Error
}

func (r *TagRepository) DeleteTagByName(name string) error {
	return r.db.Where("name = ?", name).Delete(&entity.QuestionTag{}).Error
}
