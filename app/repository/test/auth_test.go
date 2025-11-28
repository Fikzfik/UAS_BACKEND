package repository_test

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"UAS_GO/app/repository"
	"UAS_GO/database"

	"github.com/DATA-DOG/go-sqlmock"
)

func setupDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	// replace global DB used by repository
	database.PSQL = db
	return db, mock
}

func TestFindUserByEmail(t *testing.T) {
	db, mock := setupDB(t)
	defer db.Close()

	tests := []struct {
		name      string
		email     string
		setupMock func()
		wantErr   bool
	}{
		{
			name:  "Success",
			email: "alice@example.com",
			setupMock: func() {
				now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
				rows := sqlmock.NewRows([]string{
					"id", "username", "email", "password_hash", "full_name", "role_id", "is_active", "created_at", "updated_at",
				}).AddRow("u-1", "alice", "alice@example.com", "hash", "Alice", "r-1", true, now, now)

				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`)).WithArgs("alice@example.com").WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:  "NotFound",
			email: "noone@example.com",
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`)).WithArgs("noone@example.com").WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			u, err := repository.FindUserByEmail(tc.email)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.Email != tc.email {
				t.Fatalf("expected email %s got %s", tc.email, u.Email)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}

func TestGetUserProfile(t *testing.T) {
	db, mock := setupDB(t)
	defer db.Close()

	tests := []struct {
		name      string
		id        string
		setupMock func()
		wantErr   bool
	}{
		{
			name: "Success",
			id:   "u-1",
			setupMock: func() {
				now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
				rows := sqlmock.NewRows([]string{
					"id", "username", "email", "password_hash", "full_name", "role_id", "is_active", "created_at", "updated_at",
				}).AddRow("u-1", "bob", "bob@example.com", "hash", "Bob", "r-2", true, now, now)

				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`)).WithArgs("u-1").WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "NotFound",
			id:   "no-id",
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`)).WithArgs("no-id").WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			u, err := repository.GetUserProfile(tc.id)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.ID != tc.id {
				t.Fatalf("expected id %s got %s", tc.id, u.ID)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}

func TestGetRoleNameByID(t *testing.T) {
	db, mock := setupDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT name FROM roles WHERE id = $1")).WithArgs("r-1").
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("admin"))

	name, err := repository.GetRoleNameByID("r-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "admin" {
		t.Fatalf("expected admin, got %s", name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// you can add more subtests for other repository functions following the same pattern
