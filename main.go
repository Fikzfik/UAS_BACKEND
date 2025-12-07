package main

import (
	"UAS_GO/config"
	"UAS_GO/database"
	"UAS_GO/route"

	_ "UAS_GO/docs" // <- wajib: package yang dibuat swag
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title Sistem Pelaporan Prestasi Mahasiswa API
// @version 1.0
// @description Backend untuk sistem pelaporan prestasi mahasiswa sesuai SRS.
// @contact.name Pengembang Backend
// @host localhost:3000
// @BasePath /api/v1
// @schemes http
func main() {
	config.LoadEnv()

	database.ConnectPostgres()
	database.ConnectMongoDB()
	// database.AutoMigrate()
	// database.MigrateTesting(database.PSQL) // uncomment jika perlu

	app := config.NewApp()
	route.RegisterRoutes(app)
	app.Static("/static", "./uploads")
	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	port := config.GetEnv("APP_PORT", "3000")
	app.Listen(":" + port)
}
