package main

import (
	"fmt"
	"log"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"app-sistem-akuntansi/models"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🔍 Finding and fixing PPN account conflicts...")

	// Database connection
	dsn := "host=localhost user=postgres password=password dbname=accounting_system port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		log.Println("Please check your database connection settings")
		return
	}

	if err := analyzeAndFixPPNConflicts(db); err != nil {
		log.Fatalf("❌ Failed to fix PPN conflicts: %v", err)
	}

	log.Println("✅ PPN conflict analysis and fixes completed!")
}

func analyzeAndFixPPNConflicts(db *gorm.DB) error {
	log.Println("📋 Analyzing PPN account conflicts...")

	// 1. Check for duplicate PPN KELUARAN accounts (code 2103)
	if err := checkPPNKeluaranAccounts(db); err != nil {
		return fmt.Errorf("failed to check PPN Keluaran accounts: %v", err)
	}

	// 2. Check for duplicate PPN MASUKAN accounts (code 1240)
	if err := checkPPNMasukanAccounts(db); err != nil {
		return fmt.Errorf("failed to check PPN Masukan accounts: %v", err)
	}

	// 3. Clean up any conflicting files
	if err := cleanupConflictingFiles(); err != nil {
		return fmt.Errorf("failed to cleanup conflicting files: %v", err)
	}

	return nil
}

func checkPPNKeluaranAccounts(db *gorm.DB) error {
	log.Println("🔸 Checking PPN KELUARAN accounts (code 2103)...")

	var ppnKeluaranAccounts []models.Account
	err := db.Where("code = ? OR name ILIKE ? OR name ILIKE ?", 
		"2103", "%PPN KELUARAN%", "%OUTPUT VAT%").Find(&ppnKeluaranAccounts).Error
	if err != nil {
		return fmt.Errorf("failed to find PPN Keluaran accounts: %v", err)
	}

	log.Printf("Found %d PPN KELUARAN accounts:", len(ppnKeluaranAccounts))
	for _, account := range ppnKeluaranAccounts {
		log.Printf("  - ID: %d, Code: %s, Name: %s, Balance: %.2f, Active: %v", 
			account.ID, account.Code, account.Name, account.Balance, account.IsActive)
	}

	// If multiple accounts found, consolidate them
	if len(ppnKeluaranAccounts) > 1 {
		log.Println("⚠️  Multiple PPN KELUARAN accounts found - consolidating...")
		return consolidatePPNKeluaranAccounts(db, ppnKeluaranAccounts)
	} else if len(ppnKeluaranAccounts) == 0 {
		log.Println("❌ No PPN KELUARAN account found - creating default one...")
		return createDefaultPPNKeluaranAccount(db)
	} else {
		log.Println("✅ Single PPN KELUARAN account found - OK")
	}

	return nil
}

func checkPPNMasukanAccounts(db *gorm.DB) error {
	log.Println("🔸 Checking PPN MASUKAN accounts (code 1240)...")

	var ppnMasukanAccounts []models.Account
	err := db.Where("code = ? OR name ILIKE ? OR name ILIKE ?", 
		"1240", "%PPN MASUKAN%", "%INPUT VAT%").Find(&ppnMasukanAccounts).Error
	if err != nil {
		return fmt.Errorf("failed to find PPN Masukan accounts: %v", err)
	}

	log.Printf("Found %d PPN MASUKAN accounts:", len(ppnMasukanAccounts))
	for _, account := range ppnMasukanAccounts {
		log.Printf("  - ID: %d, Code: %s, Name: %s, Balance: %.2f, Active: %v", 
			account.ID, account.Code, account.Name, account.Balance, account.IsActive)
	}

	// If multiple accounts found, consolidate them
	if len(ppnMasukanAccounts) > 1 {
		log.Println("⚠️  Multiple PPN MASUKAN accounts found - consolidating...")
		return consolidatePPNMasukanAccounts(db, ppnMasukanAccounts)
	} else if len(ppnMasukanAccounts) == 0 {
		log.Println("❌ No PPN MASUKAN account found - creating default one...")
		return createDefaultPPNMasukanAccount(db)
	} else {
		log.Println("✅ Single PPN MASUKAN account found - OK")
	}

	return nil
}

func consolidatePPNKeluaranAccounts(db *gorm.DB, accounts []models.Account) error {
	log.Println("🔧 Consolidating PPN KELUARAN accounts...")

	// Find the "best" account to keep (prefer active, with standard code, etc.)
	var keepAccount *models.Account
	var totalBalance float64

	for i := range accounts {
		account := &accounts[i]
		totalBalance += account.Balance

		// Keep account with standard code 2103 and is active
		if account.Code == "2103" && account.IsActive {
			keepAccount = account
		} else if keepAccount == nil && account.IsActive {
			keepAccount = account
		} else if keepAccount == nil {
			keepAccount = account
		}
	}

	if keepAccount == nil {
		return fmt.Errorf("no suitable account to keep")
	}

	log.Printf("Keeping account: ID %d, Code: %s, Name: %s", keepAccount.ID, keepAccount.Code, keepAccount.Name)

	// Update kept account with consolidated balance
	keepAccount.Balance = totalBalance
	keepAccount.Code = "2103" // Ensure standard code
	keepAccount.Name = "PPN Keluaran" // Standardize name
	keepAccount.Type = "LIABILITY"
	keepAccount.Category = "CURRENT_LIABILITY"
	keepAccount.IsActive = true

	err := db.Save(keepAccount).Error
	if err != nil {
		return fmt.Errorf("failed to update kept account: %v", err)
	}

	// Update all journal entries to use the kept account
	for _, account := range accounts {
		if account.ID != keepAccount.ID {
			log.Printf("Updating journal entries from account %d to %d", account.ID, keepAccount.ID)
			err = db.Model(&models.JournalLine{}).
				Where("account_id = ?", account.ID).
				Update("account_id", keepAccount.ID).Error
			if err != nil {
				log.Printf("⚠️  Warning: Failed to update journal lines for account %d: %v", account.ID, err)
			}

			// Delete the duplicate account
			err = db.Delete(&account).Error
			if err != nil {
				log.Printf("⚠️  Warning: Failed to delete duplicate account %d: %v", account.ID, err)
			}
		}
	}

	log.Printf("✅ PPN KELUARAN accounts consolidated with balance: %.2f", totalBalance)
	return nil
}

func consolidatePPNMasukanAccounts(db *gorm.DB, accounts []models.Account) error {
	log.Println("🔧 Consolidating PPN MASUKAN accounts...")

	// Find the "best" account to keep (prefer active, with standard code, etc.)
	var keepAccount *models.Account
	var totalBalance float64

	for i := range accounts {
		account := &accounts[i]
		totalBalance += account.Balance

		// Keep account with standard code 1240 and is active
		if account.Code == "1240" && account.IsActive {
			keepAccount = account
		} else if keepAccount == nil && account.IsActive {
			keepAccount = account
		} else if keepAccount == nil {
			keepAccount = account
		}
	}

	if keepAccount == nil {
		return fmt.Errorf("no suitable account to keep")
	}

	log.Printf("Keeping account: ID %d, Code: %s, Name: %s", keepAccount.ID, keepAccount.Code, keepAccount.Name)

	// Update kept account with consolidated balance
	keepAccount.Balance = totalBalance
	keepAccount.Code = "1240" // Ensure standard code
	keepAccount.Name = "PPN Masukan" // Standardize name
	keepAccount.Type = "ASSET"
	keepAccount.Category = "CURRENT_ASSET"
	keepAccount.IsActive = true

	err := db.Save(keepAccount).Error
	if err != nil {
		return fmt.Errorf("failed to update kept account: %v", err)
	}

	// Update all journal entries to use the kept account
	for _, account := range accounts {
		if account.ID != keepAccount.ID {
			log.Printf("Updating journal entries from account %d to %d", account.ID, keepAccount.ID)
			err = db.Model(&models.JournalLine{}).
				Where("account_id = ?", account.ID).
				Update("account_id", keepAccount.ID).Error
			if err != nil {
				log.Printf("⚠️  Warning: Failed to update journal lines for account %d: %v", account.ID, err)
			}

			// Delete the duplicate account
			err = db.Delete(&account).Error
			if err != nil {
				log.Printf("⚠️  Warning: Failed to delete duplicate account %d: %v", account.ID, err)
			}
		}
	}

	log.Printf("✅ PPN MASUKAN accounts consolidated with balance: %.2f", totalBalance)
	return nil
}

func createDefaultPPNKeluaranAccount(db *gorm.DB) error {
	log.Println("🔧 Creating default PPN KELUARAN account...")

	account := &models.Account{
		Code:        "2103",
		Name:        "PPN Keluaran",
		Type:        "LIABILITY",
		Category:    "CURRENT_LIABILITY",
		IsActive:    true,
		Balance:     0,
		Description: "Pajak Pertambahan Nilai Keluaran (Output VAT)",
	}

	err := db.Create(account).Error
	if err != nil {
		return fmt.Errorf("failed to create default PPN Keluaran account: %v", err)
	}

	log.Printf("✅ Created default PPN KELUARAN account: ID %d", account.ID)
	return nil
}

func createDefaultPPNMasukanAccount(db *gorm.DB) error {
	log.Println("🔧 Creating default PPN MASUKAN account...")

	account := &models.Account{
		Code:        "1240",
		Name:        "PPN Masukan",
		Type:        "ASSET",
		Category:    "CURRENT_ASSET",
		IsActive:    true,
		Balance:     0,
		Description: "Pajak Pertambahan Nilai Masukan (Input VAT)",
	}

	err := db.Create(account).Error
	if err != nil {
		return fmt.Errorf("failed to create default PPN Masukan account: %v", err)
	}

	log.Printf("✅ Created default PPN MASUKAN account: ID %d", account.ID)
	return nil
}

func cleanupConflictingFiles() error {
	log.Println("🧹 Cleaning up conflicting PPN files...")

	conflictingFiles := []string{
		"fix_ppn_keluaran.go",
		"fix_ppn_keluaran_clean.go", 
		"ppn_validation_service.go",
		"test_ppn_validation.go",
		"verify_ppn_integration.go",
		"temp_check_ppn.go",
		"activate_ppn_protection.go",
		"check_ppn_mapping.go",
		"fix_ppn_hierarchy.go",
	}

	log.Println("ℹ️  The following files should be reviewed and potentially removed:")
	for _, file := range conflictingFiles {
		log.Printf("  - %s", file)
	}

	log.Println("ℹ️  These files contain legacy PPN handling code that may conflict with the unified services")
	log.Println("ℹ️  Please review and remove them if they're no longer needed")

	return nil
}