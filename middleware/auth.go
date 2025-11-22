package middleware

import (
	"UAS_GO/app/repository"
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

        roleName, err := repository.GetRoleNameByID(claims.Role)
        if err != nil {
            return helper.Unauthorized(c, "Role tidak ditemukan")
        }

        c.Locals("user_id", claims.UserID)
        c.Locals("email", claims.Email)
        c.Locals("role", roleName)

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

func OwnerOrAdvisorOrAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleVal := c.Locals("role")
		role, _ := roleVal.(string)

		// admin always allowed
		if role == "admin" {
			return c.Next()
		}

		// get student id from path
		studentID := c.Params("id")
		if studentID == "" {
			return helper.BadRequest(c, "Student id is required")
		}

		// if mahasiswa -> ensure it's their own resource
		if role == "mahasiswa" {
			userIDVal := c.Locals("user_id")
			userID, _ := userIDVal.(string)
			if userID == "" {
				return helper.Unauthorized(c, "Unauthorized")
			}
			// convert user -> student.id
			sid, err := repository.GetStudentIDByUserID(userID)
			if err != nil {
				return helper.Forbidden(c, "Student profile not found")
			}
			if sid == studentID {
				return c.Next()
			}
			return helper.Forbidden(c, "You are not allowed to access this student's data")
		}

		// if lecturer/dosen_wali -> check advisor relation
		if role == "dosen_wali" || role == "lecturer" {
			userIDVal := c.Locals("user_id")
			userID, _ := userIDVal.(string)
			if userID == "" {
				return helper.Unauthorized(c, "Unauthorized")
			}
			lecturerID, err := repository.GetLecturerIDByUserID(userID)
			if err != nil {
				return helper.Forbidden(c, "Lecturer profile not found")
			}
			isAdvisor, err := repository.IsLecturerAdvisorOfStudent(lecturerID, studentID)
			if err != nil {
				return helper.InternalError(c, "Error checking advisor relation")
			}
			if isAdvisor {
				return c.Next()
			}
			return helper.Forbidden(c, "You are not the academic advisor for this student")
		}

		// others not allowed
		return helper.Forbidden(c, "Access denied")
	}
}

func LecturerOrAdminForLecturerResource() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleVal := c.Locals("role")
		role, _ := roleVal.(string)

		// admin allowed
		if role == "admin" {
			return c.Next()
		}

		if role != "dosen_wali" {
			return helper.Forbidden(c, "Akses ditolak. Hanya admin atau dosen yang diizinkan.")
		}

		paramLecturerID := c.Params("id")
		if paramLecturerID == "" {
			return helper.BadRequest(c, "Lecturer id required")
		}

		// get lecturer id by current user
		userIDVal := c.Locals("user_id")
		userID, _ := userIDVal.(string)
		if userID == "" {
			return helper.Unauthorized(c, "Unauthorized")
		}

		lecturerID, err := repository.GetLecturerIDByUserID(userID)
		if err != nil {
			return helper.Forbidden(c, "Lecturer profile not found")
		}

		if lecturerID != paramLecturerID {
			return helper.Forbidden(c, "You are not allowed to access other lecturer's resources")
		}

		return c.Next()
	}
}