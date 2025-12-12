package main

import (
	"UAS_GO/config"
	"UAS_GO/database"
	_ "UAS_GO/docs" // <- wajib: package yang dibuat swag
	"UAS_GO/route"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title Sistem Pelaporan Prestasi Mahasiswa API
// @version 1.0
// @description Backend untuk sistem pelaporan prestasi mahasiswa sesuai SRS.
// @contact.name Pengembang Backend
// @host localhost:3000
// @BasePath /api/v1
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	config.LoadEnv()

	database.ConnectPostgres()
	database.ConnectMongoDB()
	// database.AutoMigrate()
	// database.MigrateTesting(database.PSQL) // uncomment jika perlu

	app := config.NewApp()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // atau contoh: "http://localhost:3000,http://localhost:8080"
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		// AllowCredentials: true, // aktifkan jika butuh cookies/credentials
	}))

	route.RegisterRoutes(app)
	app.Static("/static", "./uploads")
	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	port := config.GetEnv("APP_PORT", "3000")
	app.Listen(":" + port)
}
