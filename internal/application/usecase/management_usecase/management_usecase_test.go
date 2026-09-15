package management_usecase_test

import (
	"os"
	"testing"

	mnguc "solvi/internal/application/usecase/management_usecase"
	"solvi/internal/domain/entity"
	repomock "solvi/internal/domain/interface/repository/mock"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	os.Setenv("SYSTEM_USER_EMAIL", "system@solvi.local")
	os.Exit(m.Run())
}

func TestListUsers_IncludesAdminExcludesSystem(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)

	adminUUID := uuid.New()
	userRepo.EXPECT().ListAll().Return([]entity.User{
		{Name: "System", Email: "system@solvi.local"},
		{Name: "Admin", Email: "admin@solvi.local", UUID: adminUUID, IsAdmin: true},
		{Name: "User", Email: "user@example.com", UUID: uuid.New()},
	}, nil)

	uc := mnguc.NewManagementUsecase(userRepo)
	users, err := uc.ListUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Email == "system@solvi.local" || users[1].Email == "system@solvi.local" {
		t.Fatal("system user should be excluded")
	}
}

func TestUpdateUser_PreventsSelfDemotion(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)

	targetUUID := uuid.New().String()
	userRepo.EXPECT().GetByUUID(targetUUID).Return(&entity.User{
		UUID: uuid.MustParse(targetUUID), IsAdmin: true,
	}, nil)

	uc := mnguc.NewManagementUsecase(userRepo)
	err := uc.UpdateUser(targetUUID, targetUUID, mnguc.UserUpdateInput{
		IsAdmin: boolPtr(false),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateUser_PreventsRemovingLastAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)

	targetUUID := uuid.New().String()
	userRepo.EXPECT().GetByUUID(targetUUID).Return(&entity.User{
		UUID: uuid.MustParse(targetUUID), IsAdmin: true,
	}, nil)
	userRepo.EXPECT().CountAdmins().Return(1, nil)

	uc := mnguc.NewManagementUsecase(userRepo)
	err := uc.UpdateUser(targetUUID, uuid.New().String(), mnguc.UserUpdateInput{
		IsAdmin: boolPtr(false),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateUser_UpdatesFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)

	targetUUID := uuid.New().String()
	user := &entity.User{UUID: uuid.MustParse(targetUUID), IsAdmin: false, IsSupporter: false}
	userRepo.EXPECT().GetByUUID(targetUUID).Return(user, nil)
	userRepo.EXPECT().Update(gomock.Any()).DoAndReturn(func(u *entity.User) error {
		if !u.IsAdmin || !u.IsSupporter || !u.IsSupporterOverridden {
			t.Fatalf("unexpected user: %+v", u)
		}
		return nil
	})

	uc := mnguc.NewManagementUsecase(userRepo)
	if err := uc.UpdateUser(targetUUID, uuid.New().String(), mnguc.UserUpdateInput{
		IsAdmin:     boolPtr(true),
		IsSupporter: boolPtr(true),
	}); err != nil {
		t.Fatal(err)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
