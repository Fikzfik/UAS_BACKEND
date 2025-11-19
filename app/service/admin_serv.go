package service

import (
	"UAS_GO/app/models"
	"UAS_GO/app/repository"
	"UAS_GO/helper"

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
		return helper.BadRequest(c, "invalid request body")
	}
	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.InternalError(c, "failed to hash password")
	}
	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		RoleID:       req.RoleID,
		IsActive:     true,
	}
	user, err = repository.CreateUser(user)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}
	return helper.APIResponse(c, 201, "user created", user)
}

func AdminUpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "invalid request body")
	}
	user := &models.User{
		Email:    req.Email,
		RoleID:   req.RoleID,
		IsActive: req.IsActive,
	}
	user, err := repository.UpdateUser(id, user)
	if err != nil {
		return helper.NotFound(c, "user not found")
	}
	return helper.APIResponse(c, 200, "user updated", user)
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
