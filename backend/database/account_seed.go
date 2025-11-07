package database

import (
	"errors"
	"fmt"
	"log"
	"strings"
	
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// SeedAccounts creates initial chart of accounts
func SeedAccounts(db *gorm.DB) error {
	log.Println("🌱 Starting account seeding (idempotent mode)...")
	log.Println("   Note: Accounts from migrations will be preserved, not recreated")
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

	// Set parent relationships based on account hierarchy
	accountMap := make(map[string]uint)

	// First pass: create accounts to get IDs (avoid DB errors in logs by pre-checking)
	for _, account := range accounts {
		var existingAccount models.Account
		// Normalize name to UPPERCASE consistently
		normalizedName := strings.ToUpper(account.Name)
		
		findErr := db.Where("code = ? AND deleted_at IS NULL", account.Code).First(&existingAccount).Error
		if findErr == nil {
			// Already exists -> skip without causing an INSERT error
			log.Printf("   ⏭️  Skipping account %s - already exists", account.Code)
			accountMap[account.Code] = existingAccount.ID
			continue
		}
		if !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to lookup account %s: %v", account.Code, findErr)
		}
		
		// Not found -> create
		toCreate := models.Account{
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
		if err := db.Create(&toCreate).Error; err != nil {
			return fmt.Errorf("failed to create account %s: %v", account.Code, err)
		}
		log.Printf("✅ Created account: %s - %s", toCreate.Code, toCreate.Name)
		accountMap[account.Code] = toCreate.ID
	}

	// Define parent-child relationships
	parentChildMap := map[string]string{
		"1100": "1000", // CURRENT ASSETS -> ASSETS
		"1101": "1100", // Kas -> CURRENT ASSETS
		"1102": "1100", // Bank -> CURRENT ASSETS
		"1200": "1100", // ACCOUNTS RECEIVABLE -> CURRENT ASSETS
		"1201": "1200", // Piutang Usaha -> ACCOUNTS RECEIVABLE
		"1114": "1200", // PPh 21 Dibayar Dimuka -> ACCOUNTS RECEIVABLE
		"1115": "1200", // PPh 23 Dibayar Dimuka -> ACCOUNTS RECEIVABLE
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

	// Second pass: set parent relationships
	for childCode, parentCode := range parentChildMap {
		if childID, childExists := accountMap[childCode]; childExists {
			if parentID, parentExists := accountMap[parentCode]; parentExists {
				if err := db.Model(&models.Account{}).Where("id = ?", childID).Update("parent_id", parentID).Error; err != nil {
					return err
				}
			}
		}
	}

	// Count how many accounts exist vs created
	var totalAccounts int64
	db.Model(&models.Account{}).Where("deleted_at IS NULL").Count(&totalAccounts)
	
	log.Printf("✅ Account seeding completed: %d accounts ready", totalAccounts)
	
	// Capitalize existing account names that are not already uppercase
	if err := CapitalizeExistingAccounts(db); err != nil {
		log.Printf("⚠️  Warning: Failed to capitalize existing accounts: %v", err)
	}
	
	return nil
}

// FixAccountHierarchies fixes incorrect account hierarchies in existing databases
func FixAccountHierarchies(db *gorm.DB) error {
	log.Println("🔧 Fixing account hierarchies for existing databases...")
	
	// Define fixes needed for incorrect hierarchies
	hierarchyFixes := []struct {
		Code        string
		ParentCode  string
		Description string
	}{
		{
			Code:        "2103",
			ParentCode:  "2100",
			Description: "Fix PPN Keluaran (LIABILITY) to be under CURRENT LIABILITIES",
		},
		{
			Code:        "1240",
			ParentCode:  "1100",
			Description: "Fix PPN Masukan (ASSET) to be under CURRENT ASSETS",
		},
		{
			Code:        "1114",
			ParentCode:  "1200",
			Description: "Fix PPh 21 Dibayar Dimuka to be under ACCOUNTS RECEIVABLE",
		},
		{
			Code:        "1115",
			ParentCode:  "1200",
			Description: "Fix PPh 23 Dibayar Dimuka to be under ACCOUNTS RECEIVABLE",
		},
		{
			Code:        "2104",
			ParentCode:  "2100",
			Description: "Fix PPh Yang Dipotong to be under CURRENT LIABILITIES",
		},
	}
	
	for _, fix := range hierarchyFixes {
		log.Printf("🔧 Processing fix: %s", fix.Description)
		
		// Find the account to fix
		var account models.Account
		result := db.Where("code = ?", fix.Code).First(&account)
		if result.Error != nil {
			log.Printf("⚠️  Account %s not found, skipping fix", fix.Code)
			continue
		}
		
		// Find the target parent
		var parent models.Account
		result = db.Where("code = ?", fix.ParentCode).First(&parent)
		if result.Error != nil {
			log.Printf("⚠️  Parent account %s not found, skipping fix", fix.ParentCode)
			continue
		}
		
		// Check if fix is needed
		if account.ParentID != nil && *account.ParentID == parent.ID {
			log.Printf("✅ Account %s (%s) already has correct parent %s", 
				account.Code, account.Name, parent.Code)
			continue
		}
		
		// Apply the fix
		oldParentID := account.ParentID
		newLevel := parent.Level + 1
		
		// Update account with correct parent and level
		result = db.Model(&account).Updates(map[string]interface{}{
			"parent_id": parent.ID,
			"level":     newLevel,
		})
		
		if result.Error != nil {
			log.Printf("❌ Failed to fix account %s: %v", fix.Code, result.Error)
			continue
		}
		
		// Ensure parent is marked as header
		if !parent.IsHeader {
			db.Model(&parent).Update("is_header", true)
		}
		
		log.Printf("✅ Fixed: %s (%s) moved from parent %v to %s (level %d)", 
			account.Code, account.Name, oldParentID, parent.Code, newLevel)
	}
	
	log.Println("✅ Account hierarchy fixes completed")
	return nil
}

// CapitalizeExistingAccounts converts all existing account names to UPPERCASE
// This ensures consistency across all accounts, including those created before this update
func CapitalizeExistingAccounts(db *gorm.DB) error {
	log.Println("🔤 Capitalizing existing account names to ensure consistency...")
	
	// Get all non-deleted accounts
	var accounts []models.Account
	if err := db.Where("deleted_at IS NULL").Find(&accounts).Error; err != nil {
		return fmt.Errorf("failed to fetch accounts: %v", err)
	}
	
	updatedCount := 0
	for _, account := range accounts {
		// Check if name needs capitalization
		upperName := strings.ToUpper(account.Name)
		if account.Name != upperName {
			// Update to UPPERCASE
			if err := db.Model(&account).Update("name", upperName).Error; err != nil {
				log.Printf("⚠️  Failed to capitalize account %s (%s): %v", account.Code, account.Name, err)
				continue
			}
			log.Printf("✅ Capitalized: %s - %s → %s", account.Code, account.Name, upperName)
			updatedCount++
		}
	}
	
	if updatedCount > 0 {
		log.Printf("✅ Capitalized %d account names", updatedCount)
	} else {
		log.Println("✅ All account names are already capitalized")
	}
	
	return nil
}
