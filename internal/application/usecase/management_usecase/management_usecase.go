package management_usecase

import (
	"context"
	"fmt"
	"log/slog"

	"solvi/internal/application/converter"
	"solvi/internal/application/usecase"
	"solvi/internal/application/model/output_model"
	repo "solvi/internal/domain/interface/repository"
	"solvi/internal/shared/config"
	logutils "solvi/internal/shared/logUtils"
)

const opManagement = "management_usecase"

type ManagementUsecase struct {
	userRepo repo.UserRepository
}

func NewManagementUsecase(userRepo repo.UserRepository) *ManagementUsecase {
	return &ManagementUsecase{userRepo: userRepo}
}

func (uc *ManagementUsecase) ListUsers(ctx context.Context) ([]outputmodel.UserOutput, error) {
	const op = opManagement + ".ListUsers"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始")
	users, err := uc.userRepo.ListAll(ctx)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "ユーザ一覧取得失敗", err)
		return nil, err
	}
	sysEmail := config.GetSystemUserEmail()
	out := make([]outputmodel.UserOutput, 0, len(users))
	for _, u := range users {
		if u.Email == sysEmail {
			continue
		}
		out = append(out, converter.UserEntityToOutput(&u))
	}
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理完了", slog.Int("count", len(out)))
	return out, nil
}

type UserUpdateInput struct {
	IsSupporter *bool
	IsAdmin     *bool
}

func (uc *ManagementUsecase) UpdateUser(ctx context.Context, targetUUID, actorUUID string, input UserUpdateInput) error {
	const op = opManagement + ".UpdateUser"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("target_uuid", targetUUID), slog.String("actor_uuid", actorUUID))
	if input.IsSupporter == nil && input.IsAdmin == nil {
		usecase.LogBusinessWarn(ctx, op, "更新項目なし", fmt.Errorf("更新項目がありません"), slog.String("target_uuid", targetUUID))
		return fmt.Errorf("更新項目がありません")
	}

	user, err := uc.userRepo.GetByUUID(ctx, targetUUID)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "ユーザ取得失敗", err, slog.String("target_uuid", targetUUID))
		return err
	}

	if input.IsAdmin != nil {
		if targetUUID == actorUUID && user.IsAdmin && !*input.IsAdmin {
			usecase.LogBusinessWarn(ctx, op, "管理者権限解除不可", fmt.Errorf("自分自身の管理者権限は解除できません"), slog.String("target_uuid", targetUUID))
			return fmt.Errorf("自分自身の管理者権限は解除できません")
		}
		if user.IsAdmin && !*input.IsAdmin {
			count, err := uc.userRepo.CountAdmins(ctx)
			if err != nil {
				usecase.LogRepoPropagation(ctx, op, "管理者数取得失敗", err)
				return err
			}
			if count <= 1 {
				usecase.LogBusinessWarn(ctx, op, "最後の管理者解除不可", fmt.Errorf("最後の管理者は解除できません"), slog.String("target_uuid", targetUUID))
				return fmt.Errorf("最後の管理者は解除できません")
			}
		}
		user.IsAdmin = *input.IsAdmin
	}

	if input.IsSupporter != nil {
		user.IsSupporter = *input.IsSupporter
		user.IsSupporterOverridden = true
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		usecase.LogRepoPropagation(ctx, op, "ユーザ更新失敗", err, slog.String("target_uuid", targetUUID))
		return err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "ユーザ更新成功", slog.String("target_uuid", targetUUID), slog.Uint64("user_id", uint64(user.ID)))
	return nil
}
