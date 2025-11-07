package database

import (
	"fmt"
	"log"
	
	"gorm.io/gorm"
)

// RunPreventDuplicateAccountsMigration applies the duplicate account prevention migration
func RunPreventDuplicateAccountsMigration(db *gorm.DB) error {
	log.Println("🔒 Starting duplicate accounts prevention migration...")
	
	// Don't execute the full SQL file - use inline migration instead
	// The SQL file contains multiple commands that can't be executed in prepared statement
	log.Println("Using inline migration for better control...")
	return applyInlineMigration(db)
}

// applyInlineMigration applies the migration without reading from file
func applyInlineMigration(db *gorm.DB) error {
	log.Println("📝 Applying inline migration for duplicate prevention...")
	
	// Step 1: Check for existing duplicates
	type DuplicateCheck struct {
		Code  string
		Count int
	}
	
	var duplicates []DuplicateCheck
	err := db.Raw(`
		SELECT code, COUNT(*) as count
		FROM accounts
		WHERE deleted_at IS NULL AND is_header = false
		GROUP BY code
		HAVING COUNT(*) > 1
	`).Scan(&duplicates).Error
	
	if err != nil {
		return fmt.Errorf("failed to check duplicates: %w", err)
	}
	
	if len(duplicates) > 0 {
		log.Printf("⚠️  Found %d duplicate account codes", len(duplicates))
		for _, dup := range duplicates {
			log.Printf("   - Code %s has %d instances", dup.Code, dup.Count)
		}
		log.Println("🔧 Merging duplicate accounts...")
		
		// Merge duplicates
		if err := mergeDuplicateAccounts(db); err != nil {
			return fmt.Errorf("failed to merge duplicates: %w", err)
		}
	} else {
		log.Println("✅ No duplicate accounts found")
	}
	
	// Step 2: Create unique index (case-insensitive)
	log.Println("🔒 Creating unique constraint on account code...")
	
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_code_unique_active
		ON accounts (LOWER(code))
		WHERE deleted_at IS NULL AND is_header = false
	`).Error; err != nil {
		log.Printf("⚠️  Warning: Could not create unique index: %v", err)
	} else {
		log.Println("✅ Unique index created successfully")
	}
	
	// Step 3: Add validation trigger (execute statements separately)
	log.Println("🔧 Installing validation trigger...")
	
	// Create function
	if err := db.Exec(`
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
				RAISE EXCEPTION 'Account code % already exists (case-insensitive check)', NEW.code
					USING HINT = 'Use unique account codes only',
						 ERRCODE = '23505';  -- unique_violation error code
			END IF;
			
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql
	`).Error; err != nil {
		log.Printf("⚠️  Warning: Could not create function: %v", err)
		return fmt.Errorf("failed to create function: %w", err)
	}
	
	// Drop existing trigger
	if err := db.Exec(`DROP TRIGGER IF EXISTS trg_prevent_duplicate_account_code ON accounts`).Error; err != nil {
		log.Printf("⚠️  Warning: Could not drop old trigger: %v", err)
	}
	
	// Create trigger
	if err := db.Exec(`
		CREATE TRIGGER trg_prevent_duplicate_account_code
			BEFORE INSERT OR UPDATE OF code ON accounts
			FOR EACH ROW
			EXECUTE FUNCTION prevent_duplicate_account_code()
	`).Error; err != nil {
		log.Printf("⚠️  Warning: Could not create trigger: %v", err)
		return fmt.Errorf("failed to create trigger: %w", err)
	}
	
	log.Println("✅ Validation trigger installed successfully")
	
	// Step 4: Create monitoring view
	log.Println("📊 Creating monitoring view...")
	
	if err := db.Exec(`
		CREATE OR REPLACE VIEW v_potential_duplicate_accounts AS
		SELECT 
			a1.code,
			a1.id as id1,
			a1.name as name1,
			a2.id as id2,
			a2.name as name2,
			a1.created_at as created_at1,
			a2.created_at as created_at2
		FROM accounts a1
		INNER JOIN accounts a2 ON a1.code = a2.code AND a1.id < a2.id
		WHERE a1.deleted_at IS NULL AND a2.deleted_at IS NULL
		ORDER BY a1.code
	`).Error; err != nil {
		log.Printf("⚠️  Warning: Could not create monitoring view: %v", err)
	} else {
		log.Println("✅ Monitoring view created successfully")
	}
	
	log.Println("✅ Inline migration completed successfully!")
	return nil
}

// mergeDuplicateAccounts merges duplicate accounts
func mergeDuplicateAccounts(db *gorm.DB) error {
	type Duplicate struct {
		Code  string
		Count int
	}
	
	var duplicates []Duplicate
	if err := db.Raw(`
		SELECT code, COUNT(*) as count
		FROM accounts
		WHERE deleted_at IS NULL AND is_header = false
		GROUP BY code
		HAVING COUNT(*) > 1
	`).Scan(&duplicates).Error; err != nil {
		return err
	}
	
	for _, dup := range duplicates {
		log.Printf("🔄 Merging duplicate account code: %s (%d instances)", dup.Code, dup.Count)
		
		// Get primary account (oldest)
		var primaryID uint
		if err := db.Raw(`
			SELECT id FROM accounts
			WHERE code = ? AND deleted_at IS NULL
			ORDER BY id ASC
			LIMIT 1
		`, dup.Code).Scan(&primaryID).Error; err != nil {
			log.Printf("❌ Failed to find primary account for %s: %v", dup.Code, err)
			continue
		}
		
		// Update unified journal lines
		if err := db.Exec(`
			UPDATE unified_journal_lines
			SET account_id = ?
			WHERE account_id IN (
				SELECT id FROM accounts 
				WHERE code = ? AND id != ? AND deleted_at IS NULL
			)
		`, primaryID, dup.Code, primaryID).Error; err != nil {
			log.Printf("⚠️  Warning: Could not update unified journal lines for %s: %v", dup.Code, err)
		}
		
		// Update legacy journal lines
		if err := db.Exec(`
			UPDATE journal_lines
			SET account_id = ?
			WHERE account_id IN (
				SELECT id FROM accounts 
				WHERE code = ? AND id != ? AND deleted_at IS NULL
			)
		`, primaryID, dup.Code, primaryID).Error; err != nil {
			log.Printf("⚠️  Warning: Could not update legacy journal lines for %s: %v", dup.Code, err)
		}
		
		// Soft delete duplicates
		result := db.Exec(`
			UPDATE accounts
			SET deleted_at = NOW(),
				name = name || ' (MERGED - DUPLICATE)',
				is_active = false
			WHERE code = ? AND id != ? AND deleted_at IS NULL
		`, dup.Code, primaryID)
		
		if result.Error != nil {
			log.Printf("❌ Failed to soft delete duplicates for %s: %v", dup.Code, result.Error)
			continue
		}
		
		log.Printf("✅ Merged %d duplicate(s) for account %s into ID %d", result.RowsAffected, dup.Code, primaryID)
	}
	
	return nil
}

// CheckDuplicateAccounts checks for any duplicate accounts
func CheckDuplicateAccounts(db *gorm.DB) (bool, error) {
	var count int64
	err := db.Raw(`
		SELECT COUNT(*)
		FROM (
			SELECT code
			FROM accounts
			WHERE deleted_at IS NULL AND is_header = false
			GROUP BY code
			HAVING COUNT(*) > 1
		) duplicates
	`).Count(&count).Error
	
	return count > 0, err
}

