package main

import (
	"UAS_GO/config"
	"UAS_GO/database"
	"UAS_GO/route"
)

func main() {
	config.LoadEnv()

	database.ConnectPostgres()
	database.ConnectMongoDB()
	// database.AutoMigrate()
	// database.MigrateTesting(database.PSQL) // uncomment jika perlu

	app := config.NewApp()
	route.RegisterRoutes(app)
	app.Static("/static", "./uploads")

	port := config.GetEnv("APP_PORT", "3000")
	app.Listen(":" + port)
}
