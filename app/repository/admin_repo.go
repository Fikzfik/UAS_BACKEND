package repository

import (
	"UAS_GO/app/models"
	"UAS_GO/database"
	"database/sql"
	"errors"
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
	query := `INSERT INTO users (id, email, password_hash, role_id, is_active, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) 
			  RETURNING id, email, role_id, is_active, created_at, updated_at`
	
	row := database.PSQL.QueryRow(query, user.ID, user.Email, user.PasswordHash, user.RoleID, user.IsActive)
	if err := row.Scan(&user.ID, &user.Email, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}

	return user, nil
}

func UpdateUser(id string, user *models.User) (*models.User, error) {
	query := `UPDATE users SET email = $1, role_id = $2, is_active = $3, updated_at = NOW() 
			  WHERE id = $4 
			  RETURNING id, email, role_id, is_active, created_at, updated_at`
	
	row := database.PSQL.QueryRow(query, user.Email, user.RoleID, user.IsActive, id)
	if err := row.Scan(&user.ID, &user.Email, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
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
