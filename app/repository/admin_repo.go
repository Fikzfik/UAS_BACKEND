package repository

import (
	"UAS_GO/app/models"
	"UAS_GO/database"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

func GetAllUsers() ([]models.User, error) {
	query := `SELECT id, email, role_id, is_active, created_at, updated_at FROM users`
	rows, err := database.PSQL.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func GetUserByID(id string) (*models.User, error) {
	query := `SELECT id, email, role_id, is_active, created_at, updated_at FROM users WHERE id = $1`
	row := database.PSQL.QueryRow(query, id)

	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func CreateUser(user *models.User) (*models.User, error) {
	user.ID = uuid.New().String()

	query := `INSERT INTO users (id, username, full_name, email, password_hash, role_id, is_active, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW()) 
              RETURNING id, username, full_name, email, role_id, is_active, created_at, updated_at`

	row := database.PSQL.QueryRow(query, user.ID, user.Username, user.FullName, user.Email, user.PasswordHash, user.RoleID, user.IsActive)

	if err := row.Scan(&user.ID, &user.Username, &user.FullName, &user.Email, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}

	return user, nil
}

func UpdateUser(id string, user *models.User) (*models.User, error) {
	var row *sql.Row

	if user.PasswordHash != "" {
		// update termasuk password
		query := `UPDATE users 
                  SET email = $1, username = $2, full_name = $3, role_id = $4, is_active = $5, 
                      password_hash = $6, updated_at = NOW()
                  WHERE id = $7
                  RETURNING id, email, username, full_name, role_id, is_active, created_at, updated_at`

		row = database.PSQL.QueryRow(query,
			user.Email,
			user.Username,
			user.FullName,
			user.RoleID,
			user.IsActive,
			user.PasswordHash,
			id,
		)

	} else {
		// update tanpa password
		query := `UPDATE users 
                  SET email = $1, username = $2, full_name = $3, role_id = $4, is_active = $5,
                      updated_at = NOW()
                  WHERE id = $6
                  RETURNING id, email, username, full_name, role_id, is_active, created_at, updated_at`

		row = database.PSQL.QueryRow(query,
			user.Email,
			user.Username,
			user.FullName,
			user.RoleID,
			user.IsActive,
			id,
		)
	}

	if err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FullName,
		&user.RoleID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {

		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

func DeleteUser(id string) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := database.PSQL.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func UpdateUserRole(id string, roleID string) (*models.User, error) {
	query := `UPDATE users SET role_id = $1, updated_at = NOW() 
			  WHERE id = $2 
			  RETURNING id, email, role_id, is_active, created_at, updated_at`

	row := database.PSQL.QueryRow(query, roleID, id)
	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func IsEmailExistsForOtherUser(id, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(
        SELECT 1 FROM users WHERE email=$1 AND id <> $2
    )`
	err := database.PSQL.QueryRow(query, email, id).Scan(&exists)
	return exists, err
}

func IsRoleExists(roleID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM roles WHERE id=$1)`
	err := database.PSQL.QueryRow(query, roleID).Scan(&exists)
	return exists, err
}
