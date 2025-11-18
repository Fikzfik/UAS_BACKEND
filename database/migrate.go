package database

import (
	"UAS_GO/helper"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
)

func MigrateTesting(DB *sql.DB) {

    fmt.Println("🔄 Running MigrateTesting...")

    // 1. Hapus semua data lama (urutan foreign key)
    DB.Exec("DELETE FROM achievement_references")
    DB.Exec("DELETE FROM students")
    DB.Exec("DELETE FROM lecturers")
    DB.Exec("DELETE FROM role_permissions")
    DB.Exec("DELETE FROM permissions")
    DB.Exec("DELETE FROM users")
    DB.Exec("DELETE FROM roles")

    fmt.Println("🧹 Menghapus semua data lama selesai.")

    // 2. Insert default roles
    adminRoleID := uuid.New().String()
    mhsRoleID := uuid.New().String()
    dosenRoleID := uuid.New().String()

    _, err := DB.Exec(`
        INSERT INTO roles (id, name, description)
        VALUES 
        ($1, 'admin', 'Administrator Sistem'),
        ($2, 'mahasiswa', 'User Mahasiswa'),
        ($3, 'dosen_wali', 'Dosen Pembimbing Akademik')
    `, adminRoleID, mhsRoleID, dosenRoleID)

    if err != nil {
        log.Fatalf("====== Gagal insert roles: %v\n ======", err)
    }

    fmt.Println("====== Insert roles berhasil ======.")

    // 3. Insert permissions
    permissionIDs := map[string]string{
        "achievement:create": uuid.New().String(),
        "achievement:read"  : uuid.New().String(),
        "achievement:update": uuid.New().String(),
        "achievement:delete": uuid.New().String(),
        "achievement:verify": uuid.New().String(),
        "user:manage"        : uuid.New().String(),
    }

    for name, id := range permissionIDs {

        resource := name[:len(name)-len(name[strings.Index(name, ":"):])]
        action   := name[strings.Index(name, ":")+1:]

        _, err := DB.Exec(`
            INSERT INTO permissions (id, name, resource, action, description)
            VALUES ($1, $2, $3, $4, $5)
        `, id, name, resource, action, "default permission")

        if err != nil {
            log.Fatalf(" Gagal insert permission %s: %v\n", name, err)
        }
    }

    fmt.Println("====== Insert permissions berhasil. ======")

    // 4. Hubungkan semua permission ke admin
    for _, permID := range permissionIDs {
        _, err := DB.Exec(`
            INSERT INTO role_permissions (role_id, permission_id)
            VALUES ($1, $2)
        `, adminRoleID, permID)

        if err != nil {
            log.Fatalf(" Gagal assign permission ke admin: %v\n", err)
        }
    }

    fmt.Println("======Semua permissions diberikan ke admin======.")

    // 5. Insert admin user
    adminUserID := uuid.New().String()
    adminHash, _ := helper.HashPassword("123456")

    _, err = DB.Exec(`
        INSERT INTO users (id, username, email, password_hash, full_name, role_id)
        VALUES ($1, 'admin', 'admin@example.com', $2, 'Administrator', $3)
    `, adminUserID, adminHash, adminRoleID)

    if err != nil {
        log.Fatalf(" Gagal insert admin user: %v\n", err)
    }

    fmt.Println("====== Admin user berhasil dibuat.====== ")

    // 6. Insert dosen
    dosenUserID := uuid.New().String()
    dosenHash, _ := helper.HashPassword("123456")
    lecturerID := uuid.New().String()

    _, err = DB.Exec(`
        INSERT INTO users (id, username, email, password_hash, full_name, role_id)
        VALUES ($1, 'dosen1', 'dosen1@example.com', $2, 'Dosen Wali 1', $3)
    `, dosenUserID, dosenHash, dosenRoleID)

    if err != nil {
        log.Fatalf(" Gagal insert dosen user: %v\n", err)
    }

    _, err = DB.Exec(`
        INSERT INTO lecturers (id, user_id, lecturer_id, department)
        VALUES ($1, $2, 'DSN001', 'Teknik Informatika')
    `, lecturerID, dosenUserID)

    if err != nil {
        log.Fatalf(" Gagal insert lecturer: %v\n", err)
    }

    fmt.Println(" Dosen wali berhasil dibuat.")

    // 7. Insert Mahasiswa
    mhsUserID := uuid.New().String()
    mhsHash, _ := helper.HashPassword("123456")
    studentID := uuid.New().String()

    _, err = DB.Exec(`
        INSERT INTO users (id, username, email, password_hash, full_name, role_id)
        VALUES ($1, 'mhs1', 'mhs1@example.com', $2, 'Mahasiswa Satu', $3)
    `, mhsUserID, mhsHash, mhsRoleID)

    if err != nil {
        log.Fatalf(" Gagal insert mahasiswa user: %v\n", err)
    }

    _, err = DB.Exec(`
        INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id)
        VALUES ($1, $2, '20230001', 'Teknik Informatika', '2023', $3)
    `, studentID, mhsUserID, lecturerID)

    if err != nil {
        log.Fatalf(" Gagal insert mahasiswa: %v\n", err)
    }

    fmt.Println("🎓 Mahasiswa berhasil dibuat.")

    fmt.Println("🎉 MigrateTesting selesai TANPA ERROR!")
}
