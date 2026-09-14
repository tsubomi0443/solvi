package converter

import (
	"solvi/internal/application/model/output_model"
	"solvi/internal/domain/entity"
)

func UserEntityToOutput(u *entity.User) outputmodel.UserOutput {
	icon := ""
	if u.Icon != nil {
		icon = *u.Icon
	}
	return outputmodel.UserOutput{
		UUID:           u.UUID.String(),
		Name:           u.Name,
		Email:          u.Email,
		DepartmentName: u.DepartmentName,
		Icon:           icon,
		IsSupporter:    u.IsSupporter,
		IsAdmin:        u.IsAdmin(),
	}
}
