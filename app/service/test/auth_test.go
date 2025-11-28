package service_test

import (
	"database/sql"

	"regexp"
	"testing"
	
	"UAS_GO/app/service"
	"UAS_GO/database"

	"UAS_GO/helper"
	"UAS_GO/app/models"

	"github.com/DATA-DOG/go-sqlmock"
	"bou.ke/monkey"
	"golang.org/x/crypto/bcrypt"
)

func setupDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	database.PSQL = db
	return db, mock
}

func TestAuthService_Login(t *testing.T) {
	// Patch helper.GenerateToken with the exact signature used in your code.
	// According to the panic, GenerateToken accepts models.User.
	monkey.Patch(helper.GenerateToken, func(u models.User) (string, error) {
		return "fixed-token", nil
	})
	// Patch CheckPassword (likely signature func(password, hash string) bool)
	monkey.Patch(helper.CheckPassword, func(password, hash string) bool {
		// Use bcrypt compare to allow creating hashed passwords in tests
		return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
	})
	defer monkey.UnpatchAll()

	db, mock := setupDB(t)
	defer db.Close()

	tests := []struct {
		name       string
		byNIM      bool
		identifier string
		password   string
		setupMock  func()
		wantErr    bool
		errText    string
	}{
		{
			name:       "UserNotFoundByEmail",
			byNIM:      false,
			identifier: "notfound@example.com",
			password:   "any",
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, password_hash, role_id, is_active FROM users WHERE email = $1`)).
					WithArgs("notfound@example.com").WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errText: "user not found",
		},
		{
			name:       "InvalidPassword",
			byNIM:      false,
			identifier: "bob@example.com",
			password:   "wrongpass",
			setupMock: func() {
				hashed, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, password_hash, role_id, is_active FROM users WHERE email = $1`)).
					WithArgs("bob@example.com").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "role_id", "is_active"}).
						AddRow("u-1", "bob@example.com", string(hashed), "r-1", true))
			},
			wantErr: true,
			errText: "invalid password",
		},
		{
			name:       "InactiveUser",
			byNIM:      false,
			identifier: "inactive@example.com",
			password:   "secret",
			setupMock: func() {
				hashed, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, password_hash, role_id, is_active FROM users WHERE email = $1`)).
					WithArgs("inactive@example.com").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "role_id", "is_active"}).
						AddRow("u-2", "inactive@example.com", string(hashed), "r-1", false))
			},
			wantErr: true,
			errText: "user account is inactive",
		},
		{
			name:       "Success",
			byNIM:      false,
			identifier: "alice@example.com",
			password:   "secret",
			setupMock: func() {
				hashed, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, password_hash, role_id, is_active FROM users WHERE email = $1`)).
					WithArgs("alice@example.com").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "role_id", "is_active"}).
						AddRow("u-1", "alice@example.com", string(hashed), "r-1", true))

				// repository.GetPermissionsByRoleID runs a SQL query inside Login -> mock it
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT p.name
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = $1
	`)).WithArgs("r-1").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("achievement:create"))
			},
			wantErr: false,
		},
	}

	svc := service.NewAuthService()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			resp, err := svc.Login(tc.identifier, tc.password, tc.byNIM)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tc.errText != "" && err.Error() != tc.errText {
					// allow substring match
					if !regexp.MustCompile(regexp.QuoteMeta(tc.errText)).MatchString(err.Error()) {
						t.Fatalf("expected error containing %q, got %v", tc.errText, err)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil || resp.Token == "" {
				t.Fatalf("expected valid token in response")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}
