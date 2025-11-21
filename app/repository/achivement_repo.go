package repository

import (
	"UAS_GO/app/models"
	"UAS_GO/database"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func AchievementSoftDeleteMongo(id string) error {
	collection := database.MongoDB.Collection("achievements")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"isDeleted": true,
			"deletedAt": time.Now(),
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

func AchievementHardDeleteMongo(id string) error {
	collection := database.MongoDB.Collection("achievements")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
