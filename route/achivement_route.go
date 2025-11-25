package route

import (
	"UAS_GO/app/service"
	"UAS_GO/middleware"
	"github.com/gofiber/fiber/v2"
)

func registerAchivementRoutes(api fiber.Router) {
	r := api.Group("/achievements", middleware.AuthRequired())

	r.Get("/",middleware.PermissionRequired("achievement:read"),service.GetAllAchievements)
	r.Get("/:id",middleware.PermissionRequired("achievement:read"),service.GetAchievementById)
	r.Post("/",middleware.PermissionRequired("achievement:create"),service.CreateAchievement)
	r.Put("/:id",middleware.PermissionRequired("achievement:update"),service.UpdateAchievement)
	r.Delete("/:id",middleware.PermissionRequired("achievement:delete"),service.DeleteAchievement)
	r.Post("/:id/submit",middleware.PermissionRequired("achievement:submit"),service.SubmitAchievement)
	r.Post("/:id/verify",middleware.PermissionRequired("achievement:verify"),service.VerifyAchievement)
	r.Post("/:id/reject",middleware.PermissionRequired("achievement:reject"),service.RejectAchievement)
	r.Get("/:id/history",middleware.PermissionRequired("achievement:read"),service.GetAchievementHistory)
	r.Post("/:id/attachments",middleware.PermissionRequired("achievement:update"),service.UploadAchievementFile)

}
