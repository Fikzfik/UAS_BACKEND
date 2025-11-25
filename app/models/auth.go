package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Struct User sesuai collection users
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	FullName     string    `json:"full_name"`
	RoleID       string    `json:"role_id"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Request body untuk login
type LoginRequest struct {
    Email    string `json:"email"`
    NIM      string `json:"nim"`
    Password string `json:"password"`
}

// Response login (user info + token)
type LoginResponse struct {
    User        User     `json:"user"`
    Token       string   `json:"token"`
    Permissions []string `json:"permissions"` // <- tambahkan ini jika belum ada
}


// Payload JWT
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type RefreshTokenRequest struct {
	Token string `json:"token"`
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required"`
	FullName string `json:"full_name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	RoleID   string `json:"role_id" validate:"required"`
	IsActive bool   `json:"is_active" validate:"required"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	RoleID   string `json:"role_id"`
	IsActive bool   `json:"is_active"`
	Password string `json:"password"`
}

type UpdateUserRoleRequest struct {
	RoleID string `json:"role_id" validate:"required"`
}
