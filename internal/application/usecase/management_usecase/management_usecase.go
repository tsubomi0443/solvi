package management_usecase

import (
	"fmt"

	"solvi/internal/application/converter"
	"solvi/internal/application/model/output_model"
	repo "solvi/internal/domain/interface/repository"
	"solvi/internal/shared/config"
)

type ManagementUsecase struct {
	userRepo repo.UserRepository
}

func NewManagementUsecase(userRepo repo.UserRepository) *ManagementUsecase {
	return &ManagementUsecase{userRepo: userRepo}
}

func (uc *ManagementUsecase) ListUsers() ([]outputmodel.UserOutput, error) {
	users, err := uc.userRepo.ListAll()
	if err != nil {
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
	return out, nil
}

type UserUpdateInput struct {
	IsSupporter *bool
	IsAdmin     *bool
}

func (uc *ManagementUsecase) UpdateUser(targetUUID, actorUUID string, input UserUpdateInput) error {
	if input.IsSupporter == nil && input.IsAdmin == nil {
		return fmt.Errorf("更新項目がありません")
	}

	user, err := uc.userRepo.GetByUUID(targetUUID)
	if err != nil {
		return err
	}

	if input.IsAdmin != nil {
		if targetUUID == actorUUID && user.IsAdmin && !*input.IsAdmin {
			return fmt.Errorf("自分自身の管理者権限は解除できません")
		}
		if user.IsAdmin && !*input.IsAdmin {
			count, err := uc.userRepo.CountAdmins()
			if err != nil {
				return err
			}
			if count <= 1 {
				return fmt.Errorf("最後の管理者は解除できません")
			}
		}
		user.IsAdmin = *input.IsAdmin
	}

	if input.IsSupporter != nil {
		user.IsSupporter = *input.IsSupporter
		user.IsSupporterOverridden = true
	}

	return uc.userRepo.Update(user)
}
