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

        // claims.Role diasumsikan adalah role ID (UUID). Ambil nama role untuk convenience.
        roleName, err := repository.GetRoleNameByID(claims.Role)
        if err != nil {
            return helper.Unauthorized(c, "Role tidak ditemukan")
        }

        c.Locals("user_id", claims.UserID)
        c.Locals("email", claims.Email)
        c.Locals("role", roleName)
        c.Locals("role_id", claims.Role) // <<-- simpan role id juga



        return c.Next()
    }
}
func PermissionRequired(permission string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 1) Fast path: cek apakah permissions ada di Locals (mis. dari JWT claims)
        if permsVal := c.Locals("permissions"); permsVal != nil {
            if permsSlice, ok := permsVal.([]string); ok {
                for _, p := range permsSlice {
                    if p == permission {
                        return c.Next()
                    }
                }
            }
        }

        // 2) Ambil role_id dari locals
        roleIDVal := c.Locals("role_id")
        roleID, _ := roleIDVal.(string)
        if roleID == "" {
            // jika role_id kosong, kemungkinan AuthRequired belum dipanggil
            return helper.Unauthorized(c, "Unauthorized")
        }

        // 3) Cek lewat repository (DB)
        has, err := repository.RoleHasPermission(roleID, permission)
        if err != nil {
            return helper.InternalError(c, "Error checking permissions")
        }
        if !has {
            return helper.Forbidden(c, "Akses ditolak. Permission diperlukan: "+permission)
        }

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

func AdminOrLecturerOrOwnerStudent() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role").(string)
		userID := c.Locals("user_id").(string)
		studentID := c.Params("id")

		if studentID == "" {
			return helper.BadRequest(c, "student id is required")
		}

		// 1. Admin boleh semua
		if role == "admin" {
			return c.Next()
		}

		// 2. Mahasiswa hanya boleh akses dirinya sendiri
		if role == "mahasiswa" {
			myStudentID, err := repository.GetStudentIDByUserID(userID)
			if err != nil {
				return helper.Forbidden(c, "Student profile not found")
			}

			if myStudentID != studentID {
				return helper.Forbidden(c, "You are not allowed to access other student's data")
			}

			return c.Next()
		}

		// 3. Dosen Wali boleh akses mahasiswa bimbingannya
		if role == "dosen_wali" {
			lecturerID, err := repository.GetLecturerIDByUserID(userID)
			if err != nil {
				return helper.Forbidden(c, "Lecturer profile not found")
			}

			// cek apakah student ini adalah bimbingan dosen tersebut
			isAdvisee, err := repository.IsStudentAdviseeOfLecturer(studentID, lecturerID)
			if err != nil {
				return helper.InternalError(c, err.Error())
			}

			if !isAdvisee {
				return helper.Forbidden(c, "You are not allowed to access this student's data")
			}

			return c.Next()
		}

		// 4. Role lain ditolak
		return helper.Forbidden(c, "Access denied")
	}
}