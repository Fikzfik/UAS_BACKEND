package config

import "github.com/gofiber/fiber/v2"

func NewApp() *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB
	})
	return app
}
