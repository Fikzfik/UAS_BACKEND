package service

import (
	"UAS_GO/app/models"
	"UAS_GO/app/repository"
	"UAS_GO/helper"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrNotFound = mongo.ErrNoDocuments

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
	// parse body as generic map first so we can detect forbidden fields
	var bodyMap map[string]any
	if err := c.BodyParser(&bodyMap); err != nil {
		return helper.BadRequest(c, "Invalid JSON body")
	}

	// list of forbidden fields that client must not send
	forbidden := []string{
		"_id", "id",
		"studentId", "student_id",
		"points",
		"status",
		"verifiedAt", "verified_at",
		"verifiedBy", "verified_by",
		"rejectionNote", "rejection_note",
		"createdAt", "created_at",
		"updatedAt", "updated_at",
	}

	// collect any forbidden fields present in request
	var present []string
	for _, f := range forbidden {
		if _, ok := bodyMap[f]; ok {
			present = append(present, f)
		}
	}
	if len(present) > 0 {
		return helper.BadRequest(c, fmt.Sprintf(
			"You are not allowed to set the following fields: %s",
			strings.Join(present, ", "),
		))
	}

	// Get authenticated user_id from JWT context
	currentUserID, ok := c.Locals("user_id").(string)
	if !ok || currentUserID == "" {
		return helper.Unauthorized(c, "Unauthorized")
	}

	// Convert user_id -> students.id (UUID) from Postgres
	studentID, err := repository.GetStudentIDByUserID(currentUserID)
	if err != nil {
		return helper.Forbidden(c, "Student profile not found")
	}

	// Convert bodyMap to models.Achievement
	var req models.Achievement
	b, err := json.Marshal(bodyMap)
	if err != nil {
		return helper.BadRequest(c, "Invalid request payload")
	}
	if err := json.Unmarshal(b, &req); err != nil {
		return helper.BadRequest(c, "Invalid request payload structure")
	}

	// Basic validation
	if strings.TrimSpace(req.Title) == "" {
		return helper.BadRequest(c, "title is required")
	}

	// Assign server-controlled fields
	req.StudentID = studentID
	now := time.Now()
	req.CreatedAt = now
	req.UpdatedAt = now
	req.Points = 0 // force points to zero, cannot be overridden

	// Insert to MongoDB
	mongoID, err := repository.AchievementInsertMongo(&req)
	if err != nil {
		return helper.InternalError(c, "Failed to create achievement")
	}

	// Insert reference to Postgres
	if err := repository.AchievementInsertReference(studentID, mongoID); err != nil {
		return helper.InternalError(c, "Failed to create achievement reference")
	}

	return helper.APIResponse(c, fiber.StatusCreated, "Achievement submitted",
		map[string]any{"id": mongoID})
}


// update achievement oleh mahasiswa (partial update)
func UpdateAchievement(c *fiber.Ctx) error {
	id := c.Params("id")

	var reqMap map[string]any
	if err := c.BodyParser(&reqMap); err != nil {
		return helper.BadRequest(c, "Invalid JSON body")
	}

	// blocked fields mahasiswa tidak boleh ubah
	blocked := []string{
		"_id", "id",
		"createdAt", "created_at",
		"studentId", "student_id",
		"points",
		"status",
		"verifiedAt", "verified_at",
		"verifiedBy", "verified_by",
		"rejectionNote", "rejection_note",
	}

	// cek apakah ada yang diblock
	var presentBlocked []string
	for _, f := range blocked {
		if _, ok := reqMap[f]; ok {
			presentBlocked = append(presentBlocked, f)
		}
	}
	if len(presentBlocked) > 0 {
		return helper.BadRequest(c, fmt.Sprintf(
			"You are not allowed to update the following fields: %s",
			strings.Join(presentBlocked, ", "),
		))
	}

	// menghapus yang sudah diblock
	for _, f := range blocked {
		delete(reqMap, f)
	}

	// ambil user_id dari JWT context
	currentUserID, ok := c.Locals("user_id").(string)
	if !ok || currentUserID == "" {
		return helper.Unauthorized(c, "Unauthorized")
	}

	// konversi user_id -> student.id (UUID) dari Postgres
	studentID, err := repository.GetStudentIDByUserID(currentUserID)
	if err != nil {
		// jika tidak ada profil mahasiswa, tolak akses
		return helper.Forbidden(c, "Student profile not found")
	}

	// ambil data achievement dari MongoDB
	existing, err := repository.GetAchievementByIdMongo(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return helper.NotFound(c, "Achievement not found")
		}
		// ObjectID invalid atau error lain -> anggap id invalid
		return helper.BadRequest(c, "Invalid ID format")
	}

	// cek kepemilikan: compare existing.StudentID (nilai di Mongo) dengan studentID dari Postgres
	if existing.StudentID != studentID {
		return helper.Forbidden(c, "You are not allowed to update this achievement")
	}

	// if after removals there is nothing to update, return informative error
	if len(reqMap) == 0 {
		return helper.BadRequest(c, "No updatable fields provided")
	}

	// set updatedAt
	reqMap["updatedAt"] = time.Now()

	// lakukan update partial
	if err := repository.AchievementUpdateMongoMap(id, reqMap); err != nil {
		// tangani ObjectID invalid
		if err.Error() == "string is not a valid ObjectID" || err.Error() == "the provided hex string is not a valid ObjectID" {
			return helper.NotFound(c, "Achievement not found: Invalid ID format")
		}
		if err == mongo.ErrNoDocuments {
			return helper.NotFound(c, "Achievement not found")
		}
		return helper.InternalError(c, "Failed to update achievement")
	}

	return helper.APIResponse(c, fiber.StatusOK, "Achievement updated successfully", nil)
}

// update achievement oleh mahasiswa (partial update)
func DeleteAchievement(c *fiber.Ctx) error {
	id := c.Params("id")

	// Ambil user_id dari JWT
	currentUserID, ok := c.Locals("user_id").(string)
	if !ok || currentUserID == "" {
		return helper.Unauthorized(c, "Unauthorized")
	}

	// Konversi user_id -> student.id
	studentID, err := repository.GetStudentIDByUserID(currentUserID)
	if err != nil {
		return helper.Forbidden(c, "Student profile not found")
	}

	// Ambil achievement dari MongoDB
	existing, err := repository.GetAchievementByIdMongo(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return helper.NotFound(c, "Achievement not found")
		}
		return helper.BadRequest(c, "Invalid ID format")
	}

	// Cek kepemilikan
	if existing.StudentID != studentID {
		return helper.Forbidden(c, "You are not allowed to delete this achievement")
	}

	// Cek status boleh dihapus hanya draft
	ref, err := repository.GetAchievementReferenceByMongoID(id)
	if err != nil {
		return helper.NotFound(c, "Achievement reference not found")
	}

	if ref.Status != "draft" {
		return helper.BadRequest(c, "Only draft achievements can be deleted")
	}

	// Soft delete Mongo
	if err := repository.AchievementSoftDeleteMongo(id); err != nil {
		return helper.InternalError(c, "Failed to delete achievement")
	}

	// Soft delete reference in Postgres
	if err := repository.AchievementHardDeleteMongo(id); err != nil {
		return helper.InternalError(c, "Failed to delete achievement reference")
	}

	return helper.APIResponse(c, fiber.StatusOK, "Achievement deleted successfully", nil)
}
