package models

import (
    "github.com/google/uuid"
)

type Permission struct {
    ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    Name        string    `gorm:"size:100;unique;not null"`
    Resource    string    `gorm:"size:50;not null"`
    Action      string    `gorm:"size:50;not null"`
    Description string
}

type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
}