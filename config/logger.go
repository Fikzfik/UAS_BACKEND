package config

import (
	"log"
	"os"
)

var Logger *log.Logger

func InitLogger() {
	file, err := os.OpenFile("logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(" Gagal membuka file log:", err)
	}
	Logger = log.New(file, "APP_LOG: ", log.Ldate|log.Ltime|log.Lshortfile)
}
