package repository

import (
	"UAS_GO/app/models"
	"UAS_GO/database"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func GetAllUsers() ([]models.User, error) {
	query := `SELECT id,username, email,password_hash,full_name, role_id, is_active, created_at, updated_at FROM users`
	rows, err := database.PSQL.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func GetUserByID(id string) (*models.User, error) {
	query := `SELECT id,username, email,password_hash,full_name, role_id, is_active, created_at, updated_at FROM users WHERE id = $1`
	row := database.PSQL.QueryRow(query, id)

	var user models.User
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func CreateUser(user *models.User) (*models.User, error) {
	user.ID = uuid.New().String()

	query := `
		INSERT INTO users (id, username, full_name, email, password_hash, role_id, is_active, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW()) 
		RETURNING id, username, full_name, email, role_id, is_active, created_at, updated_at
	`

	row := database.PSQL.QueryRow(
		query,
		user.ID,
		user.Username,
		user.FullName,
		user.Email,
		user.PasswordHash,
		user.RoleID,
		user.IsActive,
	)

	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.FullName,
		&user.Email,
		&user.RoleID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}
	Rolenow, err := GetRoleNameByID(user.RoleID)
	if err != nil {
		return nil, err
	}
	// INSERT DOSEN
	if Rolenow == "dosen_wali" {

		nextID, err := generateNextLecturerID()
		if err != nil {
			return nil, errors.New("role tidak ada")
		}

		lecturerID := uuid.New().String()

		_, err = database.PSQL.Exec(`
			INSERT INTO lecturers (id, user_id, lecturer_id, department)
			VALUES ($1, $2, $3, $4)
		`, lecturerID, user.ID, nextID, "Teknik Informatika")

		if err != nil {
			return nil, err
		}
	}

	// INSERT MAHASISWA
	if Rolenow == "mahasiswa" {

		advisorID, err := getRandomAdvisorID()
		if err != nil {
			return nil, errors.New("no available advisor")
		}

		studentUUID := uuid.New().String()

		nextStudentID, err := generateNextStudentID()
		if err != nil {
			return nil, err
		}

		_, err = database.PSQL.Exec(`
			INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id)
			VALUES ($1, $2, $3, 'Teknik Informatika', '2023', $4)`,
			studentUUID, user.ID, nextStudentID, advisorID)

		if err != nil {
			return nil, err
		}

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

func generateNextLecturerID() (string, error) {
	var lastID string

	err := database.PSQL.QueryRow(`
		SELECT lecturer_id 
		FROM lecturers 
		ORDER BY lecturer_id DESC 
		LIMIT 1
	`).Scan(&lastID)

	if err != nil {
		// Tidak ada dosen → mulai dari DSN001
		if errors.Is(err, sql.ErrNoRows) {
			return "DSN001", nil
		}
		return "", err
	}

	// Ambil angka terakhir
	var num int
	fmt.Sscanf(lastID, "DSN%d", &num)

	next := fmt.Sprintf("DSN%03d", num+1)
	return next, nil
}

func getRandomAdvisorID() (string, error) {
	rows, err := database.PSQL.Query(`SELECT id FROM lecturers`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var list []string

	for rows.Next() {
		var id string
		rows.Scan(&id)
		list = append(list, id)
	}

	if len(list) == 0 {
		return "", errors.New("no lecturers available for advisor")
	}

	rand.Seed(time.Now().UnixNano())
	return list[rand.Intn(len(list))], nil
}

func generateNextStudentID() (string, error) {
	var lastID string

	err := database.PSQL.QueryRow(`
		SELECT student_id 
		FROM students 
		ORDER BY student_id DESC 
		LIMIT 1
	`).Scan(&lastID)

	if err != nil {
		// Tidak ada mahasiswa → mulai dari STU001
		if errors.Is(err, sql.ErrNoRows) {
			return "STU001", nil
		}
		return "", err
	}

	// Ambil angka dari format STU001
	var num int
	fmt.Sscanf(lastID, "STU%d", &num)

	next := fmt.Sprintf("STU%03d", num+1)
	return next, nil
}
