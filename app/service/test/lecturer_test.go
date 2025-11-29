// file: app/service/test/lecturer_test.go
package service_test

import (
	"UAS_GO/app/repository"
	"UAS_GO/app/service"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	bm "bou.ke/monkey"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// NOTE: repository functions return maps in this codebase:
// - GetAllLecturers() ([]map[string]interface{}, error)
// - GetAdviseeAchievementsByLecturerID(string, int, int) ([]map[string]interface{}, error)

func TestLecturerHandlers(t *testing.T) {
	app := fiber.New()

	app.Get("/lecturers", func(c *fiber.Ctx) error {
		return service.GetAllLecturers(c)
	})
	app.Get("/lecturers/:id/advisees", func(c *fiber.Ctx) error {
		return service.GetLecturerAdvisees(c)
	})

	t.Run("GetAllLecturers_Success", func(t *testing.T) {
		patch := bm.Patch(repository.GetAllLecturers,
			func() ([]map[string]interface{}, error) {
				return []map[string]interface{}{
					{
						"id":          "lec-1",
						"user_id":     "user-1",
						"lecturer_id": "L1",
						"department":  "Dept A",
						"created_at":  "2025-01-01T00:00:00Z",
					},
					{
						"id":          "lec-2",
						"user_id":     "user-2",
						"lecturer_id": "L2",
						"department":  "Dept B",
						"created_at":  "2025-01-02T00:00:00Z",
					},
				}, nil
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/lecturers", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		bodyBytes, _ := io.ReadAll(resp.Body)
		require.Contains(t, string(bodyBytes), "Dept A")
	})

	t.Run("GetAllLecturers_RepoError", func(t *testing.T) {
		patch := bm.Patch(repository.GetAllLecturers,
			func() ([]map[string]interface{}, error) {
				return nil, errors.New("db fail")
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/lecturers", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 500, resp.StatusCode)
	})

	t.Run("GetLecturerAdvisees_Success_DefaultPagination", func(t *testing.T) {
		patch := bm.Patch(repository.GetAdviseeAchievementsByLecturerID,
			func(lecturerID string, limit, offset int) ([]map[string]interface{}, error) {
				// validate input in test
				if lecturerID != "lec-1" {
					return nil, errors.New("wrong lecturer id")
				}
				if limit != 10 || offset != 0 {
					return nil, errors.New("unexpected pagination")
				}
				return []map[string]interface{}{
					{
						"id":         "507f1f77bcf86cd799439011",
						"studentId":  "stu-1",
						"title":      "A1",
						"points":     5,
						"createdAt":  "2025-01-01T00:00:00Z",
						"updatedAt":  "2025-01-02T00:00:00Z",
					},
				}, nil
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/lecturers/lec-1/advisees", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		b, _ := io.ReadAll(resp.Body)
		require.Contains(t, string(b), "A1")
	})

	t.Run("GetLecturerAdvisees_Success_CustomPagination", func(t *testing.T) {
		patch := bm.Patch(repository.GetAdviseeAchievementsByLecturerID,
			func(lecturerID string, limit, offset int) ([]map[string]interface{}, error) {
				if lecturerID != "lec-2" {
					return nil, errors.New("wrong lecturer id")
				}
				if limit != 5 || offset != 5 {
					return nil, errors.New("unexpected pagination")
				}
				return []map[string]interface{}{
					{
						"id":        "507f1f77bcf86cd799439012",
						"studentId": "stu-2",
						"title":     "A2",
					},
				}, nil
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/lecturers/lec-2/advisees?page=2&limit=5", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		b, _ := io.ReadAll(resp.Body)
		require.Contains(t, string(b), "A2")
	})

	t.Run("GetLecturerAdvisees_RepoError", func(t *testing.T) {
		patch := bm.Patch(repository.GetAdviseeAchievementsByLecturerID,
			func(lecturerID string, limit, offset int) ([]map[string]interface{}, error) {
				return nil, errors.New("db fail")
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/lecturers/lec-3/advisees", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 500, resp.StatusCode)
	})
}
