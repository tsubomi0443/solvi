package tag_usecase_test

import (
	"context"
	"testing"

	taguc "solvi/internal/application/usecase/tag_usecase"
	repomock "solvi/internal/domain/interface/repository/mock"
	repo "solvi/internal/domain/interface/repository"

	"go.uber.org/mock/gomock"
)

func TestList_ReturnsStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	tagRepo := repomock.NewMockTagRepository(ctrl)
	tagRepo.EXPECT().ListTagStats(gomock.Any()).Return([]repo.TagStat{
		{Name: "給与", Count: 3},
	}, nil)

	uc := taguc.NewTagUsecase(tagRepo)
	items, err := uc.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Name != "給与" || items[0].Count != 3 {
		t.Fatalf("unexpected: %+v", items)
	}
}

func TestRename_ValidatesEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	tagRepo := repomock.NewMockTagRepository(ctrl)
	uc := taguc.NewTagUsecase(tagRepo)
	if err := uc.Rename(context.Background(), "", "x"); err == nil {
		t.Fatal("expected error")
	}
	if err := uc.Rename(context.Background(), "a", "   "); err == nil {
		t.Fatal("expected error when to is blank")
	}
}

func TestRename_SameTagNoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	tagRepo := repomock.NewMockTagRepository(ctrl)

	uc := taguc.NewTagUsecase(tagRepo)
	if err := uc.Rename(context.Background(), " 総務 ", "総務"); err != nil {
		t.Fatalf("expected nil when from and to match, got %v", err)
	}
}

func TestRename_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	tagRepo := repomock.NewMockTagRepository(ctrl)
	tagRepo.EXPECT().RenameTag(gomock.Any(), "古いタグ", "新しいタグ").Return(nil)

	uc := taguc.NewTagUsecase(tagRepo)
	if err := uc.Rename(context.Background(), " 古いタグ ", " 新しいタグ "); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDelete_ValidatesEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	tagRepo := repomock.NewMockTagRepository(ctrl)

	uc := taguc.NewTagUsecase(tagRepo)
	if err := uc.Delete(context.Background(), "   "); err == nil {
		t.Fatal("expected error for empty tag name")
	}
}

func TestDelete_CallsRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	tagRepo := repomock.NewMockTagRepository(ctrl)
	tagRepo.EXPECT().DeleteTagByName(gomock.Any(), "勤怠").Return(nil)

	uc := taguc.NewTagUsecase(tagRepo)
	if err := uc.Delete(context.Background(), "勤怠"); err != nil {
		t.Fatal(err)
	}
}
