package repository

import (
	"solvi/internal/domain/entity"
)

//go:generate go tool mockgen -typed -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

type UserRepository interface {
	GetByID(id uint) (*entity.User, error)
	GetByUUID(uuid string) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	ListAll() ([]entity.User, error)
	CountAdmins() (int64, error)
	Create(user *entity.User) error
	Update(user *entity.User) error
}
