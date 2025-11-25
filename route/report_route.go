package route

import (
	"UAS_GO/app/service"
	"UAS_GO/middleware"
	"github.com/gofiber/fiber/v2"
)

func registerReportRoutes(api fiber.Router) {
	r := api.Group("/reports", middleware.AuthRequired())

	r.Get("/statistics", middleware.PermissionRequired("report:statistics"), service.GetGlobalStatistics)
	r.Get("/student/:id", middleware.PermissionRequired("report:student"), service.GetStudentReport)
}
