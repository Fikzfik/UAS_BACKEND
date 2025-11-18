package models
import (
    "time"
    "github.com/google/uuid"
)


type AchievementReference struct {
    ID                uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    StudentID         uuid.UUID `gorm:"type:uuid;not null"`
    MongoAchievementID string    `gorm:"size:24;not null"`
    Status            string    `gorm:"type:enum('draft','submitted','verified','rejected')"`
    SubmittedAt       *time.Time
    VerifiedAt        *time.Time
    VerifiedBy        *uuid.UUID `gorm:"type:uuid"`
    RejectionNote     string

    Student Student `gorm:"foreignKey:StudentID"`
}
