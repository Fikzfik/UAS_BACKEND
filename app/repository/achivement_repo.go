package repository

import (
	"UAS_GO/app/models"
	"UAS_GO/database"
	"context"
	"database/sql"
	"fmt"
	"strings"

	// "database/sql"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	// "database/sql"
	// "errors"
	// "github.com/google/uuid"
)

// Get all achievements with optional filters and status from PostgreSQL
func GetAllAchievements(studentId string, achType string) ([]models.Achievement, error) {
	collection := database.MongoDB.Collection("achievements")

	filter := bson.M{}

	if studentId != "" {
		filter["studentId"] = studentId
	}

	if achType != "" {
		filter["achievementType"] = achType
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	var results []models.Achievement
	if err := cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}

	// Fetch status from PostgreSQL achievement_references for each achievement
	for i := range results {
		mongoID := results[i].ID.Hex()

		// Query PostgreSQL to get the status from achievement_references
		var status sql.NullString
		query := `
			SELECT status 
			FROM achievement_references 
			WHERE mongo_achievement_id = $1
		`
		err := database.PSQL.QueryRow(query, mongoID).Scan(&status)

		// If reference exists, add status to the Details map
		if err == nil && status.Valid {
			if results[i].Details == nil {
				results[i].Details = make(map[string]any)
			}
			results[i].Details["status"] = status.String
		} else if err != nil && err != sql.ErrNoRows {
			// Only return error for actual errors, not for missing references
			return nil, fmt.Errorf("error fetching status for achievement %s: %w", mongoID, err)
		}
	}

	return results, nil
}

// Get achievement by its MongoDB ID
func GetAchievementById(id string) (*models.Achievement, error) {
	collection := database.MongoDB.Collection("achievements")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var result models.Achievement
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func AchievementInsertMongo(a *models.Achievement) (primitive.ObjectID, error) {
	collection := database.MongoDB.Collection("achievements")

	a.ID = primitive.NewObjectID()
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()

	_, err := collection.InsertOne(context.Background(), a)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return a.ID, nil
}

// Insert reference ke PostgreSQL
func AchievementInsertReference(studentID string, mongoID primitive.ObjectID) error {
	query := `
		INSERT INTO achievement_references 
		(id, student_id, mongo_achievement_id, status, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, 'draft', NOW(), NOW())
	`

	_, err := database.PSQL.Exec(query, studentID, mongoID.Hex())
	return err
}

var ErrNotFound = mongo.ErrNoDocuments

// GetAchievementByIdMongo returns the Achievement document by its hex id.
func GetAchievementByIdMongo(id string) (*models.Achievement, error) {
	collection := database.MongoDB.Collection("achievements")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// return the raw error so caller can detect invalid id format
		return nil, err
	}

	var a models.Achievement
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&a)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &a, nil
}

// AchievementUpdateMongoMap updates only the fields present in the updates map.
func AchievementUpdateMongoMap(id string, updates map[string]any) error {
	collection := database.MongoDB.Collection("achievements")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// validate ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// make sure prohibited fields are not changed
	delete(updates, "_id")
	delete(updates, "id")
	delete(updates, "createdAt")
	delete(updates, "created_at")
	delete(updates, "studentId")  // ownership shouldn't be changed here
	delete(updates, "student_id") // alternative key

	// ensure updatedAt is present
	updates["updatedAt"] = time.Now()

	// If updates is empty now (shouldn't happen), return an error
	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	updateDoc := bson.M{"$set": updates}
	filter := bson.M{"_id": objID}

	result, err := collection.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return err
	}

	// if no match => not found
	if result.MatchedCount == 0 {
		return ErrNotFound
	}

	// matched but ModifiedCount == 0 => fields were identical; treat as success
	return nil
}

// Get reference by mongo_achievement_id
func GetAchievementReferenceByMongoID(mongoID string) (*models.AchievementReference, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, verified_at, verified_by,
		       rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE mongo_achievement_id = $1
		LIMIT 1
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

// Hard delete reference by its reference UUID (NOT by mongoID)
func AchievementSoftDeleteReference(referenceID string) error {
	if referenceID == "" {
		return errors.New("reference id required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// pastikan record ada dan ambil status/updated_at semula (untuk rollback jika perlu)
	var prevStatus sql.NullString
	var prevUpdatedAt sql.NullTime
	err := database.PSQL.QueryRowContext(ctx,
		"SELECT status, updated_at FROM achievement_references WHERE id = $1", referenceID).
		Scan(&prevStatus, &prevUpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		return err
	}

	// Update menjadi soft-deleted
	res, err := database.PSQL.ExecContext(ctx,
		`UPDATE achievement_references
		 SET status = 'deleted',
		     deleted_at = NOW(),
		     updated_at = NOW()
		 WHERE id = $1`, referenceID)
	if err != nil {
		return err
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func AchievementSoftDeleteMongo(mongoID string) error {
	collection := database.MongoDB.Collection("achievements")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(mongoID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     "deleted",
			"deleted_at": time.Now(), // menandai kapan di-soft-delete
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func SubmitAchievement(referenceID string, studentID string) error {
	query := `
		UPDATE achievement_references
		SET status = 'submitted',
		    submitted_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
		AND student_id = $2
		RETURNING id
	`

	var id string
	err := database.PSQL.QueryRow(query, referenceID, studentID).Scan(&id)
	if err != nil {
		return err
	}

	return nil
}

func UpdateReferenceStatusSubmitted(mongoID string) error {
	query := `
		UPDATE achievement_references
		SET status = 'submitted',
		    submitted_at = NOW(),
		    updated_at = NOW()
		WHERE mongo_achievement_id = $1
	`

	_, err := database.PSQL.Exec(query, mongoID)
	return err
}

func VerifyAchievementMongo(id string, points int, dosenID string) error {
	collection := database.MongoDB.Collection("achievements")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"points":     points,
			"verifiedAt": time.Now(),
			"verifiedBy": dosenID,
			"updatedAt":  time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func VerifyAchievementReference(refID string, dosenID string) error {
	query := `
        UPDATE achievement_references
        SET status = 'verified',
            verified_at = NOW(),
            verified_by = $2,
            updated_at = NOW()
        WHERE id = $1
    `

	fmt.Println("DEBUG VerifyAchievementReference")
	fmt.Println("Query:", query)
	fmt.Println("refID:", refID)
	fmt.Println("dosenID:", dosenID)

	res, err := database.PSQL.Exec(query, refID, dosenID)
	if err != nil {
		fmt.Println("Postgres ERROR:", err.Error()) // ERROR ASLI di log
		return err
	}

	rows, err := res.RowsAffected()
	fmt.Println("RowsAffected:", rows, "err:", err)

	if rows == 0 {
		return errors.New("ERROR: no rows affected — reference ID not found OR lecturer not authorized")
	}

	return nil
}

func RejectAchievementMongo(id, note, dosenID string) error {
	collection := database.MongoDB.Collection("achievements")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"rejectionNote": note,
			"verifiedAt":    time.Now(),
			"verifiedBy":    dosenID,
			"updatedAt":     time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func RejectAchievementReference(refID string, note, dosenID string) error {
	query := `
        UPDATE achievement_references
        SET status = 'rejected',
            rejection_note = $2,
            verified_by = $3,
            verified_at = NOW(),
            updated_at = NOW()
        WHERE id = $1
    `
	res, err := database.PSQL.Exec(query, refID, note, dosenID)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("reference not found or not updated")
	}

	return nil
}

func IsLecturerAdvisorOfStudent(lecturerID string, studentID string) (bool, error) {
	var exists bool
	query := `
        SELECT EXISTS (
            SELECT 1 FROM students
            WHERE id = $1 AND advisor_id = $2
        )
    `
	err := database.PSQL.QueryRow(query, studentID, lecturerID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func AddAchievementAttachment(mongoID string, att models.Attachment) error {
	collection := database.MongoDB.Collection("achievements")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(mongoID)
	if err != nil {
		return err
	}

	attDoc := bson.M{
		"fileName":   att.FileName,
		"fileUrl":    att.FileURL,
		"fileType":   att.FileType,
		"uploadedAt": att.UploadedAt,
	}

	// Try push normally
	update := bson.M{
		"$push": bson.M{"attachments": attDoc},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	res, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err == nil {
		if res.MatchedCount == 0 {
			return errors.New("achievement not found")
		}
		return nil
	}

	// If attachments is null or not array → fix it
	if strings.Contains(err.Error(), "must be an array") ||
		strings.Contains(err.Error(), "is of type null") {

		fix := bson.M{
			"$set": bson.M{
				"attachments": []bson.M{},
				"updatedAt":   time.Now(),
			},
		}

		if _, fixErr := collection.UpdateOne(ctx, bson.M{"_id": objID}, fix); fixErr != nil {
			return fixErr
		}

		// Retry push
		res2, err2 := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err2 != nil {
			return err2
		}
		if res2.MatchedCount == 0 {
			return errors.New("achievement not found on retry")
		}
		return nil
	}

	return err
}
