package service

import (
	"UAS_GO/app/repository"
	"UAS_GO/helper"

	"github.com/gofiber/fiber/v2"
)

// GetAllStudents handler
func GetAllStudents(c *fiber.Ctx) error {
	advisorId := c.Query("advisorId")
	q := c.Query("q")

	students, err := repository.GetAllStudents(advisorId, q)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}
	return helper.APIResponse(c, fiber.StatusOK, "Success", students)
}

// GetStudentByID handler
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

// GetStudentAchievements handler (enriches Postgres refs with Mongo doc)
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

// UpdateStudentAdvisor handler
func UpdateStudentAdvisor(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		AdvisorId *string `json:"advisorId"`
	}
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
