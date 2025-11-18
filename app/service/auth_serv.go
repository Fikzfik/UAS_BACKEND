package service

import (
	"UAS_GO/app/models"
	"UAS_GO/database"
	"UAS_GO/helper" // <-- Pastikan ini diimpor
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v2"
	// Tidak perlu import "github.com/golang-jwt/jwt/v5" di sini lagi
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func AuthLoginHandler(c *fiber.Ctx) error {
	// Instansiasi service di sini (atau bisa di-pass melalui dependency injection)
	authService := NewAuthService()

	var req models.LoginRequest

	// Parsing body request
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Invalid request format")
	}

	// Validasi input dasar
	if req.Email == "" || req.Password == "" {
		return helper.BadRequest(c, "Email and password are required")
	}

	// Memanggil service untuk logika bisnis (termasuk verifikasi password dan generate token)
	resp, err := authService.Login(req.Email, req.Password)
	if err != nil {
		// Menggunakan helper.Unauthorized untuk error otentikasi
		return helper.Unauthorized(c, err.Error())
	}

	// Respons sukses
	return helper.APIResponse(c, fiber.StatusOK, "Login successful", resp)
}

func (s *AuthService) Login(email, password string) (*models.LoginResponse, error) {
	user, err := s.findUserByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Panggil fungsi CheckPassword dari package helper (Asumsi ada di helper/password.go)
	// Asumsi helper.CheckPassword ada dan melakukan verifikasi hash
	if !helper.CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("invalid password")
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// PANGGIL HELPER BARU UNTUK GENERATE TOKEN
	token, err := helper.GenerateToken(*user)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		User: *user,
		Token: token,
	}, nil
}

func (s *AuthService) findUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	err := database.PSQL.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.RoleID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}

// Catatan: Fungsi generateJWTToken telah dihapus dari file ini.