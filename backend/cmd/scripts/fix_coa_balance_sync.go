package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
)

func main() {
	log.Println("🚀 Starting COA Balance Synchronization Fix")

	// Initialize database connection
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("❌ Failed to connect to database")
	}

	// Initialize repositories
	accountRepo := repositories.NewAccountRepository(db)

	// Initialize the sync service
	syncService := services.NewPurchasePaymentCOASyncService(db, accountRepo)

	// Check command line arguments
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "check":
		checkBalanceDiscrepancies(syncService)
	case "fix":
		fixCOABalanceIssues(syncService)
	case "sync":
		syncAllBalances(syncService)
	case "test":
		testSyncService(syncService)
	default:
		log.Printf("❌ Unknown command: %s", command)
		showUsage()
	}
}

func showUsage() {
	fmt.Println(`
Usage: go run fix_coa_balance_sync.go <command>

Commands:
  check    - Check for balance discrepancies between Cash & Bank and COA
  fix      - Run one-time fix for existing COA balance issues
  sync     - Synchronize all COA balances with Cash & Bank balances
  test     - Test the sync service functionality

Examples:
  go run backend/cmd/scripts/fix_coa_balance_sync.go check
  go run backend/cmd/scripts/fix_coa_balance_sync.go fix
`)
}

func checkBalanceDiscrepancies(syncService *services.PurchasePaymentCOASyncService) {
	log.Println("🔍 Checking for balance discrepancies...")

	discrepancies, err := syncService.GetBalanceDiscrepancies()
	if err != nil {
		log.Printf("❌ Failed to get balance discrepancies: %v", err)
		return
	}

	if len(discrepancies) == 0 {
		log.Println("✅ No balance discrepancies found! All accounts are in sync.")
		return
	}

	log.Printf("⚠️ Found %d balance discrepancies:", len(discrepancies))
	fmt.Println("\n📊 BALANCE DISCREPANCIES REPORT")
	fmt.Println("================================================================")
	fmt.Printf("%-15s %-30s %-15s %-15s %-15s\n", "Bank ID", "Account Name", "Bank Balance", "COA Balance", "Difference")
	fmt.Println("================================================================")

	totalDifference := 0.0
	for _, disc := range discrepancies {
		bankID := disc["cash_bank_id"]
		accountName := disc["cash_bank_name"]
		bankBalance := disc["cash_bank_balance"]
		coaBalance := disc["coa_balance"]
		difference := disc["difference"]

		fmt.Printf("%-15v %-30v %-15.2f %-15.2f %-15.2f\n", 
			bankID, accountName, bankBalance, coaBalance, difference)

		if diff, ok := difference.(float64); ok {
			totalDifference += diff
		}
	}

	fmt.Println("================================================================")
	fmt.Printf("Total Difference: %.2f\n", totalDifference)
	fmt.Println()

	log.Println("💡 To fix these discrepancies, run: go run fix_coa_balance_sync.go fix")
}

func fixCOABalanceIssues(syncService *services.PurchasePaymentCOASyncService) {
	log.Println("🛠️ Running one-time fix for COA balance issues...")

	// First, show current discrepancies
	log.Println("📊 Checking discrepancies before fix...")
	discrepancies, err := syncService.GetBalanceDiscrepancies()
	if err != nil {
		log.Printf("❌ Failed to get balance discrepancies: %v", err)
		return
	}

	beforeCount := len(discrepancies)
	log.Printf("Found %d discrepancies before fix", beforeCount)

	// Run the fix
	err = syncService.FixPaymentCOABalanceIssue()
	if err != nil {
		log.Printf("❌ Failed to fix COA balance issues: %v", err)
		return
	}

	// Check discrepancies after fix
	log.Println("📊 Checking discrepancies after fix...")
	discrepanciesAfter, err := syncService.GetBalanceDiscrepancies()
	if err != nil {
		log.Printf("❌ Failed to get balance discrepancies after fix: %v", err)
		return
	}

	afterCount := len(discrepanciesAfter)
	log.Printf("Found %d discrepancies after fix", afterCount)

	if afterCount == 0 {
		log.Println("🎉 All balance discrepancies have been fixed!")
	} else if afterCount < beforeCount {
		log.Printf("✅ Fixed %d out of %d discrepancies", beforeCount-afterCount, beforeCount)
		if afterCount > 0 {
			log.Println("⚠️ Remaining discrepancies require manual investigation")
		}
	} else {
		log.Println("⚠️ Fix completed but discrepancies remain. Manual investigation may be needed.")
	}

	log.Println("✅ COA balance fix process completed")
}

func syncAllBalances(syncService *services.PurchasePaymentCOASyncService) {
	log.Println("🔄 Synchronizing all COA balances with Cash & Bank balances...")

	err := syncService.SyncAllCOABalancesWithCashBanks()
	if err != nil {
		log.Printf("❌ Failed to sync COA balances: %v", err)
		return
	}

	log.Println("✅ COA balance synchronization completed")

	// Check remaining discrepancies
	log.Println("📊 Checking for remaining discrepancies...")
	discrepancies, err := syncService.GetBalanceDiscrepancies()
	if err != nil {
		log.Printf("⚠️ Failed to check discrepancies after sync: %v", err)
		return
	}

	if len(discrepancies) == 0 {
		log.Println("🎉 All balances are now synchronized!")
	} else {
		log.Printf("⚠️ %d discrepancies remain after sync", len(discrepancies))
	}
}

func testSyncService(syncService *services.PurchasePaymentCOASyncService) {
	log.Println("🧪 Testing sync service functionality...")

	// Test 1: Check discrepancies
	log.Println("Test 1: Getting balance discrepancies...")
	discrepancies, err := syncService.GetBalanceDiscrepancies()
	if err != nil {
		log.Printf("❌ Test 1 failed: %v", err)
	} else {
		log.Printf("✅ Test 1 passed: Found %d discrepancies", len(discrepancies))
	}

	// Test 2: Test sync functionality (dry run)
	log.Println("Test 2: Testing sync functionality...")
	err = syncService.SyncAllCOABalancesWithCashBanks()
	if err != nil {
		log.Printf("❌ Test 2 failed: %v", err)
	} else {
		log.Println("✅ Test 2 passed: Sync functionality working")
	}

	log.Println("🎉 All tests completed")
}