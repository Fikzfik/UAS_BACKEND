package route

import (
	"UAS_GO/app/service"
	"UAS_GO/middleware"
	"github.com/gofiber/fiber/v2"
)

func registerStudentRoutes(api fiber.Router) {
	r := api.Group("/students", middleware.AuthRequired())

	r.Get("/", middleware.PermissionRequired("student:read"), service.GetAllStudents)
	r.Get("/:id", middleware.PermissionRequired("student:read"), service.GetStudentByID)
	r.Get("/:id/achievements", middleware.PermissionRequired("student:read"), service.GetStudentAchievements)
	r.Put("/:id/advisor", middleware.PermissionRequired("student:update"), service.UpdateStudentAdvisor)
}
