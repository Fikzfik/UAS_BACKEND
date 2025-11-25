package service

import (
	"UAS_GO/app/models"
	"UAS_GO/app/repository"
	"UAS_GO/database"
	"UAS_GO/helper" 
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v2"
)

type AuthService struct{}

func (s *AuthService) VerifyToken(param1 string) (any, any) {
	panic("unimplemented")
}

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
func (s *AuthService) Login(email, password string) (*models.LoginResponse, error) {
	// Query user from PostgreSQL
	query := `SELECT id, email, password_hash, role_id, is_active FROM users WHERE email = $1`
	user := &models.User{}
	err := database.PSQL.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.RoleID, &user.IsActive)
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
		// jika gagal ambil permissions, jangan gagal total login; log error dan return tanpa permissions
		// tapi untuk sekarang kembalikan error supaya kelihatan apa yang terjadi
		return nil, err
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

	return &models.LoginResponse{
		User:  *user,
		Token: newToken,
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