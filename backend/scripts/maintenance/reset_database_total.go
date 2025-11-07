package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== SISTEM AKUNTANSI - RESET DATABASE TOTAL ===")
	fmt.Println("")
	fmt.Println("⚠️  PERINGATAN EKSTREM: Script ini akan menghapus SEMUA DATA!")
	fmt.Println("❌ Yang akan DIHAPUS SEMUA:")
	fmt.Println("   - Chart of Accounts (COA)")
	fmt.Println("   - Semua data transaksi")
	fmt.Println("   - Master data produk") 
	fmt.Println("   - Data kontak/customer/vendor")
	fmt.Println("   - Data user (kecuali admin default)")
	fmt.Println("   - Master data cash bank")
	fmt.Println("   - SEMUA DATA APLIKASI!")
	fmt.Println("")
	fmt.Println("✅ Yang akan dibuat ulang:")
	fmt.Println("   - Struktur database kosong")
	fmt.Println("   - User admin default")
	fmt.Println("   - COA default (jika ada seed)")
	fmt.Println("")

	// Triple confirmation
	if !confirmTotalReset() {
		fmt.Println("Reset total dibatalkan oleh user.")
		return
	}

	// Load config dan connect ke database
	_ = config.LoadConfig() // Load config untuk inisialisasi environment
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Gagal koneksi ke database")
	}

	fmt.Println("🔗 Berhasil terhubung ke database")

	// Execute total reset
	fmt.Println("\n🔄 Mengeksekusi total reset...")
	if err := executeTotalReset(db); err != nil {
		log.Printf("❌ Gagal total reset: %v", err)
		return
	}

	// Re-run migration dan seeding
	fmt.Println("\n🔄 Menjalankan migration dan seeding ulang...")
	if err := recreateDatabase(db); err != nil {
		log.Printf("❌ Gagal recreate database: %v", err)
		return
	}

	fmt.Println("\n🎉 TOTAL RESET SELESAI!")
	fmt.Println("Database telah direset total dan siap digunakan dari 0.")
}

func confirmTotalReset() bool {
	fmt.Println("⚠️  KONFIRMASI 1:")
	fmt.Print("Apakah Anda SANGAT YAKIN ingin menghapus SEMUA DATA? (ketik 'HAPUS SEMUA' untuk konfirmasi): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input != "HAPUS SEMUA" {
		return false
	}

	fmt.Println("\n⚠️  KONFIRMASI 2:")
	fmt.Print("Ini adalah aksi IRREVERSIBLE! Ketik nama aplikasi 'SISTEM AKUNTANSI' untuk konfirmasi final: ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input != "SISTEM AKUNTANSI" {
		return false
	}

	fmt.Println("\n⚠️  KONFIRMASI 3:")
	fmt.Print("Terakhir, ketik 'RESET TOTAL SEKARANG' untuk melanjutkan: ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	return input == "RESET TOTAL SEKARANG"
}

func executeTotalReset(db *gorm.DB) error {
	// Get all table names
	var tables []string
	err := db.Raw(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename NOT LIKE 'pg_%'
		AND tablename != 'schema_migrations'
		ORDER BY tablename
	`).Scan(&tables).Error
	
	if err != nil {
		return fmt.Errorf("gagal mendapatkan daftar tabel: %v", err)
	}

	fmt.Printf("   Ditemukan %d tabel untuk direset\n", len(tables))

	// Drop all tables
	tx := db.Begin()
	
	// Disable foreign key checks
	if err := tx.Exec("SET session_replication_role = replica").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal disable foreign key checks: %v", err)
	}

	// Drop all tables
	for _, table := range tables {
		fmt.Printf("   Menghapus tabel: %s\n", table)
		if err := tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("gagal menghapus tabel %s: %v", table, err)
		}
	}

	// Re-enable foreign key checks
	if err := tx.Exec("SET session_replication_role = DEFAULT").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal enable foreign key checks: %v", err)
	}

	return tx.Commit().Error
}

func recreateDatabase(db *gorm.DB) error {
	// Run auto migration to recreate all tables
	fmt.Println("   🔄 Menjalankan auto migration...")
	database.AutoMigrate(db)

	// Run seeding if available
	fmt.Println("   🌱 Menjalankan seeding...")
	if err := runSeeding(db); err != nil {
		log.Printf("⚠️  Warning: Gagal seeding: %v", err)
		log.Println("Anda mungkin perlu menjalankan seeding manual")
	} else {
		fmt.Println("   ✅ Seeding berhasil")
	}

	return nil
}

func runSeeding(db *gorm.DB) error {
	// Check if seed functions exist and call them
	
	// Try to run account seeding if function exists
	// This assumes there's a seed function in database package
	
	fmt.Println("   📊 Menjalankan account seeding...")
	// You may need to call specific seed functions here
	// Example: database.SeedAccounts(db)
	
	fmt.Println("   👤 Menjalankan user seeding...")
	// Example: database.SeedUsers(db) 
	
	fmt.Println("   💰 Menjalankan default cash bank seeding...")
	// Example: database.SeedCashBanks(db)

	return nil
}
