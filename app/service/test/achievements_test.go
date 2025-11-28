package service_test

import (
	"UAS_GO/app/models"
	"UAS_GO/app/repository"
	"UAS_GO/app/service"
	"UAS_GO/helper"
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	bm "bou.ke/monkey"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// helper membuat Fiber ctx
func makeCtx(app *fiber.App, method, path string, body []byte) *fiber.Ctx {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Jalankan request dan dapatkan response
	resp, err := app.Test(req)
	if err != nil {
		panic(err)
	}

	// Fiber tidak mengembalikan *fiber.Ctx dari Test,
	// jadi kita parsing ulang dengan fasthttp.
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	ctx.Request().Header.SetMethod(method)
	ctx.Request().SetRequestURI(path)
	ctx.Request().SetBody(body)

	// Copy response dari Test agar bisa dibaca
	ctx.Response().SetStatusCode(resp.StatusCode)

	return ctx
}

// helper buat ObjectID dari hex
func oid(hex string) primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex(hex)
	return id
}

///////////////////////////////////////////////////////////////////////////
// 1. TEST GET ALL ACHIEVEMENTS
///////////////////////////////////////////////////////////////////////////

func TestGetAllAchievements(t *testing.T) {
	app := fiber.New()

	t.Run("Success", func(t *testing.T) {

		// monkey patch repository
		patch := bm.Patch(repository.GetAllAchievements,
			func(studentId string, achType string) ([]models.Achievement, error) {
				return []models.Achievement{
					{
						ID:        oid("507f1f77bcf86cd799439011"),
						StudentID: "stu-1",
						Title:     "Achievement A",
					},
				}, nil
			},
		)
		defer patch.Unpatch()

		c := makeCtx(app, "GET", "/achievements?studentId=stu-1&type=academic", nil)
		err := service.GetAllAchievements(c)
		require.NoError(t, err)
		require.Equal(t, 200, c.Response().StatusCode())
	})

	t.Run("RepositoryError", func(t *testing.T) {

		patch := bm.Patch(repository.GetAllAchievements,
			func(studentId string, achType string) ([]models.Achievement, error) {
				return nil, fiber.ErrInternalServerError
			},
		)
		defer patch.Unpatch()

		c := makeCtx(app, "GET", "/achievements", nil)
		err := service.GetAllAchievements(c)
		require.NoError(t, err)
		require.Equal(t, 500, c.Response().StatusCode())
	})
}

///////////////////////////////////////////////////////////////////////////
// 2. TEST GET ACHIEVEMENT BY ID
///////////////////////////////////////////////////////////////////////////

func TestGetAchievementById(t *testing.T) {

	app := fiber.New()
	app.Get("/achievements/:id", service.GetAchievementById)

	t.Run("Success", func(t *testing.T) {

		p := bm.Patch(repository.GetAchievementById,
			func(id string) (*models.Achievement, error) {
				return &models.Achievement{
					ID:        oid("507f1f77bcf86cd799439011"),
					StudentID: "stu-1",
					Title:     "Demo",
				}, nil
			})
		defer p.Unpatch()

		req := httptest.NewRequest("GET",
			"/achievements/507f1f77bcf86cd799439011", nil)

		resp, err := app.Test(req)

		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("InvalidID", func(t *testing.T) {
		p := bm.Patch(repository.GetAchievementById,
			func(id string) (*models.Achievement, error) {
				return nil, fiber.ErrBadRequest
			})
		defer p.Unpatch()

		req := httptest.NewRequest("GET", "/achievements/abc", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		require.Equal(t, 500, resp.StatusCode)
	})

	t.Run("NotFound", func(t *testing.T) {
		p := bm.Patch(repository.GetAchievementById,
			func(id string) (*models.Achievement, error) {
				return nil, fiber.ErrNotFound
			})
		defer p.Unpatch()

		req := httptest.NewRequest("GET",
			"/achievements/507f1f77bcf86cd79943123", nil)

		resp, err := app.Test(req)

		require.NoError(t, err)
		require.Equal(t, 500, resp.StatusCode)
	})
}

///////////////////////////////////////////////////////////////////////////
// 3. TEST CREATE ACHIEVEMENT
///////////////////////////////////////////////////////////////////////////

func TestCreateAchievement(t *testing.T) {
	app := fiber.New()

	// MIDDLEWARE FIX (WAJIB!)
	app.Use(func(c *fiber.Ctx) error {
		if uid := c.Get("user_id"); uid != "" {
			c.Locals("user_id", uid)
		}
		return c.Next()
	})

	// ROUTE
	app.Post("/achievements", service.CreateAchievement)

	t.Run("Success", func(t *testing.T) {

		p1 := bm.Patch(repository.GetStudentIDByUserID,
			func(userID string) (string, error) {
				return "stu-1", nil
			})
		defer p1.Unpatch()

		p2 := bm.Patch(repository.AchievementInsertMongo,
			func(a *models.Achievement) (primitive.ObjectID, error) {
				return oid("507f1f77bcf86cd799439011"), nil
			})
		defer p2.Unpatch()

		p3 := bm.Patch(repository.AchievementInsertReference,
			func(studentID string, mongoID primitive.ObjectID) error {
				return nil
			})
		defer p3.Unpatch()

		body, _ := json.Marshal(map[string]any{"title": "New A"})

		req := httptest.NewRequest("POST", "/achievements", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("user_id", "user-1") // pakai middleware dummy

		// Fiber membaca locals dari ctx, kita isi via middleware
		app.Use(func(c *fiber.Ctx) error {
			uid := c.Get("user_id")
			if uid != "" {
				c.Locals("user_id", uid)
			}
			return c.Next()
		})

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 201, resp.StatusCode)
	})

	t.Run("Unauthorized", func(t *testing.T) {

		body, _ := json.Marshal(map[string]any{"title": "A"})
		req := httptest.NewRequest("POST", "/achievements", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 401, resp.StatusCode)
	})

	t.Run("MissingTitle", func(t *testing.T) {

		p := bm.Patch(repository.GetStudentIDByUserID,
			func(userID string) (string, error) {
				return "stu-1", nil
			})
		defer p.Unpatch()

		body, _ := json.Marshal(map[string]any{})
		req := httptest.NewRequest("POST", "/achievements", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("user_id", "x")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)
	})

	t.Run("ForbiddenField", func(t *testing.T) {

		p := bm.Patch(repository.GetStudentIDByUserID,
			func(userID string) (string, error) {
				return "stu-1", nil
			})
		defer p.Unpatch()

		body, _ := json.Marshal(map[string]any{"id": "123"})
		req := httptest.NewRequest("POST", "/achievements", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("user_id", "x")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)
	})
}
func makeReq(method, path string, body any) *http.Request {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestUpdateAchievement(t *testing.T) {
	app := fiber.New()
	app.Patch("/achievements/:id", func(c *fiber.Ctx) error {
		return service.UpdateAchievement(c)
	})

	t.Run("Success_UpdateTitle", func(t *testing.T) {

		p1 := bm.Patch(repository.GetStudentIDByUserID,
			func(uid string) (string, error) { return "stu-1", nil })
		defer p1.Unpatch()

		p2 := bm.Patch(repository.GetAchievementByIdMongo,
			func(id string) (*models.Achievement, error) {
				return &models.Achievement{
					ID:        oid(id),
					StudentID: "stu-1",
					Title:     "Old",
				}, nil
			})
		defer p2.Unpatch()

		p3 := bm.Patch(repository.AchievementUpdateMongoMap,
			func(id string, m map[string]any) error { return nil })
		defer p3.Unpatch()

		body := map[string]any{"title": "Updated"}

		req := makeReq("PATCH", "/achievements/507f1f77bcf86cd799439011", body)
		req.Header.Set("user_id", "user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("BlockedFieldPresent", func(t *testing.T) {

		p := bm.Patch(repository.GetStudentIDByUserID,
			func(uid string) (string, error) { return "stu-x", nil })
		defer p.Unpatch()

		body := map[string]any{"points": 123}

		req := makeReq("PATCH", "/achievements/507f1f77bcf86cd799439011", body)
		req.Header.Set("user_id", "user-x")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)
	})

	t.Run("NotOwner", func(t *testing.T) {

		p1 := bm.Patch(repository.GetStudentIDByUserID,
			func(uid string) (string, error) { return "stu-A", nil })
		defer p1.Unpatch()

		p2 := bm.Patch(repository.GetAchievementByIdMongo,
			func(id string) (*models.Achievement, error) {
				return &models.Achievement{
					ID:        oid(id),
					StudentID: "stu-B", // bukan owner
				}, nil
			})
		defer p2.Unpatch()

		body := map[string]any{"title": "AAA"}

		req := makeReq("PATCH", "/achievements/507f1f77bcf86cd799439022", body)
		req.Header.Set("user_id", "user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 403, resp.StatusCode)
	})

	t.Run("EmptyBody_NoUpdatableFields", func(t *testing.T) {

		p1 := bm.Patch(repository.GetStudentIDByUserID,
			func(uid string) (string, error) { return "stu-1", nil })
		defer p1.Unpatch()

		p2 := bm.Patch(repository.GetAchievementByIdMongo,
			func(id string) (*models.Achievement, error) {
				return &models.Achievement{ID: oid(id), StudentID: "stu-1"}, nil
			})
		defer p2.Unpatch()

		req := makeReq("PATCH", "/achievements/507f1f77bcf86cd799439066", map[string]any{})
		req.Header.Set("user_id", "user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)
	})

	t.Run("InvalidID_Format", func(t *testing.T) {

		p1 := bm.Patch(repository.GetStudentIDByUserID,
			func(uid string) (string, error) { return "stu-1", nil })
		defer p1.Unpatch()

		p2 := bm.Patch(repository.GetAchievementByIdMongo,
			func(id string) (*models.Achievement, error) {
				return nil, errors.New("string is not a valid ObjectID")
			})
		defer p2.Unpatch()

		body := map[string]any{"title": "X"}

		req := makeReq("PATCH", "/achievements/invalid-id-here", body)
		req.Header.Set("user_id", "user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)
	})

	t.Run("RepoUpdateError", func(t *testing.T) {

		p1 := bm.Patch(repository.GetStudentIDByUserID,
			func(uid string) (string, error) { return "stu-1", nil })
		defer p1.Unpatch()

		p2 := bm.Patch(repository.GetAchievementByIdMongo,
			func(id string) (*models.Achievement, error) {
				return &models.Achievement{ID: oid(id), StudentID: "stu-1"}, nil
			})
		defer p2.Unpatch()

		p3 := bm.Patch(repository.AchievementUpdateMongoMap,
			func(id string, m map[string]any) error { return errors.New("update failed") })
		defer p3.Unpatch()

		body := map[string]any{"title": "X"}

		req := makeReq("PATCH", "/achievements/507f1f77bcf86cd799439011", body)
		req.Header.Set("user_id", "user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 500, resp.StatusCode)
	})
}

/*** DELETE ACHIEVEMENT TESTS ***/
func DeleteAchievement(c *fiber.Ctx) error {
	id := c.Params("id") // mongoID

	// Ambil user_id dari JWT atau header (helper.GetUserID memeriksa header lalu locals)
	currentUserID := helper.GetUserID(c)
	if currentUserID == "" {
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
	if err := repository.AchievementSoftDeleteMongo(id); err != nil {
		return helper.InternalError(c, "Failed to delete achievement in MongoDB")
	}

	// 2) HAPUS reference Postgres PAKAI reference ID
	if err := repository.AchievementSoftDeleteReference(ref.ID); err != nil {
		return helper.InternalError(c, "Failed to delete achievement reference")
	}

	return helper.APIResponse(c, fiber.StatusOK, "Achievement deleted successfully", nil)
}

func makeMultipartReq(method, path, fieldName, fileName, contentType string, fileContent []byte) (*http.Request, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile(fieldName, fileName)
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(fileContent); err != nil {
		return nil, err
	}
	w.Close()

	req := httptest.NewRequest(method, path, &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, nil
}

/*** PART 3: Submit, Verify, Reject, Upload, History ***/

func TestSubmitVerifyRejectUploadHistory(t *testing.T) {
	app := fiber.New()

	// register routes (use same param names as service expects)
	app.Post("/achievements/:id/submit", func(c *fiber.Ctx) error {
		return service.SubmitAchievement(c)
	})
	app.Post("/achievements/:id/verify", func(c *fiber.Ctx) error {
		return service.VerifyAchievement(c)
	})
	app.Post("/achievements/:id/reject", func(c *fiber.Ctx) error {
		return service.RejectAchievement(c)
	})
	app.Post("/achievements/:id/upload", func(c *fiber.Ctx) error {
		return service.UploadAchievementFile(c)
	})
	app.Get("/achievements/:id/history", func(c *fiber.Ctx) error {
		return service.GetAchievementHistory(c)
	})

	// -------------------------------
	// 1) SubmitAchievement
	// -------------------------------
	t.Run("SubmitAchievement_Success", func(t *testing.T) {
		// Patch GetUserID mapping -> student id
		p1 := bm.Patch(repository.GetStudentIDByUserID,
			func(uid string) (string, error) { return "stu-1", nil })
		defer p1.Unpatch()

		// Patch GetAchievementByIdMongo -> owned by same student
		p2 := bm.Patch(repository.GetAchievementByIdMongo,
			func(id string) (*models.Achievement, error) {
				return &models.Achievement{
					ID:        oid(id),
					StudentID: "stu-1",
				}, nil
			})
		defer p2.Unpatch()

		// Patch GetAchievementReferenceByMongoID -> exists and draft
		p3 := bm.Patch(repository.GetAchievementReferenceByMongoID,
			func(mongoID string) (*models.AchievementReference, error) {
				now := time.Now()
				return &models.AchievementReference{
					ID:                 "ref-1",
					StudentID:          "stu-1",
					MongoAchievementID: mongoID,
					Status:             "draft",
					CreatedAt:          now,
					UpdatedAt:          now,
				}, nil
			})
		defer p3.Unpatch()

		// Patch AchievementUpdateMongoMap -> success (no-op)
		p4 := bm.Patch(repository.AchievementUpdateMongoMap,
			func(id string, updates map[string]any) error { return nil })
		defer p4.Unpatch()

		// Patch UpdateReferenceStatusSubmitted -> success
		p5 := bm.Patch(repository.UpdateReferenceStatusSubmitted,
			func(mongoID string) error { return nil })
		defer p5.Unpatch()

		// Buat request & set header user_id (helper.GetUserID membaca header)
		req := makeReq("POST", "/achievements/507f1f77bcf86cd799439011/submit", nil)
		req.Header.Set("user_id", "user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("SubmitAchievement_AlreadySubmitted", func(t *testing.T) {
		p1 := bm.Patch(repository.GetStudentIDByUserID,
			func(uid string) (string, error) { return "stu-1", nil })
		defer p1.Unpatch()

		// Mongo doc ownership ok
		p2 := bm.Patch(repository.GetAchievementByIdMongo,
			func(id string) (*models.Achievement, error) {
				return &models.Achievement{ID: oid(id), StudentID: "stu-1"}, nil
			})
		defer p2.Unpatch()

		// Reference already submitted
		p3 := bm.Patch(repository.GetAchievementReferenceByMongoID,
			func(mongoID string) (*models.AchievementReference, error) {
				now := time.Now()
				sa := now
				return &models.AchievementReference{
					ID:                 "ref-2",
					StudentID:          "stu-1",
					MongoAchievementID: mongoID,
					Status:             "submitted",
					SubmittedAt:        &sa,
					CreatedAt:          now,
					UpdatedAt:          now,
				}, nil
			})
		defer p3.Unpatch()

		req := makeReq("POST", "/achievements/507f1f77bcf86cd799439011/submit", nil)
		req.Header.Set("user_id", "user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		// service returns BadRequest when already submitted
		require.Equal(t, 400, resp.StatusCode)
	})

	// -------------------------------
	// 2) VerifyAchievement
	// -------------------------------
	t.Run("VerifyAchievement_Success", func(t *testing.T) {
		// lecturer user
		pL := bm.Patch(repository.GetLecturerIDByUserID,
			func(userID string) (string, error) { return "lec-1", nil })
		defer pL.Unpatch()

		// reference exists & status submitted
		pRef := bm.Patch(repository.GetAchievementReferenceByMongoID,
			func(mongoID string) (*models.AchievementReference, error) {
				now := time.Now()
				return &models.AchievementReference{
					ID:                 "ref-3",
					StudentID:          "stu-1",
					MongoAchievementID: mongoID,
					Status:             "submitted",
					SubmittedAt:        &now,
					CreatedAt:          now,
					UpdatedAt:          now,
				}, nil
			})
		defer pRef.Unpatch()

		// lecturer is advisor of student
		pAdvisor := bm.Patch(repository.IsLecturerAdvisorOfStudent,
			func(lecturerID, studentID string) (bool, error) { return true, nil })
		defer pAdvisor.Unpatch()

		// VerifyAchievementMongo -> success
		pVM := bm.Patch(repository.VerifyAchievementMongo,
			func(id string, points int, dosenID string) error { return nil })
		defer pVM.Unpatch()

		// VerifyAchievementReference -> success
		pVR := bm.Patch(repository.VerifyAchievementReference,
			func(refID string, dosenID string) error { return nil })
		defer pVR.Unpatch()

		// body: points
		body := map[string]any{"points": 10}
		req := makeReq("POST", "/achievements/507f1f77bcf86cd799439011/verify", body)
		req.Header.Set("user_id", "lecturer-user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("VerifyAchievement_NotAdvisor", func(t *testing.T) {
		pL := bm.Patch(repository.GetLecturerIDByUserID,
			func(userID string) (string, error) { return "lec-2", nil })
		defer pL.Unpatch()

		pRef := bm.Patch(repository.GetAchievementReferenceByMongoID,
			func(mongoID string) (*models.AchievementReference, error) {
				now := time.Now()
				return &models.AchievementReference{
					ID:                 "ref-4",
					StudentID:          "stu-99",
					MongoAchievementID: mongoID,
					Status:             "submitted",
					CreatedAt:          now,
					UpdatedAt:          now,
				}, nil
			})
		defer pRef.Unpatch()

		pAdvisor := bm.Patch(repository.IsLecturerAdvisorOfStudent,
			func(lecturerID, studentID string) (bool, error) { return false, nil })
		defer pAdvisor.Unpatch()

		body := map[string]any{"points": 5}
		req := makeReq("POST", "/achievements/507f1f77bcf86cd799439011/verify", body)
		req.Header.Set("user_id", "lecturer-user-2")

		resp, err := app.Test(req)
		require.NoError(t, err)
		// forbidden
		require.Equal(t, 403, resp.StatusCode)
	})

	t.Run("VerifyAchievement_InvalidPoints", func(t *testing.T) {
		// lecturer id ok and advisor ok
		pL := bm.Patch(repository.GetLecturerIDByUserID,
			func(userID string) (string, error) { return "lec-3", nil })
		defer pL.Unpatch()
		pRef := bm.Patch(repository.GetAchievementReferenceByMongoID,
			func(mongoID string) (*models.AchievementReference, error) {
				now := time.Now()
				return &models.AchievementReference{
					ID:                 "ref-5",
					StudentID:          "stu-1",
					MongoAchievementID: mongoID,
					Status:             "submitted",
					CreatedAt:          now,
					UpdatedAt:          now,
				}, nil
			})
		defer pRef.Unpatch()
		pAdvisor := bm.Patch(repository.IsLecturerAdvisorOfStudent,
			func(lecturerID, studentID string) (bool, error) { return true, nil })
		defer pAdvisor.Unpatch()

		body := map[string]any{"points": 0} // invalid, service requires >0
		req := makeReq("POST", "/achievements/507f1f77bcf86cd799439011/verify", body)
		req.Header.Set("user_id", "lecturer-user-3")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)
	})

	// -------------------------------
	// 3) RejectAchievement
	// -------------------------------
	t.Run("RejectAchievement_Success", func(t *testing.T) {
		pL := bm.Patch(repository.GetLecturerIDByUserID,
			func(userID string) (string, error) { return "lec-4", nil })
		defer pL.Unpatch()

		pRef := bm.Patch(repository.GetAchievementReferenceByMongoID,
			func(mongoID string) (*models.AchievementReference, error) {
				now := time.Now()
				return &models.AchievementReference{
					ID:                 "ref-6",
					StudentID:          "stu-2",
					MongoAchievementID: mongoID,
					Status:             "submitted",
					CreatedAt:          now,
					UpdatedAt:          now,
				}, nil
			})
		defer pRef.Unpatch()

		pAdv := bm.Patch(repository.IsLecturerAdvisorOfStudent,
			func(lecturerID, studentID string) (bool, error) { return true, nil })
		defer pAdv.Unpatch()

		pRM := bm.Patch(repository.RejectAchievementMongo,
			func(id, note, dosenID string) error { return nil })
		defer pRM.Unpatch()
		pRR := bm.Patch(repository.RejectAchievementReference,
			func(refID, note, dosenID string) error { return nil })
		defer pRR.Unpatch()

		body := map[string]any{"note": "not sufficient"}
		req := makeReq("POST", "/achievements/507f1f77bcf86cd799439011/reject", body)
		req.Header.Set("user_id", "lecturer-user-4")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("RejectAchievement_MissingNote", func(t *testing.T) {
		pL := bm.Patch(repository.GetLecturerIDByUserID,
			func(userID string) (string, error) { return "lec-5", nil })
		defer pL.Unpatch()

		pRef := bm.Patch(repository.GetAchievementReferenceByMongoID,
			func(mongoID string) (*models.AchievementReference, error) {
				now := time.Now()
				return &models.AchievementReference{
					ID:                 "ref-7",
					StudentID:          "stu-2",
					MongoAchievementID: mongoID,
					Status:             "submitted",
					CreatedAt:          now,
					UpdatedAt:          now,
				}, nil
			})
		defer pRef.Unpatch()

		pAdv := bm.Patch(repository.IsLecturerAdvisorOfStudent,
			func(lecturerID, studentID string) (bool, error) { return true, nil })
		defer pAdv.Unpatch()

		body := map[string]any{} // missing note
		req := makeReq("POST", "/achievements/507f1f77bcf86cd799439011/reject", body)
		req.Header.Set("user_id", "lecturer-user-5")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 400, resp.StatusCode)
	})

	// -------------------------------
	// 4) UploadAchievementFile
	// -------------------------------
	t.Run("UploadAchievementFile_Success", func(t *testing.T) {
		// student auth + owner + existing doc
		pS := bm.Patch(repository.GetStudentIDByUserID,
			func(uid string) (string, error) { return "stu-u", nil })
		defer pS.Unpatch()

		pDoc := bm.Patch(repository.GetAchievementByIdMongo,
			func(id string) (*models.Achievement, error) {
				return &models.Achievement{ID: oid(id), StudentID: "stu-u"}, nil
			})
		defer pDoc.Unpatch()

		pAdd := bm.Patch(repository.AddAchievementAttachment,
			func(mongoID string, att models.Attachment) error { return nil })
		defer pAdd.Unpatch()

		// make multipart request (small text file)
		req, err := makeMultipartReq("POST", "/achievements/507f1f77bcf86cd799439011/upload",
			"file", "test.txt", "text/plain", []byte("hello"))
		require.NoError(t, err)
		req.Header.Set("user_id", "user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 201, resp.StatusCode)
	})

	t.Run("UploadAchievementFile_NotOwner", func(t *testing.T) {
		pS := bm.Patch(repository.GetStudentIDByUserID,
			func(uid string) (string, error) { return "stu-other", nil })
		defer pS.Unpatch()

		pDoc := bm.Patch(repository.GetAchievementByIdMongo,
			func(id string) (*models.Achievement, error) {
				return &models.Achievement{ID: oid(id), StudentID: "stu-owner"}, nil
			})
		defer pDoc.Unpatch()

		req, err := makeMultipartReq("POST", "/achievements/507f1f77bcf86cd799439011/upload",
			"file", "test.txt", "text/plain", []byte("hello"))
		require.NoError(t, err)
		req.Header.Set("user_id", "user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		// forbidden
		require.Equal(t, 403, resp.StatusCode)
	})

	// -------------------------------
	// 5) GetAchievementHistory
	// -------------------------------
	t.Run("GetAchievementHistory_Success", func(t *testing.T) {
		// reference present
		pRef := bm.Patch(repository.GetAchievementReferenceByMongoID,
			func(mongoID string) (*models.AchievementReference, error) {
				now := time.Now()
				sub := now
				ver := now
				note := "no"
				vby := "lec-1"
				return &models.AchievementReference{
					ID:                 "ref-h-1",
					StudentID:          "stu-1",
					MongoAchievementID: mongoID,
					Status:             "verified",
					SubmittedAt:        &sub,
					VerifiedAt:         &ver,
					VerifiedBy:         &vby,
					RejectionNote:      &note,
					CreatedAt:          now,
					UpdatedAt:          now,
				}, nil
			})
		defer pRef.Unpatch()

		// achievement doc with attachments
		pDoc := bm.Patch(repository.GetAchievementByIdMongo,
			func(id string) (*models.Achievement, error) {
				now := time.Now()
				return &models.Achievement{
					ID:        oid(id),
					StudentID: "stu-1",
					Title:     "T",
					Attachments: []models.Attachment{
						{
							FileName:   "a.txt",
							FileURL:    "/static/a.txt",
							FileType:   "text/plain",
							UploadedAt: now,
						},
					},
					Points:    5,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			})
		defer pDoc.Unpatch()

		req := makeReq("GET", "/achievements/507f1f77bcf86cd799439011/history", nil)
		// no auth required for history - but set header anyway
		req.Header.Set("user_id", "user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)

		// optional: check response body contains "history"
		bodyBytes, _ := io.ReadAll(resp.Body)
		require.True(t, strings.Contains(string(bodyBytes), "history"))
	})

	t.Run("GetAchievementHistory_ReferenceNotFound", func(t *testing.T) {
		pRef := bm.Patch(repository.GetAchievementReferenceByMongoID,
			func(mongoID string) (*models.AchievementReference, error) {
				return nil, errors.New("not found")
			})
		defer pRef.Unpatch()

		req := makeReq("GET", "/achievements/507f1f77bcf86cd799439011/history", nil)
		req.Header.Set("user_id", "user-1")

		resp, err := app.Test(req)
		require.NoError(t, err)
		// service returns NotFound
		require.Equal(t, 404, resp.StatusCode)
	})
}
