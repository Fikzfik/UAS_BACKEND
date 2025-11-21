package repository

import (
	"UAS_GO/app/models"
	"UAS_GO/database"
	"context"
	"database/sql"
	"errors"
	"time"
)

func FindUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	err := database.PSQL.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.RoleID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}

func GetUserProfile(userID string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	err := database.PSQL.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.RoleID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}

func GetRoleNameByID(roleID string) (string, error) {
    var roleName string

    err := database.PSQL.QueryRow(
        "SELECT name FROM roles WHERE id = $1",
        roleID,
    ).Scan(&roleName)

    if err != nil {
        return "", err
    }

    return roleName, nil
}

func LogoutUser(userID string) error {
	// Token invalidation logic can be implemented here
	// For now, logout is handled client-side by removing the token
	return nil
}

func GetStudentIDByUserID(userID string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    query := `
        SELECT id 
        FROM students 
        WHERE user_id = $1
        LIMIT 1
    `

    var studentID string
    err := database.PSQL.QueryRowContext(ctx, query, userID).Scan(&studentID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return "", errors.New("student not found")
        }
        return "", err
    }

    return studentID, nil
}

func GetAchievementReferenceByMongoID(mongoID string) (*models.AchievementReference, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, verified_at, verified_by,
		       rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE mongo_achievement_id = $1
	`

	var r models.AchievementReference
	err := database.PSQL.QueryRowContext(ctx, query, mongoID).Scan(
		&r.ID, &r.StudentID, &r.MongoAchievementID, &r.Status,
		&r.SubmittedAt, &r.VerifiedAt, &r.VerifiedBy,
		&r.RejectionNote, &r.CreatedAt, &r.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &r, nil
}
