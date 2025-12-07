package service

import (
	"UAS_GO/app/repository"
	"UAS_GO/helper"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// GetGlobalStatistics godoc
// @Summary Get global statistics (role-based)
// @Description
//   Mengambil statistik global prestasi:
//   - admin: melihat semua data
//   - dosen_wali: hanya prestasi mahasiswa bimbingannya
//   - mahasiswa: hanya prestasi milik sendiri.
// @Tags Statistics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "envelope {status,message,data} berisi statistik global"
// @Failure 401 {object} map[string]interface{} "Unauthorized (role/user_id tidak tersedia)"
// @Failure 403 {object} map[string]interface{} "Forbidden (profil tidak ditemukan)"
// @Failure 500 {object} map[string]interface{} "error response"
// @Router /statistics/global [get]
func GetGlobalStatistics(c *fiber.Ctx) error {
	role, ok := c.Locals("role").(string)
	if !ok || role == "" {
		return helper.Unauthorized(c, "Unauthorized: role not found")
	}
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return helper.Unauthorized(c, "Unauthorized: user_id not found")
	}

	var filter bson.M

	// Admin → semua data
	if role == "admin" {
		filter = bson.M{}
	}

	// Dosen Wali → prestasi mahasiswa bimbingannya
	if role == "dosen_wali" {
		lecturerID, err := repository.GetLecturerIDByUserID(userID)
		if err != nil {
			return helper.Forbidden(c, "Lecturer data not found")
		}

		// ambil student advisees
		advisees, _ := repository.GetAdviseeIDsByLecturer(lecturerID)
		filter = bson.M{"studentId": bson.M{"$in": advisees}}
	}

	// Mahasiswa → prestasi sendiri
	if role == "mahasiswa" {
		studentID, _ := repository.GetStudentIDByUserID(userID)
		filter = bson.M{"studentId": studentID}
	}

	stats, err := repository.GetStatistics(filter)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}

	return helper.APIResponse(c, fiber.StatusOK, "Get Data Global Statistic Succesfully", stats)
}

// GetStudentReport godoc
// @Summary Get single student report
// @Description Mengambil statistik dan laporan prestasi untuk satu mahasiswa berdasarkan student ID (UUID Postgres).
// @Tags Statistics
// @Accept json
// @Produce json
// @Param id path string true "Student ID (UUID)"
// @Success 200 {object} map[string]interface{} "envelope {status,message,data} berisi statistik student"
// @Failure 500 {object} map[string]interface{} "error response"
// @Router /statistics/students/{id} [get]
func GetStudentReport(c *fiber.Ctx) error {
	studentID := c.Params("id")

	stats, err := repository.GetStudentStatistics(studentID)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}

	return helper.APIResponse(c, fiber.StatusOK, "Get Student Report Succesfully", stats)
}
