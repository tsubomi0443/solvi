package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func assignUUIDIfZero(dest *uuid.UUID) error {
	if *dest == uuid.Nil {
		*dest = uuid.New()
	}
	return nil
}

func (u *User) BeforeCreate(tx *gorm.DB) error { return assignUUIDIfZero(&u.UUID) }

func (q *Question) BeforeCreate(tx *gorm.DB) error { return assignUUIDIfZero(&q.UUID) }

func (c *QuestionContent) BeforeCreate(tx *gorm.DB) error { return assignUUIDIfZero(&c.UUID) }

func (t *QuestionTag) BeforeCreate(tx *gorm.DB) error { return assignUUIDIfZero(&t.UUID) }

func (a *QuestionAnswer) BeforeCreate(tx *gorm.DB) error { return assignUUIDIfZero(&a.UUID) }

func (m *QuestionMemo) BeforeCreate(tx *gorm.DB) error { return assignUUIDIfZero(&m.UUID) }

func (r *QuestionRefer) BeforeCreate(tx *gorm.DB) error { return assignUUIDIfZero(&r.UUID) }

func (s *QuestionSummary) BeforeCreate(tx *gorm.DB) error { return assignUUIDIfZero(&s.UUID) }

func (r *QuestionSummaryReference) BeforeCreate(tx *gorm.DB) error {
	return assignUUIDIfZero(&r.UUID)
}
