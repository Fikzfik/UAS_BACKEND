// --- admin handlers tests (append to your existing *_test.go) ---
package service_test

import (
	"UAS_GO/app/models"
	"UAS_GO/app/repository"
	"UAS_GO/app/service"
	"UAS_GO/helper"
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	bm "bou.ke/monkey"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// helper makeReq already exists in your file; reuse it.

// Register admin routes and tests
func TestAdminHandlers(t *testing.T) {
	app := fiber.New()

	// register routes
	app.Get("/admin/users", func(c *fiber.Ctx) error { return service.AdminGetAllUsers(c) })
	app.Get("/admin/users/:id", func(c *fiber.Ctx) error { return service.AdminGetUserByID(c) })
	app.Post("/admin/users", func(c *fiber.Ctx) error { return service.AdminCreateUser(c) })
	app.Put("/admin/users/:id", func(c *fiber.Ctx) error { return service.AdminUpdateUser(c) })
	app.Delete("/admin/users/:id", func(c *fiber.Ctx) error { return service.AdminDeleteUser(c) })
	app.Put("/admin/users/:id/role", func(c *fiber.Ctx) error { return service.AdminUpdateUserRole(c) })

	t.Run("AdminGetAllUsers_Success", func(t *testing.T) {
		patch := bm.Patch(repository.GetAllUsers,
			func() ([]models.User, error) {
				return []models.User{
					{Username: "u1", Email: "u1@example.com"},
					{Username: "u2", Email: "u2@example.com"},
				}, nil
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/admin/users", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("AdminGetAllUsers_RepoError", func(t *testing.T) {
		patch := bm.Patch(repository.GetAllUsers,
			func() ([]models.User, error) {
				return nil, errors.New("db error")
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/admin/users", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 500, resp.StatusCode)
	})

	t.Run("AdminGetUserByID_Success", func(t *testing.T) {
		patch := bm.Patch(repository.GetUserByID,
			func(id string) (*models.User, error) {
				return &models.User{Username: "u1", Email: "u1@example.com"}, nil
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/admin/users/any-id", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("AdminGetUserByID_NotFound", func(t *testing.T) {
		patch := bm.Patch(repository.GetUserByID,
			func(id string) (*models.User, error) {
				return nil, errors.New("not found")
			})
		defer patch.Unpatch()

		req := httptest.NewRequest("GET", "/admin/users/any-id", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		// service returns 404 when repo error in this handler
		require.Equal(t, 404, resp.StatusCode)
	})

t.Run("AdminCreateUser_Success", func(t *testing.T) {
	// patch HashPassword => tidak perlu hashing nyata
	ph := bm.Patch(helper.HashPassword,
		func(password string) (string, error) {
			return "hashed-pass", nil
		})
	defer ph.Unpatch()

	// patch CreateUser => kembalikan user dengan ID
	pc := bm.Patch(repository.CreateUser,
		func(u *models.User) (*models.User, error) {
			u.ID = "new-id"
			return u, nil
		})
	defer pc.Unpatch()

	// --- gunakan snake_case (ubah jika models menggunakan camelCase) ---
	payload := map[string]any{
		"username":  "testuser",
		"full_name": "Test User",
		"email":     "test@example.com",
		"password":  "Secret123!",    // pastikan valid menurut validator (min length, dsb.)
		"role_id":   "role-student",
		"is_active": true,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/admin/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)


	require.Equal(t, 201, resp.StatusCode)
})


	t.Run("AdminCreateUser_ValidationError", func(t *testing.T) {
		// missing required fields -> validator should fail
		payload := map[string]any{
			"username": "", // invalid
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/admin/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		// service uses validator and returns 400 on validation error
		require.Equal(t, 400, resp.StatusCode)
	})

	t.Run("AdminUpdateUser_Success", func(t *testing.T) {
		// patch email check & role check & update
		p1 := bm.Patch(repository.IsEmailExistsForOtherUser,
			func(id string, email string) (bool, error) { return false, nil })
		defer p1.Unpatch()

		p2 := bm.Patch(repository.IsRoleExists,
			func(roleID string) (bool, error) { return true, nil })
		defer p2.Unpatch()

		p3 := bm.Patch(repository.UpdateUser,
			func(id string, u *models.User) (*models.User, error) {
				u.ID = id
				return u, nil
			})
		defer p3.Unpatch()

		payload := map[string]any{
			"username": "updated",
			"fullName": "Updated Name",
			"email":    "updated@example.com",
			"roleId":   "role-1",
			"isActive": true,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("PUT", "/admin/users/uid-1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("AdminUpdateUser_EmailAlreadyUsed", func(t *testing.T) {
		p1 := bm.Patch(repository.IsEmailExistsForOtherUser,
			func(id string, email string) (bool, error) { return true, nil })
		defer p1.Unpatch()

		payload := map[string]any{
			"username": "updated",
			"fullName": "Updated Name",
			"email":    "taken@example.com",
			"roleId":   "role-1",
			"isActive": true,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("PUT", "/admin/users/uid-1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		// service returns 400 when email exists
		require.Equal(t, 400, resp.StatusCode)
	})

	t.Run("AdminDeleteUser_Success", func(t *testing.T) {
		p := bm.Patch(repository.DeleteUser,
			func(id string) error { return nil })
		defer p.Unpatch()

		req := httptest.NewRequest("DELETE", "/admin/users/uid-1", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("AdminDeleteUser_NotFound", func(t *testing.T) {
		p := bm.Patch(repository.DeleteUser,
			func(id string) error { return errors.New("not found") })
		defer p.Unpatch()

		req := httptest.NewRequest("DELETE", "/admin/users/uid-1", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 404, resp.StatusCode)
	})

	t.Run("AdminUpdateUserRole_Success", func(t *testing.T) {
		p := bm.Patch(repository.UpdateUserRole,
			func(id string, roleID string) (*models.User, error) {
				return &models.User{ID: id, RoleID: roleID}, nil
			})
		defer p.Unpatch()

		body, _ := json.Marshal(map[string]any{"roleId": "role-2"})
		req := httptest.NewRequest("PUT", "/admin/users/uid-1/role", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
	})

	t.Run("AdminUpdateUserRole_NotFound", func(t *testing.T) {
		p := bm.Patch(repository.UpdateUserRole,
			func(id string, roleID string) (*models.User, error) {
				return nil, errors.New("not found")
			})
		defer p.Unpatch()

		body, _ := json.Marshal(map[string]any{"roleId": "role-x"})
		req := httptest.NewRequest("PUT", "/admin/users/uid-1/role", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 404, resp.StatusCode)
	})
}
