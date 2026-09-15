package setting_usecase

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"solvi/internal/application/converter"
	outputmodel "solvi/internal/application/model/output_model"
	repo "solvi/internal/domain/interface/repository"

	"github.com/google/uuid"
)

type SettingUsecase struct {
	userRepo  repo.UserRepository
	uploadDir string
}

func NewSettingUsecase(userRepo repo.UserRepository, uploadDir string) *SettingUsecase {
	return &SettingUsecase{userRepo: userRepo, uploadDir: uploadDir}
}

func (uc *SettingUsecase) UpdateProfile(userID uint, name, email string) (outputmodel.UserOutput, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return outputmodel.UserOutput{}, err
	}
	user.Name = name
	user.Email = email
	if err := uc.userRepo.Update(user); err != nil {
		return outputmodel.UserOutput{}, err
	}
	return converter.UserEntityToOutput(user), nil
}

func (uc *SettingUsecase) UploadIcon(userID uint, filename string, r io.Reader) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	iconName := uuid.NewString() + filepath.Ext(filename)
	path := filepath.Join(uc.uploadDir, iconName)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	user.Icon = &iconName
	return uc.userRepo.Update(user)
}

func (uc *SettingUsecase) DeleteIcon(userID uint) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user.Icon != nil && strings.TrimSpace(*user.Icon) != "" {
		_ = os.Remove(filepath.Join(uc.uploadDir, *user.Icon))
	}
	user.Icon = nil
	return uc.userRepo.Update(user)
}

func (uc *SettingUsecase) GetProfile(userID uint) (outputmodel.UserOutput, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return outputmodel.UserOutput{}, fmt.Errorf("ユーザが見つかりません")
	}
	return converter.UserEntityToOutput(user), nil
}

func (uc *SettingUsecase) GetProfileByUUID(uuid string) (outputmodel.UserOutput, error) {
	user, err := uc.userRepo.GetByUUID(uuid)
	if err != nil {
		return outputmodel.UserOutput{}, fmt.Errorf("ユーザが見つかりません")
	}
	return converter.UserEntityToOutput(user), nil
}
