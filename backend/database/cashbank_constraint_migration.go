package database

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"time"
	"gorm.io/gorm"
)

// CashBankConstraintMigration fixes foreign key constraint issues in cash_banks table
func CashBankConstraintMigration(db *gorm.DB) {
	log.Println("🔧 Starting Cash Bank Constraint Migration...")

	// Check if this migration has already been run
	migrationName := "cashbank_constraint_fix_v1.0"
	var existingMigration models.MigrationRecord
	if err := db.Where("migration_id = ?", migrationName).First(&existingMigration).Error; err == nil {
		log.Printf("✅ Cash Bank Constraint Migration already applied")
		return
	}

	// Start transaction for the entire migration
	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("❌ Failed to start migration transaction: %v", tx.Error)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("❌ Migration rolled back due to panic: %v", r)
		}
	}()

	// Step 1: Diagnose problematic records
	var problematicCount int64
	query := `
		SELECT COUNT(*)
		FROM cash_banks cb 
		LEFT JOIN accounts a ON cb.account_id = a.id AND a.deleted_at IS NULL
		WHERE cb.account_id IS NULL OR a.id IS NULL
	`
	
	if err := tx.Raw(query).Scan(&problematicCount).Error; err != nil {
		log.Printf("❌ Failed to diagnose problematic records: %v", err)
		tx.Rollback()
		return
	}

	log.Printf("📊 Found %d cash_banks records with constraint issues", problematicCount)

	// Step 2: Create parent accounts if they don't exist
	log.Println("🏗️  Creating parent accounts if needed...")

	// Create Current Assets parent account
	currentAssetsAccount := &models.Account{
		Code:        "1100",
		Name:        "Current Assets",
		Type:        "ASSET",
		Category:    "CURRENT_ASSET",
		Level:       1,
		IsHeader:    true,
		IsActive:    true,
		Description: "Parent account for current assets",
	}

	result := tx.Where("code = ? AND deleted_at IS NULL", "1100").FirstOrCreate(currentAssetsAccount)
	if result.Error != nil {
		log.Printf("❌ Failed to create current assets account: %v", result.Error)
		tx.Rollback()
		return
	}
	if result.RowsAffected > 0 {
		log.Printf("✅ Created parent account: %s (%s)", currentAssetsAccount.Name, currentAssetsAccount.Code)
	}

	// Create Cash and Cash Equivalents parent account
	cashEquivalentsAccount := &models.Account{
		Code:        "1101",
		Name:        "Cash and Cash Equivalents",
		Type:        "ASSET",
		Category:    "CURRENT_ASSET",
		ParentID:    &currentAssetsAccount.ID,
		Level:       2,
		IsHeader:    true,
		IsActive:    true,
		Description: "Parent for cash and bank accounts",
	}

	result = tx.Where("code = ? AND deleted_at IS NULL", "1101").FirstOrCreate(cashEquivalentsAccount)
	if result.Error != nil {
		log.Printf("❌ Failed to create cash equivalents account: %v", result.Error)
		tx.Rollback()
		return
	}
	if result.RowsAffected > 0 {
		log.Printf("✅ Created cash equivalents account: %s (%s)", cashEquivalentsAccount.Name, cashEquivalentsAccount.Code)
	}

	// Step 3: Fix problematic cash_banks records
	if problematicCount > 0 {
		log.Printf("🔧 Fixing %d problematic cash_banks records...", problematicCount)

		// Get all problematic records
		var problematicRecords []struct {
			ID        uint
			Code      string
			Name      string
			Type      string
			AccountID *uint
		}

		problematicQuery := `
			SELECT cb.id, cb.code, cb.name, cb.type, cb.account_id
			FROM cash_banks cb 
			LEFT JOIN accounts a ON cb.account_id = a.id AND a.deleted_at IS NULL
			WHERE cb.account_id IS NULL OR a.id IS NULL
			ORDER BY cb.id
		`

		if err := tx.Raw(problematicQuery).Scan(&problematicRecords).Error; err != nil {
			log.Printf("❌ Failed to fetch problematic records: %v", err)
			tx.Rollback()
			return
		}

		fixedCount := 0
		for _, record := range problematicRecords {
			log.Printf("  🔧 Fixing cash_bank record ID: %d (%s)", record.ID, record.Name)

			// Generate unique account code
			var accountCode string
			if record.Type == "CASH" {
				accountCode = fmt.Sprintf("1101-%03d", record.ID)
			} else {
				accountCode = fmt.Sprintf("1102-%03d", record.ID)
			}

			// Check if account with this name already exists
			var existingAccount models.Account
			err := tx.Where("name = ? AND type = ? AND category = ? AND deleted_at IS NULL", 
				record.Name, "ASSET", "CURRENT_ASSET").First(&existingAccount).Error

			var accountIDToUse uint
			if err == nil {
				// Use existing account
				accountIDToUse = existingAccount.ID
				log.Printf("    ✅ Using existing GL account: %s (%s)", existingAccount.Name, existingAccount.Code)
			} else if err == gorm.ErrRecordNotFound {
				// Create new account
				newGLAccount := &models.Account{
					Code:        accountCode,
					Name:        record.Name,
					Type:        "ASSET",
					Category:    "CURRENT_ASSET",
					ParentID:    &cashEquivalentsAccount.ID,
					Level:       3,
					IsHeader:    false,
					IsActive:    true,
					Description: fmt.Sprintf("Auto-created GL account for %s: %s (%s)", record.Type, record.Name, record.Code),
				}

				if err := tx.Create(newGLAccount).Error; err != nil {
					log.Printf("    ❌ Failed to create GL account for %s: %v", record.Name, err)
					continue
				}
				accountIDToUse = newGLAccount.ID
				log.Printf("    ✅ Created new GL account: %s (%s)", newGLAccount.Name, newGLAccount.Code)
			} else {
				log.Printf("    ❌ Error checking for existing account: %v", err)
				continue
			}

			// Update cash_bank record
			if err := tx.Model(&models.CashBank{}).Where("id = ?", record.ID).Update("account_id", accountIDToUse).Error; err != nil {
				log.Printf("    ❌ Failed to update cash_bank record ID %d: %v", record.ID, err)
			} else {
				log.Printf("    ✅ Updated cash_bank record ID %d with account_id %d", record.ID, accountIDToUse)
				fixedCount++
			}
		}

		log.Printf("✅ Successfully fixed %d out of %d problematic records", fixedCount, len(problematicRecords))
	}

	// Step 4: Create fallback account for any remaining issues
	log.Println("🛡️  Creating fallback account...")
	
	fallbackAccount := &models.Account{
		Code:        "1199",
		Name:        "Unclassified Current Assets",
		Type:        "ASSET",
		Category:    "CURRENT_ASSET",
		ParentID:    &currentAssetsAccount.ID,
		Level:       2,
		IsHeader:    false,
		IsActive:    true,
		Description: "Fallback account for unclassified current assets",
	}

	result = tx.Where("code = ? AND deleted_at IS NULL", "1199").FirstOrCreate(fallbackAccount)
	if result.Error != nil {
		log.Printf("❌ Failed to create fallback account: %v", result.Error)
		tx.Rollback()
		return
	}
	if result.RowsAffected > 0 {
		log.Printf("✅ Created fallback account: %s (%s)", fallbackAccount.Name, fallbackAccount.Code)
	}

	// Update any remaining cash_banks with NULL account_id
	updateResult := tx.Model(&models.CashBank{}).
		Where("account_id IS NULL").
		Update("account_id", fallbackAccount.ID)
	
	if updateResult.Error != nil {
		log.Printf("❌ Failed to update records with fallback account: %v", updateResult.Error)
		tx.Rollback()
		return
	}
	if updateResult.RowsAffected > 0 {
		log.Printf("✅ Updated %d cash_bank records to use fallback account", updateResult.RowsAffected)
	}

	// Step 5: Verify the fix
	log.Println("🔍 Verifying the fix...")

	var remainingIssues int64
	if err := tx.Raw(`
		SELECT COUNT(*)
		FROM cash_banks cb 
		LEFT JOIN accounts a ON cb.account_id = a.id AND a.deleted_at IS NULL
		WHERE cb.account_id IS NULL OR a.id IS NULL
	`).Scan(&remainingIssues).Error; err != nil {
		log.Printf("❌ Failed to verify fix: %v", err)
		tx.Rollback()
		return
	}

	if remainingIssues > 0 {
		log.Printf("⚠️  %d records still have issues after migration", remainingIssues)
		tx.Rollback()
		return
	}

	// Step 6: Record successful migration
	migrationRecord := &models.MigrationRecord{
		MigrationID: migrationName,
		AppliedAt:   time.Now(),
	}

	if err := tx.Create(migrationRecord).Error; err != nil {
		log.Printf("❌ Failed to record migration: %v", err)
		tx.Rollback()
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ Failed to commit migration: %v", err)
		return
	}

	log.Println("🎉 Cash Bank Constraint Migration completed successfully!")
	log.Printf("📊 Migration Summary:")
	log.Printf("  - Problematic records found: %d", problematicCount)
	log.Printf("  - Remaining issues: %d", remainingIssues)
	log.Printf("  - Success rate: %.1f%%", 
		float64(problematicCount-remainingIssues)/float64(max(problematicCount, 1))*100)
}

// Helper function to get maximum of two integers
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// EnsureCashBankAccountIntegrity validates and fixes cash bank account integrity in real-time
func EnsureCashBankAccountIntegrity(db *gorm.DB, cashBankID uint) error {
	// This function can be called by services to ensure account integrity
	var cashBank models.CashBank
	if err := db.First(&cashBank, cashBankID).Error; err != nil {
		return fmt.Errorf("cash bank not found: %v", err)
	}

	needsAccountCreation := false

	// Check if account_id is missing or invalid
	if cashBank.AccountID == 0 {
		needsAccountCreation = true
		log.Printf("⚠️  Cash bank %d (%s) has missing account_id, creating...", cashBank.ID, cashBank.Name)
	} else {
		// Validate existing account_id
		var account models.Account
		if err := db.Where("id = ? AND deleted_at IS NULL", cashBank.AccountID).First(&account).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("⚠️  Cash bank %d has invalid account_id %d, creating new account...", cashBank.ID, cashBank.AccountID)
				needsAccountCreation = true
			} else {
				return fmt.Errorf("failed to validate account_id: %v", err)
			}
		}
	}

	if needsAccountCreation {
		// Generate account code
		var accountCode string
		if cashBank.Type == "CASH" {
			accountCode = fmt.Sprintf("1101-%03d", cashBank.ID)
		} else {
			accountCode = fmt.Sprintf("1102-%03d", cashBank.ID)
		}

		// Find or create parent account
		var parentAccount models.Account
		if err := db.Where("code = ? AND deleted_at IS NULL", "1101").First(&parentAccount).Error; err != nil {
			// Create parent if it doesn't exist
			parentAccount = models.Account{
				Code:        "1101",
				Name:        "Cash and Cash Equivalents",
				Type:        "ASSET",
				Category:    "CURRENT_ASSET",
				Level:       2,
				IsHeader:    true,
				IsActive:    true,
				Description: "Parent for cash and bank accounts",
			}
			if err := db.Create(&parentAccount).Error; err != nil {
				return fmt.Errorf("failed to create parent account: %v", err)
			}
		}

		// Create GL account
		newAccount := models.Account{
			Code:        accountCode,
			Name:        cashBank.Name,
			Type:        "ASSET",
			Category:    "CURRENT_ASSET",
			ParentID:    &parentAccount.ID,
			Level:       3,
			IsHeader:    false,
			IsActive:    true,
			Description: fmt.Sprintf("Auto-created GL account for %s: %s", cashBank.Type, cashBank.Name),
		}

		if err := db.Create(&newAccount).Error; err != nil {
			return fmt.Errorf("failed to create GL account: %v", err)
		}

		// Update cash bank with explicit field assignment to avoid GORM issues
		cashBank.AccountID = newAccount.ID
		if err := db.Save(&cashBank).Error; err != nil {
			return fmt.Errorf("failed to update cash bank account_id: %v", err)
		}

		log.Printf("✅ Created GL account %s (%s) for cash bank %s", newAccount.Code, newAccount.Name, cashBank.Name)
	}

	return nil
}
