package models
import (
	"github.com/google/uuid"
)

type Student struct {
	ID				uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID  		uuid.UUID `gorm:"type:uuid;not null;unique"`
	StudentID  		string    `gorm:"size:20;unique;not null"`
	ProgramStudy	string    `gorm:"size:100;not null"`
	AcademicYear	string    `gorm:"size:20;not null"`
	AdvisorID		uuid.UUID `gorm:"type:uuid"`

	User		User	`gorm:"foreignKey:UserID"`
	Advisor		Lecturer	`gorm:"foreignKey:AdvisorID"`
}