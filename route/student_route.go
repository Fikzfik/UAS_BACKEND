package route

import (
	"UAS_GO/app/service"
	"UAS_GO/middleware"
	"github.com/gofiber/fiber/v2"

)

func registerStudentRoutes(api fiber.Router) {
	r := api.Group("/students", middleware.AuthRequired())

	r.Get("/", middleware.AdminOnly(), service.GetAllStudents) // admin only
	r.Get("/:id", middleware.OwnerOrAdvisorOrAdmin(), service.GetStudentByID)
	r.Get("/:id/achievements", middleware.OwnerOrAdvisorOrAdmin(), service.GetStudentAchievements)
	r.Put("/:id/advisor", middleware.AdminOnly(), service.UpdateStudentAdvisor)
}

