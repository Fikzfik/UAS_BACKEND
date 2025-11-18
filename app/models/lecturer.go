package models

import (
	"github.com/google/uuid"
)	

type Lecturer struct {
	ID         	uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID		uuid.UUID `gorm:"type:uuid;not null;unique"`
	LecturerID	string    `gorm:"size:20;unique;not null"`
	Department	string    `gorm:"size:100;not null"`
	
	User		User	`gorm:"foreignKey:UserID"`
}