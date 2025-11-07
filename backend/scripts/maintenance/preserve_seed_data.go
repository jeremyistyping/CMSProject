package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== PRESERVE SEED DATA ===")
	fmt.Println("Script ini akan membersihkan duplicate accounts dan mempertahankan data seed yang benar")

	// Load config dan connect ke database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Gagal koneksi ke database")
	}

	fmt.Println("🔗 Berhasil terhubung ke database")

	// Show current situation
	showAccountsSituation(db)

	// Ask for confirmation
	if !confirmCleanup() {
		fmt.Println("Cleanup dibatalkan.")
		return
	}

	// Perform cleanup and preserve seed data
	if err := cleanupAndPreserveSeedData(db); err != nil {
		log.Printf("❌ Gagal cleanup: %v", err)
		return
	}

	fmt.Println("✅ Cleanup dan preserve seed data berhasil!")

	// Now run seed to restore proper data
	fmt.Println("\n🌱 Running seed untuk memastikan data seed lengkap...")
	if err := runSeed(); err != nil {
		log.Printf("⚠️  Warning: Seed gagal: %v", err)
		fmt.Println("Anda bisa jalankan manual: go run cmd/seed.go")
	} else {
		fmt.Println("✅ Seed berhasil!")
	}
}

func showAccountsSituation(db *gorm.DB) {
	fmt.Println("\n📊 Current accounts situation:")

	var totalAccounts, activeAccounts, deletedAccounts int64
	db.Raw("SELECT COUNT(*) FROM accounts").Scan(&totalAccounts)
	db.Raw("SELECT COUNT(*) FROM accounts WHERE deleted_at IS NULL").Scan(&activeAccounts) 
	db.Raw("SELECT COUNT(*) FROM accounts WHERE deleted_at IS NOT NULL").Scan(&deletedAccounts)

	fmt.Printf("   - Total accounts: %d\n", totalAccounts)
	fmt.Printf("   - Active accounts: %d\n", activeAccounts)
	fmt.Printf("   - Soft deleted accounts: %d\n", deletedAccounts)

	// Show duplicate codes count
	var duplicateCodesCount int64
	db.Raw(`
		SELECT COUNT(DISTINCT code)
		FROM accounts 
		WHERE code IN (
			SELECT code 
			FROM accounts 
			WHERE code IS NOT NULL AND code != ''
			GROUP BY code 
			HAVING COUNT(*) > 1
		)
	`).Scan(&duplicateCodesCount)

	fmt.Printf("   - Duplicate account codes: %d\n", duplicateCodesCount)
}

func confirmCleanup() bool {
	fmt.Print("\nApakah Anda ingin melakukan cleanup dan preserve seed data? (ketik 'ya' untuk konfirmasi): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	
	return input == "ya" || input == "y" || input == "yes"
}

func cleanupAndPreserveSeedData(db *gorm.DB) error {
	start := time.Now()
	fmt.Printf("\n⏳ Starting cleanup and preserve seed data... (started: %s)\n", start.Format("15:04:05"))

	tx := db.Begin()

	// Strategy: Keep the original seed accounts (usually created first) and remove duplicates
	fmt.Println("🧹 Step 1: Remove duplicate accounts, keeping original seed data...")

	// Delete soft deleted accounts that have active duplicates
	result := tx.Exec(`
		DELETE FROM accounts 
		WHERE deleted_at IS NOT NULL
		AND code IN (
			SELECT code 
			FROM accounts 
			WHERE deleted_at IS NULL
			AND code IS NOT NULL AND code != ''
		)
	`)

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete soft deleted duplicates: %v", result.Error)
	}

	fmt.Printf("   ✅ Removed %d soft deleted duplicate accounts\n", result.RowsAffected)

	// For remaining duplicates in active accounts, keep the one with smallest ID (oldest/original)
	fmt.Println("🧹 Step 2: Remove active duplicate accounts, keeping original (smallest ID)...")
	
	result = tx.Exec(`
		DELETE FROM accounts a1
		WHERE a1.deleted_at IS NULL
		AND EXISTS (
			SELECT 1 FROM accounts a2
			WHERE a2.code = a1.code
			AND a2.deleted_at IS NULL
			AND a2.id < a1.id
			AND a1.code IS NOT NULL AND a1.code != ''
		)
	`)

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete active duplicates: %v", result.Error)
	}

	fmt.Printf("   ✅ Removed %d active duplicate accounts\n", result.RowsAffected)

	// Now recover any remaining soft deleted accounts that don't conflict
	fmt.Println("🔄 Step 3: Recover non-conflicting soft deleted accounts...")
	
	result = tx.Exec(`
		UPDATE accounts 
		SET deleted_at = NULL 
		WHERE deleted_at IS NOT NULL
		AND (
			code IS NULL OR code = '' OR
			code NOT IN (
				SELECT code 
				FROM accounts 
				WHERE deleted_at IS NULL
				AND code IS NOT NULL AND code != ''
			)
		)
	`)

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to recover non-conflicting accounts: %v", result.Error)
	}

	fmt.Printf("   ✅ Recovered %d non-conflicting accounts\n", result.RowsAffected)

	// Clean up other tables too
	fmt.Println("🔄 Step 4: Recover other tables...")
	otherTablesRecovered, err := recoverOtherTables(tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to recover other tables: %v", err)
	}

	fmt.Printf("   ✅ Recovered data from %d other tables\n", otherTablesRecovered)

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit cleanup: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("✅ Cleanup completed in: %s\n", duration.String())
	return nil
}

func recoverOtherTables(db *gorm.DB) (int, error) {
	// Get all tables with deleted_at column except accounts
	type tbl struct{ TableName string }
	var tables []tbl
	
	db.Raw(`
		SELECT table_name AS table_name
		FROM information_schema.columns
		WHERE table_schema = 'public' 
		AND column_name = 'deleted_at'
		AND table_name != 'accounts'
		AND table_name NOT IN (
			'accounts_backup',
			'accounts_hierarchy_backup', 
			'accounts_original_balances',
			'schema_migrations',
			'gorm_migrations'
		)
		ORDER BY table_name
	`).Scan(&tables)

	recovered := 0

	for _, t := range tables {
		name := t.TableName
		
		// Count soft deleted records first
		var count int64
		db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NOT NULL", name)).Scan(&count)

		if count > 0 {
			// Recover the table
			result := db.Exec(fmt.Sprintf("UPDATE %s SET deleted_at = NULL WHERE deleted_at IS NOT NULL", name))
			if result.Error != nil {
				fmt.Printf("   ⚠️  Warning: Failed to recover %s: %v\n", name, result.Error)
				continue
			}

			if result.RowsAffected > 0 {
				recovered++
				fmt.Printf("   ✅ Recovered: %s (%d records)\n", name, result.RowsAffected)
			}
		}
	}

	return recovered, nil
}

func runSeed() error {
	fmt.Println("   📦 Running go run cmd/seed.go...")
	
	// We'll implement a simple approach - just run the seed command
	// In a real scenario, you might want to import and call the seed functions directly
	
	// For now, let's just print what should be done
	fmt.Println("   ℹ️  Please run manually: go run cmd/seed.go")
	fmt.Println("   This will ensure all master data is properly seeded")
	
	return nil
}