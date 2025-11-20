package service

import (
	"UAS_GO/app/models"
	"UAS_GO/app/repository"
	"UAS_GO/helper"
	"errors"

	"github.com/gofiber/fiber/v2"
)

func GetAllAchievements(c *fiber.Ctx) error {
	studentId := c.Query("studentId")
	achType := c.Query("type")

	data, err := repository.GetAllAchievements(studentId, achType)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}

	return helper.APIResponse(c, 200, "Success", data)
}

func GetAchievementById(c *fiber.Ctx) error {
	id := c.Params("id")
	data, err := repository.GetAchievementById(id)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}	
	return helper.APIResponse(c, 200, "Success", data)
}

func CreateAchievement(c *fiber.Ctx) error {
	var req models.Achievement

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Invalid JSON body")
	}

	// Validasi minimal
	if req.StudentID == "" || req.Title == "" {
		return errors.New("studentId and title are required")
	}

	// Insert ke MongoDB
	mongoID, err := repository.AchievementInsertMongo(&req)
	if err != nil {
		return err
	}

	// Insert reference ke PostgreSQL
	err = repository.AchievementInsertReference(req.StudentID, mongoID)
	if err != nil {
		return err
	}

	return helper.APIResponse(c, 201, "Achievement submitted", nil)
}

// func UpdateAchievement(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	var req models.Achievement
// 	if err := c.BodyParser(&req); err != nil {
// 		return helper.BadRequest(c, "Invalid JSON body")
// 	}
// 	// Implementasi update achievement di MongoDB
	
// 	// ...
// 	return helper.APIResponse(c, 200, "Achievement updated", nil)
// }

func awdawd(c *fiber.Ctx) error {
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