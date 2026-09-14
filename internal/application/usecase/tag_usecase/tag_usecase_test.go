package tag_usecase_test

import (
	"testing"

	taguc "solvi/internal/application/usecase/tag_usecase"
	repomock "solvi/internal/domain/interface/repository/mock"
	repo "solvi/internal/domain/interface/repository"

	"go.uber.org/mock/gomock"
)

func TestList_ReturnsStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	tagRepo := repomock.NewMockTagRepository(ctrl)
	tagRepo.EXPECT().ListTagStats().Return([]repo.TagStat{
		{Name: "給与", Count: 3},
	}, nil)

	uc := taguc.NewTagUsecase(tagRepo)
	items, err := uc.List()
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
	if err := uc.Rename("", "x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDelete_CallsRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	tagRepo := repomock.NewMockTagRepository(ctrl)
	tagRepo.EXPECT().DeleteTagByName("勤怠").Return(nil)

	uc := taguc.NewTagUsecase(tagRepo)
	if err := uc.Delete("勤怠"); err != nil {
		t.Fatal(err)
	}
}
