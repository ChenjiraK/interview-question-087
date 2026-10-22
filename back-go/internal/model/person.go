package model

import (
	"errors"
	"time"
)

var ErrPersonNotFound = errors.New("person not found")

// Person maps to the schema managed by the SQL migrations.
type Person struct {
	ID        int64      `gorm:"primaryKey;autoIncrement"`
	Firstname string     `gorm:"column:firstname;not null"`
	Lastname  string     `gorm:"column:lastname;not null"`
	Birthday  *time.Time `gorm:"type:date"`
	Address   *string
	CreatedAt time.Time `gorm:"type:timestamptz;not null;autoCreateTime:false;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null;autoUpdateTime:false;default:CURRENT_TIMESTAMP"`
}

func (Person) TableName() string {
	return "public.person"
}
