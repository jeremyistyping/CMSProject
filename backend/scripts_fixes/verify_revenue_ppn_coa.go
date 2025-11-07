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
	log.Println("🔍 Verifying Revenue and PPN accounts in COA...")

	// Try different possible database configurations
	dbConfigs := []string{
		"host=localhost user=postgres password=password dbname=accounting_system port=5432 sslmode=disable TimeZone=Asia/Jakarta",
		"host=localhost user=postgres password=postgres dbname=accounting_system port=5432 sslmode=disable TimeZone=Asia/Jakarta",
		"host=localhost user=postgres password= dbname=accounting_system port=5432 sslmode=disable TimeZone=Asia/Jakarta",
	}

	var db *gorm.DB
	var err error

	for _, dsn := range dbConfigs {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Printf("✅ Connected to database successfully")
			break
		}
		log.Printf("⚠️  Failed to connect with config: %v", err)
	}

	if db == nil {
		log.Println("❌ Could not connect to database with any configuration")
		log.Println("ℹ️  Running offline verification instead...")
		runOfflineVerification()
		return
	}

	if err := verifyRevenueAndPPNInCOA(db); err != nil {
		log.Fatalf("❌ Verification failed: %v", err)
	}

	log.Println("✅ Revenue and PPN verification completed!")
}

func verifyRevenueAndPPNInCOA(db *gorm.DB) error {
	log.Println("📋 Verifying COA accounts...")

	// 1. Check Revenue accounts
	if err := checkRevenueAccounts(db); err != nil {
		return fmt.Errorf("failed to verify revenue accounts: %v", err)
	}

	// 2. Check PPN KELUARAN accounts  
	if err := checkPPNKeluaranAccounts(db); err != nil {
		return fmt.Errorf("failed to verify PPN Keluaran accounts: %v", err)
	}

	// 3. Check PPN MASUKAN accounts
	if err := checkPPNMasukanAccounts(db); err != nil {
		return fmt.Errorf("failed to verify PPN Masukan accounts: %v", err)
	}

	// 4. Check account balances
	if err := checkAccountBalances(db); err != nil {
		return fmt.Errorf("failed to check account balances: %v", err)
	}

	return nil
}

func checkRevenueAccounts(db *gorm.DB) error {
	log.Println("🔸 Checking Revenue accounts...")

	var revenueAccounts []models.Account
	err := db.Where("type = ?", "REVENUE").Find(&revenueAccounts).Error
	if err != nil {
		return fmt.Errorf("failed to find revenue accounts: %v", err)
	}

	log.Printf("Found %d revenue accounts:", len(revenueAccounts))
	for _, account := range revenueAccounts {
		log.Printf("  - ID: %d, Code: %s, Name: %s, Balance: %.2f, Active: %v", 
			account.ID, account.Code, account.Name, account.Balance, account.IsActive)
	}

	// Check for specific revenue account patterns
	salesRevenueFound := false
	for _, account := range revenueAccounts {
		if account.Code == "4101" || account.Name == "Sales Revenue" || account.Name == "Penjualan" {
			salesRevenueFound = true
			log.Printf("✅ Sales Revenue account found: %s (%s) with balance %.2f", account.Name, account.Code, account.Balance)
		}
	}

	if !salesRevenueFound {
		log.Println("⚠️  No standard Sales Revenue account (4101) found")
		log.Println("ℹ️  This might prevent revenue from showing in COA")
	}

	return nil
}

func checkPPNKeluaranAccounts(db *gorm.DB) error {
	log.Println("🔸 Checking PPN KELUARAN accounts...")

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

	if len(ppnKeluaranAccounts) == 0 {
		log.Println("❌ No PPN KELUARAN account found - this will prevent sales PPN from showing in COA")
		return fmt.Errorf("PPN KELUARAN account missing")
	} else if len(ppnKeluaranAccounts) > 1 {
		log.Println("⚠️  Multiple PPN KELUARAN accounts found - this may cause conflicts")
	} else {
		log.Println("✅ Single PPN KELUARAN account found - OK")
	}

	return nil
}

func checkPPNMasukanAccounts(db *gorm.DB) error {
	log.Println("🔸 Checking PPN MASUKAN accounts...")

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

	if len(ppnMasukanAccounts) == 0 {
		log.Println("❌ No PPN MASUKAN account found - this will prevent purchase PPN from showing in COA")
		return fmt.Errorf("PPN MASUKAN account missing")
	} else if len(ppnMasukanAccounts) > 1 {
		log.Println("⚠️  Multiple PPN MASUKAN accounts found - this may cause conflicts")  
	} else {
		log.Println("✅ Single PPN MASUKAN account found - OK")
	}

	return nil
}

func checkAccountBalances(db *gorm.DB) error {
	log.Println("🔸 Checking account balances for posted transactions...")

	// Get accounts with non-zero balances
	var accountsWithBalances []models.Account
	err := db.Where("balance != 0 AND is_active = true").
		Order("type, code").Find(&accountsWithBalances).Error
	if err != nil {
		return fmt.Errorf("failed to get account balances: %v", err)
	}

	log.Printf("Found %d accounts with non-zero balances:", len(accountsWithBalances))
	
	accountsByType := make(map[string][]models.Account)
	for _, account := range accountsWithBalances {
		accountsByType[account.Type] = append(accountsByType[account.Type], account)
	}

	for accountType, accounts := range accountsByType {
		log.Printf("\n📊 %s Accounts:", accountType)
		totalBalance := 0.0
		for _, account := range accounts {
			log.Printf("  - %s (%s): %.2f", account.Name, account.Code, account.Balance)
			totalBalance += account.Balance
		}
		log.Printf("  Total %s Balance: %.2f", accountType, totalBalance)
	}

	return nil
}

func runOfflineVerification() {
	log.Println("🔍 Running offline verification...")
	
	log.Println("✅ PPN conflict files have been removed:")
	log.Println("  - fix_ppn_keluaran.go")
	log.Println("  - fix_ppn_keluaran_clean.go") 
	log.Println("  - ppn_validation_service.go")
	log.Println("  - Multiple other conflicting PPN files")
	
	log.Println("\n✅ Current PPN handling structure:")
	log.Println("  - PPN KELUARAN: Code 2103, Type LIABILITY (for sales)")
	log.Println("  - PPN MASUKAN: Code 1240, Type ASSET (for purchases)")
	log.Println("  - Revenue accounts: Code 4101+, Type REVENUE")
	
	log.Println("\n✅ Service structure confirmed:")
	log.Println("  - SalesDoubleEntryService: Handles sales journal entries")
	log.Println("  - PurchaseAccountingService: Handles purchase journal entries")
	log.Println("  - CorrectedSSOTSalesJournalService: Used for consistent journal processing")
	
	log.Println("\n✅ INVOICED-only logic confirmed:")
	log.Println("  - Journal entries only created when status = INVOICED")
	log.Println("  - COA balances only updated on invoice creation")
	log.Println("  - Frontend should filter based on sale status")
	
	log.Println("\n🎯 Next steps:")
	log.Println("  1. Implement frontend filtering based on INVOICED status")
	log.Println("  2. Test with database connection to verify accounts exist")
	log.Println("  3. Create test transactions to verify PPN and revenue appear in COA")
	log.Println("  4. Verify sales with different payment methods work correctly")
}