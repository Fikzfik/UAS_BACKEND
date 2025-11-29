// file: app/service/test/statistics_test.go
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
	"go.mongodb.org/mongo-driver/bson" // <- gunakan bson.M
)

// middleware helper: copy headers "role" and "user_id" into locals
func registerAppWithAuthLocals() *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		if r := c.Get("role"); r != "" {
			c.Locals("role", r)
		}
		if uid := c.Get("user_id"); uid != "" {
			c.Locals("user_id", uid)
		}
		return c.Next()
	})
	// register handlers
	app.Get("/stats/global", func(c *fiber.Ctx) error {
		return service.GetGlobalStatistics(c)
	})
	app.Get("/stats/students/:id", func(c *fiber.Ctx) error {
		return service.GetStudentReport(c)
	})
	return app
}

func TestStatisticsHandlers(t *testing.T) {
	app := registerAppWithAuthLocals()

	// --- GLOBAL PATCH: stub GetStatistics so no test hits real MongoDB ---
	// This prevents nil-pointer panic if any repo function (directly or indirectly)
	// calls GetStatistics during tests.
pSS := bm.Patch(repository.GetStudentStatistics,
    func(studentID string) (map[string]interface{}, error) {
        return map[string]interface{}{"student_id": studentID, "points": 12}, nil
    })
defer pSS.Unpatch()


	t.Run("GetGlobalStatistics_AdminSuccess", func(t *testing.T) {
		// Patch GetStatistics for this subtest to return meaningful data
		pStats := bm.Patch(repository.GetStatistics,
			func(filter bson.M) (map[string]interface{}, error) {
				return map[string]interface{}{
					"total_achievements": 100,
					"total_students":     50,
				}, nil
			})
		defer pStats.Unpatch()

		req := httptest.NewRequest("GET", "/stats/global", nil)
		req.Header.Set("role", "admin")
		req.Header.Set("user_id", "admin-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		b, _ := io.ReadAll(resp.Body)
		require.Contains(t, string(b), "total_achievements")
	})

	t.Run("GetGlobalStatistics_DosenWaliSuccess", func(t *testing.T) {
		pL := bm.Patch(repository.GetLecturerIDByUserID,
			func(userID string) (string, error) {
				if userID == "lec-user-1" {
					return "lec-1", nil
				}
				return "", errors.New("not found")
			})
		defer pL.Unpatch()

		pAdv := bm.Patch(repository.GetAdviseeIDsByLecturer,
			func(lecturerID string) ([]string, error) {
				if lecturerID == "lec-1" {
					return []string{"stu-1", "stu-2"}, nil
				}
				return nil, nil
			})
		defer pAdv.Unpatch()

		pStats := bm.Patch(repository.GetStatistics,
			func(filter bson.M) (map[string]interface{}, error) {
				return map[string]interface{}{
					"total_achievements": 10,
					"scope":              "advisees",
				}, nil
			})
		defer pStats.Unpatch()

		req := httptest.NewRequest("GET", "/stats/global", nil)
		req.Header.Set("role", "dosen_wali")
		req.Header.Set("user_id", "lec-user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		b, _ := io.ReadAll(resp.Body)
		require.Contains(t, string(b), "advisees")
	})

	t.Run("GetGlobalStatistics_DosenWali_LecturerNotFound", func(t *testing.T) {
		pL := bm.Patch(repository.GetLecturerIDByUserID,
			func(userID string) (string, error) {
				return "", errors.New("not found")
			})
		defer pL.Unpatch()

		req := httptest.NewRequest("GET", "/stats/global", nil)
		req.Header.Set("role", "dosen_wali")
		req.Header.Set("user_id", "lec-user-x")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 403, resp.StatusCode)
	})

	t.Run("GetGlobalStatistics_MahasiswaSuccess", func(t *testing.T) {
		pS := bm.Patch(repository.GetStudentIDByUserID,
			func(userID string) (string, error) {
				if userID == "stu-user-1" {
					return "stu-1", nil
				}
				return "", errors.New("not found")
			})
		defer pS.Unpatch()

		pStats := bm.Patch(repository.GetStatistics,
			func(filter bson.M) (map[string]interface{}, error) {
				return map[string]interface{}{
					"total_achievements": 2,
					"scope":              "self",
				}, nil
			})
		defer pStats.Unpatch()

		req := httptest.NewRequest("GET", "/stats/global", nil)
		req.Header.Set("role", "mahasiswa")
		req.Header.Set("user_id", "stu-user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		b, _ := io.ReadAll(resp.Body)
		require.Contains(t, string(b), "self")
	})

	t.Run("GetGlobalStatistics_RepoError", func(t *testing.T) {
		pStats := bm.Patch(repository.GetStatistics,
			func(filter bson.M) (map[string]interface{}, error) {
				return nil, errors.New("db error")
			})
		defer pStats.Unpatch()

		req := httptest.NewRequest("GET", "/stats/global", nil)
		req.Header.Set("role", "admin")
		req.Header.Set("user_id", "admin-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 500, resp.StatusCode)
	})

	// -------------------------
	// GetStudentReport
	// -------------------------
	t.Run("GetStudentReport_Success", func(t *testing.T) {
		// patch GetStudentStatistics so handler won't call real repo code
		pSS := bm.Patch(repository.GetStudentStatistics,
			func(studentID string) (map[string]interface{}, error) {
				if studentID == "stu-1" {
					return map[string]interface{}{
						"student_id": "stu-1",
						"points":     12,
					}, nil
				}
				return nil, errors.New("not found")
			})
		defer pSS.Unpatch()

		req := httptest.NewRequest("GET", "/stats/students/stu-1", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		b, _ := io.ReadAll(resp.Body)
		require.Contains(t, string(b), "points")
	})

	t.Run("GetStudentReport_RepoError", func(t *testing.T) {
		pSS := bm.Patch(repository.GetStudentStatistics,
			func(studentID string) (map[string]interface{}, error) {
				return nil, errors.New("db error")
			})
		defer pSS.Unpatch()

		req := httptest.NewRequest("GET", "/stats/students/stu-1", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 500, resp.StatusCode)
	})
}
