package service

import (
	"UAS_GO/app/models"
	"UAS_GO/app/repository"
	"UAS_GO/helper"

	"github.com/gofiber/fiber/v2"
)


// GetAllStudents godoc
// @Summary      Get all students
// @Description  Mengambil daftar mahasiswa.
// @Description  Optional filter by advisorId dan free-text query.
// @Tags         Students
// @Accept       json
// @Produce      json
// @Param        advisorId  query  string  false  "Filter by advisor UUID"  format(uuid)
// @Param        q          query  string  false  "Free-text search (nama, NIM, dll)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "envelope {status,message,data}"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden"
// @Failure      500  {object}  map[string]interface{}  "error response"
// @Router       /students [get]
func GetAllStudents(c *fiber.Ctx) error {
	advisorId := c.Query("advisorId")
	q := c.Query("q")

	students, err := repository.GetAllStudents(advisorId, q)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}
	return helper.APIResponse(c, fiber.StatusOK, "Success", students)
}

// GetStudentByID godoc
// @Summary      Get student by ID
// @Description  Mengambil detail student berdasarkan ID (UUID).
// @Tags         Students
// @Accept       json
// @Produce      json
// @Param        id   path   string  true  "Student ID (UUID)"  format(uuid)
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "envelope {status,message,data}"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden"
// @Failure      404  {object}  map[string]interface{}  "Student not found"
// @Failure      500  {object}  map[string]interface{}  "error response"
// @Router       /students/{id} [get]
func GetStudentByID(c *fiber.Ctx) error {
	id := c.Params("id")

	s, err := repository.GetStudentByID(id)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}
	if s == nil {
		return helper.NotFound(c, "Student not found")
	}
	return helper.APIResponse(c, fiber.StatusOK, "Success", s)
}

// GetStudentAchievements godoc
// @Summary      Get student's achievements
// @Description  Mengambil achievement references milik student.
// @Description  Jika reference memiliki mongoId, maka akan di-enrich dengan dokumen MongoDB.
// @Tags         Students, Achievements
// @Accept       json
// @Produce      json
// @Param        id      path   string  true   "Student ID (UUID)"  format(uuid)
// @Param        status  query  string  false  "Optional status filter (draft, submitted, verified, rejected)"
// @Security     BearerAuth
// @Success      200  {array}   map[string]interface{}  "array of { reference: object, achievement: object|null }"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden"
// @Failure      404  {object}  map[string]interface{}  "Student not found"
// @Failure      500  {object}  map[string]interface{}  "error response"
// @Router       /students/{id}/achievements [get]
func GetStudentAchievements(c *fiber.Ctx) error {
	id := c.Params("id")
	status := c.Query("status") // optional

	// validate student exists
	s, err := repository.GetStudentByID(id)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}
	if s == nil {
		return helper.NotFound(c, "Student not found")
	}

	refs, err := repository.GetAchievementReferencesByStudentID(id, status)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}

	// enrich with Mongo docs (if available)
	results := make([]map[string]any, 0, len(refs))
	for _, ref := range refs {
		item := map[string]any{"reference": ref}

		// extract mongo id safely from ref map (key is "mongoId")
		var mongoID string
		if v, ok := ref["mongoId"]; ok && v != nil {
			// postgres json_build_object returns strings for IDs
			if sID, ok2 := v.(string); ok2 {
				mongoID = sID
			}
		}

		if mongoID != "" {
			ach, err := repository.GetAchievementByIdMongo(mongoID)
			if err == nil && ach != nil {
				item["achievement"] = map[string]any{
					"id":              ach.ID.Hex(),
					"title":           ach.Title,
					"description":     ach.Description,
					"achievementType": ach.AchievementType,
					"attachments":     ach.Attachments,
					"tags":            ach.Tags,
					"points":          ach.Points,
					"createdAt":       ach.CreatedAt,
					"updatedAt":       ach.UpdatedAt,
				}
			} else {
				item["achievement"] = nil
			}
		} else {
			item["achievement"] = nil
		}

		results = append(results, item)
	}

	return helper.APIResponse(c, fiber.StatusOK, "Success", results)
}

// UpdateStudentAdvisor godoc
// @Summary      Update student's advisor
// @Description  Update advisorId milik student.
// @Description  Body JSON: { "advisorId": "<uuid|null>" }.
// @Tags         Students
// @Accept       json
// @Produce      json
// @Param        id    path   string  true  "Student ID (UUID)"  format(uuid)
// @Param        body  body   models.UpdateStudentAdvisorRequest  true  "Payload: advisorId (uuid or null)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Student advisor updated (envelope)"
// @Failure      400  {object}  map[string]interface{}  "Bad request (invalid JSON/body validation)"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden"
// @Failure      500  {object}  map[string]interface{}  "error response"
// @Router       /students/{id}/advisor [put]
func UpdateStudentAdvisor(c *fiber.Ctx) error {
	id := c.Params("id")

	var body models.UpdateStudentAdvisorRequest
	if err := c.BodyParser(&body); err != nil {
		return helper.BadRequest(c, "Invalid JSON body")
	}

	// basic validation
	if body.AdvisorId != nil && len(*body.AdvisorId) == 0 {
		return helper.BadRequest(c, "advisorId must be non-empty UUID or null")
	}

	if err := repository.UpdateStudentAdvisor(id, body.AdvisorId); err != nil {
		return helper.InternalError(c, err.Error())
	}

	return helper.APIResponse(c, fiber.StatusOK, "Student advisor updated", nil)
}
