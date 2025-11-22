package repository

import (
	"UAS_GO/database"
	"context"
	"encoding/json"
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
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
func GetAdviseeAchievementsByLecturerID(lecturerID string, limit, offset int) ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// STEP 1 — ambil student IDs
	queryStudents := `
        SELECT id FROM students WHERE advisor_id = $1
    `
	rows, err := database.PSQL.QueryContext(ctx, queryStudents, lecturerID)
	if err != nil {
		return nil, err
	}

	var studentIDs []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		studentIDs = append(studentIDs, id)
	}
	rows.Close()

	if len(studentIDs) == 0 {
		return []map[string]any{}, nil // dosen tidak punya bimbingan
	}

	// STEP 2 — ambil achievement_references untuk student tersebut
	queryAchievements := `
        SELECT id, student_id, mongo_achievement_id, status,
               submitted_at, verified_at, verified_by, rejection_note,
               created_at, updated_at
        FROM achievement_references
        WHERE student_id = ANY($1)
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows2, err := database.PSQL.QueryContext(ctx, queryAchievements, pq.Array(studentIDs), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	type Reference struct {
		ID         string
		StudentID  string
		MongoID    string
		Status     string
		Submitted  *time.Time
		VerifiedAt *time.Time
		VerifiedBy *string
		Note       *string
		CreatedAt  time.Time
		UpdatedAt  time.Time
	}

	var refs []Reference

	for rows2.Next() {
		var r Reference
		err := rows2.Scan(
			&r.ID, &r.StudentID, &r.MongoID, &r.Status,
			&r.Submitted, &r.VerifiedAt, &r.VerifiedBy, &r.Note,
			&r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		refs = append(refs, r)
	}

	// STEP 3 — fetch detail dari MongoDB
	coll := database.MongoDB.Collection("achievements")
	var out []map[string]any

	for _, r := range refs {
		var achievement map[string]any
		objectID, err := primitive.ObjectIDFromHex(r.MongoID)
		if err != nil {
			continue // skip jika ID invalid
		}

		err = coll.FindOne(ctx, bson.M{"_id": objectID}).Decode(&achievement)
		if err != nil {
			continue // skip jika tidak ditemukan
		}

		// STEP 4 — merge Postgres + MongoDB
		merged := map[string]any{
			"referenceId":   r.ID,
			"studentId":     r.StudentID,
			"mongoId":       r.MongoID,
			"status":        r.Status,
			"submittedAt":   r.Submitted,
			"verifiedAt":    r.VerifiedAt,
			"verifiedBy":    r.VerifiedBy,
			"rejectionNote": r.Note,
			"createdAt":     r.CreatedAt,
			"updatedAt":     r.UpdatedAt,
			"achievement":   achievement, // nested object dari MongoDB
		}

		out = append(out, merged)
	}

	return out, nil
}
