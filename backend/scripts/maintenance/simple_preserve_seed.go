package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== SIMPLE PRESERVE SEED DATA ===")
	fmt.Println("Script ini akan menghapus SEMUA data soft deleted dan kemudian menjalankan seed")

	// Load config dan connect ke database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Gagal koneksi ke database")
	}

	fmt.Println("🔗 Berhasil terhubung ke database")

	// Show current situation
	showSummary(db)

	fmt.Println("\n📝 Strategy:")
	fmt.Println("   1. Hard delete semua record yang soft deleted")
	fmt.Println("   2. Jalankan seed untuk memastikan data master lengkap")
	fmt.Println("   3. Data yang tersisa adalah data aktif + data seed baru")

	// Perform cleanup
	if err := hardDeleteSoftDeleted(db); err != nil {
		log.Printf("❌ Gagal cleanup: %v", err)
		return
	}

	fmt.Println("\n✅ Cleanup berhasil! Data sekarang bersih tanpa duplikat.")
	fmt.Println("🌱 Silakan jalankan seed manual: go run cmd/seed.go")
}

func showSummary(db *gorm.DB) {
	fmt.Println("\n📊 Current situation:")

	var totalAccounts, activeAccounts, deletedAccounts int64
	db.Raw("SELECT COUNT(*) FROM accounts").Scan(&totalAccounts)
	db.Raw("SELECT COUNT(*) FROM accounts WHERE deleted_at IS NULL").Scan(&activeAccounts) 
	db.Raw("SELECT COUNT(*) FROM accounts WHERE deleted_at IS NOT NULL").Scan(&deletedAccounts)

	fmt.Printf("   - Total accounts: %d\n", totalAccounts)
	fmt.Printf("   - Active accounts: %d\n", activeAccounts)
	fmt.Printf("   - Soft deleted accounts: %d (akan dihapus permanen)\n", deletedAccounts)

	// Show other soft deleted data
	type tbl struct{ TableName string }
	var tables []tbl
	
	db.Raw(`
		SELECT table_name AS table_name
		FROM information_schema.columns
		WHERE table_schema = 'public' 
		AND column_name = 'deleted_at'
		AND table_name NOT IN (
			'accounts_backup',
			'accounts_hierarchy_backup', 
			'accounts_original_balances',
			'schema_migrations',
			'gorm_migrations'
		)
		ORDER BY table_name
	`).Scan(&tables)

	totalSoftDeleted := 0
	for _, t := range tables {
		var count int64
		db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NOT NULL", t.TableName)).Scan(&count)
		totalSoftDeleted += int(count)
	}

	fmt.Printf("   - Total soft deleted records across all tables: %d (akan dihapus permanen)\n", totalSoftDeleted)
}

func hardDeleteSoftDeleted(db *gorm.DB) error {
	start := time.Now()
	fmt.Printf("\n⏳ Starting hard delete of soft deleted records... (started: %s)\n", start.Format("15:04:05"))

	// Get all tables with deleted_at column
	type tbl struct{ TableName string }
	var tables []tbl
	
	db.Raw(`
		SELECT table_name AS table_name
		FROM information_schema.columns
		WHERE table_schema = 'public' 
		AND column_name = 'deleted_at'
		AND table_name NOT IN (
			'accounts_backup',
			'accounts_hierarchy_backup', 
			'accounts_original_balances',
			'schema_migrations',
			'gorm_migrations'
		)
		ORDER BY table_name
	`).Scan(&tables)

	totalDeleted := 0

	// Process each table
	for _, t := range tables {
		name := t.TableName
		
		// Count soft deleted records first
		var count int64
		db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NOT NULL", name)).Scan(&count)

		if count > 0 {
			// Hard delete soft deleted records
			result := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE deleted_at IS NOT NULL", name))
			if result.Error != nil {
				fmt.Printf("   ⚠️  Warning: Failed to delete from %s: %v\n", name, result.Error)
				continue
			}

			totalDeleted += int(result.RowsAffected)
			fmt.Printf("   ✅ Deleted: %s (%d records)\n", name, result.RowsAffected)
		} else {
			fmt.Printf("   ℹ  Skipped: %s (no soft deleted records)\n", name)
		}
	}

	duration := time.Since(start)
	fmt.Printf("✅ Hard delete completed in %s. Total deleted: %d records\n", duration.String(), totalDeleted)
	return nil
}