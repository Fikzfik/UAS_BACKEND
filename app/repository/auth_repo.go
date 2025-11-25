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
func GetPermissionsByRoleID(roleID string) ([]string, error) {
	if roleID == "" {
		return nil, errors.New("role id is empty")
	}

	rows, err := database.PSQL.Query(`
		SELECT p.name
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = $1
	`, roleID)
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

func GetLecturerIDByUserID(userID string) (string, error) {
    var lecturerID string

    err := database.PSQL.QueryRow(`
        SELECT id FROM lecturers WHERE user_id = $1 LIMIT 1
    `, userID).Scan(&lecturerID)

    if err != nil {
        return "", err
    }
    return lecturerID, nil
}

func GetAdviseeIDsByLecturer(lecturerID string) ([]string, error) {
	rows, err := database.PSQL.Query(`
        SELECT id FROM students WHERE advisor_id = $1
    `, lecturerID)

	if err != nil {
		return nil, err
	}

	list := []string{}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		list = append(list, id)
	}
	return list, nil
}

func IsStudentAdviseeOfLecturer(studentID, lecturerID string) (bool, error) {
	var exists bool
	err := database.PSQL.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM students 
			WHERE id = $1 AND advisor_id = $2
		)
	`, studentID, lecturerID).Scan(&exists)

	return exists, err
}


func RoleHasPermission(roleID string, permissionName string) (bool, error) {
    if roleID == "" || permissionName == "" {
        return false, errors.New("invalid args")
    }

    var exists bool
    query := `
        SELECT EXISTS(
            SELECT 1
            FROM role_permissions rp
            JOIN permissions p ON rp.permission_id = p.id
            WHERE rp.role_id = $1 AND p.name = $2
        )
    `
    err := database.PSQL.QueryRow(query, roleID, permissionName).Scan(&exists)
    if err != nil {
        return false, err
    }
    return exists, nil
}

// UserHasPermission: opsi jika kamu punya mekanisme permission per-user
func UserHasPermission(userID string, permissionName string) (bool, error) {
    if userID == "" || permissionName == "" {
        return false, errors.New("invalid args")
    }

    var exists bool
    // contoh query: ambil role_id dari users lalu cek role_permissions
    query := `
        SELECT EXISTS(
            SELECT 1
            FROM users u
            JOIN roles r ON u.role_id = r.id
            JOIN role_permissions rp ON rp.role_id = r.id
            JOIN permissions p ON p.id = rp.permission_id
            WHERE u.id = $1 AND p.name = $2
        )
    `
    err := database.PSQL.QueryRow(query, userID, permissionName).Scan(&exists)
    if err != nil {
        return false, err
    }
    return exists, nil
}