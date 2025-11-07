package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Could not load .env file")
	}

	// Database connection
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("🔧 ====================================================================")
	fmt.Println("    FIX USER SEEDING - ROBUST UPSERT APPROACH")
	fmt.Println("🔧 ====================================================================")
	fmt.Println()

	// Hash password once
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Users to create/update
	users := []struct {
		Username  string
		Email     string
		Role      string
		FirstName string
		LastName  string
	}{
		{"admin", "admin@company.com", "admin", "Admin", "User"},
		{"finance", "finance@company.com", "finance", "Finance", "User"},
		{"inventory", "inventory@company.com", "inventory_manager", "Inventory", "User"},
		{"director", "director@company.com", "director", "Director", "User"},
		{"employee", "employee@company.com", "employee", "Employee", "User"},
	}

	fmt.Println("1. MENJALANKAN ROBUST USER UPSERT...")
	fmt.Println("-------------------------------------------------------")

	successCount := 0
	for _, user := range users {
		// Use PostgreSQL UPSERT (ON CONFLICT DO UPDATE)
		query := `
		INSERT INTO users (
			username, email, password, role, first_name, last_name, 
			phone, address, department, position, salary, is_active,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, 
			'', '', '', '', 0, true,
			NOW(), NOW()
		)
		ON CONFLICT (username) DO UPDATE SET
			email = EXCLUDED.email,
			role = EXCLUDED.role,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			is_active = true,
			updated_at = NOW()
		RETURNING id, username
		`

		var returnedID int
		var returnedUsername string
		
		err := db.QueryRow(query, 
			user.Username, user.Email, string(hashedPassword), 
			user.Role, user.FirstName, user.LastName,
		).Scan(&returnedID, &returnedUsername)

		if err != nil {
			fmt.Printf("❌ Error creating/updating user %s: %v\n", user.Username, err)
		} else {
			fmt.Printf("✅ User %s (ID: %d) created/updated successfully\n", returnedUsername, returnedID)
			successCount++
		}
	}

	fmt.Println("-------------------------------------------------------")
	fmt.Printf("📊 HASIL: %d/%d users berhasil diproses\n", successCount, len(users))

	if successCount == len(users) {
		fmt.Println()
		fmt.Println("🎉 SEMUA USER SEEDING BERHASIL!")
		fmt.Println("✅ Tidak akan ada lagi error duplicate key untuk user seeding")
		fmt.Println("✅ Backend sekarang bisa restart tanpa masalah")
		
		fmt.Println()
		fmt.Println("🔄 LANGKAH SELANJUTNYA:")
		fmt.Println("1. Update database/seed.go dengan logika yang lebih robust")
		fmt.Println("2. Restart backend aplikasi")
		fmt.Println("3. Verifikasi tidak ada lagi error user seeding")
	} else {
		fmt.Println()
		fmt.Println("⚠️  BEBERAPA USER GAGAL DIPROSES")
		fmt.Println("🔍 Cek error messages di atas untuk detail")
	}

	fmt.Println()
	fmt.Println("📋 USER LOGIN CREDENTIALS:")
	fmt.Println("-------------------------------------------------------")
	for _, user := range users {
		fmt.Printf("Username: %-10s | Role: %-17s | Password: password123\n", 
			user.Username, user.Role)
	}

	fmt.Println()
	fmt.Println("🔧 ====================================================================")
}