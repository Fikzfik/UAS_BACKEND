package repository

import (
	"UAS_GO/app/models"
	"UAS_GO/database"
	
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