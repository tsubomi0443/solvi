package external

type FAQReference struct {
	Name string
	URL  string
}

type FAQAnswerResult struct {
	Content    string
	References []FAQReference
}

//go:generate go tool mockgen -typed -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

type BedrockClient interface {
	AnswerQuestion(title, content string) (*FAQAnswerResult, error)
}
