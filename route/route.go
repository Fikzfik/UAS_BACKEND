package route

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")
	registerAuthRoutes(api)
	registerAdminRoutes(api)
	registerAchivementRoutes(api)
	registerStudentRoutes(api)
	registerlecturerRoutes(api)
	// registerReportRoutes(api)
}
