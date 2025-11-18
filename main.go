package main

import (
	"UAS_GO/config"
	"UAS_GO/route"
)

func main() {
	config.LoadEnv()
	config.InitLogger()
	// database.ConnectDB()
	// database.MigrateTesting(database.DB) // uncomment jika perlu

	app := config.NewApp()
	route.RegisterRoutes(app)

	port := config.GetEnv("APP_PORT", "3000")
	app.Listen(":" + port)
}
