package route

import (
	"UAS_GO/app/service"
	"UAS_GO/middleware"
	"github.com/gofiber/fiber/v2"
)

func registerAdminRoutes(api fiber.Router) {
	admin := api.Group("/users", middleware.AuthRequired(), middleware.AdminOnly())
	admin.Get("/", service.AdminGetAllUsers)
	admin.Get("/:id", service.AdminGetUserByID)
	admin.Post("/", service.AdminCreateUser)
	admin.Put("/:id", service.AdminUpdateUser)
	admin.Delete("/:id", service.AdminDeleteUser)
	admin.Put("/:id/role", service.AdminUpdateUserRole)
}