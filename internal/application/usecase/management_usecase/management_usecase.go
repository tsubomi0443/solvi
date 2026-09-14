package management_usecase

import (
	"fmt"

	"solvi/internal/application/converter"
	"solvi/internal/application/model/output_model"
	repo "solvi/internal/domain/interface/repository"
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
	out := make([]outputmodel.UserOutput, 0, len(users))
	for _, u := range users {
		if u.IsAdmin() {
			continue
		}
		out = append(out, converter.UserEntityToOutput(&u))
	}
	return out, nil
}

func (uc *ManagementUsecase) UpdateIsSupporter(uuid string, isSupporter bool) error {
	user, err := uc.userRepo.GetByUUID(uuid)
	if err != nil {
		return err
	}
	if user.IsAdmin() {
		return fmt.Errorf("管理者は変更できません")
	}
	user.IsSupporter = isSupporter
	user.IsSupporterOverridden = true
	return uc.userRepo.Update(user)
}
