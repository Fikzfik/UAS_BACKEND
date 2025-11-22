package repository

import (
	"UAS_GO/database"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// === Global Statistics (Admin / Dosen / Mahasiswa) ===
func GetStatistics(filter bson.M) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.MongoDB.Collection("achievements")

	stats := make(map[string]any)

	// 1. Total per tipe
	typeAgg := []bson.M{
		{"$match": filter},
		{"$group": bson.M{"_id": "$achievementType", "count": bson.M{"$sum": 1}}},
	}
	cursor, err := collection.Aggregate(ctx, typeAgg)
	if err != nil {
		return nil, err
	}
	var typeResult []bson.M
	cursor.All(ctx, &typeResult)
	stats["typeDistribution"] = typeResult

	// 2. Total per bulan
	periodAgg := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id": bson.M{
				"year":  bson.M{"$year": "$createdAt"},
				"month": bson.M{"$month": "$createdAt"},
			},
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"_id.year": 1, "_id.month": 1}},
	}
	cursor, err = collection.Aggregate(ctx, periodAgg)
	if err != nil {
		return nil, err
	}
	var periodResult []bson.M
	cursor.All(ctx, &periodResult)
	stats["periodDistribution"] = periodResult

	// 3. Statistik tingkat kompetisi
	levelAgg := []bson.M{
		{"$match": filter},
		{"$group": bson.M{"_id": "$details.competitionLevel", "count": bson.M{"$sum": 1}}},
	}
	cursor, err = collection.Aggregate(ctx, levelAgg)
	if err != nil {
		return nil, err
	}
	var levelResult []bson.M
	cursor.All(ctx, &levelResult)
	stats["competitionLevelDistribution"] = levelResult

	return stats, nil
}

// === Student Specific Statistics ===
func GetStudentStatistics(studentID string) (map[string]any, error) {
	filter := bson.M{"studentId": studentID}
	return GetStatistics(filter)
}
