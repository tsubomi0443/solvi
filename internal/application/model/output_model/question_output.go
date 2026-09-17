package outputmodel

type QuestionListItemOutput struct {
	UUID                   string   `json:"uuid"`
	Title                  string   `json:"title"`
	Content                string   `json:"content"`
	SupportStatus          string   `json:"supportStatus"`
	IsRequireHumanSupport  bool     `json:"isRequireHumanSupport"`
	QuestionUserUUID       string   `json:"questionUserUUID"`
	QuestionUserName       string   `json:"questionUserName"`
	QuestionUserDepartment string   `json:"questionUserDepartment"`
	AnswerDue              string   `json:"answerDue,omitempty"`
	Tags                   []string `json:"tags"`
}

type QuestionDetailOutput struct {
	UUID                   string           `json:"uuid"`
	Title                  string           `json:"title"`
	SupportStatus          string           `json:"supportStatus"`
	IsRequireHumanSupport  bool             `json:"isRequireHumanSupport"`
	AnswerDue              string           `json:"answerDue,omitempty"`
	QuestionUserUUID       string           `json:"questionUserUuid"`
	QuestionUserID         uint             `json:"-"`
	QuestionUserName       string           `json:"questionUserName"`
	QuestionUserDepartment string           `json:"questionUserDepartment"`
	Tags                   []string         `json:"tags"`
	Contents               []TimelineOutput `json:"contents"`
	Answers                []TimelineOutput `json:"answers"`
	Memos                  []TimelineOutput `json:"memos,omitempty"`
	Refers                 []ReferOutput    `json:"refers"`
	Summary                *SummaryOutput   `json:"summary,omitempty"`
}

type SummaryOutput struct {
	UUID       string                    `json:"uuid"`
	Title      string                    `json:"title"`
	Content    string                    `json:"content"`
	Answer     string                    `json:"answer"`
	References []SummaryReferenceOutput `json:"references"`
}

type SummaryReferenceOutput struct {
	UUID string `json:"uuid,omitempty"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type TimelineOutput struct {
	UUID      string `json:"uuid"`
	Content   string `json:"content"`
	UserUUID  string `json:"userUuid"`
	UserName  string `json:"userName"`
	CreatedAt string `json:"createdAt"`
}

type ReferOutput struct {
	UUID      string `json:"uuid,omitempty"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	UserUUID  string `json:"userUuid,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}
