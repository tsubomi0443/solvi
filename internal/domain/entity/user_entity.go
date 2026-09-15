package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UUID                  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()"`
	Name                  string    `gorm:"type:text;not null"`
	Email                 string    `gorm:"type:text;not null;uniqueIndex"`
	Password              *string   `gorm:"type:text"`
	DepartmentName        string    `gorm:"type:text"`
	Icon                  *string   `gorm:"type:text"`
	IconBlob              []byte    `gorm:"-"`
	IsSupporter           bool      `gorm:"not null;default:false"`
	IsSupporterOverridden bool      `gorm:"not null;default:false"`
	IsAdmin               bool      `gorm:"not null;default:false"`
}
