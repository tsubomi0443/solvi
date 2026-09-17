package repository

import (
	"context"

	"solvi/internal/domain/entity"
)

//go:generate go tool mockgen -typed -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*entity.User, error)
	GetByUUID(ctx context.Context, uuid string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	ListAll(ctx context.Context) ([]entity.User, error)
	CountAdmins(ctx context.Context) (int64, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	SoftDeleteByUUID(ctx context.Context, uuid string) error
}
