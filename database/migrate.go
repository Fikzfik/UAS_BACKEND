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

	fmt.Println("Running MigrateTesting...")

	// 1. Hapus semua data lama (urutan foreign key)
	DB.Exec("DELETE FROM achievement_references")
	DB.Exec("DELETE FROM students")
	DB.Exec("DELETE FROM lecturers")
	DB.Exec("DELETE FROM role_permissions")
	DB.Exec("DELETE FROM permissions")
	DB.Exec("DELETE FROM users")
	DB.Exec("DELETE FROM roles")

	fmt.Println("Menghapus semua data lama selesai.")

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

	// 3. Daftar permissions sesuai SQL yang diberikan (nama/resource/action/description)
	permList := []struct {
		Name, Resource, Action, Description string
	}{
		// AUTH
		{"auth:profile", "auth", "profile", "View logged-in user profile"},

		// USERS
		{"user:create", "user", "create", "Create a new user"},
		{"user:read", "user", "read", "View list or detail of users"},
		{"user:update", "user", "update", "Update user data"},
		{"user:delete", "user", "delete", "Delete user"},
		{"user:assign-role", "user", "assign-role", "Assign role to user"},
		{"user:manage", "user", "manage", "Full user management"},

		// ACHIEVEMENTS
		{"achievement:create", "achievement", "create", "Create achievement"},
		{"achievement:read", "achievement", "read", "Read achievements"},
		{"achievement:update", "achievement", "update", "Update achievement"},
		{"achievement:delete", "achievement", "delete", "Delete achievement"},
		{"achievement:submit", "achievement", "submit", "Submit for verification"},

		// VERIFICATION
		{"achievement:view-advisee", "achievement", "view-advisee", "View advisee achievements"},
		{"achievement:verify", "achievement", "verify", "Verify student achievement"},
		{"achievement:reject", "achievement", "reject", "Reject student achievement"},

		// STUDENTS & LECTURERS
		{"student:read", "student", "read", "View student list"},
		{"student:update", "student", "update", "Update student data"},
		{"lecturer:read", "lecturer", "read", "View lecturers"},
		{"lecturer:advisee-list", "lecturer", "advisee-list", "View lecturer advisee list"},

		// REPORTS
		{"report:statistics", "report", "statistics", "View achievement statistics"},
		{"report:student", "report", "student", "View student-specific report"},
	}

	// map nama permission -> id (agar bisa diassign ke role)
	permissionIDs := map[string]string{}

	for _, p := range permList {
		id := uuid.New().String()
		permissionIDs[p.Name] = id

		// decompose resource/action (redundan tapi konsisten)
		parts := strings.SplitN(p.Name, ":", 2)
		resource := ""
		action := ""
		if len(parts) > 0 {
			resource = parts[0]
		}
		if len(parts) > 1 {
			action = parts[1]
		}

		_, err := DB.Exec(`
            INSERT INTO permissions (id, name, resource, action, description)
            VALUES ($1, $2, $3, $4, $5)
        `, id, p.Name, resource, action, p.Description)

		if err != nil {
			log.Fatalf(" Gagal insert permission %s: %v\n", p.Name, err)
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

	// 4b. Berikan subset permissions ke role mahasiswa & dosen_wali
	// Permissions untuk Mahasiswa (sesuai kebutuhan)
	mhsPermNames := []string{
		"achievement:create",
		"achievement:read",
		"achievement:update",
		"achievement:delete",
		"achievement:submit",
		"report:student",
		"auth:profile",
	}

	for _, name := range mhsPermNames {
		pid, ok := permissionIDs[name]
		if !ok {
			log.Fatalf(" Permission %s tidak ditemukan saat assign ke mahasiswa", name)
		}
		_, err := DB.Exec(`
            INSERT INTO role_permissions (role_id, permission_id)
            VALUES ($1, $2)
        `, mhsRoleID, pid)
		if err != nil {
			log.Fatalf(" Gagal assign permission %s ke mahasiswa: %v\n", name, err)
		}
	}

	// Permissions untuk Dosen Wali
	dosenPermNames := []string{
		"achievement:read",
		"achievement:view-advisee",
		"achievement:verify",
		"achievement:reject",
		"lecturer:advisee-list",
		"report:statistics",
		"auth:profile",
	}

	for _, name := range dosenPermNames {
		pid, ok := permissionIDs[name]
		if !ok {
			log.Fatalf(" Permission %s tidak ditemukan saat assign ke dosen_wali", name)
		}
		_, err := DB.Exec(`
            INSERT INTO role_permissions (role_id, permission_id)
            VALUES ($1, $2)
        `, dosenRoleID, pid)
		if err != nil {
			log.Fatalf(" Gagal assign permission %s ke dosen_wali: %v\n", name, err)
		}
	}

	fmt.Println("====== Permissions untuk mahasiswa & dosen_wali berhasil diassign. ======")

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

	fmt.Println("Mahasiswa berhasil dibuat.")

	fmt.Println("MigrateTesting selesai TANPA ERROR!")
}
