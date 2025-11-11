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
		
		// COA untuk Cost Control Padel Bandung - Project Construction Management
		accounts := []models.Account{
			// ========================================
			// 1000 - ASET LANCAR (CURRENT ASSETS)
			// ========================================
			{Code: "1000", Name: strings.ToUpper("ASET LANCAR"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 1, IsHeader: true, IsActive: true},
			
			// Kas & Bank
			{Code: "1101", Name: strings.ToUpper("KAS PROYEK"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 2, IsHeader: false, IsActive: true, Balance: 0, Description: "Uang tunai site"},
			{Code: "1102", Name: strings.ToUpper("BANK"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 2, IsHeader: false, IsActive: true, Balance: 0, Description: "Transfer project"},
			{Code: "1201", Name: strings.ToUpper("DEPOSIT"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 2, IsHeader: false, IsActive: true, Balance: 0, Description: "DP supplier / sewa"},
			
			// Tax Prepaid Accounts
			{Code: "1114", Name: strings.ToUpper("PPh 21 DIBAYAR DIMUKA"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "1115", Name: strings.ToUpper("PPh 23 DIBAYAR DIMUKA"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "1240", Name: strings.ToUpper("PPN MASUKAN"), Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 2, IsHeader: false, IsActive: true, Balance: 0},

			// ========================================
			// 2000 - KEWAJIBAN (LIABILITIES)
			// ========================================
			{Code: "2000", Name: strings.ToUpper("KEWAJIBAN"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 1, IsHeader: true, IsActive: true},
			{Code: "2101", Name: strings.ToUpper("UTANG USAHA"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "2103", Name: strings.ToUpper("PPN KELUARAN"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "2104", Name: strings.ToUpper("PPh YANG DIPOTONG"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "2111", Name: strings.ToUpper("UTANG PPh 21"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "2112", Name: strings.ToUpper("UTANG PPh 23"), Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 2, IsHeader: false, IsActive: true, Balance: 0},

			// ========================================
			// 3000 - EKUITAS (EQUITY)
			// ========================================
			{Code: "3000", Name: strings.ToUpper("EKUITAS"), Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 1, IsHeader: true, IsActive: true},
			{Code: "3101", Name: strings.ToUpper("MODAL PEMILIK"), Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
			{Code: "7000", Name: strings.ToUpper("LABA / RUGI PROYEK"), Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 2, IsHeader: false, IsActive: true, Balance: 0, Description: "Selisih income - cost"},

			// ========================================
			// 4000 - PENDAPATAN PROYEK (INCOME)
			// ========================================
			{Code: "4000", Name: strings.ToUpper("PENDAPATAN PROYEK (INCOME)"), Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 1, IsHeader: true, IsActive: true, Description: "Total kontrak, termin"},
			{Code: "4101", Name: strings.ToUpper("PENDAPATAN TERMIN 1"), Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 2, IsHeader: false, IsActive: true, Balance: 0, Description: "Termin proyek"},
			{Code: "4102", Name: strings.ToUpper("PENDAPATAN TERMIN 2"), Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 2, IsHeader: false, IsActive: true, Balance: 0, Description: "Pembayaran bertahap"},
			{Code: "4201", Name: strings.ToUpper("RETENSI"), Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 2, IsHeader: false, IsActive: true, Balance: 0, Description: "Potongan retensi"},

			// ========================================
			// 5000 - BEBAN LANGSUNG PROYEK (DIRECT COST)
			// ========================================
			{Code: "5000", Name: strings.ToUpper("BEBAN LANGSUNG PROYEK (DIRECT COST)"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 1, IsHeader: true, IsActive: true, Description: "Biaya yang langsung terkait proyek"},
			
			// 5100 - Material Bangunan
			{Code: "5100", Name: strings.ToUpper("MATERIAL BANGUNAN"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: true, IsActive: true, Description: "Material, semen, baja, pipa, pasir"},
			{Code: "5101", Name: strings.ToUpper("SEMEN & PASIR"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Material utama"},
			{Code: "5102", Name: strings.ToUpper("BESI & BAJA"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Struktur"},
			{Code: "5103", Name: strings.ToUpper("PLUMBING & FITTING"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Instalasi"},
			{Code: "5104", Name: strings.ToUpper("KACA, ALUMINIUM"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Finishing"},
			{Code: "5105", Name: strings.ToUpper("CAT & FINISHING"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Pengecatan"},
			
			// 5200 - Sewa Alat Berat / Equipment Hire
			{Code: "5200", Name: strings.ToUpper("SEWA ALAT BERAT / EQUIPMENT HIRE"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: true, IsActive: true, Description: "Excavator, Crane, Mixer"},
			{Code: "5201", Name: strings.ToUpper("SEWA ALAT BERAT"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Harian/Mingguan"},
			{Code: "5202", Name: strings.ToUpper("TRANSPORT & MOBILISASI"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Mobilisasi alat"},
			
			// 5300 - Tenaga Kerja (Labour)
			{Code: "5300", Name: strings.ToUpper("TENAGA KERJA (LABOUR)"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: true, IsActive: true, Description: "Upah mandor & pekerja"},
			{Code: "5301", Name: strings.ToUpper("MANDOR"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Gaji mandor"},
			{Code: "5302", Name: strings.ToUpper("TUKANG & HELPER"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Upah pekerja lapangan"},
			{Code: "5303", Name: strings.ToUpper("OVERTIME & BONUS"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Tambahan jam kerja"},
			
			// 5400 - Biaya Operasional Site
			{Code: "5400", Name: strings.ToUpper("BIAYA OPERASIONAL SITE"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: true, IsActive: true, Description: "Air kerja, listrik kerja, tol, bensin, kosan"},
			{Code: "5401", Name: strings.ToUpper("AIR & LISTRIK KERJA"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Utilitas proyek"},
			{Code: "5402", Name: strings.ToUpper("TRANSPORTASI & TOL"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Perjalanan tim"},
			{Code: "5403", Name: strings.ToUpper("KONSUMSI & ENTERTAIN"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Konsumsi lapangan"},
			{Code: "5404", Name: strings.ToUpper("AKOMODASI (KOSAN, HOTEL)"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Tempat tinggal tim"},
			{Code: "5405", Name: strings.ToUpper("ATK & ALAT KECIL"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Meteran, spidol, helm, dll"},
			{Code: "5406", Name: strings.ToUpper("KOMPENSASI & KEAMANAN"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0, Description: "Gaji security, kompensasi"},
			
			// ========================================
			// 6000 - OVERHEAD KANTOR & ADMIN
			// ========================================
			{Code: "6000", Name: strings.ToUpper("OVERHEAD KANTOR & ADMIN"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 1, IsHeader: true, IsActive: true, Description: "Admin, pajak, fee"},
			{Code: "6101", Name: strings.ToUpper("ADMIN FEE"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0, Description: "Biaya transfer, admin bank"},
			{Code: "6102", Name: strings.ToUpper("PAJAK PROYEK (PPH, PPN)"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0, Description: "Biaya pajak"},
			{Code: "6103", Name: strings.ToUpper("FEE MARKETING"), Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0, Description: "Fee rekanan / marketing"},
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
		// ASSETS hierarchy
		"1101": "1000", // Kas Proyek -> ASET LANCAR
		"1102": "1000", // Bank -> ASET LANCAR
		"1201": "1000", // Deposit -> ASET LANCAR
		"1114": "1000", // PPh 21 Dibayar Dimuka -> ASET LANCAR
		"1115": "1000", // PPh 23 Dibayar Dimuka -> ASET LANCAR
		"1240": "1000", // PPN Masukan -> ASET LANCAR
		
		// LIABILITIES hierarchy
		"2101": "2000", // Utang Usaha -> KEWAJIBAN
		"2103": "2000", // PPN Keluaran -> KEWAJIBAN
		"2104": "2000", // PPh Yang Dipotong -> KEWAJIBAN
		"2111": "2000", // Utang PPh 21 -> KEWAJIBAN
		"2112": "2000", // Utang PPh 23 -> KEWAJIBAN
		
		// EQUITY hierarchy
		"3101": "3000", // Modal Pemilik -> EKUITAS
		"7000": "3000", // Laba / Rugi Proyek -> EKUITAS
		
		// REVENUE hierarchy
		"4101": "4000", // Pendapatan Termin 1 -> PENDAPATAN PROYEK
		"4102": "4000", // Pendapatan Termin 2 -> PENDAPATAN PROYEK
		"4201": "4000", // Retensi -> PENDAPATAN PROYEK
		
		// EXPENSES - BEBAN LANGSUNG PROYEK hierarchy
		// Material Bangunan
		"5100": "5000", // MATERIAL BANGUNAN -> BEBAN LANGSUNG PROYEK
		"5101": "5100", // Semen & Pasir -> MATERIAL BANGUNAN
		"5102": "5100", // Besi & Baja -> MATERIAL BANGUNAN
		"5103": "5100", // Plumbing & Fitting -> MATERIAL BANGUNAN
		"5104": "5100", // Kaca, Aluminium -> MATERIAL BANGUNAN
		"5105": "5100", // Cat & Finishing -> MATERIAL BANGUNAN
		
		// Sewa Alat Berat
		"5200": "5000", // SEWA ALAT BERAT -> BEBAN LANGSUNG PROYEK
		"5201": "5200", // Sewa Alat Berat -> SEWA ALAT BERAT / EQUIPMENT HIRE
		"5202": "5200", // Transport & Mobilisasi -> SEWA ALAT BERAT / EQUIPMENT HIRE
		
		// Tenaga Kerja
		"5300": "5000", // TENAGA KERJA -> BEBAN LANGSUNG PROYEK
		"5301": "5300", // Mandor -> TENAGA KERJA
		"5302": "5300", // Tukang & Helper -> TENAGA KERJA
		"5303": "5300", // Overtime & Bonus -> TENAGA KERJA
		
		// Biaya Operasional Site
		"5400": "5000", // BIAYA OPERASIONAL SITE -> BEBAN LANGSUNG PROYEK
		"5401": "5400", // Air & Listrik Kerja -> BIAYA OPERASIONAL SITE
		"5402": "5400", // Transportasi & Tol -> BIAYA OPERASIONAL SITE
		"5403": "5400", // Konsumsi & Entertain -> BIAYA OPERASIONAL SITE
		"5404": "5400", // Akomodasi -> BIAYA OPERASIONAL SITE
		"5405": "5400", // ATK & Alat Kecil -> BIAYA OPERASIONAL SITE
		"5406": "5400", // Kompensasi & Keamanan -> BIAYA OPERASIONAL SITE
		
		// OVERHEAD KANTOR & ADMIN
		"6101": "6000", // Admin Fee -> OVERHEAD KANTOR & ADMIN
		"6102": "6000", // Pajak Proyek -> OVERHEAD KANTOR & ADMIN
		"6103": "6000", // Fee Marketing -> OVERHEAD KANTOR & ADMIN
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
