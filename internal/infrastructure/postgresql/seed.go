package postgresql

import (
	"log/slog"
	"solvi/internal/domain/entity"
	"solvi/internal/shared/config"
	"solvi/internal/shared/crypto"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	email := config.GetSystemUserEmail()
	var count int64
	if err := db.Model(&entity.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		sys := entity.User{Name: "System", Email: email, IsSupporter: false}
		if err := db.Create(&sys).Error; err != nil {
			return err
		}
		slog.Info("seeded system user", "email", email)
	}

	adminEmail := "admin@solvi.local"
	if err := db.Model(&entity.User{}).Where("email = ?", adminEmail).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		hash, err := crypto.HashPassword("admin", config.GetPepper())
		if err != nil {
			return err
		}
		admin := entity.User{
			Name:                  "Administrator",
			Email:                 adminEmail,
			Password:              &hash,
			IsSupporter:           true,
			IsSupporterOverridden: true,
			IsAdmin:               true,
		}
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
		slog.Info("seeded admin user", "email", adminEmail)
	}

	if err := db.Model(&entity.User{}).
		Where("email = ? AND is_admin = ?", adminEmail, false).
		Update("is_admin", true).Error; err != nil {
		return err
	}
	if err := db.Model(&entity.User{}).
		Where("password IS NOT NULL AND password <> '' AND is_admin = ?", false).
		Update("is_admin", true).Error; err != nil {
		return err
	}

	return nil
}
