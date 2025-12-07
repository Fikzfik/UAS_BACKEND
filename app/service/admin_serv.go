package service

import (
	"UAS_GO/app/models"
	"UAS_GO/app/repository"
	"UAS_GO/helper"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AdminGetAllUsers godoc
// @Summary Get all users (admin)
// @Description Mengambil daftar semua user (hanya untuk admin).
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "envelope {status,message,data} berisi array user"
// @Failure 500 {object} map[string]interface{} "error response"
// @Router /admin/users [get]
func AdminGetAllUsers(c *fiber.Ctx) error {
	users, err := repository.GetAllUsers()
	if err != nil {
		return helper.InternalError(c, err.Error())
	}
	return helper.APIResponse(c, fiber.StatusOK, "users retrieved", users)
}

// AdminGetUserByID godoc
// @Summary Get user by ID (admin)
// @Description Mengambil detail user berdasarkan ID (UUID).
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]interface{} "envelope {status,message,data} berisi user"
// @Failure 404 {object} map[string]interface{} "user not found"
// @Failure 500 {object} map[string]interface{} "error response"
// @Router /admin/users/{id} [get]
func AdminGetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := repository.GetUserByID(id)
	if err != nil {
		return helper.NotFound(c, "user not found")
	}
	return helper.APIResponse(c, fiber.StatusOK, "user retrieved", user)
}

// AdminCreateUser godoc
// @Summary Create new user (admin)
// @Description Admin membuat user baru. Password akan di-hash sebelum disimpan.
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Param body body models.CreateUserRequest true "User payload"
// @Success 201 {object} map[string]interface{} "envelope {status,message,data} berisi user baru"
// @Failure 400 {object} map[string]interface{} "invalid request body / validation error"
// @Failure 500 {object} map[string]interface{} "error response"
// @Router /admin/users [post]
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
// @Summary Update user (admin)
// @Description Admin mengubah data user (email, username, full name, role, status aktif).
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param body body models.UpdateUserRequest true "User update payload"
// @Success 200 {object} map[string]interface{} "envelope {status,message,data} berisi user terupdate"
// @Failure 400 {object} map[string]interface{} "invalid request body / email already used / invalid role_id"
// @Failure 404 {object} map[string]interface{} "user not found"
// @Failure 500 {object} map[string]interface{} "error response"
// @Router /admin/users/{id} [put]
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

	validRole, err := repository.IsRoleExists(req.RoleID)
	if err != nil {
		return helper.InternalError(c, "failed to check role id")
	}
	if !validRole {
		return helper.BadRequest(c, "invalid role_id")
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
// @Summary Delete user (admin)
// @Description Admin menghapus user berdasarkan ID.
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]interface{} "user deleted (envelope)"
// @Failure 404 {object} map[string]interface{} "user not found"
// @Failure 500 {object} map[string]interface{} "error response"
// @Router /admin/users/{id} [delete]
func AdminDeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := repository.DeleteUser(id); err != nil {
		return helper.NotFound(c, "user not found")
	}
	return helper.APIResponse(c, fiber.StatusOK, "user deleted", nil)
}

// AdminUpdateUserRole godoc
// @Summary Update user role (admin)
// @Description Admin mengubah role user (hanya role, tanpa mengubah field lain).
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param body body models.UpdateUserRoleRequest true "Role update payload"
// @Success 200 {object} map[string]interface{} "user role updated (envelope)"
// @Failure 400 {object} map[string]interface{} "invalid request body"
// @Failure 404 {object} map[string]interface{} "user not found"
// @Failure 500 {object} map[string]interface{} "error response"
// @Router /admin/users/{id}/role [put]
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
