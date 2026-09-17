package auth_usecase_test

import (
	"context"
	"os"
	"testing"

	authuc "solvi/internal/application/usecase/auth_usecase"
	ldap_entity "solvi/internal/domain/entity/ldap"
	extmock "solvi/internal/domain/interface/external/mock"
	repomock "solvi/internal/domain/interface/repository/mock"
	"solvi/internal/domain/entity"
	"solvi/internal/shared/crypto"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	os.Setenv("JWK_KEY", "test-jwt-key")
	os.Setenv("PEPPER", "test-pepper")
	os.Exit(m.Run())
}

func TestLoginBasic_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)
	hash, err := crypto.HashPassword("admin", "test-pepper")
	if err != nil {
		t.Fatal(err)
	}
	user := &entity.User{Model: gorm.Model{ID: 1}, UUID: uuid.New(), Name: "Admin", Email: "admin@solvi.local", Password: &hash, IsSupporter: true, IsAdmin: true}
	userRepo.EXPECT().GetByEmail(gomock.Any(), "admin@solvi.local").Return(user, nil)

	uc := authuc.NewAuthUsecase(nil, userRepo)
	token, out, err := uc.LoginBasic(context.Background(), "admin@solvi.local", "admin")
	if err != nil {
		t.Fatalf("LoginBasic: %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
	if out.Email != "admin@solvi.local" || !out.IsAdmin {
		t.Fatalf("unexpected user output: %+v", out)
	}
}

func TestLoginBasic_PasswordOnlyDoesNotGrantAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)
	hash, err := crypto.HashPassword("admin", "test-pepper")
	if err != nil {
		t.Fatal(err)
	}
	user := &entity.User{Model: gorm.Model{ID: 1}, UUID: uuid.New(), Name: "User", Email: "user@solvi.local", Password: &hash, IsAdmin: false}
	userRepo.EXPECT().GetByEmail(gomock.Any(), "user@solvi.local").Return(user, nil)

	uc := authuc.NewAuthUsecase(nil, userRepo)
	_, out, err := uc.LoginBasic(context.Background(), "user@solvi.local", "admin")
	if err != nil {
		t.Fatalf("LoginBasic: %v", err)
	}
	if out.IsAdmin {
		t.Fatalf("expected non-admin output: %+v", out)
	}
}

func TestLoginBasic_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)
	hash, _ := crypto.HashPassword("admin", "test-pepper")
	user := &entity.User{Email: "admin@solvi.local", Password: &hash}
	userRepo.EXPECT().GetByEmail(gomock.Any(), "admin@solvi.local").Return(user, nil)

	uc := authuc.NewAuthUsecase(nil, userRepo)
	_, _, err := uc.LoginBasic(context.Background(), "admin@solvi.local", "wrong")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoginLDAP_CreatesSupporter(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)
	ldapClient := extmock.NewMockLDAPClient(ctrl)

	ldapClient.EXPECT().Authenticate("user@example.com", "pass").Return(&ldap_entity.LDAPUser{
		Name: "User", Mail: "user@example.com", Department: "総務部",
	}, nil)
	userRepo.EXPECT().GetByEmail(gomock.Any(), "user@example.com").Return(nil, gorm.ErrRecordNotFound)
	userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *entity.User) error {
		u.ID = 2
		if !u.IsSupporter {
			t.Fatal("expected supporter from department")
		}
		return nil
	})

	uc := authuc.NewAuthUsecase(ldapClient, userRepo)
	token, out, err := uc.LoginLDAP(context.Background(), "user@example.com", "pass")
	if err != nil {
		t.Fatalf("LoginLDAP: %v", err)
	}
	if token == "" || !out.IsSupporter {
		t.Fatalf("unexpected: token=%q out=%+v", token, out)
	}
}

func TestLoginLDAP_NilClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)

	uc := authuc.NewAuthUsecase(nil, userRepo)
	_, _, err := uc.LoginLDAP(context.Background(), "user@cap-net.co.jp", "pass")
	if err == nil {
		t.Fatal("expected error when ldapClient is nil")
	}
}

func TestLoginLDAP_StripsDomainAndHandlesAuthFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)
	ldapClient := extmock.NewMockLDAPClient(ctrl)

	ldapClient.EXPECT().Authenticate("taro", "pass").Return(nil, gorm.ErrInvalidData)

	uc := authuc.NewAuthUsecase(ldapClient, userRepo)
	_, _, err := uc.LoginLDAP(context.Background(), "taro@cap-net.co.jp", "pass")
	if err == nil {
		t.Fatal("expected error when authenticate fails")
	}
}

func TestLoginLDAP_CreatesNonSupporterWhenDepartmentMismatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)
	ldapClient := extmock.NewMockLDAPClient(ctrl)

	ldapClient.EXPECT().Authenticate("sales@example.com", "pass").Return(&ldap_entity.LDAPUser{
		Name: "Sales User", Mail: "sales@example.com", Department: "営業部",
	}, nil)
	userRepo.EXPECT().GetByEmail(gomock.Any(), "sales@example.com").Return(nil, gorm.ErrRecordNotFound)
	userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *entity.User) error {
		u.ID = 10
		if u.IsSupporter {
			t.Fatal("expected non-supporter")
		}
		return nil
	})

	uc := authuc.NewAuthUsecase(ldapClient, userRepo)
	token, out, err := uc.LoginLDAP(context.Background(), "sales@example.com", "pass")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || out.IsSupporter {
		t.Fatalf("unexpected output: token=%q out=%+v", token, out)
	}
}

func TestLoginLDAP_UpdatesExistingUserSupporterRules(t *testing.T) {
	t.Run("supporter updated when not overridden", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		userRepo := repomock.NewMockUserRepository(ctrl)
		ldapClient := extmock.NewMockLDAPClient(ctrl)

		ldapClient.EXPECT().Authenticate("user@example.com", "pass").Return(&ldap_entity.LDAPUser{
			Name: "User Updated", Mail: "user@example.com", Department: "総務部",
		}, nil)
		existing := &entity.User{
			Model:                 gorm.Model{ID: 3},
			Name:                  "Old Name",
			Email:                 "user@example.com",
			DepartmentName:        "旧部署",
			IsSupporter:           false,
			IsSupporterOverridden: false,
		}
		userRepo.EXPECT().GetByEmail(gomock.Any(), "user@example.com").Return(existing, nil)
		userRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *entity.User) error {
			if !u.IsSupporter || u.Name != "User Updated" {
				t.Fatalf("expected supporter update, got %+v", u)
			}
			return nil
		})

		uc := authuc.NewAuthUsecase(ldapClient, userRepo)
		token, out, err := uc.LoginLDAP(context.Background(), "user@example.com", "pass")
		if err != nil {
			t.Fatal(err)
		}
		if token == "" || !out.IsSupporter {
			t.Fatalf("unexpected out: %+v", out)
		}
	})

	t.Run("supporter retained when overridden", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		userRepo := repomock.NewMockUserRepository(ctrl)
		ldapClient := extmock.NewMockLDAPClient(ctrl)

		ldapClient.EXPECT().Authenticate("user2@example.com", "pass").Return(&ldap_entity.LDAPUser{
			Name: "User2", Mail: "user2@example.com", Department: "総務部",
		}, nil)
		existing := &entity.User{
			Model:                 gorm.Model{ID: 4},
			Name:                  "User2",
			Email:                 "user2@example.com",
			DepartmentName:        "旧部署",
			IsSupporter:           false,
			IsSupporterOverridden: true,
		}
		userRepo.EXPECT().GetByEmail(gomock.Any(), "user2@example.com").Return(existing, nil)
		userRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *entity.User) error {
			if u.IsSupporter {
				t.Fatalf("overridden supporter should remain false, got %+v", u)
			}
			return nil
		})

		uc := authuc.NewAuthUsecase(ldapClient, userRepo)
		_, out, err := uc.LoginLDAP(context.Background(), "user2@example.com", "pass")
		if err != nil {
			t.Fatal(err)
		}
		if out.IsSupporter {
			t.Fatalf("expected non-supporter due to override: %+v", out)
		}
	})
}

func TestLoginBasic_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)

	userRepo.EXPECT().GetByEmail(gomock.Any(), "missing@solvi.local").Return(nil, gorm.ErrRecordNotFound)

	uc := authuc.NewAuthUsecase(nil, userRepo)
	_, _, err := uc.LoginBasic(context.Background(), "missing@solvi.local", "secret")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoginBasic_NilPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)

	user := &entity.User{Email: "nopass@solvi.local", Password: nil}
	userRepo.EXPECT().GetByEmail(gomock.Any(), "nopass@solvi.local").Return(user, nil)

	uc := authuc.NewAuthUsecase(nil, userRepo)
	_, _, err := uc.LoginBasic(context.Background(), "nopass@solvi.local", "secret")
	if err == nil {
		t.Fatal("expected error")
	}
}
