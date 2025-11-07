package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🧪 Testing Auto Balance Sync System...")

	// Initialize database connection
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")

	// Test the auto balance system
	if err := testAutoBalanceSystem(db); err != nil {
		log.Fatalf("❌ Auto balance system test failed: %v", err)
	}

	log.Println("🎉 Auto balance sync system test completed successfully!")
}

func testAutoBalanceSystem(db *gorm.DB) error {
	log.Println("🔬 Running comprehensive auto balance sync tests...")

	// Initialize services
	accountRepo := repositories.NewAccountRepository(db)
	autoSyncService := services.NewAutoBalanceSyncService(db, accountRepo)

	// Test 1: Validate initial state
	log.Println("📋 Test 1: Checking initial balance consistency...")
	
	initialReport, err := autoSyncService.ValidateBalanceConsistency()
	if err != nil {
		return fmt.Errorf("initial validation failed: %w", err)
	}

	log.Printf("Initial state - Consistent: %t", initialReport.IsConsistent)
	log.Printf("  Cash Bank Issues: %d", len(initialReport.CashBankIssues))
	log.Printf("  Parent-Child Issues: %d", len(initialReport.ParentChildIssues))
	log.Printf("  Balance Equation Diff: %.2f", initialReport.BalanceEquationDifference)

	// Test 2: Create a test cash bank transaction and see if triggers work
	log.Println("💸 Test 2: Testing trigger activation with new transaction...")
	
	// Find an existing cash bank to add transaction to
	var cashBank models.CashBank
	if err := db.Where("deleted_at IS NULL AND is_active = true").First(&cashBank).Error; err != nil {
		return fmt.Errorf("no active cash bank found for testing: %w", err)
	}

	log.Printf("Using cash bank: %s (%s)", cashBank.Code, cashBank.Name)

	// Record balance before transaction
	var beforeBalance float64
	if err := db.Model(&cashBank).Select("balance").Scan(&beforeBalance).Error; err != nil {
		return fmt.Errorf("failed to get initial balance: %w", err)
	}

	// Get linked COA balance before transaction
	var beforeCOABalance float64
	if err := db.Model(&models.Account{}).Where("id = ?", cashBank.AccountID).Select("balance").Scan(&beforeCOABalance).Error; err != nil {
		return fmt.Errorf("failed to get COA balance: %w", err)
	}

	log.Printf("Before transaction - Cash Bank: %.2f, COA: %.2f", beforeBalance, beforeCOABalance)

	// Create a test transaction
	testTransaction := models.CashBankTransaction{
		CashBankID:      cashBank.ID,
		Amount:          1000.00, // Test amount
		ReferenceType:   "DEPOSIT",
		Notes:          "Auto balance sync test transaction",
		TransactionDate: time.Now(),
	}

	if err := db.Create(&testTransaction).Error; err != nil {
		return fmt.Errorf("failed to create test transaction: %w", err)
	}

	log.Printf("✅ Created test transaction: ID=%d, Amount=%.2f", testTransaction.ID, testTransaction.Amount)

	// Check if triggers updated balances automatically
	var afterBalance float64
	if err := db.Model(&cashBank).Where("id = ?", cashBank.ID).Select("balance").Scan(&afterBalance).Error; err != nil {
		return fmt.Errorf("failed to get updated balance: %w", err)
	}

	// Get linked COA balance after transaction
	var afterCOABalance float64
	if err := db.Model(&models.Account{}).Where("id = ?", cashBank.AccountID).Select("balance").Scan(&afterCOABalance).Error; err != nil {
		return fmt.Errorf("failed to get COA balance: %w", err)
	}

	log.Printf("After transaction - Cash Bank: %.2f, COA: %.2f", afterBalance, afterCOABalance)

	// Validate trigger worked
	expectedBalance := beforeBalance + testTransaction.Amount
	if afterBalance != expectedBalance {
		return fmt.Errorf("cash bank balance trigger failed: expected %.2f, got %.2f", expectedBalance, afterBalance)
	}

	if afterCOABalance != afterBalance {
		return fmt.Errorf("COA balance sync failed: cash bank %.2f != COA %.2f", afterBalance, afterCOABalance)
	}

	log.Println("✅ Triggers working correctly - balances synced automatically")

	// Test 3: Validate parent account balances updated
	log.Println("🏗️ Test 3: Checking parent account balance rollup...")

	var account models.Account
	if err := db.Where("id = ?", cashBank.AccountID).First(&account).Error; err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	if account.ParentID != nil {
		// Check if parent balance was updated
		var parentBalance float64
		if err := db.Model(&models.Account{}).Where("id = ?", *account.ParentID).Select("balance").Scan(&parentBalance).Error; err != nil {
			return fmt.Errorf("failed to get parent balance: %w", err)
		}

		// Calculate expected parent balance
		var childrenSum float64
		if err := db.Model(&models.Account{}).Where("parent_id = ?", *account.ParentID).Select("COALESCE(SUM(balance), 0)").Scan(&childrenSum).Error; err != nil {
			return fmt.Errorf("failed to calculate children sum: %w", err)
		}

		log.Printf("Parent balance: %.2f, Children sum: %.2f", parentBalance, childrenSum)

		if parentBalance != childrenSum {
			return fmt.Errorf("parent balance rollup failed: parent %.2f != children sum %.2f", parentBalance, childrenSum)
		}

		log.Println("✅ Parent account balance rollup working correctly")
	} else {
		log.Println("ℹ️ Account has no parent - skipping parent balance test")
	}

	// Test 4: Run final validation
	log.Println("🔍 Test 4: Final balance consistency validation...")
	
	finalReport, err := autoSyncService.ValidateBalanceConsistency()
	if err != nil {
		return fmt.Errorf("final validation failed: %w", err)
	}

	log.Printf("Final state - Consistent: %t", finalReport.IsConsistent)
	log.Printf("  Cash Bank Issues: %d", len(finalReport.CashBankIssues))
	log.Printf("  Parent-Child Issues: %d", len(finalReport.ParentChildIssues))
	log.Printf("  Balance Equation Diff: %.2f", finalReport.BalanceEquationDifference)

	// Clean up test transaction
	log.Println("🧹 Cleaning up test transaction...")
	if err := db.Delete(&testTransaction).Error; err != nil {
		log.Printf("⚠️ Warning: Failed to clean up test transaction: %v", err)
	} else {
		log.Println("✅ Test transaction cleaned up")
	}

	// Test 5: Verify cleanup worked
	log.Println("♻️ Test 5: Verifying cleanup...")
	
	var cleanupBalance float64
	if err := db.Model(&cashBank).Where("id = ?", cashBank.ID).Select("balance").Scan(&cleanupBalance).Error; err != nil {
		return fmt.Errorf("failed to get cleanup balance: %w", err)
	}

	if cleanupBalance != beforeBalance {
		return fmt.Errorf("cleanup failed: balance should be %.2f but is %.2f", beforeBalance, cleanupBalance)
	}

	log.Printf("✅ Cleanup successful - balance restored to %.2f", cleanupBalance)

	// Test 6: Trigger existence verification
	log.Println("🔧 Test 6: Verifying trigger existence...")
	
	var triggerCount int64
	if err := db.Raw(`
		SELECT COUNT(*) 
		FROM pg_trigger 
		WHERE tgname IN (
			'trigger_sync_cashbank_coa',
			'trigger_recalc_cashbank_balance_insert',
			'trigger_recalc_cashbank_balance_update', 
			'trigger_recalc_cashbank_balance_delete',
			'trigger_validate_account_balance'
		)
	`).Scan(&triggerCount).Error; err != nil {
		return fmt.Errorf("failed to validate triggers: %w", err)
	}

	if triggerCount != 5 {
		return fmt.Errorf("expected 5 triggers, found %d", triggerCount)
	}

	log.Printf("✅ All %d triggers are properly installed", triggerCount)

	return nil
}

// Additional helper functions for extended testing
func testManualSync(db *gorm.DB) error {
	log.Println("🔄 Testing manual sync operations...")
	
	accountRepo := repositories.NewAccountRepository(db)
	autoSyncService := services.NewAutoBalanceSyncService(db, accountRepo)

	// Find a cash bank to test manual sync
	var cashBank models.CashBank
	if err := db.Where("deleted_at IS NULL AND is_active = true").First(&cashBank).Error; err != nil {
		return fmt.Errorf("no active cash bank found: %w", err)
	}

	// Test manual sync
	if err := autoSyncService.SyncCashBankToAccount(cashBank.ID); err != nil {
		return fmt.Errorf("manual sync failed: %w", err)
	}

	log.Printf("✅ Manual sync completed for cash bank %s", cashBank.Code)

	// Test fix all balance issues
	if err := autoSyncService.FixAllBalanceIssues(); err != nil {
		return fmt.Errorf("fix all balance issues failed: %w", err)
	}

	log.Println("✅ Fix all balance issues completed")
	return nil
}

func performanceTest(db *gorm.DB) error {
	log.Println("⚡ Running performance tests...")

	// This would test performance under load
	// For now, we'll just measure single operations
	
	accountRepo := repositories.NewAccountRepository(db)
	autoSyncService := services.NewAutoBalanceSyncService(db, accountRepo)

	// Measure validation time
	start := time.Now()
	_, err := autoSyncService.ValidateBalanceConsistency()
	if err != nil {
		return fmt.Errorf("performance test validation failed: %w", err)
	}
	duration := time.Since(start)

	log.Printf("✅ Balance validation completed in %v", duration)

	if duration > time.Second*10 {
		log.Printf("⚠️ Warning: Validation took longer than expected: %v", duration)
	}

	return nil
}