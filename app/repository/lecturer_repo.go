package repository

import (
	"UAS_GO/database"
	"context"
	"encoding/json"
	"time"
)

// GetAllLecturers returns list of lecturers as raw JSON objects
func GetAllLecturers() ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
    SELECT json_build_object(
        'id', l.id,
        'userId', l.user_id,
        'lecturerId', l.lecturer_id,
        'department', l.department,
        'createdAt', l.created_at
    )
    FROM lecturers l
    ORDER BY l.created_at DESC
`

	rows, err := database.PSQL.QueryContext(ctx, query)
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

// GetAdviseesByLecturerID returns students where advisor_id = lecturerID (raw JSON)
func GetAdviseesByLecturerID(lecturerID string) ([]map[string]any, error) {
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
        WHERE s.advisor_id = $1
        ORDER BY s.created_at DESC
    `

    rows, err := database.PSQL.QueryContext(ctx, query, lecturerID)
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

