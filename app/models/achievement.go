package models

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Achievement struct {
    ID              primitive.ObjectID     `bson:"_id,omitempty"`
    StudentID       string                 `bson:"studentId"` // UUID string
    AchievementType string                 `bson:"achievementType"`
    Title           string                 `bson:"title"`
    Description     string                 `bson:"description"`

    Details         AchievementDetails      `bson:"details"`

    Attachments     []Attachment           `bson:"attachments"`
    Tags            []string               `bson:"tags"`
    Points          float64                `bson:"points"`

    CreatedAt       time.Time              `bson:"createdAt"`
    UpdatedAt       time.Time              `bson:"updatedAt"`
}

type AchievementDetails struct {
    CompetitionName  string    `bson:"competitionName,omitempty"`
    CompetitionLevel string    `bson:"competitionLevel,omitempty"`
    Rank             int       `bson:"rank,omitempty"`
    MedalType        string    `bson:"medalType,omitempty"`

    PublicationType  string    `bson:"publicationType,omitempty"`
    PublicationTitle string    `bson:"publicationTitle,omitempty"`
    Authors          []string  `bson:"authors,omitempty"`
    Publisher        string    `bson:"publisher,omitempty"`
    ISSN             string    `bson:"issn,omitempty"`

    OrganizationName string    `bson:"organizationName,omitempty"`
    Position         string    `bson:"position,omitempty"`
    Period           *Period   `bson:"period,omitempty"`

    CertificationName string   `bson:"certificationName,omitempty"`
    IssuedBy          string   `bson:"issuedBy,omitempty"`
    CertificationNumber string `bson:"certificationNumber,omitempty"`
    ValidUntil        time.Time `bson:"validUntil,omitempty"`

    EventDate         time.Time `bson:"eventDate,omitempty"`
    Location          string    `bson:"location,omitempty"`
    Organizer         string    `bson:"organizer,omitempty"`
    Score             float64   `bson:"score,omitempty"`

    CustomFields     map[string]interface{} `bson:"customFields,omitempty"`
}

type Period struct {
    Start time.Time `bson:"start"`
    End   time.Time `bson:"end"`
}
type Attachment struct {
	FileName    string `bson:"fileName"`
	FileURL     string `bson:"fileUrl"`
	FileType    string `bson:"fileType"`
	UploadedAt time.Time `bson:"uploadedAt"`
}