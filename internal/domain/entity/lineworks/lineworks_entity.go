package lineworks

import (
	"time"

	"solvi/internal/domain/valueobject"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notification は LINE WORKS への送信予定。質問の識別は QuestionUUID。
type Notification struct {
	gorm.Model
	UUID             uuid.UUID                        `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()"`
	QuestionUUID     uuid.UUID                        `gorm:"type:uuid;not null;uniqueIndex:ux_lineworks_notice"`
	Event            valueobject.LineWorksEvent       `gorm:"type:smallint;not null;uniqueIndex:ux_lineworks_notice"`
	StatusRevision   int                              `gorm:"not null;uniqueIndex:ux_lineworks_notice"`
	Burst            int                              `gorm:"not null;default:0;uniqueIndex:ux_lineworks_notice"`
	ChannelID        string                           `gorm:"type:text;not null;default:'';uniqueIndex:ux_lineworks_notice"`
	Destination      valueobject.LineWorksDestination `gorm:"type:smallint;not null"`
	RecipientLoginID string                           `gorm:"type:text"`
	Body             string                           `gorm:"type:text;not null"`
	Status           valueobject.LineWorksJobStatus   `gorm:"type:smallint;not null;index:idx_lw_due,priority:1"`
	NextAttemptAt    time.Time                        `gorm:"not null;index:idx_lw_due,priority:2"`
	AttemptCount     int                              `gorm:"not null;default:0"`
	ErrorKind        valueobject.LineWorksErrorKind   `gorm:"type:smallint;not null;default:0"`
	LastError        string                           `gorm:"type:text"`
	SentAt           *time.Time                       `gorm:"type:timestamptz"`
}

func (Notification) TableName() string { return "lineworks_notifications" }

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.UUID == uuid.Nil {
		n.UUID = uuid.New()
	}
	return nil
}

// NotificationAttempt は送信試行の履歴。秘密情報は持たない。
type NotificationAttempt struct {
	gorm.Model
	UUID             uuid.UUID                           `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()"`
	NotificationUUID uuid.UUID                           `gorm:"type:uuid;not null;index"`
	HTTPStatus       int                                 `gorm:"type:int"`
	Outcome          valueobject.LineWorksAttemptOutcome `gorm:"type:smallint;not null"`
	AttemptedAt      time.Time                           `gorm:"not null"`
}

func (NotificationAttempt) TableName() string { return "lineworks_notification_attempts" }

func (a *NotificationAttempt) BeforeCreate(tx *gorm.DB) error {
	if a.UUID == uuid.Nil {
		a.UUID = uuid.New()
	}
	return nil
}

// ReopenFollowUp は対応中コメントの集約入力。永続化しない。
type ReopenFollowUp struct {
	QuestionUUID uuid.UUID
	Comment      string
	Debounce     time.Duration
	ChannelIDs   []string
	Template     Notification
}
