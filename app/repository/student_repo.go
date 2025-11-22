package repository

import (
	"UAS_GO/database"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// GetAllStudents returns students list with optional advisorId and search term (q)
func GetAllStudents(advisorId string, q string) ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queryBase := `
		SELECT json_build_object(
			'id', s.id,
			'userId', s.user_id,
			'studentId', s.student_id,
			'programStudy', s.program_study,
			'academicYear', s.academic_year,
			'advisorId', s.advisor_id,
			'createdAt', s.created_at,
			'updatedAt', s.updated_at
		)
		FROM students s
		LEFT JOIN users u ON u.id = s.user_id
	`

	var where []string
	var args []any
	argIdx := 1

	if advisorId != "" {
		where = append(where, fmt.Sprintf("s.advisor_id = $%d", argIdx))
		args = append(args, advisorId)
		argIdx++
	}
	if q != "" {
		// search on user's full_name or email
		where = append(where, fmt.Sprintf("(u.full_name ILIKE $%d OR u.email ILIKE $%d)", argIdx, argIdx+1))
		args = append(args, "%"+q+"%", "%"+q+"%")
		argIdx += 2
	}

	query := queryBase
	if len(where) > 0 {
		query = query + " WHERE " + strings.Join(where, " AND ")
	}
	query = query + " ORDER BY s.created_at DESC"

	rows, err := database.PSQL.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var obj map[string]any
		if err := json.Unmarshal(raw, &obj); err != nil {
			return nil, err
		}
		out = append(out, obj)
	}
	return out, nil
}

// GetStudentByID returns a single student (raw json object) or nil if not found
func GetStudentByID(id string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT json_build_object(
			'id', s.id,
			'userId', s.user_id,
			'studentId', s.student_id,
			'programStudy', s.program_study,
			'academicYear', s.academic_year,
			'advisorId', s.advisor_id,
			'createdAt', s.created_at
		)
		FROM students s
		LEFT JOIN users u ON u.id = s.user_id
		WHERE s.id = $1
		LIMIT 1
	`

	var raw []byte
	err := database.PSQL.QueryRowContext(ctx, query, id).Scan(&raw)
	if err != nil {
		// pass error up (service will convert sql.ErrNoRows -> not found)
		return nil, err
	}

	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// GetAchievementReferencesByStudentID returns raw JSON array of references (optional status filter)
func GetAchievementReferencesByStudentID(studentID string, status string) ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT json_build_object(
			'id', r.id,
			'studentId', r.student_id,
			'mongoId', r.mongo_achievement_id,
			'status', r.status,
			'submittedAt', r.submitted_at,
			'verifiedAt', r.verified_at,
			'verifiedBy', r.verified_by,
			'rejectionNote', r.rejection_note,
			'createdAt', r.created_at,
			'updatedAt', r.updated_at
		)
		FROM achievement_references r
		WHERE r.student_id = $1
	`
	args := []any{studentID}
	if status != "" {
		query += " AND r.status = $2"
		args = append(args, status)
	}
	query += " ORDER BY r.created_at DESC"

	rows, err := database.PSQL.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var obj map[string]any
		if err := json.Unmarshal(raw, &obj); err != nil {
			return nil, err
		}
		out = append(out, obj)
	}
	return out, nil
}

// UpdateStudentAdvisor sets or unsets advisor for a student
func UpdateStudentAdvisor(studentID string, advisorId *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if advisorId == nil {
		res, err := database.PSQL.ExecContext(ctx, "UPDATE students SET advisor_id = NULL, updated_at = NOW() WHERE id = $1", studentID)
		if err != nil {
			return err
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			return errors.New("student not found")
		}
		return nil
	}

	res, err := database.PSQL.ExecContext(ctx, "UPDATE students SET advisor_id = $1, updated_at = NOW() WHERE id = $2", *advisorId, studentID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("student not found or not updated")
	}
	return nil
}
