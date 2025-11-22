package route

import (
	"UAS_GO/app/service"
	"UAS_GO/middleware"
	"github.com/gofiber/fiber/v2"
)


func registerReportRoutes(api fiber.Router) {
	r := api.Group("/reports", middleware.AuthRequired())

	r.Get("/statistics", service.GetGlobalStatistics)
	r.Get("/student/:id", middleware.AdminOrLecturerOrOwnerStudent(), service.GetStudentReport)
}
