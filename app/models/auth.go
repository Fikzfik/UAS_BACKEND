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
	Password string `json:"password"`
}

// Response login (user info + token)
type LoginResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

// Payload JWT
type JWTClaims struct {
	UserID 		string `json:"user_id"`
	Email  		string `json:"email"`
	Role   		string `json:"role"`
	jwt.RegisteredClaims
}

type RefreshTokenRequest struct {
    Token string `json:"token"`
}

type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
    RoleID   string `json:"role_id" validate:"required"`
}

type UpdateUserRequest struct {
    Email    string `json:"email"`
    RoleID   string `json:"role_id"`
    IsActive bool   `json:"is_active"`
}

type UpdateUserRoleRequest struct {
    RoleID string `json:"role_id" validate:"required"`
}