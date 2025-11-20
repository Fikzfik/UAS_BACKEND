package repository

import (
	"UAS_GO/app/models"
	"UAS_GO/database"
	"context"
	"time"


	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"
	// "database/sql"
	// "errors"
	// "github.com/google/uuid"
)

// Get all achievements with optional filters
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
		VALUES (uuid_generate_v4(), $1, $2, 'submitted', NOW(), NOW())
	`

	_, err := database.PSQL.Exec(query, studentID, mongoID.Hex())
	return err
}


