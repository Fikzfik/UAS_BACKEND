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
	id := c.Params("id") // mongoID

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

	// Ambil document Mongo
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

	// Cek reference di Postgres
	ref, err := repository.GetAchievementReferenceByMongoID(id)
	if err != nil {
		return helper.InternalError(c, "Reference not found")
	}

	if ref.Status != "draft" {
		return helper.BadRequest(c, "Only draft achievements can be deleted")
	}

	// 1) HAPUS Mongo DULU
	if err := repository.AchievementHardDeleteMongo(id); err != nil {
		return helper.InternalError(c, "Failed to delete achievement in MongoDB")
	}

	// 2) HAPUS reference Postgres PAKAI reference ID
	if err := repository.AchievementHardDeleteReference(ref.ID); err != nil {
		return helper.InternalError(c, "Failed to delete achievement reference")
	}

	return helper.APIResponse(c, fiber.StatusOK, "Achievement deleted successfully", nil)
}

func SubmitAchievement(c *fiber.Ctx) error {
	id := c.Params("id") // mongo achievement ID

	// ambil user_id dari JWT context
	currentUserID, ok := c.Locals("user_id").(string)
	if !ok || currentUserID == "" {
		return helper.Unauthorized(c, "Unauthorized")
	}

	// konversi user_id -> student.id PostgreSQL
	studentID, err := repository.GetStudentIDByUserID(currentUserID)
	if err != nil {
		return helper.Forbidden(c, "Student profile not found")
	}

	// cek achievement di MongoDB
	existing, err := repository.GetAchievementByIdMongo(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return helper.NotFound(c, "Achievement not found")
		}
		return helper.BadRequest(c, "Invalid achievement ID")
	}

	// cek kepemilikan
	if existing.StudentID != studentID {
		return helper.Forbidden(c, "You are not allowed to submit this achievement")
	}

	// cek reference di PostgreSQL
	ref, err := repository.GetAchievementReferenceByMongoID(id)
	if err != nil {
		return helper.NotFound(c, "Achievement reference not found")
	}

	// tidak boleh submit ulang
	if ref.Status == "submitted" || ref.Status == "verified" {
		return helper.BadRequest(c, "Achievement already submitted")
	}

	// update MongoDB: hanya update updatedAt
	updateMongo := map[string]any{
		"updatedAt": time.Now(),
	}

	err = repository.AchievementUpdateMongoMap(id, updateMongo)
	if err != nil {
		return helper.InternalError(c, "Failed to update achievement in MongoDB")
	}

	// update PostgreSQL: status + submitted_at
	err = repository.UpdateReferenceStatusSubmitted(id)
	if err != nil {
		return helper.InternalError(c, "Failed to update achievement reference")
	}

	return helper.APIResponse(
		c,
		fiber.StatusOK,
		"Achievement submitted successfully",
		nil,
	)
}

func VerifyAchievement(c *fiber.Ctx) error {
	id := c.Params("id")

	// Get dosen user ID
	currentUserID, ok := c.Locals("user_id").(string)
	if !ok || currentUserID == "" {
		return helper.Unauthorized(c, "Unauthorized")
	}

	// Get lecturer ID (from users table → lecturers.user_id)
	lecturerID, err := repository.GetLecturerIDByUserID(currentUserID)
	if err != nil {
		return helper.Forbidden(c, "Lecturer profile not found")
	}

	// Get reference
	ref, err := repository.GetAchievementReferenceByMongoID(id)
	if err != nil {
		return helper.NotFound(c, "Achievement reference not found")
	}

	// Only submitted can be verified
	if ref.Status != "submitted" {
		return helper.BadRequest(c, "Only submitted achievements can be verified")
	}
	if ref.Status != "verified" {
		return helper.BadRequest(c, "this achievements has already verified")
	}

	// Verify dosen advisor harus wali mahasiswa
	isAdvisor, err := repository.IsLecturerAdvisorOfStudent(lecturerID, ref.StudentID)
	if err != nil {
		return helper.InternalError(c, "Error verifying advisor relationship")
	}
	if !isAdvisor {
		return helper.Forbidden(c, "You are not the academic advisor for this student")
	}

	// Parse points
	var body struct {
		Points int `json:"points"`
	}
	if err := c.BodyParser(&body); err != nil {
		return helper.BadRequest(c, "Invalid JSON")
	}
	if body.Points <= 0 {
		return helper.BadRequest(c, "Points must be > 0")
	}

	// Update Mongo
	if err := repository.VerifyAchievementMongo(id, body.Points, currentUserID); err != nil {
		return helper.InternalError(c, "Failed to update MongoDB")
	}
	fmt.Println("REF DEBUG:", ref.ID, ref.MongoAchievementID, ref.Status)
	// Update Postgres
	if err := repository.VerifyAchievementReference(ref.ID, currentUserID); err != nil {
		return helper.InternalError(c, err.Error())
	}

	return helper.APIResponse(c, 200, "Achievement verified", nil)
}

func RejectAchievement(c *fiber.Ctx) error {
	id := c.Params("id")

	// Get current user (dosen)
	currentUserID, ok := c.Locals("user_id").(string)
	if !ok || currentUserID == "" {
		return helper.Unauthorized(c, "Unauthorized")
	}

	// Convert → lecturer.id
	lecturerID, err := repository.GetLecturerIDByUserID(currentUserID)
	if err != nil {
		return helper.Forbidden(c, "Lecturer profile not found")
	}

	// Get reference
	ref, err := repository.GetAchievementReferenceByMongoID(id)
	if err != nil {
		return helper.NotFound(c, "Achievement reference not found")
	}

	if ref.Status != "submitted" {
		return helper.BadRequest(c, "Only submitted achievements can be rejected")
	}

	// advisor validation
	isAdvisor, err := repository.IsLecturerAdvisorOfStudent(lecturerID, ref.StudentID)
	if err != nil {
		return helper.InternalError(c, "Error verifying advisor relationship")
	}
	if !isAdvisor {
		return helper.Forbidden(c, "You are not the academic advisor for this student")
	}

	// Parse rejection note
	var body struct {
		Note string `json:"note"`
	}
	if err := c.BodyParser(&body); err != nil {
		return helper.BadRequest(c, "Invalid JSON")
	}
	if strings.TrimSpace(body.Note) == "" {
		return helper.BadRequest(c, "Rejection note is required")
	}

	// Update Mongo
	if err := repository.RejectAchievementMongo(id, body.Note, currentUserID); err != nil {
		return helper.InternalError(c, "Failed to update MongoDB")
	}

	// Update Postgres
	if err := repository.RejectAchievementReference(ref.ID, body.Note, currentUserID); err != nil {
		return helper.InternalError(c, "Failed to update reference")
	}

	return helper.APIResponse(c, 200, "Achievement rejected", nil)
}


