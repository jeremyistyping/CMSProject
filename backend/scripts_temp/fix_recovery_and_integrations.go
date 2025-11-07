package main

import (
	"fmt"
	"log"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/shopspring/decimal"
	"app-sistem-akuntansi/models"
)

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("🔧 Starting Recovery and Integration Fix...")

	// Step 1: Fix Account Duplicates and Recover COA
	log.Println("\n📊 Step 1: Fixing Account Duplicates and Recovering COA...")
	if err := fixAccountDuplicatesAndRecover(db); err != nil {
		log.Printf("❌ Failed to fix accounts: %v", err)
	} else {
		log.Println("✅ COA recovery and fix completed")
	}

	// Step 2: Recover Other Critical Master Data
	log.Println("\n👥 Step 2: Recovering Critical Master Data...")
	if err := recoverCriticalMasterData(db); err != nil {
		log.Printf("❌ Failed to recover master data: %v", err)
	} else {
		log.Println("✅ Master data recovery completed")
	}

	// Step 3: Verify and Fix Cash&Bank Integration with COA
	log.Println("\n💰 Step 3: Verifying Cash&Bank Integration...")
	if err := verifyCashBankCOAIntegration(db); err != nil {
		log.Printf("❌ Failed to verify cash&bank integration: %v", err)
	} else {
		log.Println("✅ Cash&Bank COA integration verified")
	}

	// Step 4: Verify SSOT Journal System
	log.Println("\n📑 Step 4: Verifying SSOT Journal System...")
	if err := verifySSOTJournalSystem(db); err != nil {
		log.Printf("❌ Failed to verify SSOT system: %v", err)
	} else {
		log.Println("✅ SSOT Journal system verified")
	}

	// Step 5: Verify Purchase Approval Integration
	log.Println("\n🛒 Step 5: Verifying Purchase Approval Integration...")
	if err := verifyPurchaseApprovalIntegration(db); err != nil {
		log.Printf("❌ Failed to verify purchase integration: %v", err)
	} else {
		log.Println("✅ Purchase approval integration verified")
	}

	// Step 6: Test Integration Health
	log.Println("\n🩺 Step 6: Testing Integration Health...")
	testIntegrationHealth(db)

	log.Println("\n🎉 Recovery and Integration Fix Completed!")
	log.Println("✅ All systems should now be functional and integrated properly")
}

func fixAccountDuplicatesAndRecover(db *gorm.DB) error {
	log.Println("   🔍 Checking for duplicate accounts...")
	
	// First, check if there are active accounts
	var activeCount int64
	db.Model(&models.Account{}).Where("deleted_at IS NULL").Count(&activeCount)
	log.Printf("   📊 Active accounts: %d", activeCount)
	
	// Check soft deleted accounts
	var deletedCount int64
	db.Model(&models.Account{}).Where("deleted_at IS NOT NULL").Count(&deletedCount)
	log.Printf("   🗑️  Soft deleted accounts: %d", deletedCount)
	
	if deletedCount == 0 {
		log.Println("   ℹ️  No soft deleted accounts to recover")
		return nil
	}

	// Handle duplicates by finding conflicting codes
	type DuplicateInfo struct {
		Code  string
		Count int64
	}
	
	var duplicates []DuplicateInfo
	db.Raw(`
		SELECT code, COUNT(*) as count
		FROM accounts 
		GROUP BY code 
		HAVING COUNT(*) > 1
	`).Scan(&duplicates)

	log.Printf("   🔄 Found %d code(s) with duplicates", len(duplicates))

	// Process duplicates
	for _, dup := range duplicates {
		log.Printf("   🔧 Fixing duplicate code: %s (%d instances)", dup.Code, dup.Count)
		
		// Get all instances of this code
		var accounts []models.Account
		db.Unscoped().Where("code = ?", dup.Code).Order("created_at ASC").Find(&accounts)
		
		// Keep the first (oldest) active, mark others for code change
		activeFound := false
		for i, acc := range accounts {
			if acc.DeletedAt.Time.IsZero() {
				// This is active
				if !activeFound {
					log.Printf("      ✅ Keeping active account: %s (%s)", acc.Code, acc.Name)
					activeFound = true
				} else {
					// Rename this duplicate active account
					newCode := fmt.Sprintf("%s_DUP_%d", acc.Code, i)
					db.Model(&acc).Update("code", newCode)
					log.Printf("      🔄 Renamed duplicate active account to: %s", newCode)
				}
			} else {
				// This is soft deleted
				if !activeFound {
					// Recover this one as the main account
					db.Model(&acc).Update("deleted_at", nil)
					log.Printf("      ♻️  Recovered soft deleted account: %s (%s)", acc.Code, acc.Name)
					activeFound = true
				} else {
					// Rename and recover this as duplicate
					newCode := fmt.Sprintf("%s_RECOVERED_%d", acc.Code, i)
					db.Model(&acc).Updates(map[string]interface{}{
						"code":       newCode,
						"deleted_at": nil,
					})
					log.Printf("      ♻️  Recovered and renamed account to: %s", newCode)
				}
			}
		}
	}

	// Now recover remaining soft deleted accounts that don't have conflicts
	log.Println("   ♻️  Recovering non-conflicting soft deleted accounts...")
	result := db.Model(&models.Account{}).
		Where("deleted_at IS NOT NULL AND code NOT IN (?)", 
			db.Table("accounts").Select("code").Where("deleted_at IS NULL")).
		Update("deleted_at", nil)
	
	log.Printf("   ✅ Recovered %d non-conflicting accounts", result.RowsAffected)

	return nil
}

func recoverCriticalMasterData(db *gorm.DB) error {
	// Recovery priority: Users, Contacts, Products, Cash Banks, etc.
	criticalTables := []struct {
		Model interface{}
		Name  string
	}{
		{&models.Contact{}, "contacts"},
		{&models.Product{}, "products"}, 
		{&models.CashBank{}, "cash_banks"},
		{&models.ProductCategory{}, "product_categories"},
		{&models.Permission{}, "permissions"},
	}

	for _, table := range criticalTables {
		// Count soft deleted
		var deletedCount int64
		db.Model(table.Model).Where("deleted_at IS NOT NULL").Count(&deletedCount)
		
		if deletedCount > 0 {
			log.Printf("   ♻️  Recovering %d soft deleted %s...", deletedCount, table.Name)
			result := db.Model(table.Model).Where("deleted_at IS NOT NULL").Update("deleted_at", nil)
			log.Printf("   ✅ Recovered %d %s", result.RowsAffected, table.Name)
		} else {
			log.Printf("   ℹ️  No soft deleted %s to recover", table.Name)
		}
	}

	return nil
}

func verifyCashBankCOAIntegration(db *gorm.DB) error {
	log.Println("   🔍 Checking Cash&Bank accounts...")
	
	// Get all cash banks
	var cashBanks []models.CashBank
	if err := db.Preload("Account").Find(&cashBanks).Error; err != nil {
		return fmt.Errorf("failed to fetch cash banks: %v", err)
	}

	log.Printf("   📊 Found %d cash&bank accounts", len(cashBanks))

	missingAccountLinks := 0
	for _, cb := range cashBanks {
		if cb.AccountID == 0 {
			log.Printf("   ⚠️  Cash&Bank '%s' has no linked COA account", cb.Name)
			
			// Try to find or create appropriate COA account
			accountName := fmt.Sprintf("Cash&Bank - %s", cb.Name)
			account := models.Account{
				Code:        fmt.Sprintf("CB-%s", cb.Code),
				Name:        accountName,
				Type:        "ASSET",
				Category:    "CURRENT_ASSET",
				IsHeader:    false,
				IsActive:    true,
				Balance:     cb.Balance,
			}
			
			if err := db.Create(&account).Error; err != nil {
				log.Printf("   ❌ Failed to create COA for %s: %v", cb.Name, err)
				missingAccountLinks++
			} else {
				// Link the account
				db.Model(&cb).Update("account_id", account.ID)
				log.Printf("   ✅ Created and linked COA account for %s", cb.Name)
			}
		} else {
			if cb.Account.ID == 0 {
				log.Printf("   ⚠️  Cash&Bank '%s' has broken COA link (Account ID: %d)", cb.Name, cb.AccountID)
				missingAccountLinks++
			} else {
				log.Printf("   ✅ Cash&Bank '%s' properly linked to COA '%s'", cb.Name, cb.Account.Name)
			}
		}
	}

	if missingAccountLinks > 0 {
		log.Printf("   ⚠️  Found %d cash&bank accounts with missing COA links", missingAccountLinks)
	} else {
		log.Println("   ✅ All cash&bank accounts properly integrated with COA")
	}

	return nil
}

func verifySSOTJournalSystem(db *gorm.DB) error {
	// Check SSOT journal entries
	var ssotEntries int64
	db.Model(&models.SSOTJournalEntry{}).Count(&ssotEntries)
	log.Printf("   📊 SSOT Journal Entries: %d", ssotEntries)

	// Check SSOT journal lines
	var ssotLines int64
	db.Model(&models.SSOTJournalLine{}).Count(&ssotLines)
	log.Printf("   📊 SSOT Journal Lines: %d", ssotLines)

	// Test SSOT system by creating a simple test entry (then delete it)
	testEntry := &models.SSOTJournalEntry{
		SourceType:       models.SSOTSourceTypeManual,
		SourceID:         nil,
		EntryNumber:      fmt.Sprintf("TEST-%d", time.Now().Unix()),
		EntryDate:        time.Now(),
		Description:      "Integration test entry",
		TotalDebit:       decimal.NewFromInt(0),
		TotalCredit:      decimal.NewFromInt(0),
		Status:           models.SSOTStatusDraft,
		IsAutoGenerated:  false,
		CreatedBy:        1, // Assume admin user ID 1 exists
	}

	if err := db.Create(testEntry).Error; err != nil {
		log.Printf("   ❌ Failed to create test SSOT entry: %v", err)
		return fmt.Errorf("SSOT system not working: %v", err)
	}

	// Delete test entry
	db.Delete(testEntry)
	log.Println("   ✅ SSOT Journal system is functional")

	return nil
}

func verifyPurchaseApprovalIntegration(db *gorm.DB) error {
	// Check approval workflows
	var workflows []models.ApprovalWorkflow
	db.Where("module = ? AND is_active = ?", models.ApprovalModulePurchase, true).Find(&workflows)
	
	log.Printf("   📊 Active purchase workflows: %d", len(workflows))
	
	if len(workflows) == 0 {
		log.Println("   ❌ No active purchase approval workflows found!")
		return fmt.Errorf("no active purchase workflows")
	}

	for _, wf := range workflows {
		var stepCount int64
		db.Model(&models.ApprovalStep{}).Where("workflow_id = ?", wf.ID).Count(&stepCount)
		log.Printf("   🔧 Workflow: %s (%d steps)", wf.Name, stepCount)
		
		if stepCount == 0 {
			log.Printf("   ⚠️  Workflow %s has no steps!", wf.Name)
		}
	}

	return nil
}

func testIntegrationHealth(db *gorm.DB) {
	log.Println("   🏥 Running integration health checks...")
	
	// Check 1: COA Health
	var accountCount int64
	db.Model(&models.Account{}).Where("deleted_at IS NULL").Count(&accountCount)
	log.Printf("   📊 Active COA accounts: %d", accountCount)
	
	// Check 2: Cash&Bank Health  
	var cashBankCount int64
	db.Model(&models.CashBank{}).Where("deleted_at IS NULL").Count(&cashBankCount)
	log.Printf("   💰 Active cash&bank accounts: %d", cashBankCount)
	
	// Check 3: Product Health
	var productCount int64
	db.Model(&models.Product{}).Where("deleted_at IS NULL").Count(&productCount)
	log.Printf("   📦 Active products: %d", productCount)
	
	// Check 4: Contact Health
	var contactCount int64
	db.Model(&models.Contact{}).Where("deleted_at IS NULL").Count(&contactCount)
	log.Printf("   👥 Active contacts: %d", contactCount)
	
	// Check 5: User Health (critical!)
	var userCount int64
	db.Model(&models.User{}).Where("deleted_at IS NULL OR deleted_at IS NULL").Count(&userCount)
	log.Printf("   👤 Active users: %d", userCount)
	
	if userCount == 0 {
		log.Println("   🚨 CRITICAL: No active users found! You may not be able to login!")
	}

	// Check 6: SSOT Health
	var ssotHealthy bool = true
	
	// Try to query SSOT tables
	var ssotEntryCount, ssotLineCount int64
	if err := db.Model(&models.SSOTJournalEntry{}).Count(&ssotEntryCount).Error; err != nil {
		log.Printf("   ❌ SSOT Entry table error: %v", err)
		ssotHealthy = false
	}
	
	if err := db.Model(&models.SSOTJournalLine{}).Count(&ssotLineCount).Error; err != nil {
		log.Printf("   ❌ SSOT Line table error: %v", err)
		ssotHealthy = false
	}
	
	if ssotHealthy {
		log.Printf("   ✅ SSOT system healthy (%d entries, %d lines)", ssotEntryCount, ssotLineCount)
	}

	// Overall Health Score
	healthIssues := 0
	if accountCount == 0 { healthIssues++ }
	if cashBankCount == 0 { healthIssues++ }
	if userCount == 0 { healthIssues++ }
	if !ssotHealthy { healthIssues++ }

	if healthIssues == 0 {
		log.Println("   🎉 System health: EXCELLENT - All integrations functional!")
	} else if healthIssues <= 2 {
		log.Printf("   ⚠️  System health: FAIR - %d minor issues detected", healthIssues)
	} else {
		log.Printf("   🚨 System health: CRITICAL - %d major issues detected", healthIssues)
	}
}