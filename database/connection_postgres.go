package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var PSQL *gorm.DB

func ConnectPostgres() {
	host := os.Getenv("PG_HOST")
	user := os.Getenv("PG_USER")
	pass := os.Getenv("PG_PASSWORD")
	dbname := os.Getenv("UAS_GO")
	port := os.Getenv("PG_PORT")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, pass, dbname, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Gagal koneksi ke PostgreSQL: %v", err)
	}

	PSQL = db
	fmt.Println("✅ Berhasil terhubung ke PostgreSQL!")
}
