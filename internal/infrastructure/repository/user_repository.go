package repository

import (
	"fmt"
	"solvi/internal/domain/entity"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(id uint) (*entity.User, error) {
	var user entity.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("ユーザが見つかりません: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByUUID(uuid string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		return nil, fmt.Errorf("ユーザが見つかりません: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ListAll() ([]entity.User, error) {
	var users []entity.User
	if err := r.db.Order("id ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Update(user *entity.User) error {
	return r.db.Save(user).Error
}
