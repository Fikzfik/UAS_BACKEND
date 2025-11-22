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

    // pagination
    page := helper.GetIntQuery(c, "page", 1)
    limit := helper.GetIntQuery(c, "limit", 10)
    offset := (page - 1) * limit

    results, err := repository.GetAdviseeAchievementsByLecturerID(lecturerID, limit, offset)
    if err != nil {
        return helper.InternalError(c, err.Error())
    }

    return helper.APIResponse(c, 200, "Success", map[string]any{
        "page":    page,
        "limit":   limit,
        "results": results,
    })
}