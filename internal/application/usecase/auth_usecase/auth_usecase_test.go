package auth_usecase_test

import (
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
	userRepo.EXPECT().GetByEmail("admin@solvi.local").Return(user, nil)

	uc := authuc.NewAuthUsecase(nil, userRepo)
	token, out, err := uc.LoginBasic("admin@solvi.local", "admin")
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
	userRepo.EXPECT().GetByEmail("user@solvi.local").Return(user, nil)

	uc := authuc.NewAuthUsecase(nil, userRepo)
	_, out, err := uc.LoginBasic("user@solvi.local", "admin")
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
	userRepo.EXPECT().GetByEmail("admin@solvi.local").Return(user, nil)

	uc := authuc.NewAuthUsecase(nil, userRepo)
	_, _, err := uc.LoginBasic("admin@solvi.local", "wrong")
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
	userRepo.EXPECT().GetByEmail("user@example.com").Return(nil, gorm.ErrRecordNotFound)
	userRepo.EXPECT().Create(gomock.Any()).DoAndReturn(func(u *entity.User) error {
		u.ID = 2
		if !u.IsSupporter {
			t.Fatal("expected supporter from department")
		}
		return nil
	})

	uc := authuc.NewAuthUsecase(ldapClient, userRepo)
	token, out, err := uc.LoginLDAP("user@example.com", "pass")
	if err != nil {
		t.Fatalf("LoginLDAP: %v", err)
	}
	if token == "" || !out.IsSupporter {
		t.Fatalf("unexpected: token=%q out=%+v", token, out)
	}
}
