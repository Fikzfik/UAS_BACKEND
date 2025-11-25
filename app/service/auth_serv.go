package service

import (
	"UAS_GO/app/models"
	"UAS_GO/app/repository"
	"UAS_GO/database"
	"UAS_GO/helper"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func AuthLogin(c *fiber.Ctx) error {
	authService := NewAuthService()

	var req models.LoginRequest

	// Parsing body request
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Invalid request format")
	}
fmt.Printf("DEBUG LoginRequest: %+v\n", req)
	// Validasi input dasar: minimal email atau nim, dan password
	if req.Password == "" || (req.Email == "" && req.NIM == "") {
		return helper.BadRequest(c, "Email or NIM and password are required")
	}

	// Tentukan apakah login by NIM atau by Email
	byNIM := false
	identifier := req.Email
	if req.NIM != "" {
		byNIM = true
		identifier = req.NIM
	}

	// Memanggil service untuk logika bisnis (termasuk verifikasi password dan generate token)
	resp, err := authService.Login(identifier, req.Password, byNIM)
	if err != nil {
		// Menggunakan helper.Unauthorized untuk error otentikasi
		return helper.Unauthorized(c, err.Error())
	}

	// Respons sukses
	return helper.APIResponse(c, fiber.StatusOK, "Login successful", resp)
}

func AuthGetProfile(c *fiber.Ctx) error {

	// Extract user ID from JWT claims
	userID := c.Locals("user_id")
	if userID == nil {
		return helper.Unauthorized(c, "user not authenticated")
	}

	// Call service to get user profile
	profile, err := repository.GetUserProfile(userID.(string))
	if err != nil {
		if err == sql.ErrNoRows {
			return helper.NotFound(c, "user not found")
		}
		return helper.InternalError(c, err.Error())
	}

	return helper.APIResponse(c, fiber.StatusOK, "Profile retrieved successfully", profile)
}

func AuthLogout(c *fiber.Ctx) error {

	// Extract user ID from JWT claims (assumes middleware sets it)
	userID := c.Locals("user_id")
	if userID == nil {
		return helper.Unauthorized(c, "user not authenticated")
	}

	err := repository.LogoutUser(userID.(string))
	if err != nil {
		return helper.InternalError(c, err.Error())
	}

	return helper.APIResponse(c, fiber.StatusOK, "Logout successful", nil)
}

func (s *AuthService) Login(identifier, password string, byNIM bool) (*models.LoginResponse, error) {
	// Query user from PostgreSQL
	var query string
	var user = &models.User{}

	if byNIM {
		// Cari user berdasarkan student_id (NIM) di table students
		query = `
			SELECT u.id, u.email, u.password_hash, u.role_id, u.is_active
			FROM users u
			JOIN students s ON s.user_id = u.id
			WHERE s.student_id = $1
		`
	} else {
		// Cari user berdasarkan email
		query = `SELECT id, email, password_hash, role_id, is_active FROM users WHERE email = $1`
	}

	err := database.PSQL.QueryRow(query, identifier).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.RoleID, &user.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Verify password
	if !helper.CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("invalid password")
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Generate JWT token
	token, err := helper.GenerateToken(*user)
	if err != nil {
		return nil, err
	}

	// Ambil daftar permissions berdasarkan role_id
	perms, err := repository.GetPermissionsByRoleID(user.RoleID)
	if err != nil {
		// jangan gagalkan login hanya karena gagal ambil permissions; kembalikan tanpa permissions
		perms = []string{}
	}

	return &models.LoginResponse{
		User:        *user,
		Token:       token,
		Permissions: perms,
	}, nil
}

func (s *AuthService) RefreshToken(token string) (*models.LoginResponse, error) {
	// Verify and parse the token
	claims, err := helper.VerifyToken(token)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	// Query user from PostgreSQL to get fresh data
	query := `SELECT id, email, password_hash, role_id, is_active FROM users WHERE id = $1`
	user := &models.User{}
	err = database.PSQL.QueryRow(query, claims.UserID).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.RoleID, &user.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Generate new JWT token
	newToken, err := helper.GenerateToken(*user)
	if err != nil {
		return nil, err
	}

	// Ambil permissions untuk role (jangan gagal token refresh jika error -> kembalikan perms kosong)
	perms, err := repository.GetPermissionsByRoleID(user.RoleID)
	if err != nil {
		perms = []string{}
	}

	return &models.LoginResponse{
		User:        *user,
		Token:       newToken,
		Permissions: perms,
	}, nil
}

func AuthRefreshToken(c *fiber.Ctx) error {
	authService := NewAuthService()

	var req models.RefreshTokenRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Invalid request format")
	}

	if req.Token == "" {
		return helper.BadRequest(c, "Token is required")
	}

	// Call service to refresh token
	resp, err := authService.RefreshToken(req.Token)
	if err != nil {
		return helper.Unauthorized(c, err.Error())
	}

	return helper.APIResponse(c, fiber.StatusOK, "Token refreshed successfully", resp)
}
