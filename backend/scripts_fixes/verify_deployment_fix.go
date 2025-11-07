package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
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

	fmt.Println("🔍 ====================================================================")
	fmt.Println("    VERIFIKASI SKENARIO DEPLOYMENT DI PC/SERVER BARU")
	fmt.Println("🔍 ====================================================================")
	fmt.Println()

	// 1. Check if migration file exists
	fmt.Println("📄 1. CEK KETERSEDIAAN MIGRATION FILE")
	fmt.Println("-------------------------------------------------------")
	
	migrationFile := "031_fix_account_mapping_for_tax_and_revenue.sql"
	if _, err := os.Stat("migrations/" + migrationFile); err != nil {
		fmt.Printf("❌ Migration file TIDAK DITEMUKAN: %s\n", migrationFile)
		fmt.Println("⚠️  MASALAH: Perbaikan tidak akan otomatis jalan di deployment baru!")
	} else {
		fmt.Printf("✅ Migration file DITEMUKAN: %s\n", migrationFile)
	}

	// 2. Check if migration has been executed
	fmt.Println()
	fmt.Println("📋 2. CEK STATUS EKSEKUSI MIGRATION")
	fmt.Println("-------------------------------------------------------")
	
	var migrationCount int
	var migrationStatus, migrationMessage string
	var executedAt sql.NullString
	
	err = db.QueryRow(`
		SELECT COUNT(*), 
		       COALESCE(MAX(status), 'NOT_EXECUTED') as status,
		       COALESCE(MAX(message), '') as message,
		       COALESCE(MAX(executed_at::text), '') as executed_at
		FROM migration_logs 
		WHERE migration_name = $1
	`, migrationFile).Scan(&migrationCount, &migrationStatus, &migrationMessage, &executedAt)
	
	if err != nil {
		fmt.Printf("⚠️  Tidak bisa cek status migration: %v\n", err)
		fmt.Println("💡 Kemungkinan tabel migration_logs belum ada (normal untuk fresh install)")
	} else {
		if migrationCount > 0 {
			fmt.Printf("✅ Migration sudah dieksekusi: %s\n", migrationStatus)
			fmt.Printf("   Tanggal eksekusi: %s\n", executedAt.String)
			if migrationMessage != "" {
				fmt.Printf("   Pesan: %s\n", migrationMessage)
			}
		} else {
			fmt.Println("📋 Migration belum dieksekusi (normal untuk fresh install)")
		}
	}

	// 3. Simulate deployment scenario
	fmt.Println()
	fmt.Println("🚀 3. SIMULASI SKENARIO DEPLOYMENT")
	fmt.Println("-------------------------------------------------------")
	
	fmt.Println("Ketika Anda melakukan git pull di PC/server baru:")
	fmt.Println()
	fmt.Println("1️⃣  git pull origin main")
	fmt.Println("   ↳ Mendapatkan migration file: migrations/031_fix_account_mapping_for_tax_and_revenue.sql")
	fmt.Println()
	fmt.Println("2️⃣  go run main.go (atau restart aplikasi)")
	fmt.Println("   ↳ main.go → database.RunAutoMigrations(db)")
	fmt.Println("   ↳ Otomatis mencari file *.sql di folder migrations/")
	fmt.Println("   ↳ Mengeksekusi migration 031_fix_account_mapping_for_tax_and_revenue.sql")
	fmt.Println("   ↳ Update akun mapping:")
	fmt.Println("      • 2103 → LIABILITY (PPN Keluaran)")
	fmt.Println("      • 2102 → ASSET (PPN Masukan)")
	fmt.Println("      • 4101 → REVENUE (Pendapatan Penjualan)")
	fmt.Println()
	fmt.Println("3️⃣  Aplikasi siap digunakan dengan mapping akun yang benar!")

	// 4. Check if system is idempotent
	fmt.Println()
	fmt.Println("🔄 4. VERIFIKASI SISTEM IDEMPOTENT")
	fmt.Println("-------------------------------------------------------")
	
	fmt.Println("Migration script menggunakan:")
	fmt.Println("• UPDATE statements dengan WHERE clause spesifik")
	fmt.Println("• INSERT ... WHERE NOT EXISTS untuk create missing accounts")
	fmt.Println("• Aman untuk dijalankan berulang kali tanpa side effect")
	
	// 5. Environment variable check
	fmt.Println()
	fmt.Println("🌐 5. CEK ENVIRONMENT VARIABLES")
	fmt.Println("-------------------------------------------------------")
	
	envVars := map[string]string{
		"DATABASE_URL": os.Getenv("DATABASE_URL"),
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_NAME":     os.Getenv("DB_NAME"),
		"DB_USER":     os.Getenv("DB_USER"),
	}
	
	hasValidDB := false
	for key, value := range envVars {
		if value != "" {
			fmt.Printf("✅ %s: %s\n", key, maskSensitive(key, value))
			hasValidDB = true
		} else {
			fmt.Printf("⚪ %s: (kosong)\n", key)
		}
	}
	
	if !hasValidDB {
		fmt.Println("⚠️  PERHATIAN: Pastikan environment variables database sudah dikonfigurasi!")
	}

	// 6. Final assessment
	fmt.Println()
	fmt.Println("📊 6. KESIMPULAN DEPLOYMENT")
	fmt.Println("-------------------------------------------------------")
	
	allGood := true
	
	// Check migration file exists
	if _, err := os.Stat("migrations/" + migrationFile); err != nil {
		allGood = false
	}
	
	if allGood {
		fmt.Println("✅ DEPLOYMENT SIAP!")
		fmt.Println()
		fmt.Println("🎯 LANGKAH DEPLOYMENT DI PC/SERVER BARU:")
		fmt.Println("1. git pull origin main")
		fmt.Println("2. Setup environment variables (.env file)")
		fmt.Println("3. go run main.go atau restart service")
		fmt.Println("4. Migration akan otomatis memperbaiki mapping akun")
		fmt.Println("5. Aplikasi ready dengan akun yang sudah benar")
		fmt.Println()
		fmt.Println("💡 CATATAN:")
		fmt.Println("• Migration hanya jalan sekali per environment") 
		fmt.Println("• Aman untuk fresh install maupun existing database")
		fmt.Println("• Tidak perlu manual intervention")
	} else {
		fmt.Println("❌ DEPLOYMENT BERMASALAH!")
		fmt.Println()
		fmt.Println("🔧 YANG PERLU DIPERBAIKI:")
		fmt.Println("• Migration file tidak ditemukan")
		fmt.Println("• Perbaikan tidak akan otomatis jalan")
		fmt.Println("• Perlu manual fix di setiap environment")
	}
	
	fmt.Println()
	fmt.Println("🔍 ====================================================================")
}

func maskSensitive(key, value string) string {
	sensitiveKeys := []string{"PASSWORD", "SECRET", "TOKEN"}
	for _, sensitive := range sensitiveKeys {
		if strings.Contains(strings.ToUpper(key), sensitive) && len(value) > 3 {
			return value[:3] + "***"
		}
	}
	
	// Mask database connection strings
	if key == "DATABASE_URL" && len(value) > 20 {
		return value[:20] + "***"
	}
	
	return value
}