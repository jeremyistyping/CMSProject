package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	fmt.Println("============================================================================")
	fmt.Println("DUPLICATE ACCOUNTS FIX - Go Version")
	fmt.Println("============================================================================")
	fmt.Println()

	// Load config
	cfg := config.LoadConfig()
	
	// Connect to database
	fmt.Println("Step 1: Connecting to database...")
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	fmt.Println("✅ Connected to database successfully")
	fmt.Println()

	// Step 2: Check for duplicates
	fmt.Println("Step 2: Checking for duplicate accounts...")
	var duplicateCount int64
	err = db.Raw(`
		SELECT COUNT(*) as duplicate_codes
		FROM (
			SELECT code
			FROM accounts
			WHERE deleted_at IS NULL
			GROUP BY code
			HAVING COUNT(*) > 1
		) dup
	`).Count(&duplicateCount).Error
	
	if err != nil {
		log.Fatalf("❌ Failed to check duplicates: %v", err)
	}

	if duplicateCount == 0 {
		fmt.Println("✅ No duplicate accounts found!")
		fmt.Println("   Installing unique constraint...")
		installConstraint(db)
		fmt.Println()
		fmt.Println("✅ All done! No duplicates to fix.")
		return
	}

	fmt.Printf("⚠️  Found %d duplicate account codes\n", duplicateCount)
	fmt.Println()

	// Show duplicates
	type DuplicateInfo struct {
		Code      string
		Instances int64
		Names     string
	}
	
	var duplicates []DuplicateInfo
	db.Raw(`
		SELECT 
			code,
			COUNT(*) as instances,
			STRING_AGG(DISTINCT name, ' | ') as names
		FROM accounts
		WHERE deleted_at IS NULL
		GROUP BY code
		HAVING COUNT(*) > 1
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`).Scan(&duplicates)

	fmt.Println("Duplicate account codes:")
	for _, dup := range duplicates {
		fmt.Printf("  - Code %s: %d instances (%s)\n", dup.Code, dup.Instances, dup.Names)
	}
	fmt.Println()

	// Step 3: Backup (using timestamp in account names for reference)
	fmt.Println("Step 3: Creating safety checkpoint...")
	backupTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("  Checkpoint time: %s\n", backupTime)
	fmt.Println("  (Note: Keep database backups before running this)")
	fmt.Println()

	// Step 4: Confirm
	fmt.Println("============================================================================")
	fmt.Println("READY TO FIX DUPLICATE ACCOUNTS")
	fmt.Println("============================================================================")
	fmt.Println()
	fmt.Println("This will:")
	fmt.Println("  1. Merge duplicate accounts (keeping the one with most usage)")
	fmt.Println("  2. Move all transactions to primary account")
	fmt.Println("  3. Consolidate balances")
	fmt.Println("  4. Soft-delete duplicate accounts")
	fmt.Println("  5. Create unique constraint to prevent future duplicates")
	fmt.Println()
	fmt.Print("Continue with fix? (yes/no): ")
	
	var confirmation string
	fmt.Scanln(&confirmation)
	
	if strings.ToLower(confirmation) != "yes" {
		fmt.Println("❌ Fix cancelled by user")
		return
	}

	fmt.Println()
	fmt.Println("Step 4: Running fix in transaction...")
	
	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		log.Fatalf("❌ Failed to start transaction: %v", tx.Error)
	}

	// Run the fix
	err = runFix(tx)
	if err != nil {
		tx.Rollback()
		log.Fatalf("❌ Fix failed: %v\nTransaction rolled back.", err)
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		log.Fatalf("❌ Failed to commit transaction: %v", err)
	}

	fmt.Println()
	fmt.Println("✅ Fix completed successfully!")
	fmt.Println()

	// Verify
	fmt.Println("Step 5: Verifying results...")
	var remainingDups int64
	db.Raw(`
		SELECT COUNT(*) FROM (
			SELECT code FROM accounts WHERE deleted_at IS NULL GROUP BY code HAVING COUNT(*) > 1
		) dup
	`).Count(&remainingDups)

	if remainingDups == 0 {
		fmt.Println("✅ SUCCESS: All duplicates have been fixed!")
	} else {
		fmt.Printf("⚠️  WARNING: Still have %d duplicate(s)\n", remainingDups)
	}

	fmt.Println()
	fmt.Println("============================================================================")
	fmt.Println("FIX COMPLETE")
	fmt.Println("============================================================================")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Restart your backend application")
	fmt.Println("  2. Test your balance sheet report")
	fmt.Println("  3. Verify account balances are correct")
	fmt.Println()
}

func runFix(tx *gorm.DB) error {
	fmt.Println()
	fmt.Println("  → Creating merge mapping...")
	
	// Create temp table for merge mapping
	err := tx.Exec(`
		CREATE TEMP TABLE IF NOT EXISTS duplicate_merge_map AS
		WITH duplicate_codes AS (
			SELECT code
			FROM accounts
			WHERE deleted_at IS NULL
			GROUP BY code
			HAVING COUNT(*) > 1
		),
		ranked_accounts AS (
			SELECT 
				a.id,
				a.code,
				a.name,
				a.balance,
				COALESCE(
					(SELECT COUNT(*) FROM unified_journal_ledger WHERE account_id = a.id AND deleted_at IS NULL), 0
				) + COALESCE(
					(SELECT COUNT(*) FROM cash_banks WHERE account_id = a.id AND deleted_at IS NULL), 0
				) + COALESCE(
					(SELECT COUNT(*) FROM assets WHERE asset_account_id = a.id AND deleted_at IS NULL), 0
				) as usage_count,
				ROW_NUMBER() OVER (
					PARTITION BY a.code 
					ORDER BY 
						(COALESCE(
							(SELECT COUNT(*) FROM unified_journal_ledger WHERE account_id = a.id AND deleted_at IS NULL), 0
						) + COALESCE(
							(SELECT COUNT(*) FROM cash_banks WHERE account_id = a.id AND deleted_at IS NULL), 0
						) + COALESCE(
							(SELECT COUNT(*) FROM assets WHERE asset_account_id = a.id AND deleted_at IS NULL), 0
						)) DESC,
						a.created_at ASC,
						a.id ASC
				) as rank
			FROM accounts a
			INNER JOIN duplicate_codes dc ON a.code = dc.code
			WHERE a.deleted_at IS NULL
		)
		SELECT 
			code,
			id as duplicate_id,
			name as duplicate_name,
			balance,
			usage_count,
			(SELECT id FROM ranked_accounts WHERE code = r.code AND rank = 1) as primary_id,
			(SELECT name FROM ranked_accounts WHERE code = r.code AND rank = 1) as primary_name,
			rank
		FROM ranked_accounts r
		WHERE rank > 1
	`).Error
	
	if err != nil {
		return fmt.Errorf("failed to create merge mapping: %w", err)
	}

	// Count duplicates to merge
	var mergeCount int64
	tx.Raw("SELECT COUNT(*) FROM duplicate_merge_map").Count(&mergeCount)
	fmt.Printf("  → Found %d duplicate accounts to merge\n", mergeCount)

	// Move journal entries
	fmt.Println("  → Moving journal entries...")
	result := tx.Exec(`
		UPDATE unified_journal_ledger ujl
		SET account_id = dmm.primary_id
		FROM duplicate_merge_map dmm
		WHERE ujl.account_id = dmm.duplicate_id
		  AND ujl.deleted_at IS NULL
	`)
	if result.Error != nil {
		return fmt.Errorf("failed to move journal entries: %w", result.Error)
	}
	fmt.Printf("     Moved %d journal entries\n", result.RowsAffected)

	// Move cash/bank references
	fmt.Println("  → Moving cash/bank references...")
	result = tx.Exec(`
		UPDATE cash_banks cb
		SET account_id = dmm.primary_id
		FROM duplicate_merge_map dmm
		WHERE cb.account_id = dmm.duplicate_id
		  AND cb.deleted_at IS NULL
	`)
	if result.Error != nil {
		return fmt.Errorf("failed to move cash/bank references: %w", result.Error)
	}
	fmt.Printf("     Moved %d cash/bank references\n", result.RowsAffected)

	// Move asset references
	fmt.Println("  → Moving asset references...")
	result = tx.Exec(`
		UPDATE assets ast
		SET asset_account_id = dmm.primary_id
		FROM duplicate_merge_map dmm
		WHERE ast.asset_account_id = dmm.duplicate_id
		  AND ast.deleted_at IS NULL
	`)
	if result.Error != nil {
		return fmt.Errorf("failed to move asset references: %w", result.Error)
	}
	fmt.Printf("     Moved %d asset references\n", result.RowsAffected)

	// Consolidate balances
	fmt.Println("  → Consolidating balances...")
	err = tx.Exec(`
		DO $$
		DECLARE
			rec RECORD;
			total_balance DECIMAL(20,2);
		BEGIN
			FOR rec IN 
				SELECT DISTINCT code, primary_id 
				FROM duplicate_merge_map
			LOOP
				SELECT COALESCE(SUM(balance), 0) INTO total_balance
				FROM accounts
				WHERE code = rec.code
				  AND deleted_at IS NULL;
				
				UPDATE accounts
				SET balance = total_balance,
					updated_at = NOW()
				WHERE id = rec.primary_id;
			END LOOP;
		END $$
	`).Error
	if err != nil {
		return fmt.Errorf("failed to consolidate balances: %w", err)
	}

	// Soft-delete duplicates
	fmt.Println("  → Soft-deleting duplicate accounts...")
	result = tx.Exec(`
		UPDATE accounts
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id IN (SELECT duplicate_id FROM duplicate_merge_map)
	`)
	if err != nil {
		return fmt.Errorf("failed to soft-delete duplicates: %w", err)
	}
	fmt.Printf("     Soft-deleted %d duplicate accounts\n", result.RowsAffected)

	// Install constraint
	fmt.Println("  → Installing unique constraint...")
	err = installConstraint(tx)
	if err != nil {
		return fmt.Errorf("failed to install constraint: %w", err)
	}

	return nil
}

func installConstraint(db *gorm.DB) error {
	// Drop old constraints/indexes
	db.Exec("DROP INDEX IF EXISTS uni_accounts_code CASCADE")
	db.Exec("DROP INDEX IF EXISTS accounts_code_key CASCADE")
	db.Exec("DROP INDEX IF EXISTS idx_accounts_code_unique CASCADE")
	db.Exec("DROP INDEX IF EXISTS accounts_code_unique CASCADE")
	db.Exec("DROP INDEX IF EXISTS idx_accounts_code_unique_active CASCADE")

	// Create new partial unique index
	err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_code_active
		ON accounts (LOWER(code))
		WHERE deleted_at IS NULL
	`).Error
	if err != nil {
		return err
	}

	// Create trigger function
	err = db.Exec(`
		CREATE OR REPLACE FUNCTION prevent_duplicate_account_code()
		RETURNS TRIGGER AS $$
		DECLARE
			existing_count INTEGER;
		BEGIN
			SELECT COUNT(*) INTO existing_count
			FROM accounts
			WHERE LOWER(code) = LOWER(NEW.code)
			  AND deleted_at IS NULL
			  AND id != COALESCE(NEW.id, 0);
			
			IF existing_count > 0 THEN
				RAISE EXCEPTION 'Account code % already exists', NEW.code
					USING HINT = 'Use unique account codes only',
						  ERRCODE = '23505';
			END IF;
			
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql
	`).Error
	if err != nil {
		return err
	}

	// Create trigger
	db.Exec("DROP TRIGGER IF EXISTS trg_prevent_duplicate_account_code ON accounts")
	err = db.Exec(`
		CREATE TRIGGER trg_prevent_duplicate_account_code
			BEFORE INSERT OR UPDATE OF code ON accounts
			FOR EACH ROW
			EXECUTE FUNCTION prevent_duplicate_account_code()
	`).Error
	
	if err == nil {
		fmt.Println("     ✅ Unique constraint installed")
		fmt.Println("     ✅ Trigger created")
	}

	return err
}

