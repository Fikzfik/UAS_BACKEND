package repository

import (
	"database/sql"
	"errors"
)

var DB *sql.DB

func SetDB(db *sql.DB) {
	DB = db
}

// GetUserPermissions returns slice of permission names for the user (via role -> role_permissions -> permissions)
func GetUserPermissions(userID string) ([]string, error) {
	if DB == nil {
		return nil, errors.New("db not initialized")
	}

	query := `
		SELECT p.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		JOIN role_permissions rp ON rp.role_id = r.id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE u.id = $1
	`
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		perms = append(perms, name)
	}
	return perms, nil
}

// UserHasPermission = convenience (kept for backward compatibility)
func UserHasPermission(userID string, permissionName string) (bool, error) {
	perms, err := GetUserPermissions(userID)
	if err != nil {
		return false, err
	}
	for _, p := range perms {
		if p == permissionName {
			return true, nil
		}
	}
	return false, nil
}
