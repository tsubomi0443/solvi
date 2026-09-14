package entity

import (
	"solvi/internal/domain/valueobject"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Question struct {
	gorm.Model
	UUID                   uuid.UUID                `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()"`
	Title                  string                   `gorm:"type:text;not null"`
	AnswerDue              *time.Time               `gorm:"type:timestamptz"`
	IsRequireHumanSupport  bool                     `gorm:"not null;default:false"`
	SupportStatus          valueobject.SupportStatus `gorm:"type:smallint;not null;default:1"`
	QuestionUserID         uint                     `gorm:"not null;index"`

	QuestionUser           User                     `gorm:"foreignKey:QuestionUserID"`
	Contents               []QuestionContent        `gorm:"foreignKey:QuestionID"`
	Tags                   []QuestionTag            `gorm:"foreignKey:QuestionID"`
	Answers                []QuestionAnswer         `gorm:"foreignKey:QuestionID"`
	Memos                  []QuestionMemo           `gorm:"foreignKey:QuestionID"`
	Refers                 []QuestionRefer          `gorm:"foreignKey:QuestionID"`
	Summary                *QuestionSummary         `gorm:"foreignKey:QuestionID"`
}

type QuestionContent struct {
	gorm.Model
	UUID           uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()"`
	Content        string    `gorm:"type:text;not null"`
	QuestionUserID uint   `gorm:"not null;index"`
	QuestionID     uint   `gorm:"not null;index"`

	QuestionUser User `gorm:"foreignKey:QuestionUserID"`
}

type QuestionTag struct {
	gorm.Model
	UUID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()"`
	Name       string    `gorm:"type:text;not null"`
	QuestionID uint   `gorm:"not null;index"`
}

type QuestionAnswer struct {
	gorm.Model
	UUID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()"`
	Content      string    `gorm:"type:text;not null"`
	AnswerUserID uint   `gorm:"not null;index"`
	QuestionID   uint   `gorm:"not null;index"`

	AnswerUser User `gorm:"foreignKey:AnswerUserID"`
}

type QuestionMemo struct {
	gorm.Model
	UUID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()"`
	Content      string    `gorm:"type:text;not null"`
	QuestionID   uint   `gorm:"not null;index"`
	MemoUserID   uint   `gorm:"not null;index"`

	MemoUser User `gorm:"foreignKey:MemoUserID"`
}

type QuestionRefer struct {
	gorm.Model
	UUID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()"`
	Name       string    `gorm:"type:text;not null"`
	URL        string `gorm:"type:text;not null"`
	QuestionID uint   `gorm:"not null;index"`
	UserID     uint   `gorm:"not null;index"`

	User User `gorm:"foreignKey:UserID"`
}

type QuestionSummary struct {
	gorm.Model
	UUID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()"`
	Title      string    `gorm:"type:text;not null"`
	Content    string `gorm:"type:text;not null"`
	Answer     string `gorm:"type:text;not null"`
	QuestionID uint   `gorm:"not null;uniqueIndex"`

	References []QuestionSummaryReference `gorm:"foreignKey:QuestionSummaryID"`
}

type QuestionSummaryReference struct {
	gorm.Model
	UUID               uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()"`
	Name               string    `gorm:"type:text;not null"`
	URL                string `gorm:"type:text;not null"`
	QuestionSummaryID  uint   `gorm:"not null;index"`
}
