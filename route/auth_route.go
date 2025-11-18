package route

import (
	"UAS_GO/app/service"

	"github.com/gofiber/fiber/v2"
)

func registerAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")
	auth.Post("/login", service.AuthLoginHandler)
}
