package middleware

import (
	"UAS_GO/helper" 
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.Unauthorized(c, "Token akses diperlukan")
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return helper.Unauthorized(c, "Format token tidak valid")
		}
		
		tokenString := tokenParts[1]

		claims, err := helper.ValidateToken(tokenString)
		if err != nil {
			return helper.Unauthorized(c, "Token tidak valid atau expired")
		}


		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role) 

		return c.Next()
	}
}

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role") 
		
		if roleStr, ok := role.(string); ok && roleStr == "admin" {
			return c.Next()
		}
		
		return helper.Forbidden(c, "Akses ditolak. Hanya admin yang diizinkan.")
	}
}

func DosenWaliOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if roleStr, ok := role.(string); ok && roleStr == "dosen_wali" {
			return c.Next()
		}		
		return helper.Forbidden(c, "Akses ditolak. Hanya dosen wali yang diizinkan.")
	}		
}

func MahasiswaOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")	
		if roleStr, ok := role.(string); ok && roleStr == "mahasiswa" {
			return c.Next()
		}		
		return helper.Forbidden(c, "Akses ditolak. Hanya mahasiswa yang diizinkan.")
	}	
}