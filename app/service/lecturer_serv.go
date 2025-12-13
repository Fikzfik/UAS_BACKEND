package service

import (
	"UAS_GO/app/repository"
	"UAS_GO/helper"

	"github.com/gofiber/fiber/v2"
)

// GetAllLecturers godoc
// @Summary      Get all lecturers
// @Description  Mengambil daftar semua dosen.
// @Tags         Lecturers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "envelope {status,message,data} berisi array lecturers"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden"
// @Failure      500  {object}  map[string]interface{}  "error response"
// @Router       /lecturers [get]
func GetAllLecturers(c *fiber.Ctx) error {
	lects, err := repository.GetAllLecturers()
	if err != nil {
		return helper.InternalError(c, err.Error())
	}
	return helper.APIResponse(c, fiber.StatusOK, "Success", lects)
}

// GetLecturerAdvisees godoc
// @Summary      Get lecturer's advisees and achievements
// @Description  Mengambil daftar mahasiswa bimbingan seorang dosen beserta prestasinya (pagination).
// @Tags         Lecturers
// @Accept       json
// @Produce      json
// @Param        id     path   string  true   "Lecturer ID (UUID)"
// @Param        page   query  int     false  "Page number (default 1)"
// @Param        limit  query  int     false  "Items per page (default 10)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "envelope {status,message,data:{page,limit,results}}"
// @Failure      400  {object}  map[string]interface{}  "Invalid lecturer ID"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden (not advisor)"
// @Failure      404  {object}  map[string]interface{}  "Lecturer not found"
// @Failure      500  {object}  map[string]interface{}  "error response"
// @Router       /lecturers/{id}/advisees [get]
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

	return helper.APIResponse(c, fiber.StatusOK, "Success", map[string]any{
		"page":    page,
		"limit":   limit,
		"results": results,
	})
}
