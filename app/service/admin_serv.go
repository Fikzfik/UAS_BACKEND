package service

import (
	"UAS_GO/app/models"
	"UAS_GO/app/repository"
	"UAS_GO/helper"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func AdminGetAllUsers(c *fiber.Ctx) error {
	users, err := repository.GetAllUsers()
	if err != nil {
		return helper.InternalError(c, err.Error())
	}
	return helper.APIResponse(c, 200, "users retrieved", users)
}

func AdminGetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := repository.GetUserByID(id)
	if err != nil {
		return helper.NotFound(c, "user not found")
	}
	return helper.APIResponse(c, 200, "user retrieved", user)
}
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

	return helper.APIResponse(c, 201, "User created successfully", user)
}

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

	return helper.APIResponse(c, 200, "user updated", updatedUser)
}

func AdminDeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := repository.DeleteUser(id); err != nil {
		return helper.NotFound(c, "user not found")
	}
	return helper.APIResponse(c, 200, "user deleted", nil)
}

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
	return helper.APIResponse(c, 200, "user role updated", user)
}
