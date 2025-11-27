package service_test

import (
	"database/sql"
	"testing"

	"UAS_GO/app/service"
	// "UAS_GO/app/models"
	"UAS_GO/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// ===== Mock helper functions =====

// bcrypt helper
func hashPassword(t *testing.T, p string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(h)
}

func TestAuthLogin(t *testing.T) {

	t.Run("Valid_Email", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		database.PSQL = db

		email := "mhs1@example.com"
		pass := "123456"
		passHash := hashPassword(t, pass)
		roleID := "role-uuid-1"

		// Query user by email
		mock.ExpectQuery(`SELECT id, email, password_hash, role_id, is_active FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows(
				[]string{"id", "email", "password_hash", "role_id", "is_active"},
			).AddRow("user-uuid-1", email, passHash, roleID, true))

		// Permissions query
		mock.ExpectQuery(`SELECT p.name FROM role_permissions`).
			WithArgs(roleID).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).
				AddRow("achievement:create").
				AddRow("auth:profile"),
			)

		s := service.NewAuthService()
		resp, err := s.Login(email, pass, false)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Token)
		require.Equal(t, email, resp.User.Email)
		require.Equal(t, roleID, resp.User.RoleID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Valid_NIM", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		database.PSQL = db

		nim := "20230001"
		pass := "123456"
		passHash := hashPassword(t, pass)
		roleID := "role-uuid-2"

		// Query join students -> users
		mock.ExpectQuery(`SELECT u.id, u.email, u.password_hash, u.role_id, u.is_active FROM users u`).
			WithArgs(nim).
			WillReturnRows(sqlmock.NewRows(
				[]string{"id", "email", "password_hash", "role_id", "is_active"},
			).AddRow("user-uuid-2", "nimuser@example.com", passHash, roleID, true))

		mock.ExpectQuery(`SELECT p.name FROM role_permissions`).
			WithArgs(roleID).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("achievement:create"))

		s := service.NewAuthService()
		resp, err := s.Login(nim, pass, true)
		require.NoError(t, err)
		require.Equal(t, roleID, resp.User.RoleID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Wrong_Password", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		database.PSQL = db

		email := "mhs1@example.com"
		passHash := hashPassword(t, "WRONGPASS")

		mock.ExpectQuery(`SELECT id, email, password_hash, role_id, is_active FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows(
				[]string{"id", "email", "password_hash", "role_id", "is_active"},
			).AddRow("uuid-x", email, passHash, "role-x", true))

		s := service.NewAuthService()
		_, err := s.Login(email, "123456", false)
		require.Error(t, err)
		require.Equal(t, "invalid password", err.Error())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("User_Inactive", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		database.PSQL = db

		email := "inactive@example.com"
		passHash := hashPassword(t, "123456")

		mock.ExpectQuery(`SELECT id, email, password_hash, role_id, is_active FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows(
				[]string{"id", "email", "password_hash", "role_id", "is_active"},
			).AddRow("uuid-inactive", email, passHash, "role-inactive", false))

		s := service.NewAuthService()
		_, err := s.Login(email, "123456", false)

		require.Error(t, err)
		require.Contains(t, err.Error(), "inactive")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("User_Not_Found", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		database.PSQL = db

		mock.ExpectQuery(`SELECT id, email, password_hash, role_id, is_active FROM users WHERE email = \$1`).
			WithArgs("notfound@example.com").
			WillReturnError(sql.ErrNoRows)

		s := service.NewAuthService()
		_, err := s.Login("notfound@example.com", "whatever", false)

		require.Error(t, err)
		require.Equal(t, "user not found", err.Error())

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
