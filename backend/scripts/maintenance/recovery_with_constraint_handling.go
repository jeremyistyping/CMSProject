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
	fmt.Println("=== RECOVERY DENGAN CONSTRAINT HANDLING ===")
	fmt.Println("Script ini akan menangani recovery data dengan memperhatikan unique constraints")

	// Load config dan connect ke database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Gagal koneksi ke database")
	}

	fmt.Println("🔗 Berhasil terhubung ke database")

	// Tampilkan summary data yang soft deleted
	showSoftDeletedSummary(db)

	// Recovery step by step dengan constraint handling
	if err := recoverWithConstraintHandling(db); err != nil {
		log.Printf("❌ Gagal recovery: %v", err)
		return
	}

	fmt.Println("✅ Recovery selesai!")
}

func showSoftDeletedSummary(db *gorm.DB) {
	fmt.Println("\n📊 Summary data yang soft deleted:")

	// Check accounts specifically
	var accountsDeleted int64
	db.Raw("SELECT COUNT(*) FROM accounts WHERE deleted_at IS NOT NULL").Scan(&accountsDeleted)
	fmt.Printf("   - Accounts (soft deleted): %d\n", accountsDeleted)

	// Check for duplicate codes that might cause constraint issues
	type duplicateCode struct {
		Code  string
		Count int64
	}

	var duplicates []duplicateCode
	db.Raw(`
		SELECT code, COUNT(*) as count
		FROM accounts 
		WHERE code IS NOT NULL AND code != ''
		GROUP BY code 
		HAVING COUNT(*) > 1
		ORDER BY count DESC
	`).Scan(&duplicates)

	if len(duplicates) > 0 {
		fmt.Printf("   ⚠️  Found %d duplicate account codes:\n", len(duplicates))
		for _, dup := range duplicates {
			fmt.Printf("      - Code '%s': %d records\n", dup.Code, dup.Count)
		}
	} else {
		fmt.Println("   ✅ No duplicate account codes found")
	}
}

func recoverWithConstraintHandling(db *gorm.DB) error {
	start := time.Now()
	fmt.Printf("\n⏳ Starting recovery with constraint handling... (started: %s)\n", start.Format("15:04:05"))

	// Step 1: Handle accounts table separately due to unique constraints
	if err := recoverAccountsWithConstraintHandling(db); err != nil {
		return fmt.Errorf("failed to recover accounts: %v", err)
	}

	// Step 2: Recover other tables that don't have constraint issues
	if err := recoverOtherTables(db); err != nil {
		return fmt.Errorf("failed to recover other tables: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("✅ Recovery completed in: %s\n", duration.String())
	return nil
}

func recoverAccountsWithConstraintHandling(db *gorm.DB) error {
	fmt.Println("🔄 Recovering accounts table with constraint handling...")

	// First, check the constraint that's causing issues
	var constraintName string
	db.Raw(`
		SELECT conname 
		FROM pg_constraint 
		WHERE conrelid = 'accounts'::regclass 
		AND contype = 'u'
		AND conname LIKE '%code%'
	`).Scan(&constraintName)

	fmt.Printf("   Found constraint: %s\n", constraintName)

	// Strategy: Temporarily disable the constraint, then re-enable
	tx := db.Begin()

	// Option 1: Try to recover accounts by handling duplicates
	var duplicateAccounts []struct {
		ID       uint   `json:"id"`
		Code     string `json:"code"`
		Name     string `json:"name"`
		DeletedAt *time.Time `json:"deleted_at"`
	}

	// Find accounts that would conflict
	tx.Raw(`
		SELECT id, code, name, deleted_at
		FROM accounts 
		WHERE deleted_at IS NOT NULL
		AND code IN (
			SELECT code 
			FROM accounts 
			WHERE deleted_at IS NULL
			AND code IS NOT NULL AND code != ''
		)
		ORDER BY code, created_at
	`).Scan(&duplicateAccounts)

	if len(duplicateAccounts) > 0 {
		fmt.Printf("   ⚠️  Found %d conflicting accounts that cannot be recovered as-is\n", len(duplicateAccounts))
		for _, acc := range duplicateAccounts {
			fmt.Printf("      - ID: %d, Code: %s, Name: %s\n", acc.ID, acc.Code, acc.Name)
		}

		fmt.Println("   📝 These accounts will need manual intervention")
		
		// For now, recover non-conflicting accounts only
		result := tx.Exec(`
			UPDATE accounts 
			SET deleted_at = NULL 
			WHERE deleted_at IS NOT NULL
			AND code NOT IN (
				SELECT code 
				FROM accounts 
				WHERE deleted_at IS NULL
				AND code IS NOT NULL AND code != ''
			)
		`)

		if result.Error != nil {
			tx.Rollback()
			return fmt.Errorf("failed to recover non-conflicting accounts: %v", result.Error)
		}

		fmt.Printf("   ✅ Recovered %d non-conflicting accounts\n", result.RowsAffected)
	} else {
		// No conflicts, safe to recover all
		result := tx.Exec("UPDATE accounts SET deleted_at = NULL WHERE deleted_at IS NOT NULL")
		if result.Error != nil {
			tx.Rollback()
			return fmt.Errorf("failed to recover accounts: %v", result.Error)
		}

		fmt.Printf("   ✅ Recovered %d accounts\n", result.RowsAffected)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit accounts recovery: %v", err)
	}

	return nil
}

func recoverOtherTables(db *gorm.DB) error {
	fmt.Println("🔄 Recovering other tables...")

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
	totalRecovered := 0

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

			totalRecovered += int(result.RowsAffected)
			fmt.Printf("   ✅ Recovered: %s (%d records)\n", name, result.RowsAffected)
		} else {
			fmt.Printf("   ℹ  Skipped: %s (no deleted records)\n", name)
		}
		recovered++
	}

	fmt.Printf("✅ Processed %d tables, recovered %d total records\n", recovered, totalRecovered)
	return nil
}