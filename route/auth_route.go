package route

import (
	"UAS_GO/app/service"
	"UAS_GO/middleware"
	"github.com/gofiber/fiber/v2"
)

func registerAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")
	auth.Post("/login", service.AuthLogin)

	protected := auth.Use(middleware.AuthRequired())

	protected.Get("/profile", service.AuthGetProfile)
	protected.Post("/logout", service.AuthLogout)
	protected.Post("/refresh", service.AuthRefreshToken)
}
