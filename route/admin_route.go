package route

import (
	"UAS_GO/app/service"
	"UAS_GO/middleware"
	"github.com/gofiber/fiber/v2"
)

func registerAdminRoutes(api fiber.Router) {
	admin := api.Group("/users", middleware.AuthRequired(), middleware.PermissionRequired("user:manage"))

	admin.Get("/", middleware.PermissionRequired("user:read"), service.AdminGetAllUsers)
	admin.Get("/:id", middleware.PermissionRequired("user:read"), service.AdminGetUserByID)
	admin.Post("/", middleware.PermissionRequired("user:create"), service.AdminCreateUser)
	admin.Put("/:id", middleware.PermissionRequired("user:update"), service.AdminUpdateUser)
	admin.Delete("/:id", middleware.PermissionRequired("user:delete"), service.AdminDeleteUser)
	admin.Put("/:id/role", middleware.PermissionRequired("user:assign-role"), service.AdminUpdateUserRole)
}
