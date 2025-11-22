package service

import (
	"UAS_GO/helper"
	"UAS_GO/app/repository"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func GetGlobalStatistics(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	userID := c.Locals("user_id").(string)

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

func GetStudentReport(c *fiber.Ctx) error {
	studentID := c.Params("id")

	stats, err := repository.GetStudentStatistics(studentID)
	if err != nil {
		return helper.InternalError(c, err.Error())
	}

	return helper.APIResponse(c, fiber.StatusOK, "Get Student Report Succesfully", stats)
}
