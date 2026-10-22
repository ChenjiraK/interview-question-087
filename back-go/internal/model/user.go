package model

import (
	"errors"
	"time"
)

var ErrUserNotFound = errors.New("user not found")

// User maps to the schema managed by the SQL migrations.
type User struct {
	ID        int64      `gorm:"primaryKey;autoIncrement"`
	Firstname string     `gorm:"column:firstname;not null"`
	Lastname  string     `gorm:"column:lastname;not null"`
	Birthday  *time.Time `gorm:"type:date"`
	Address   *string
	CreatedAt time.Time `gorm:"type:timestamptz;not null;autoCreateTime:false;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null;autoUpdateTime:false;default:CURRENT_TIMESTAMP"`
}

func (User) TableName() string {
	return "public.user"
}
