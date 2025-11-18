package models

import (
    "time"

    "github.com/google/uuid"
)

type Role struct {
    ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    Name        string    `gorm:"size:50;unique;not null"`
    Description string
    CreatedAt   time.Time `gorm:"autoCreateTime"`

    Permissions []Permission `gorm:"many2many:role_permissions"`
}
