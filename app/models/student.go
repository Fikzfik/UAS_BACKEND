package models

import "time"

type Student struct {
    ID            string    `json:"id"`
    UserID        string    `json:"user_id"`
    StudentID     string    `json:"student_id"`
    ProgramStudy  string    `json:"program_study"`
    AcademicYear  string    `json:"academic_year"`
    AdvisorID     string    `json:"advisor_id"`
    CreatedAt     time.Time `json:"created_at"`
}


// dipakai utk PUT /students/{id}/advisor
type UpdateStudentAdvisorRequest struct {
	AdvisorId *string `json:"advisorId"`
}

// dipakai utk POST /achievements
type CreateAchievementRequest struct {
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	AchievementType string   `json:"achievementType"`
	Details         string   `json:"details,omitempty"`
	Tags            []string `json:"tags,omitempty"`
}

// dipakai utk PATCH /achievements/{id}
type UpdateAchievementRequest struct {
	Title           *string   `json:"title,omitempty"`
	Description     *string   `json:"description,omitempty"`
	AchievementType *string   `json:"achievementType,omitempty"`
	Details         *string   `json:"details,omitempty"`
	Tags            *[]string `json:"tags,omitempty"`
}

// dipakai utk POST /achievements/{id}/verify
type VerifyAchievementRequest struct {
	Points int `json:"points"`
}

// dipakai utk POST /achievements/{id}/reject
type RejectAchievementRequest struct {
	Note string `json:"note"`
}
