package database

import (
    "UAS_GO/app/models"
    "fmt"
)

func AutoMigrate() {
    err := PSQL.AutoMigrate(
        &models.User{},
        &models.Role{},
        &models.Permission{},
        &models.Student{},
        &models.Lecturer{},
        &models.AchievementReference{},
    )

    if err != nil {
        fmt.Println("❌ AutoMigrate gagal:", err)
    } else {
        fmt.Println("✅ AutoMigrate sukses!")
    }
}
