package service

import (
	"UAS_GO/app/models"
	"UAS_GO/app/repository"
	"UAS_GO/helper"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AdminGetAllUsers godoc
// @Summary      Get all users (admin)
// @Description  Mengambil daftar semua user (khusus admin)
// @Tags         Admin - Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "envelope {status,message,data:[users]}"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden (not admin)"
// @Failure      500  {object}  map[string]interface{}  "error response"
// @Router       /admin/users [get]
func AdminGetAllUsers(c *fiber.Ctx) error {
	users, err := repository.GetAllUsers()
	if err != nil {
		return helper.InternalError(c, err.Error())
	}
	return helper.APIResponse(c, fiber.StatusOK, "users retrieved", users)
}

// AdminGetUserByID godoc
// @Summary      Get user by ID (admin)
// @Description  Mengambil detail user berdasarkan ID (UUID)
// @Tags         Admin - Users
// @Accept       json
// @Produce      json
// @Param        id   path   string  true  "User ID (UUID)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "envelope {status,message,data:user}"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden (not admin)"
// @Failure      404  {object}  map[string]interface{}  "User not found"
// @Failure      500  {object}  map[string]interface{}  "error response"
// @Router       /admin/users/{id} [get]
func AdminGetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := repository.GetUserByID(id)
	if err != nil {
		return helper.NotFound(c, "user not found")
	}
	return helper.APIResponse(c, fiber.StatusOK, "user retrieved", user)
}

// AdminCreateUser godoc
// @Summary      Create new user (admin)
// @Description  Admin membuat user baru. Password akan di-hash sebelum disimpan.
// @Tags         Admin - Users
// @Accept       json
// @Produce      json
// @Param        body  body   models.CreateUserRequest  true  "User payload"
// @Security     BearerAuth
// @Success      201  {object}  map[string]interface{}  "envelope {status,message,data:user}"
// @Failure      400  {object}  map[string]interface{}  "Validation error / invalid payload"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden (not admin)"
// @Failure      500  {object}  map[string]interface{}  "error response"
// @Router       /admin/users [post]
func AdminCreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Invalid request body")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return helper.BadRequest(c, err.Error())
	}

	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.InternalError(c, "Failed to hash password")
	}

	user := &models.User{
		Username:     req.Username,
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		RoleID:       req.RoleID,
		IsActive:     req.IsActive,
	}

	user, err = repository.CreateUser(user)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}

	return helper.APIResponse(c, fiber.StatusCreated, "User created successfully", user)
}

// AdminUpdateUser godoc
// @Summary      Update user (admin)
// @Description  Admin mengubah data user (email, username, full name, role, status aktif)
// @Tags         Admin - Users
// @Accept       json
// @Produce      json
// @Param        id    path   string  true  "User ID (UUID)"
// @Param        body  body   models.UpdateUserRequest  true  "User update payload"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "envelope {status,message,data:user}"
// @Failure      400  {object}  map[string]interface{}  "Invalid payload / email already used / invalid role"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden (not admin)"
// @Failure      404  {object}  map[string]interface{}  "User not found"
// @Failure      500  {object}  map[string]interface{}  "error response"
// @Router       /admin/users/{id} [put]
func AdminUpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.UpdateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "invalid request body")
	}

	// Cek email sudah dipakai user lain atau belum
	exists, err := repository.IsEmailExistsForOtherUser(id, req.Email)
	if err != nil {
		return helper.InternalError(c, "failed to check email")
	}
	if exists {
		return helper.BadRequest(c, "email already used by another user")
	}

	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		FullName: req.FullName,
		RoleID:   req.RoleID,
		IsActive: req.IsActive,
	}

	updatedUser, err := repository.UpdateUser(id, user)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}

	return helper.APIResponse(c, fiber.StatusOK, "user updated", updatedUser)
}

// AdminDeleteUser godoc
// @Summary      Delete user (admin)
// @Description  Admin menghapus user berdasarkan ID.
// @Tags         Admin - Users
// @Accept       json
// @Produce      json
// @Param        id   path   string  true  "User ID (UUID)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "user deleted (envelope)"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden (not admin)"
// @Failure      404  {object}  map[string]interface{}  "user not found"
// @Failure      500  {object}  map[string]interface{}  "error response"
// @Router       /admin/users/{id} [delete]
func AdminDeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	log.Println("DELETE USER ID:", id)
	if err := repository.DeleteUser(id); err != nil {
		log.Println("DELETE ERROR:", err)
		return helper.NotFound(c, "user not found")
	}
	return helper.APIResponse(c, fiber.StatusOK, "user deleted", nil)
}

// AdminUpdateUserRole godoc
// @Summary      Update user role (admin)
// @Description  Admin mengubah role user (hanya role, tanpa mengubah field lain).
// @Tags         Admin - Users
// @Accept       json
// @Produce      json
// @Param        id    path   string  true  "User ID (UUID)"
// @Param        body  body   models.UpdateUserRoleRequest  true  "Role update payload"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "user role updated (envelope)"
// @Failure      400  {object}  map[string]interface{}  "invalid request body"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden (not admin)"
// @Failure      404  {object}  map[string]interface{}  "user not found"
// @Failure      500  {object}  map[string]interface{}  "error response"
// @Router       /admin/users/{id}/role [put]
func AdminUpdateUserRole(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.UpdateUserRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "invalid request body")
	}
	user, err := repository.UpdateUserRole(id, req.RoleID)
	if err != nil {
		return helper.NotFound(c, "user not found")
	}
	return helper.APIResponse(c, fiber.StatusOK, "user role updated", user)
}
