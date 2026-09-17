package auth_usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"solvi/internal/application/converter"
	"solvi/internal/application/usecase"
	outputmodel "solvi/internal/application/model/output_model"
	"solvi/internal/domain/entity"
	ext "solvi/internal/domain/interface/external"
	repo "solvi/internal/domain/interface/repository"
	"solvi/internal/shared/auth"
	"solvi/internal/shared/config"
	"solvi/internal/shared/crypto"
	logutils "solvi/internal/shared/logUtils"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthUsecase struct {
	ldapClient ext.LDAPClient
	userRepo   repo.UserRepository
}

const (
	domain = "@cap-net.co.jp"
	opAuth = "auth_usecase"
)

func NewAuthUsecase(ldapClient ext.LDAPClient, userRepo repo.UserRepository) *AuthUsecase {
	return &AuthUsecase{ldapClient: ldapClient, userRepo: userRepo}
}

func (uc *AuthUsecase) LoginLDAP(ctx context.Context, email, password string) (string, outputmodel.UserOutput, error) {
	const op = opAuth + ".LoginLDAP"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("email", email))
	if uc.ldapClient == nil {
		logutils.Warn(ctx, logutils.LayerUsecase, op, "LDAP未設定", slog.String("email", email))
		return "", outputmodel.UserOutput{}, fmt.Errorf("LDAP認証は設定されていません")
	}
	if strings.HasSuffix(email, domain) {
		email, _ = strings.CutSuffix(email, domain)
	}

	ldapUser, err := uc.ldapClient.Authenticate(email, password)
	if err != nil {
		logutils.Error(ctx, logutils.LayerUsecase, op, "LDAP認証失敗", slog.String("email", email), slog.String("err", err.Error()))
		return "", outputmodel.UserOutput{}, fmt.Errorf("LDAP認証に失敗しました: %w", err)
	}
	user, err := uc.userRepo.GetByEmail(ctx, ldapUser.Mail)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			usecase.LogRepoPropagation(ctx, op, "ユーザ取得失敗", err, slog.String("email", ldapUser.Mail))
			return "", outputmodel.UserOutput{}, err
		}
		logutils.Debug(ctx, logutils.LayerUsecase, op, "新規ユーザ作成", slog.String("email", ldapUser.Mail))
		user = &entity.User{Name: ldapUser.Name, Email: ldapUser.Mail, DepartmentName: ldapUser.Department}
		if shouldBeSupporter(ldapUser.Department) {
			user.IsSupporter = true
		}
		if err := uc.userRepo.Create(ctx, user); err != nil {
			usecase.LogRepoPropagation(ctx, op, "ユーザ作成失敗", err, slog.String("email", ldapUser.Mail))
			return "", outputmodel.UserOutput{}, err
		}
	} else {
		logutils.Debug(ctx, logutils.LayerUsecase, op, "既存ユーザ更新", slog.Uint64("user_id", uint64(user.ID)))
		user.Name = ldapUser.Name
		user.DepartmentName = ldapUser.Department
		if !user.IsSupporterOverridden && shouldBeSupporter(ldapUser.Department) {
			user.IsSupporter = true
		}
		if err := uc.userRepo.Update(ctx, user); err != nil {
			usecase.LogRepoPropagation(ctx, op, "ユーザ更新失敗", err, slog.Uint64("user_id", uint64(user.ID)))
			return "", outputmodel.UserOutput{}, err
		}
	}
	token, err := issueToken(user)
	if err != nil {
		logutils.Error(ctx, logutils.LayerUsecase, op, "トークン発行失敗", slog.Uint64("user_id", uint64(user.ID)), slog.String("err", err.Error()))
		return "", outputmodel.UserOutput{}, err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "ログイン成功", slog.String("method", "ldap"), slog.String("email", user.Email), slog.Uint64("user_id", uint64(user.ID)))
	return token, converter.UserEntityToOutput(user), nil
}

func (uc *AuthUsecase) LoginBasic(ctx context.Context, email, password string) (string, outputmodel.UserOutput, error) {
	const op = opAuth + ".LoginBasic"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("email", email))
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			usecase.LogBusinessWarn(ctx, op, "ログイン失敗", fmt.Errorf("アカウントが存在しません"), slog.String("email", email))
			return "", outputmodel.UserOutput{}, fmt.Errorf("アカウントが存在しません")
		}
		usecase.LogRepoPropagation(ctx, op, "ユーザ取得失敗", err, slog.String("email", email))
		return "", outputmodel.UserOutput{}, err
	}
	if user.Password == nil {
		usecase.LogBusinessWarn(ctx, op, "ログイン失敗", fmt.Errorf("パスワードが入力されていません"), slog.String("email", email))
		return "", outputmodel.UserOutput{}, fmt.Errorf("パスワードが入力されていません")
	}
	ok, err := crypto.VerifyPassword(password, config.GetPepper(), *user.Password)
	if err != nil || !ok {
		usecase.LogBusinessWarn(ctx, op, "ログイン失敗", fmt.Errorf("メールアドレスかパスワードが誤っています"), slog.String("email", email))
		return "", outputmodel.UserOutput{}, fmt.Errorf("メールアドレスかパスワードが誤っています")
	}
	token, err := issueToken(user)
	if err != nil {
		logutils.Error(ctx, logutils.LayerUsecase, op, "トークン発行失敗", slog.Uint64("user_id", uint64(user.ID)), slog.String("err", err.Error()))
		return "", outputmodel.UserOutput{}, err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "ログイン成功", slog.String("method", "basic"), slog.String("email", user.Email), slog.Uint64("user_id", uint64(user.ID)))
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
		IsAdmin:     user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.GetJWTKey()))
}
