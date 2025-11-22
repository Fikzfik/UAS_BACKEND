package route

import (
	"UAS_GO/app/service"
	"UAS_GO/middleware"

	"github.com/gofiber/fiber/v2"
)

func registerlecturerRoutes(api fiber.Router) {
	r := api.Group("/lecturers", middleware.AuthRequired())

	r.Get("/", middleware.AdminOnly(), service.GetAllLecturers)
	r.Get("/:id/advisees", middleware.LecturerOrAdminForLecturerResource(), service.GetLecturerAdvisees)
}
