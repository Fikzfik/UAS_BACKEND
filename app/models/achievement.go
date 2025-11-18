package models

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Achievement struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    StudentID string             `bson:"studentId"`
    Title     string             `bson:"title"`
    Description string           `bson:"description"`
    AchievementType string       `bson:"achievementType"`
    Details map[string]any       `bson:"details"`
    Attachments []Attachment     `bson:"attachments"`
    Tags []string                `bson:"tags"`
    Points int                   `bson:"points"`
    CreatedAt time.Time          `bson:"createdAt"`
    UpdatedAt time.Time          `bson:"updatedAt"`
}

type Attachment struct {
	FileName    string `bson:"fileName"`
	FileURL     string `bson:"fileUrl"`
	FileType    string `bson:"fileType"`
	UploadedAt time.Time `bson:"uploadedAt"`
}