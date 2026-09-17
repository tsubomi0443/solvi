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

	oldIcon := ""
	if user.Icon != nil {
		oldIcon = *user.Icon
	}

	iconName := uuid.NewString() + filepath.Ext(filename)
	path := filepath.Join(uc.uploadDir, iconName)
	f, err := os.Create(path)
	if err != nil {
		logutils.Error(ctx, logutils.LayerUsecase, op, "アイコンファイル作成失敗", slog.Uint64("user_id", uint64(userID)), slog.String("err", err.Error()))
		return err
	}

	var writeErr error
	if _, err := io.Copy(f, r); err != nil {
		writeErr = err
	}
	if closeErr := f.Close(); closeErr != nil && writeErr == nil {
		writeErr = closeErr
	}

	if writeErr != nil {
		logutils.Error(ctx, logutils.LayerUsecase, op, "アイコンファイル書き込み失敗", slog.Uint64("user_id", uint64(userID)), slog.String("err", writeErr.Error()))
		_ = uc.safeRemoveIconFile(ctx, iconName)
		return writeErr
	}

	user.Icon = &iconName
	if err := uc.userRepo.Update(ctx, user); err != nil {
		usecase.LogRepoPropagation(ctx, op, "アイコン更新失敗", err, slog.Uint64("user_id", uint64(userID)))
		_ = uc.safeRemoveIconFile(ctx, iconName)
		return err
	}

	if oldIcon != "" && oldIcon != iconName {
		if err := uc.safeRemoveIconFile(ctx, oldIcon); err != nil {
			logutils.Error(ctx, logutils.LayerUsecase, op, "旧アイコンファイル削除失敗", slog.Uint64("user_id", uint64(userID)), slog.String("old_icon", oldIcon), slog.String("err", err.Error()))
			return err
		}
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
		if err := uc.safeRemoveIconFile(ctx, *user.Icon); err != nil {
			logutils.Error(ctx, logutils.LayerUsecase, op, "アイコンファイル削除失敗", slog.Uint64("user_id", uint64(userID)), slog.String("icon", *user.Icon), slog.String("err", err.Error()))
			return err
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

func (uc *SettingUsecase) safeRemoveIconFile(ctx context.Context, iconName string) error {
	const op = opSetting + ".safeRemoveIconFile"
	trimmed := strings.TrimSpace(iconName)
	if trimmed == "" {
		return nil
	}
	base := filepath.Base(trimmed)
	if base == "." || base == ".." || base != trimmed {
		err := fmt.Errorf("不正なアイコンファイル名です: %s", iconName)
		logutils.Error(ctx, logutils.LayerUsecase, op, "パストラバーサル検知", slog.String("icon_name", iconName), slog.String("err", err.Error()))
		return err
	}

	cleanUploadDir := filepath.Clean(uc.uploadDir)
	targetPath := filepath.Clean(filepath.Join(cleanUploadDir, base))

	rel, err := filepath.Rel(cleanUploadDir, targetPath)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		err := fmt.Errorf("無効なアイコンパスです: %s", iconName)
		logutils.Error(ctx, logutils.LayerUsecase, op, "不正なパス", slog.String("icon_name", iconName), slog.String("err", err.Error()))
		return err
	}

	info, err := os.Lstat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			logutils.Debug(ctx, logutils.LayerUsecase, op, "削除対象ファイルが存在しないためスキップ", slog.String("path", targetPath))
			return nil
		}
		logutils.Error(ctx, logutils.LayerUsecase, op, "アイコンファイル状態確認失敗", slog.String("path", targetPath), slog.String("err", err.Error()))
		return fmt.Errorf("アイコンファイル状態確認失敗: %w", err)
	}

	if !info.Mode().IsRegular() {
		err := fmt.Errorf("削除対象が通常ファイルではありません: %s (mode: %v)", targetPath, info.Mode())
		logutils.Error(ctx, logutils.LayerUsecase, op, "通常ファイル外検出", slog.String("path", targetPath), slog.String("err", err.Error()))
		return err
	}

	if err := os.Remove(targetPath); err != nil {
		if os.IsNotExist(err) {
			logutils.Debug(ctx, logutils.LayerUsecase, op, "削除時にファイルが存在しなくなっていたためスキップ", slog.String("path", targetPath))
			return nil
		}
		logutils.Error(ctx, logutils.LayerUsecase, op, "アイコンファイル削除失敗", slog.String("path", targetPath), slog.String("err", err.Error()))
		return fmt.Errorf("アイコンファイル削除失敗: %w", err)
	}

	logutils.Info(ctx, logutils.LayerUsecase, op, "アイコンファイル削除完了", slog.String("path", targetPath))
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
