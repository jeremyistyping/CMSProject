package database

import (
	"fmt"
	"log"
	"strings"
	
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SeedAccountsImproved creates initial chart of accounts with improved duplicate handling
// This version includes:
// - Explicit soft-delete filtering
// - Better error messages
// - Duplicate detection and reporting
// - Transaction support for atomic operations
func SeedAccountsImproved(db *gorm.DB) error {
	log.Println("🔒 PRODUCTION MODE: Seeding accounts with improved duplicate handling...")
	
	// Start a transaction for atomic operations
	return db.Transaction(func(tx *gorm.DB) error {
		// First, check for existing duplicates
		if err := checkExistingDuplicates(tx); err != nil {
			return fmt.Errorf("pre-seed duplicate check failed: %v", err)
		}
		
		// Normalize all account names to UPPERCASE to ensure consistency
		accounts := []models.Account{
			// ASSETS (1xxx)
			{Code: "1000", Name: strings.ToUpper("ASSETS"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 1, IsHeader: true, IsActive: true},
			{Code: "1100", Name: strings.ToUpper("CURRENT ASSETS"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 2, IsHeader: true, IsActive: true},
			{Code: "1101", Name: strings.ToUpper("KAS"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: true, IsActive: true, Balance: 0},
			{Code: "1102", Name: strings.ToUpper("BANK"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: true, IsActive: true, Balance: 0},
			{Code: "1200", Name: strings.ToUpper("ACCOUNTS RECEIVABLE"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 2, IsHeader: true, IsActive: true},
			{Code: "1201", Name: strings.ToUpper("PIUTANG USAHA"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			
			// Tax Prepaid Accounts (Prepaid taxes/Input VAT)
			{Code: "1114", Name: strings.ToUpper("PPh 21 DIBAYAR DIMUKA"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "1115", Name: strings.ToUpper("PPh 23 DIBAYAR DIMUKA"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "1116", Name: strings.ToUpper("POTONGAN PAJAK LAINNYA DIBAYAR DIMUKA"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "1240", Name: strings.ToUpper("PPN MASUKAN"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			
			// Inventory
			{Code: "1301", Name: strings.ToUpper("PERSEDIAAN BARANG DAGANGAN"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},

			{Code: "1500", Name: strings.ToUpper("FIXED ASSETS"), Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 2, IsHeader: true, IsActive: true},
			{Code: "1501", Name: strings.ToUpper("PERALATAN KANTOR"), Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "1502", Name: strings.ToUpper("KENDARAAN"), Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "1503", Name: strings.ToUpper("BANGUNAN"), Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "1509", Name: strings.ToUpper("TRUK"), Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},

			// LIABILITIES (2xxx)
			{Code: "2000", Name: strings.ToUpper("LIABILITIES"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 1, IsHeader: true, IsActive: true},
			{Code: "2100", Name: strings.ToUpper("CURRENT LIABILITIES"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 2, IsHeader: true, IsActive: true},
			{Code: "2101", Name: strings.ToUpper("UTANG USAHA"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "2103", Name: strings.ToUpper("PPN KELUARAN"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "2104", Name: strings.ToUpper("PPh YANG DIPOTONG"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "2107", Name: strings.ToUpper("PEMOTONGAN PAJAK LAINNYA"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "2108", Name: strings.ToUpper("PENAMBAHAN PAJAK LAINNYA"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			
			// Specific withholding tax accounts for purchases
			{Code: "2111", Name: strings.ToUpper("UTANG PPh 21"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "2112", Name: strings.ToUpper("UTANG PPh 23"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},

			// EQUITY (3xxx)
			{Code: "3000", Name: strings.ToUpper("EQUITY"), Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 1, IsHeader: true, IsActive: true},
			{Code: "3101", Name: strings.ToUpper("MODAL PEMILIK"), Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "3201", Name: strings.ToUpper("LABA DITAHAN"), Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 2, IsHeader: false, IsActive: true, Balance: 0},

			// REVENUE (4xxx)
			{Code: "4000", Name: strings.ToUpper("REVENUE"), Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 1, IsHeader: true, IsActive: true},
			{Code: "4101", Name: strings.ToUpper("PENDAPATAN PENJUALAN"), Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "4102", Name: strings.ToUpper("PENDAPATAN JASA/ONGKIR"), Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "4201", Name: strings.ToUpper("PENDAPATAN LAIN-LAIN"), Type: models.AccountTypeRevenue, Category: models.CategoryOtherIncome, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "4900", Name: strings.ToUpper("OTHER INCOME"), Type: models.AccountTypeRevenue, Category: models.CategoryOtherIncome, Level: 2, IsHeader: false, IsActive: true, Balance: 0},

			// EXPENSES (5xxx)
			{Code: "5000", Name: strings.ToUpper("EXPENSES"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 1, IsHeader: true, IsActive: true},
			{Code: "5101", Name: strings.ToUpper("HARGA POKOK PENJUALAN"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "5201", Name: strings.ToUpper("BEBAN GAJI"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "5202", Name: strings.ToUpper("BEBAN LISTRIK"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "5203", Name: strings.ToUpper("BEBAN TELEPON"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "5204", Name: strings.ToUpper("BEBAN TRANSPORTASI"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "5900", Name: strings.ToUpper("GENERAL EXPENSE"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		}

		// Verify no duplicates in seed data itself
		if err := verifyNoDuplicatesInSeed(accounts); err != nil {
			return err
		}

		accountMap := make(map[string]uint)

		// Create accounts with improved duplicate handling
		for _, account := range accounts {
			accountID, created, err := upsertAccount(tx, account)
			if err != nil {
				return fmt.Errorf("failed to upsert account %s: %v", account.Code, err)
			}
			
			if created {
				log.Printf("✅ Created new account: %s - %s", account.Code, account.Name)
			} else {
				log.Printf("🔒 Account exists: %s - %s (preserving balance)", account.Code, account.Name)
			}
			
			accountMap[account.Code] = accountID
		}

		// Set parent relationships
		if err := setParentRelationships(tx, accountMap); err != nil {
			return fmt.Errorf("failed to set parent relationships: %v", err)
		}

		log.Println("✅ Account seeding completed successfully")
		return nil
	})
}

// checkExistingDuplicates checks for duplicate accounts before seeding
func checkExistingDuplicates(tx *gorm.DB) error {
	var duplicates []struct {
		Code  string
		Count int64
	}
	
	err := tx.Model(&models.Account{}).
		Select("code, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("code").
		Having("COUNT(*) > 1").
		Scan(&duplicates).Error
	
	if err != nil {
		return err
	}
	
	if len(duplicates) > 0 {
		log.Println("⚠️  WARNING: Found duplicate accounts in database:")
		for _, dup := range duplicates {
			log.Printf("   - Code %s has %d instances", dup.Code, dup.Count)
		}
		return fmt.Errorf("database has %d duplicate account codes - please clean up first", len(duplicates))
	}
	
	return nil
}

// verifyNoDuplicatesInSeed checks seed data for duplicates
func verifyNoDuplicatesInSeed(accounts []models.Account) error {
	seen := make(map[string]bool)
	duplicates := []string{}
	
	for _, account := range accounts {
		if seen[account.Code] {
			duplicates = append(duplicates, account.Code)
		}
		seen[account.Code] = true
	}
	
	if len(duplicates) > 0 {
		return fmt.Errorf("seed data contains duplicate codes: %v", duplicates)
	}
	
	return nil
}

// upsertAccount creates or updates an account atomically
func upsertAccount(tx *gorm.DB, account models.Account) (uint, bool, error) {
	var existingAccount models.Account
	
	// Always normalize name to UPPERCASE
	normalizedName := strings.ToUpper(account.Name)
	
	// Use FOR UPDATE lock to prevent race conditions
	// Suppress logging for expected "record not found" by using Session
	err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(4)}). // 4 = Silent mode
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("code = ?", account.Code).
		Where("deleted_at IS NULL").
		First(&existingAccount).Error
	
	if err == gorm.ErrRecordNotFound {
		// Account doesn't exist, create it
		newAccount := models.Account{
			Code:        account.Code,
			Name:        normalizedName,
			Type:        account.Type,
			Category:    account.Category,
			Level:       account.Level,
			IsHeader:    account.IsHeader,
			IsActive:    account.IsActive,
			Balance:     account.Balance,
			Description: account.Description,
		}
		
		if err := tx.Create(&newAccount).Error; err != nil {
			return 0, false, err
		}
		
		return newAccount.ID, true, nil
	} else if err != nil {
		return 0, false, err
	}
	
	// Account exists - check if it's system critical before updating protected fields
	updates := map[string]interface{}{
		"name":        normalizedName,
		"level":       account.Level,
		"is_header":   account.IsHeader,
		"description": account.Description,
	}
	
	// Only update type, category, and is_active if account is NOT system critical
	// to avoid triggering the database trigger that blocks critical account changes
	if !existingAccount.IsSystemCritical {
		updates["type"] = account.Type
		updates["category"] = account.Category
		updates["is_active"] = account.IsActive
	}
	
	if err := tx.Model(&existingAccount).Updates(updates).Error; err != nil {
		return 0, false, err
	}
	
	return existingAccount.ID, false, nil
}

// setParentRelationships sets parent-child relationships for accounts
func setParentRelationships(tx *gorm.DB, accountMap map[string]uint) error {
	parentChildMap := map[string]string{
		"1100": "1000", // CURRENT ASSETS -> ASSETS
		"1101": "1100", // Kas -> CURRENT ASSETS
		"1102": "1100", // Bank -> CURRENT ASSETS
		"1200": "1100", // ACCOUNTS RECEIVABLE -> CURRENT ASSETS
		"1201": "1200", // Piutang Usaha -> ACCOUNTS RECEIVABLE
		"1114": "1200", // PPh 21 Dibayar Dimuka -> ACCOUNTS RECEIVABLE
		"1115": "1200", // PPh 23 Dibayar Dimuka -> ACCOUNTS RECEIVABLE
		"1116": "1200", // Potongan Pajak Lainnya Dibayar Dimuka -> ACCOUNTS RECEIVABLE
		"1240": "1100", // PPN Masukan -> CURRENT ASSETS
		"1301": "1100", // Persediaan Barang Dagangan -> CURRENT ASSETS
		"1500": "1000", // FIXED ASSETS -> ASSETS
		"1501": "1500", // Peralatan Kantor -> FIXED ASSETS
		"1502": "1500", // Kendaraan -> FIXED ASSETS
		"1503": "1500", // Bangunan -> FIXED ASSETS
		"1509": "1500", // TRUK -> FIXED ASSETS
		"2100": "2000", // CURRENT LIABILITIES -> LIABILITIES
		"2101": "2100", // Utang Usaha -> CURRENT LIABILITIES
		"2103": "2100", // PPN Keluaran -> CURRENT LIABILITIES
		"2104": "2100", // PPh Yang Dipotong -> CURRENT LIABILITIES
		"2107": "2100", // Pemotongan Pajak Lainnya -> CURRENT LIABILITIES
		"2108": "2100", // Penambahan Pajak Lainnya -> CURRENT LIABILITIES
		"2111": "2100", // Utang PPh 21 -> CURRENT LIABILITIES
		"2112": "2100", // Utang PPh 23 -> CURRENT LIABILITIES
		"3101": "3000", // Modal Pemilik -> EQUITY
		"3201": "3000", // Laba Ditahan -> EQUITY
		"4101": "4000", // Pendapatan Penjualan -> REVENUE
		"4102": "4000", // Pendapatan Jasa/Ongkir -> REVENUE
		"4201": "4000", // Pendapatan Lain-lain -> REVENUE
		"4900": "4000", // Other Income -> REVENUE
		"5101": "5000", // Harga Pokok Penjualan -> EXPENSES
		"5201": "5000", // Beban Gaji -> EXPENSES
		"5202": "5000", // Beban Listrik -> EXPENSES
		"5203": "5000", // Beban Telepon -> EXPENSES
		"5204": "5000", // Beban Transportasi -> EXPENSES
		"5900": "5000", // General Expense -> EXPENSES
	}

	for childCode, parentCode := range parentChildMap {
		childID, childExists := accountMap[childCode]
		parentID, parentExists := accountMap[parentCode]
		
		if !childExists {
			log.Printf("⚠️  Child account %s not found in map, skipping relationship", childCode)
			continue
		}
		
		if !parentExists {
			log.Printf("⚠️  Parent account %s not found in map, skipping relationship", parentCode)
			continue
		}
		
		err := tx.Model(&models.Account{}).
			Where("id = ?", childID).
			Update("parent_id", parentID).Error
		
		if err != nil {
			return fmt.Errorf("failed to set parent for %s -> %s: %v", childCode, parentCode, err)
		}
	}
	
	return nil
}

// CleanDuplicateAccounts removes duplicate accounts, keeping the oldest one
// WARNING: This should only be run after backing up the database!
func CleanDuplicateAccounts(db *gorm.DB, dryRun bool) error {
	log.Println("🧹 Starting duplicate account cleanup...")
	
	if dryRun {
		log.Println("📋 DRY RUN MODE - No changes will be made")
	}
	
	// Find duplicates
	var duplicates []struct {
		Code string
		IDs  string
	}
	
	err := db.Raw(`
		SELECT 
			code,
			STRING_AGG(id::text, ',') as ids
		FROM accounts
		WHERE deleted_at IS NULL
		GROUP BY code
		HAVING COUNT(*) > 1
	`).Scan(&duplicates).Error
	
	if err != nil {
		return fmt.Errorf("failed to find duplicates: %v", err)
	}
	
	if len(duplicates) == 0 {
		log.Println("✅ No duplicates found!")
		return nil
	}
	
	log.Printf("⚠️  Found %d duplicate account codes", len(duplicates))
	
	for _, dup := range duplicates {
		ids := strings.Split(dup.IDs, ",")
		if len(ids) <= 1 {
			continue
		}
		
		// Keep the first (oldest) ID, delete the rest
		keepID := ids[0]
		deleteIDs := ids[1:]
		
		log.Printf("   Code %s: Keeping ID %s, deleting %v", dup.Code, keepID, deleteIDs)
		
		if !dryRun {
			// Soft delete duplicates
			err := db.Model(&models.Account{}).
				Where("code = ?", dup.Code).
				Where("id IN ?", deleteIDs).
				Update("deleted_at", gorm.Expr("NOW()")).Error
			
			if err != nil {
				return fmt.Errorf("failed to delete duplicates for code %s: %v", dup.Code, err)
			}
		}
	}
	
	if dryRun {
		log.Println("📋 DRY RUN COMPLETE - Run with dryRun=false to actually clean")
	} else {
		log.Println("✅ Duplicate cleanup completed!")
	}
	
	return nil
}
