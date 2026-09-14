package outputmodel

type QuestionListItemOutput struct {
	UUID                  string `json:"uuid"`
	Title                 string `json:"title"`
	SupportStatus         string `json:"supportStatus"`
	IsRequireHumanSupport bool   `json:"isRequireHumanSupport"`
	QuestionUserName      string `json:"questionUserName"`
	AnswerDue             string `json:"answerDue,omitempty"`
	Tags                  []string `json:"tags"`
}

type QuestionDetailOutput struct {
	UUID                  string              `json:"uuid"`
	Title                 string              `json:"title"`
	SupportStatus         string              `json:"supportStatus"`
	IsRequireHumanSupport bool                `json:"isRequireHumanSupport"`
	AnswerDue             string              `json:"answerDue,omitempty"`
	QuestionUserUUID      string              `json:"questionUserUuid"`
	QuestionUserID        uint                `json:"-"`
	QuestionUserName      string              `json:"questionUserName"`
	Tags                  []string            `json:"tags"`
	Contents              []TimelineOutput    `json:"contents"`
	Answers               []TimelineOutput    `json:"answers"`
	Memos                 []TimelineOutput    `json:"memos,omitempty"`
	Refers                []ReferOutput       `json:"refers"`
}

type TimelineOutput struct {
	UUID     string `json:"uuid"`
	Content  string `json:"content"`
	UserName string `json:"userName"`
	CreatedAt string `json:"createdAt"`
}

type ReferOutput struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
