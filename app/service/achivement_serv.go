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
