package repository

import (
	"context"
)

//go:generate go run go.uber.org/mock/mockgen@latest -typed -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

type TagStat struct {
	Name  string
	Count int64
}

type TagRepository interface {
	ListTagStats(ctx context.Context) ([]TagStat, error)
	RenameTag(ctx context.Context, from, to string) error
	DeleteTagByName(ctx context.Context, name string) error
}
