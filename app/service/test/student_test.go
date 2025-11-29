// file: app/service/test/student_test.go
package service_test

import (
	"UAS_GO/app/models"
	"UAS_GO/app/repository"
	"UAS_GO/app/service"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	bm "bou.ke/monkey"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// helper buat ObjectID dari hex
func oid(hex string) primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex(hex)
	return id
}

// create a small app with student routes registered
func registerStudentApp() *fiber.App {
	app := fiber.New()
	app.Get("/students", func(c *fiber.Ctx) error {
		return service.GetAllStudents(c)
	})
	app.Get("/students/:id", func(c *fiber.Ctx) error {
		return service.GetStudentByID(c)
	})
	app.Get("/students/:id/achievements", func(c *fiber.Ctx) error {
		return service.GetStudentAchievements(c)
	})
	app.Put("/students/:id/advisor", func(c *fiber.Ctx) error {
		return service.UpdateStudentAdvisor(c)
	})
	return app
}

func TestStudentHandlers(t *testing.T) {
	app := registerStudentApp()

	// -------------------------
	// GetAllStudents
	// -------------------------
	t.Run("GetAllStudents_Success", func(t *testing.T) {
		patch := bm.Patch(repository.GetAllStudents,
			func(advisorId string, q string) ([]map[string]interface{}, error) {
				return []map[string]interface{}{
					{"id": "stu-1", "full_name": "Satu"},
					{"id": "stu-2", "full_name": "Dua"},
				}, nil
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/students", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		b, _ := io.ReadAll(resp.Body)
		require.Contains(t, string(b), "Satu")
	})

	t.Run("GetAllStudents_WithQuery", func(t *testing.T) {
		patch := bm.Patch(repository.GetAllStudents,
			func(advisorId string, q string) ([]map[string]interface{}, error) {
				// expect query passed through
				if q != "search" {
					return nil, nil
				}
				return []map[string]interface{}{
					{"id": "stu-x", "full_name": "FindMe"},
				}, nil
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/students?q=search", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		b, _ := io.ReadAll(resp.Body)
		require.Contains(t, string(b), "FindMe")
	})

	t.Run("GetAllStudents_RepoError", func(t *testing.T) {
		patch := bm.Patch(repository.GetAllStudents,
			func(advisorId string, q string) ([]map[string]interface{}, error) {
				return nil, errors.New("db fail")
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/students", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 500, resp.StatusCode)
	})

	// -------------------------
	// GetStudentByID
	// -------------------------
	t.Run("GetStudentByID_Success", func(t *testing.T) {
		patch := bm.Patch(repository.GetStudentByID,
			func(id string) (map[string]interface{}, error) {
				if id == "stu-1" {
					return map[string]interface{}{
						"id":        "stu-1",
						"full_name": "Satu",
						"email":     "satu@example.com",
					}, nil
				}
				return nil, nil
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/students/stu-1", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		b, _ := io.ReadAll(resp.Body)
		require.Contains(t, string(b), "satu@example.com")
	})

	t.Run("GetStudentByID_NotFound", func(t *testing.T) {
		patch := bm.Patch(repository.GetStudentByID,
			func(id string) (map[string]interface{}, error) {
				return nil, nil
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/students/unknown", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 404, resp.StatusCode)
	})

	t.Run("GetStudentByID_RepoError", func(t *testing.T) {
		patch := bm.Patch(repository.GetStudentByID,
			func(id string) (map[string]interface{}, error) {
				return nil, errors.New("db error")
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/students/stu-err", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 500, resp.StatusCode)
	})

	// -------------------------
	// GetStudentAchievements (enrich flow)
	// -------------------------
	t.Run("GetStudentAchievements_Success_WithMongo", func(t *testing.T) {
		// student exists
		pS := bm.Patch(repository.GetStudentByID,
			func(id string) (map[string]interface{}, error) {
				return map[string]interface{}{"id": id, "full_name": "Satu"}, nil
			})
		defer pS.Unpatch()

		// reference list includes a mongoId
		now := time.Now()
		pRefs := bm.Patch(repository.GetAchievementReferencesByStudentID,
			func(studentID string, status string) ([]map[string]interface{}, error) {
				return []map[string]interface{}{
					{
						"id":      "ref-1",
						"mongoId": "507f1f77bcf86cd799439011",
						"status":  "verified",
						"created_at": now,
					},
				}, nil
			})
		defer pRefs.Unpatch()

		// mongo document exists
		pAch := bm.Patch(repository.GetAchievementByIdMongo,
			func(id string) (*models.Achievement, error) {
				return &models.Achievement{
					ID:        oid("507f1f77bcf86cd799439011"),
					StudentID: "stu-1",
					Title:     "Achievement A",
					Attachments: []models.Attachment{
						{FileName: "a.txt", FileURL: "/static/a.txt", FileType: "text/plain", UploadedAt: now},
					},
					Points:    5,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			})
		defer pAch.Unpatch()

		req := httptest.NewRequest("GET", "/students/stu-1/achievements", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		b, _ := io.ReadAll(resp.Body)
		require.Contains(t, string(b), "Achievement A")
		require.Contains(t, string(b), "ref-1")
	})

	t.Run("GetStudentAchievements_Success_NoMongoDoc", func(t *testing.T) {
		// student exists
		pS := bm.Patch(repository.GetStudentByID,
			func(id string) (map[string]interface{}, error) {
				return map[string]interface{}{"id": id, "full_name": "Satu"}, nil
			})
		defer pS.Unpatch()

		// refs present but mongo doc missing
		pRefs := bm.Patch(repository.GetAchievementReferencesByStudentID,
			func(studentID string, status string) ([]map[string]interface{}, error) {
				return []map[string]interface{}{
					{"id": "ref-2", "mongoId": "507f1f77bcf86cd799439099", "status": "submitted"},
				}, nil
			})
		defer pRefs.Unpatch()

		// GetAchievementByIdMongo returns error (not found)
		pAch := bm.Patch(repository.GetAchievementByIdMongo,
			func(id string) (*models.Achievement, error) {
				return nil, errors.New("not found")
			})
		defer pAch.Unpatch()

		req := httptest.NewRequest("GET", "/students/stu-1/achievements", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		b, _ := io.ReadAll(resp.Body)
		// still should return reference and achievement=null in result
		require.Contains(t, string(b), "ref-2")
		require.Contains(t, string(b), "achievement")
	})

	t.Run("GetStudentAchievements_StudentNotFound", func(t *testing.T) {
		pS := bm.Patch(repository.GetStudentByID,
			func(id string) (map[string]interface{}, error) {
				return nil, nil
			})
		defer pS.Unpatch()

		req := httptest.NewRequest("GET", "/students/unknown/achievements", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 404, resp.StatusCode)
	})

	t.Run("GetStudentAchievements_RefsRepoError", func(t *testing.T) {
		pS := bm.Patch(repository.GetStudentByID,
			func(id string) (map[string]interface{}, error) {
				return map[string]interface{}{"id": id}, nil
			})
		defer pS.Unpatch()

		pRefs := bm.Patch(repository.GetAchievementReferencesByStudentID,
			func(studentID string, status string) ([]map[string]interface{}, error) {
				return nil, errors.New("db fail")
			})
		defer pRefs.Unpatch()

		req := httptest.NewRequest("GET", "/students/stu-1/achievements", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 500, resp.StatusCode)
	})

	// -------------------------
	// UpdateStudentAdvisor
	// -------------------------
	t.Run("UpdateStudentAdvisor_Success_Null", func(t *testing.T) {
		patch := bm.Patch(repository.UpdateStudentAdvisor,
			func(id string, advisorId *string) error {
				// expect advisorId == nil
				if advisorId != nil {
					return errors.New("expected nil")
				}
				return nil
			})
		defer patch.Unpatch()

		// send {"advisorId": null}
		body, _ := json.Marshal(map[string]any{"advisorId": nil})
		req := httptest.NewRequest("PUT", "/students/stu-1/advisor", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("UpdateStudentAdvisor_Success_Set", func(t *testing.T) {
		patch := bm.Patch(repository.UpdateStudentAdvisor,
			func(id string, advisorId *string) error {
				if advisorId == nil || *advisorId != "lec-1" {
					return errors.New("unexpected advisor id")
				}
				return nil
			})
		defer patch.Unpatch()

		body, _ := json.Marshal(map[string]any{"advisorId": "lec-1"})
		req := httptest.NewRequest("PUT", "/students/stu-1/advisor", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("UpdateStudentAdvisor_BadRequest_EmptyString", func(t *testing.T) {
		// body with advisorId empty string should be rejected by handler
		body, _ := json.Marshal(map[string]any{"advisorId": ""})
		req := httptest.NewRequest("PUT", "/students/stu-1/advisor", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)
	})

	t.Run("UpdateStudentAdvisor_RepoError", func(t *testing.T) {
		patch := bm.Patch(repository.UpdateStudentAdvisor,
			func(id string, advisorId *string) error {
				return errors.New("db fail")
			})
		defer patch.Unpatch()

		body, _ := json.Marshal(map[string]any{"advisorId": "lec-1"})
		req := httptest.NewRequest("PUT", "/students/stu-1/advisor", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 500, resp.StatusCode)
	})
}
