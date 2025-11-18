package database

import (
    "database/sql"
    "fmt"
    "log"
    "UAS_GO/config"

    _ "github.com/lib/pq"
)

var PSQL *sql.DB

func ConnectPostgres() {
    host := config.GetEnv("DB_HOST", "localhost")
    port := config.GetEnv("DB_PORT", "5432")
    user := config.GetEnv("DB_USER", "postgres")
    pass := config.GetEnv("DB_PASS", "")
    name := config.GetEnv("DB_NAME", "UAS_GO")

    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, pass, name,
    )

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatalf(" Gagal konek PostgreSQL: %v", err)
    }

    if err := db.Ping(); err != nil {
        log.Fatalf(" PostgreSQL tidak merespon: %v", err)
    }

    fmt.Println(" Berhasil konek PostgreSQL!")

    PSQL = db
}
