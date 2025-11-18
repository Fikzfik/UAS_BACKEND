package route

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")
	registerAuthRoutes(api)
	// registerUserRoutes(api)
	// registerAchivRoutes(api)
	// registerStudentRoutes(api)
	// registerlecturerRoutes(api)
	// registerReportRoutes(api)
}
