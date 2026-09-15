package repository

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"solvi/internal/domain/entity"
	"solvi/internal/shared/config"

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
	if user.Icon != nil {
		icon, err := getIconBlob(&user)
		if err != nil {
			return nil, fmt.Errorf("アイコン画像の取得に失敗しました: %w", err)
		}
		user.IconBlob = icon
	}

	return &user, nil
}

func (r *UserRepository) GetByUUID(uuid string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		return nil, fmt.Errorf("ユーザが見つかりません: %w", err)
	}
	if user.Icon != nil {
		icon, err := getIconBlob(&user)
		if err != nil {
			return nil, fmt.Errorf("アイコン画像の取得に失敗しました: %w", err)
		}
		user.IconBlob = icon
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("ユーザが見つかりません: %w", err)
	}
	if user.Icon != nil {
		icon, err := getIconBlob(&user)
		if err != nil {
			return nil, fmt.Errorf("アイコン画像の取得に失敗しました: %w", err)
		}
		user.IconBlob = icon
	}
	return &user, nil
}

func (r *UserRepository) CountAdmins() (int64, error) {
	var count int64
	if err := r.db.Model(&entity.User{}).Where("is_admin = ?", true).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("管理者数の取得に失敗しました: %w", err)
	}
	return count, nil
}

func (r *UserRepository) ListAll() ([]entity.User, error) {
	var users []entity.User
	if err := r.db.Order("id ASC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("ユーザが取得できません: %w", err)
	}
	for _, user := range users {
		if user.Icon != nil {
			icon, err := getIconBlob(&user)
			if err != nil {
				return nil, fmt.Errorf("アイコン画像の取得に失敗しました: %w", err)
			}
			user.IconBlob = icon
		}
	}
	return users, nil
}

func (r *UserRepository) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Update(user *entity.User) error {
	return r.db.Save(user).Error
}

func getIconBlob(user *entity.User) ([]byte, error) {
	iconPath := filepath.Join(config.GetUploadDir(), *user.Icon)
	iconFile, err := os.Open(iconPath)
	if err != nil {
		return nil, fmt.Errorf("プロフィールアイコン画像の取得に失敗しました: %w", err)
	}
	defer iconFile.Close()

	var blob bytes.Buffer
	if _, err := io.Copy(&blob, iconFile); err != nil {
		return nil, fmt.Errorf("アイコン画像のデータ読み取りに失敗しました: %w", err)
	}

	return blob.Bytes(), nil
}
