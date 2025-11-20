package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"UAS_GO/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoDB *mongo.Database // Variabel global untuk database MongoDB

// ConnectDB menghubungkan ke MongoDB dan menginisialisasi variabel MongoDB
func ConnectMongoDB() {
	// Ambil konfigurasi dari .env
	mongoURI := config.GetEnv("MONGO_URI", "mongodb://localhost:27017")
	dbName := config.GetEnv("MONGO_DB_NAME", "UAS_GO")

	// Siapkan client MongoDB
	clientOptions := options.Client().ApplyURI(mongoURI)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf(" Gagal koneksi ke MongoDB: %v", err)
	}

	// Ping MongoDB untuk memastikan koneksi
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf(" MongoDB tidak bisa di-ping: %v", err)
	}

	fmt.Println(" Berhasil terhubung ke MongoDB!")

	// Simpan database global
	MongoDB = client.Database(dbName)
}
