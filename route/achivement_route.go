package route

import (
	"UAS_GO/app/service"
	"UAS_GO/middleware"
	"github.com/gofiber/fiber/v2"
)

func registerAchivementRoutes(api fiber.Router) {
	r := api.Group("/achievements", middleware.AuthRequired())

	r.Get("/", service.GetAllAchievements)         // All, filtered by role
	// r.Get("/:id", service.GetAchievementDetail) // Detail

	r.Post("/", middleware.MahasiswaOnly(), service.CreateAchievement) // Create
	// r.Put("/:id", middleware.MahasiswaOnly(), service.UpdateAchievement)
	// r.Delete("/:id", middleware.MahasiswaOnly(), service.DeleteAchievement)

	// r.Post("/:id/submit", middleware.MahasiswaOnly(), service.SubmitAchievement)

	// r.Post("/:id/verify", middleware.DosenWaliOnly(), service.VerifyAchievement)
	// r.Post("/:id/reject", middleware.DosenWaliOnly(), service.RejectAchievement)

	// r.Get("/:id/history", service.GetAchievementHistory)

	// r.Post("/:id/attachments", middleware.MahasiswaOnly(), service.UploadAchievementFile)
}
