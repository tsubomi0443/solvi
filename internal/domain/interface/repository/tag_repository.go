package repository

//go:generate go run go.uber.org/mock/mockgen@latest -typed -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

type TagStat struct {
	Name  string
	Count int64
}

type TagRepository interface {
	ListTagStats() ([]TagStat, error)
	RenameTag(from, to string) error
	DeleteTagByName(name string) error
}
