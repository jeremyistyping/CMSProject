package main

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Database configuration
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "root:@tcp(localhost:3306)/sistem_akuntansi?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// Connect to database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("🔧 Starting Cash Bank Constraint Fix...")

	// Step 1: Diagnose problematic records
	var problematicRecords []struct {
		ID        uint
		Code      string
		Name      string
		Type      string
		AccountID *uint
		Status    string
	}

	query := `
		SELECT 
			cb.id,
			cb.code,
			cb.name,
			cb.type,
			cb.account_id,
			CASE 
				WHEN cb.account_id IS NULL THEN 'NULL account_id'
				WHEN a.id IS NULL THEN 'Invalid account_id reference'
				ELSE 'Valid reference'
			END as status
		FROM cash_banks cb 
		LEFT JOIN accounts a ON cb.account_id = a.id AND a.deleted_at IS NULL
		WHERE cb.account_id IS NULL OR a.id IS NULL
		ORDER BY cb.id
	`

	if err := db.Raw(query).Scan(&problematicRecords).Error; err != nil {
		log.Fatalf("Failed to diagnose problematic records: %v", err)
	}

	log.Printf("📊 Found %d problematic cash_banks records", len(problematicRecords))

	if len(problematicRecords) > 0 {
		log.Println("📋 Problematic records:")
		for _, record := range problematicRecords {
			accountID := "NULL"
			if record.AccountID != nil {
				accountID = fmt.Sprintf("%d", *record.AccountID)
			}
			log.Printf("  - ID: %d, Code: %s, Name: %s, Type: %s, AccountID: %s, Status: %s",
				record.ID, record.Code, record.Name, record.Type, accountID, record.Status)
		}
	}

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

	result := db.Where("code = ? AND deleted_at IS NULL", "1100").FirstOrCreate(currentAssetsAccount)
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

	result = db.Where("code = ? AND deleted_at IS NULL", "1101").FirstOrCreate(cashEquivalentsAccount)
	if result.RowsAffected > 0 {
		log.Printf("✅ Created cash equivalents account: %s (%s)", cashEquivalentsAccount.Name, cashEquivalentsAccount.Code)
	}

	// Step 3: Fix each problematic record
	for _, record := range problematicRecords {
		log.Printf("🔧 Fixing cash_bank record ID: %d (%s)", record.ID, record.Name)

		// Generate unique account code
		var accountCode string
		if record.Type == "CASH" {
			accountCode = fmt.Sprintf("1101-%03d", record.ID)
		} else {
			accountCode = fmt.Sprintf("1102-%03d", record.ID)
		}

		// Create GL account
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

		// Check if account with this name already exists
		var existingAccount models.Account
		err := db.Where("name = ? AND type = ? AND category = ? AND deleted_at IS NULL", 
			record.Name, "ASSET", "CURRENT_ASSET").First(&existingAccount).Error

		var accountIDToUse uint
		if err == nil {
			// Use existing account
			accountIDToUse = existingAccount.ID
			log.Printf("  ✅ Using existing GL account: %s (%s)", existingAccount.Name, existingAccount.Code)
		} else {
			// Create new account
			if err := db.Create(newGLAccount).Error; err != nil {
				log.Printf("  ❌ Failed to create GL account for %s: %v", record.Name, err)
				continue
			}
			accountIDToUse = newGLAccount.ID
			log.Printf("  ✅ Created new GL account: %s (%s)", newGLAccount.Name, newGLAccount.Code)
		}

		// Update cash_bank record
		if err := db.Model(&models.CashBank{}).Where("id = ?", record.ID).Update("account_id", accountIDToUse).Error; err != nil {
			log.Printf("  ❌ Failed to update cash_bank record ID %d: %v", record.ID, err)
		} else {
			log.Printf("  ✅ Updated cash_bank record ID %d with account_id %d", record.ID, accountIDToUse)
		}
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

	result = db.Where("code = ? AND deleted_at IS NULL", "1199").FirstOrCreate(fallbackAccount)
	if result.RowsAffected > 0 {
		log.Printf("✅ Created fallback account: %s (%s)", fallbackAccount.Name, fallbackAccount.Code)
	}

	// Update any remaining cash_banks with NULL account_id
	updateResult := db.Model(&models.CashBank{}).
		Where("account_id IS NULL").
		Update("account_id", fallbackAccount.ID)
	
	if updateResult.RowsAffected > 0 {
		log.Printf("✅ Updated %d cash_bank records to use fallback account", updateResult.RowsAffected)
	}

	// Step 5: Verify the fix
	log.Println("🔍 Verifying the fix...")

	var verificationRecords []struct {
		ID           uint
		Code         string
		Name         string
		Type         string
		AccountID    uint
		AccountCode  string
		AccountName  string
		Status       string
	}

	verifyQuery := `
		SELECT 
			cb.id,
			cb.code,
			cb.name,
			cb.type,
			cb.account_id,
			COALESCE(a.code, 'N/A') as account_code,
			COALESCE(a.name, 'N/A') as account_name,
			CASE 
				WHEN cb.account_id IS NULL THEN 'ERROR: Still NULL'
				WHEN a.id IS NULL THEN 'ERROR: Invalid reference'
				ELSE 'OK: Valid reference'
			END as status
		FROM cash_banks cb 
		LEFT JOIN accounts a ON cb.account_id = a.id AND a.deleted_at IS NULL
		ORDER BY cb.id
	`

	if err := db.Raw(verifyQuery).Scan(&verificationRecords).Error; err != nil {
		log.Fatalf("Failed to verify fix: %v", err)
	}

	// Count statuses
	statusCounts := make(map[string]int)
	for _, record := range verificationRecords {
		statusCounts[record.Status]++
		if record.Status != "OK: Valid reference" {
			log.Printf("  ⚠️  ID: %d, Status: %s", record.ID, record.Status)
		}
	}

	log.Println("📊 Verification Results:")
	for status, count := range statusCounts {
		log.Printf("  %s: %d records", status, count)
	}

	// Step 6: Check constraint status
	log.Println("🔍 Checking foreign key constraints...")
	
	var constraints []struct {
		ConstraintName        string
		TableName            string
		ColumnName           string
		ReferencedTableName  string
		ReferencedColumnName string
	}

	constraintQuery := `
		SELECT 
			CONSTRAINT_NAME,
			TABLE_NAME,
			COLUMN_NAME,
			REFERENCED_TABLE_NAME,
			REFERENCED_COLUMN_NAME
		FROM information_schema.KEY_COLUMN_USAGE 
		WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'cash_banks' 
			AND REFERENCED_TABLE_NAME IS NOT NULL
	`

	if err := db.Raw(constraintQuery).Scan(&constraints).Error; err != nil {
		log.Printf("⚠️  Could not check constraints: %v", err)
	} else {
		log.Printf("✅ Found %d foreign key constraints for cash_banks table", len(constraints))
		for _, constraint := range constraints {
			log.Printf("  - %s: %s.%s -> %s.%s", 
				constraint.ConstraintName,
				constraint.TableName, constraint.ColumnName,
				constraint.ReferencedTableName, constraint.ReferencedColumnName)
		}
	}

	// Final summary
	totalCashBanks := len(verificationRecords)
	validReferences := statusCounts["OK: Valid reference"]
	
	log.Println("🎉 Fix Complete!")
	log.Printf("📊 Final Summary:")
	log.Printf("  - Total cash_banks records: %d", totalCashBanks)
	log.Printf("  - Records with valid account_id: %d", validReferences)
	log.Printf("  - Success rate: %.1f%%", float64(validReferences)/float64(totalCashBanks)*100)

	if validReferences == totalCashBanks {
		log.Println("✅ All cash_banks records now have valid account_id references!")
		log.Println("🚀 The foreign key constraint issue should be resolved.")
	} else {
		log.Printf("⚠️  %d records still have issues. Manual intervention may be required.", totalCashBanks-validReferences)
	}
}
