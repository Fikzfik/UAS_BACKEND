package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Struct User sesuai collection users
type User struct {
    ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    Username     string    `gorm:"size:50;unique;not null"`
    Email        string    `gorm:"size:100;unique;not null"`
    PasswordHash string    `gorm:"size:255;not null"`
    FullName     string    `gorm:"size:100;not null"`
    RoleID       uuid.UUID `gorm:"type:uuid"`
    IsActive     bool      `gorm:"default:true"`

    Role      Role      `gorm:"foreignKey:RoleID"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// Request body untuk login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Response login (user info + token)
type LoginResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

// Payload JWT
type JWTClaims struct {
	UserID string `json:"user_id"` // gunakan string karena ObjectID di Mongo
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
