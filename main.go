package main

import (
	"UAS_GO/config"
	"UAS_GO/route"
	"UAS_GO/database"
)

func main() {
	config.LoadEnv()
	
	database.ConnectPostgres()
	database.ConnectMongoDB()
	// database.AutoMigrate()
	database.MigrateTesting(database.PSQL) // uncomment jika perlu

	app := config.NewApp()
	route.RegisterRoutes(app)

	port := config.GetEnv("APP_PORT", "3000")
	app.Listen(":" + port)
}
