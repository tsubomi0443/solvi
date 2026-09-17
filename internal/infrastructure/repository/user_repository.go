package repository

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"solvi/internal/domain/entity"
	"solvi/internal/shared/config"
	logutils "solvi/internal/shared/logUtils"

	"gorm.io/gorm"
)

const opUserRepo = "UserRepository"

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	const op = opUserRepo + ".GetByID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索", slog.Uint64("user_id", uint64(id)))
	var user entity.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logutils.Debug(ctx, logutils.LayerRepository, op, "レコードなし", slog.Uint64("user_id", uint64(id)))
			return nil, fmt.Errorf("ユーザが見つかりません: %w", err)
		}
		logutils.Error(ctx, logutils.LayerRepository, op, "DB検索失敗", slog.Uint64("user_id", uint64(id)), slog.String("err", err.Error()))
		return nil, fmt.Errorf("ユーザが見つかりません: %w", err)
	}
	if user.Icon != nil {
		icon, err := getIconBlob(&user)
		if err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "アイコン取得失敗", slog.Uint64("user_id", uint64(id)), slog.String("err", err.Error()))
			return nil, fmt.Errorf("アイコン画像の取得に失敗しました: %w", err)
		}
		user.IconBlob = icon
	}
	return &user, nil
}

func (r *UserRepository) GetByUUID(ctx context.Context, uuid string) (*entity.User, error) {
	const op = opUserRepo + ".GetByUUID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索", slog.String("user_uuid", uuid))
	var user entity.User
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logutils.Debug(ctx, logutils.LayerRepository, op, "レコードなし", slog.String("user_uuid", uuid))
			return nil, fmt.Errorf("ユーザが見つかりません: %w", err)
		}
		logutils.Error(ctx, logutils.LayerRepository, op, "DB検索失敗", slog.String("user_uuid", uuid), slog.String("err", err.Error()))
		return nil, fmt.Errorf("ユーザが見つかりません: %w", err)
	}
	if user.Icon != nil {
		icon, err := getIconBlob(&user)
		if err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "アイコン取得失敗", slog.String("user_uuid", uuid), slog.String("err", err.Error()))
			return nil, fmt.Errorf("アイコン画像の取得に失敗しました: %w", err)
		}
		user.IconBlob = icon
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	const op = opUserRepo + ".GetByEmail"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索", slog.String("email", email))
	var user entity.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logutils.Debug(ctx, logutils.LayerRepository, op, "レコードなし", slog.String("email", email))
			return nil, fmt.Errorf("ユーザが見つかりません: %w", err)
		}
		logutils.Error(ctx, logutils.LayerRepository, op, "DB検索失敗", slog.String("email", email), slog.String("err", err.Error()))
		return nil, fmt.Errorf("ユーザが見つかりません: %w", err)
	}
	if user.Icon != nil {
		icon, err := getIconBlob(&user)
		if err != nil {
			logutils.Error(ctx, logutils.LayerRepository, op, "アイコン取得失敗", slog.String("email", email), slog.String("err", err.Error()))
			return nil, fmt.Errorf("アイコン画像の取得に失敗しました: %w", err)
		}
		user.IconBlob = icon
	}
	return &user, nil
}

func (r *UserRepository) CountAdmins(ctx context.Context) (int64, error) {
	const op = opUserRepo + ".CountAdmins"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索")
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Where("is_admin = ?", true).Count(&count).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB検索失敗", slog.String("err", err.Error()))
		return 0, fmt.Errorf("管理者数の取得に失敗しました: %w", err)
	}
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索完了", slog.Int64("count", count))
	return count, nil
}

func (r *UserRepository) ListAll(ctx context.Context) ([]entity.User, error) {
	const op = opUserRepo + ".ListAll"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索")
	var users []entity.User
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&users).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB検索失敗", slog.String("err", err.Error()))
		return nil, fmt.Errorf("ユーザが取得できません: %w", err)
	}
	for _, user := range users {
		if user.Icon != nil {
			icon, err := getIconBlob(&user)
			if err != nil {
				logutils.Error(ctx, logutils.LayerRepository, op, "アイコン取得失敗", slog.Uint64("user_id", uint64(user.ID)), slog.String("err", err.Error()))
				return nil, fmt.Errorf("アイコン画像の取得に失敗しました: %w", err)
			}
			user.IconBlob = icon
		}
	}
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB検索完了", slog.Int("count", len(users)))
	return users, nil
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	const op = opUserRepo + ".Create"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB作成", slog.String("email", user.Email))
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB作成失敗", slog.String("email", user.Email), slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	const op = opUserRepo + ".Update"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB更新", slog.Uint64("user_id", uint64(user.ID)))
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB更新失敗", slog.Uint64("user_id", uint64(user.ID)), slog.String("err", err.Error()))
		return err
	}
	return nil
}

func (r *UserRepository) SoftDeleteByUUID(ctx context.Context, uuid string) error {
	const op = opUserRepo + ".SoftDeleteByUUID"
	logutils.Debug(ctx, logutils.LayerRepository, op, "DB削除", slog.String("user_uuid", uuid))
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&entity.User{}).Error; err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "DB削除失敗", slog.String("user_uuid", uuid), slog.String("err", err.Error()))
		return err
	}
	return nil
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
