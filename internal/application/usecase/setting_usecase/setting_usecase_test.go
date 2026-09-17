package setting_usecase_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	setuc "solvi/internal/application/usecase/setting_usecase"
	"solvi/internal/domain/entity"
	repomock "solvi/internal/domain/interface/repository/mock"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestUploadIcon_ReplacesOldIcon(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)
	uploadDir := t.TempDir()

	oldIconName := uuid.NewString() + ".png"
	oldPath := filepath.Join(uploadDir, oldIconName)
	if err := os.WriteFile(oldPath, []byte("old-icon-content"), 0644); err != nil {
		t.Fatal(err)
	}

	user := &entity.User{
		Icon: &oldIconName,
	}
	user.ID = 1

	userRepo.EXPECT().GetByID(gomock.Any(), uint(1)).Return(user, nil)
	userRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *entity.User) error {
		if u.Icon == nil || *u.Icon == oldIconName {
			t.Fatalf("expected new icon name, got %v", u.Icon)
		}
		return nil
	})

	uc := setuc.NewSettingUsecase(userRepo, uploadDir)
	newContent := []byte("new-icon-content")
	if err := uc.UploadIcon(context.Background(), 1, "avatar.png", bytes.NewReader(newContent)); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("expected old icon to be removed, but stat returned: %v", err)
	}

	if user.Icon == nil {
		t.Fatal("user.Icon is nil")
	}
	newPath := filepath.Join(uploadDir, *user.Icon)
	data, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("failed to read new icon: %v", err)
	}
	if !bytes.Equal(data, newContent) {
		t.Fatalf("expected new content, got %s", data)
	}
}

func TestUploadIcon_WithoutExistingIcon(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)
	uploadDir := t.TempDir()

	user := &entity.User{
		Icon: nil,
	}
	user.ID = 1

	userRepo.EXPECT().GetByID(gomock.Any(), uint(1)).Return(user, nil)
	userRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	uc := setuc.NewSettingUsecase(userRepo, uploadDir)
	if err := uc.UploadIcon(context.Background(), 1, "first.png", bytes.NewReader([]byte("first"))); err != nil {
		t.Fatal(err)
	}

	if user.Icon == nil {
		t.Fatal("user.Icon is nil")
	}
	if _, err := os.Stat(filepath.Join(uploadDir, *user.Icon)); err != nil {
		t.Fatalf("expected uploaded file to exist: %v", err)
	}
}

func TestDeleteIcon_RemovesExistingFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)
	uploadDir := t.TempDir()

	iconName := uuid.NewString() + ".png"
	iconPath := filepath.Join(uploadDir, iconName)
	if err := os.WriteFile(iconPath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	user := &entity.User{
		Icon: &iconName,
	}
	user.ID = 1

	userRepo.EXPECT().GetByID(gomock.Any(), uint(1)).Return(user, nil)
	userRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *entity.User) error {
		if u.Icon != nil {
			t.Fatalf("expected user.Icon to be nil, got %v", *u.Icon)
		}
		return nil
	})

	uc := setuc.NewSettingUsecase(userRepo, uploadDir)
	if err := uc.DeleteIcon(context.Background(), 1); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(iconPath); !os.IsNotExist(err) {
		t.Fatalf("expected file to be deleted, got: %v", err)
	}
}

func TestDeleteIcon_MissingFileClearsDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)
	uploadDir := t.TempDir()

	iconName := uuid.NewString() + ".png"
	user := &entity.User{
		Icon: &iconName,
	}
	user.ID = 1

	userRepo.EXPECT().GetByID(gomock.Any(), uint(1)).Return(user, nil)
	userRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *entity.User) error {
		if u.Icon != nil {
			t.Fatalf("expected user.Icon to be nil, got %v", *u.Icon)
		}
		return nil
	})

	uc := setuc.NewSettingUsecase(userRepo, uploadDir)
	if err := uc.DeleteIcon(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteIcon_RejectsPathTraversal(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repomock.NewMockUserRepository(ctrl)
	uploadDir := t.TempDir()

	traversalName := "../secret.txt"
	user := &entity.User{
		Icon: &traversalName,
	}
	user.ID = 1

	userRepo.EXPECT().GetByID(gomock.Any(), uint(1)).Return(user, nil)

	uc := setuc.NewSettingUsecase(userRepo, uploadDir)
	err := uc.DeleteIcon(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for path traversal icon name")
	}
	if user.Icon == nil || *user.Icon != traversalName {
		t.Fatalf("user.Icon should have remained unchanged, got %v", user.Icon)
	}
}
