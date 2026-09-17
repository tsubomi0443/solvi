package setting_usecase

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"solvi/internal/application/converter"
	"solvi/internal/application/usecase"
	outputmodel "solvi/internal/application/model/output_model"
	repo "solvi/internal/domain/interface/repository"
	logutils "solvi/internal/shared/logUtils"

	"github.com/google/uuid"
)

const opSetting = "setting_usecase"

type SettingUsecase struct {
	userRepo  repo.UserRepository
	uploadDir string
}

func NewSettingUsecase(userRepo repo.UserRepository, uploadDir string) *SettingUsecase {
	return &SettingUsecase{
		userRepo:  userRepo,
		uploadDir: uploadDir,
	}
}

func (uc *SettingUsecase) UpdateProfile(ctx context.Context, userID uint, name, email string) (outputmodel.UserOutput, error) {
	const op = opSetting + ".UpdateProfile"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.Uint64("user_id", uint64(userID)))
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "ユーザ取得失敗", err, slog.Uint64("user_id", uint64(userID)))
		return outputmodel.UserOutput{}, err
	}
	user.Name = name
	user.Email = email
	if err := uc.userRepo.Update(ctx, user); err != nil {
		usecase.LogRepoPropagation(ctx, op, "プロフィール更新失敗", err, slog.Uint64("user_id", uint64(userID)))
		return outputmodel.UserOutput{}, err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "プロフィール更新成功", slog.Uint64("user_id", uint64(userID)))
	return converter.UserEntityToOutput(user), nil
}

func (uc *SettingUsecase) UploadIcon(ctx context.Context, userID uint, filename string, r io.Reader) error {
	const op = opSetting + ".UploadIcon"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.Uint64("user_id", uint64(userID)), slog.String("filename", filename))
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "ユーザ取得失敗", err, slog.Uint64("user_id", uint64(userID)))
		return err
	}
	iconName := uuid.NewString() + filepath.Ext(filename)
	path := filepath.Join(uc.uploadDir, iconName)
	f, err := os.Create(path)
	if err != nil {
		logutils.Error(ctx, logutils.LayerUsecase, op, "アイコンファイル作成失敗", slog.Uint64("user_id", uint64(userID)), slog.String("err", err.Error()))
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		logutils.Error(ctx, logutils.LayerUsecase, op, "アイコンファイル書き込み失敗", slog.Uint64("user_id", uint64(userID)), slog.String("err", err.Error()))
		return err
	}
	user.Icon = &iconName
	if err := uc.userRepo.Update(ctx, user); err != nil {
		usecase.LogRepoPropagation(ctx, op, "アイコン更新失敗", err, slog.Uint64("user_id", uint64(userID)))
		return err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "アイコンアップロード成功", slog.Uint64("user_id", uint64(userID)))
	return nil
}

func (uc *SettingUsecase) DeleteIcon(ctx context.Context, userID uint) error {
	const op = opSetting + ".DeleteIcon"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.Uint64("user_id", uint64(userID)))
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "ユーザ取得失敗", err, slog.Uint64("user_id", uint64(userID)))
		return err
	}
	if user.Icon != nil && strings.TrimSpace(*user.Icon) != "" {
		if err := os.Remove(filepath.Join(uc.uploadDir, *user.Icon)); err != nil {
			logutils.Warn(ctx, logutils.LayerUsecase, op, "アイコンファイル削除失敗", slog.Uint64("user_id", uint64(userID)), slog.String("err", err.Error()))
		}
	}
	user.Icon = nil
	if err := uc.userRepo.Update(ctx, user); err != nil {
		usecase.LogRepoPropagation(ctx, op, "アイコン削除失敗", err, slog.Uint64("user_id", uint64(userID)))
		return err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "アイコン削除成功", slog.Uint64("user_id", uint64(userID)))
	return nil
}

func (uc *SettingUsecase) GetProfile(ctx context.Context, userID uint) (outputmodel.UserOutput, error) {
	const op = opSetting + ".GetProfile"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.Uint64("user_id", uint64(userID)))
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "ユーザ取得失敗", err, slog.Uint64("user_id", uint64(userID)))
		return outputmodel.UserOutput{}, fmt.Errorf("ユーザが見つかりません")
	}
	return converter.UserEntityToOutput(user), nil
}

func (uc *SettingUsecase) GetProfileByUUID(ctx context.Context, uuid string) (outputmodel.UserOutput, error) {
	const op = opSetting + ".GetProfileByUUID"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("user_uuid", uuid))
	user, err := uc.userRepo.GetByUUID(ctx, uuid)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "ユーザ取得失敗", err, slog.String("user_uuid", uuid))
		return outputmodel.UserOutput{}, fmt.Errorf("ユーザが見つかりません")
	}
	return converter.UserEntityToOutput(user), nil
}
