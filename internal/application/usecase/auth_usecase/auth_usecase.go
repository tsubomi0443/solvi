package auth_usecase

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"solvi/internal/application/converter"
	outputmodel "solvi/internal/application/model/output_model"
	"solvi/internal/domain/entity"
	ext "solvi/internal/domain/interface/external"
	repo "solvi/internal/domain/interface/repository"
	"solvi/internal/shared/auth"
	"solvi/internal/shared/config"
	"solvi/internal/shared/crypto"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthUsecase struct {
	ldapClient ext.LDAPClient
	userRepo   repo.UserRepository
}

const (
	domain = "@cap-net.co.jp"
)

func NewAuthUsecase(ldapClient ext.LDAPClient, userRepo repo.UserRepository) *AuthUsecase {
	return &AuthUsecase{ldapClient: ldapClient, userRepo: userRepo}
}

func (uc *AuthUsecase) LoginLDAP(email, password string) (string, outputmodel.UserOutput, error) {
	if uc.ldapClient == nil {
		return "", outputmodel.UserOutput{}, fmt.Errorf("LDAP認証は設定されていません")
	}
	if strings.HasSuffix(email, domain) {
		email, _ = strings.CutSuffix(email, domain)
	}

	ldapUser, err := uc.ldapClient.Authenticate(email, password)
	if err != nil {
		return "", outputmodel.UserOutput{}, fmt.Errorf("LDAP認証に失敗しました: %w", err)
	}
	user, err := uc.userRepo.GetByEmail(ldapUser.Mail)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", outputmodel.UserOutput{}, err
		}
		user = &entity.User{Name: ldapUser.Name, Email: ldapUser.Mail, DepartmentName: ldapUser.Department}
		if shouldBeSupporter(ldapUser.Department) {
			user.IsSupporter = true
		}
		if err := uc.userRepo.Create(user); err != nil {
			return "", outputmodel.UserOutput{}, err
		}
	} else {
		user.Name = ldapUser.Name
		user.DepartmentName = ldapUser.Department
		if !user.IsSupporterOverridden && shouldBeSupporter(ldapUser.Department) {
			user.IsSupporter = true
		}
		if err := uc.userRepo.Update(user); err != nil {
			return "", outputmodel.UserOutput{}, err
		}
	}
	token, err := issueToken(user)
	if err != nil {
		return "", outputmodel.UserOutput{}, err
	}
	return token, converter.UserEntityToOutput(user), nil
}

func (uc *AuthUsecase) LoginBasic(email, password string) (string, outputmodel.UserOutput, error) {
	user, err := uc.userRepo.GetByEmail(email)
	if err != nil {
		return "", outputmodel.UserOutput{}, fmt.Errorf("認証に失敗しました")
	}
	if user.Password == nil {
		return "", outputmodel.UserOutput{}, fmt.Errorf("認証に失敗しました")
	}
	ok, err := crypto.VerifyPassword(password, config.GetPepper(), *user.Password)
	if err != nil || !ok {
		return "", outputmodel.UserOutput{}, fmt.Errorf("認証に失敗しました")
	}
	token, err := issueToken(user)
	if err != nil {
		return "", outputmodel.UserOutput{}, err
	}
	return token, converter.UserEntityToOutput(user), nil
}

func shouldBeSupporter(department string) bool {
	for _, d := range config.GetSupporterDepartments() {
		if strings.Contains(department, d) {
			return true
		}
	}
	return false
}

func issueToken(user *entity.User) (string, error) {
	claims := auth.CustomClaims{
		UserID:      user.ID,
		UUID:        user.UUID.String(),
		Email:       user.Email,
		Name:        user.Name,
		IsSupporter: user.IsSupporter,
		IsAdmin:     user.IsAdmin(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.GetJWTKey()))
}
