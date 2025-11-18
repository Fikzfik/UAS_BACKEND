package route

import (
	// "UAS_GO/app/service"
	// // "crud-alumni/helper"
	// "UAS_GO/middleware"

	"github.com/gofiber/fiber/v2"
)

func registerAuthRoutes(api fiber.Router) {
	alumni := api.Group("/alumni")

	alumni.Get("/profile")
	alumni.Post("/login")
	alumni.Post("/logout")
	alumni.Post("/refresh")
}	
