package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Struct User sesuai collection users
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password_hash,omitempty" json:"-"`
	Role         string             `bson:"role" json:"role"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
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
