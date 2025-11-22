package service

import (
	"UAS_GO/app/repository"
	"UAS_GO/helper"

	"github.com/gofiber/fiber/v2"
)


func GetAllLecturers(c *fiber.Ctx) error {
	lects, err := repository.GetAllLecturers()
	if err != nil {
		return helper.InternalError(c, err.Error())
	}
	return helper.APIResponse(c, fiber.StatusOK, "Success", lects)
}

func GetLecturerAdvisees(c *fiber.Ctx) error {
	lecturerID := c.Params("id")
	if lecturerID == "" {
		return helper.BadRequest(c, "Lecturer id is required")
	}

	advisees, err := repository.GetAdviseesByLecturerID(lecturerID)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}
	return helper.APIResponse(c, fiber.StatusOK, "Success", advisees)
}
