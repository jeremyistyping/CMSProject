package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("=== Testing CashBank SSOT Integration Service ===")
	fmt.Println("================================================")

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Database connection established")

	// Initialize repositories
	accountRepo := repositories.NewAccountRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)

	// Initialize services
	cashBankService := services.NewCashBankService(db, cashBankRepo, accountRepo)
	unifiedJournalService := services.NewUnifiedJournalService(db)

	// Initialize integrated service
	integratedService := services.NewCashBankIntegratedService(
		db,
		cashBankService,
		unifiedJournalService,
		accountRepo,
	)

	fmt.Println("✅ Services initialized")

	// Test 1: Get integrated summary
	fmt.Println("\n--- Test 1: Get Integrated Summary ---")
	summary, err := integratedService.GetIntegratedSummary()
	if err != nil {
		log.Printf("❌ Failed to get integrated summary: %v", err)
	} else {
		fmt.Printf("✅ Summary retrieved successfully\n")
		fmt.Printf("   Total Accounts: %d\n", summary.SyncStatus.TotalAccounts)
		fmt.Printf("   Synced Accounts: %d\n", summary.SyncStatus.SyncedAccounts)
		fmt.Printf("   Variance Accounts: %d\n", summary.SyncStatus.VarianceAccounts)
		fmt.Printf("   Total Balance: %s\n", summary.Summary.TotalBalance.String())
		fmt.Printf("   Total SSOT Balance: %s\n", summary.Summary.TotalSSOTBalance.String())
		fmt.Printf("   Balance Variance: %s\n", summary.Summary.BalanceVariance.String())

		if len(summary.Accounts) > 0 {
			fmt.Printf("\n   Sample Account:\n")
			account := summary.Accounts[0]
			fmt.Printf("   - ID: %d\n", account.ID)
			fmt.Printf("   - Name: %s\n", account.Name)
			fmt.Printf("   - Type: %s\n", account.Type)
			fmt.Printf("   - Balance: %s\n", account.Balance.String())
			fmt.Printf("   - SSOT Balance: %s\n", account.SSOTBalance.String())
			fmt.Printf("   - Variance: %s\n", account.Variance.String())
			fmt.Printf("   - Status: %s\n", account.ReconciliationStatus)

			// Test 2: Get integrated account details
			fmt.Printf("\n--- Test 2: Get Account Details (ID: %d) ---\n", account.ID)
			details, err := integratedService.GetIntegratedAccountDetails(account.ID)
			if err != nil {
				log.Printf("❌ Failed to get account details: %v", err)
			} else {
				fmt.Printf("✅ Account details retrieved successfully\n")
				fmt.Printf("   Account Name: %s\n", details.Account.Name)
				fmt.Printf("   CashBank Balance: %.2f\n", details.Account.Balance)
				fmt.Printf("   SSOT Balance: %s\n", details.SSOTBalance.String())
				fmt.Printf("   Balance Difference: %s\n", details.BalanceDifference.String())
				fmt.Printf("   Reconciliation Status: %s\n", details.ReconciliationStatus)
				fmt.Printf("   Recent Transactions: %d\n", len(details.RecentTransactions))
				fmt.Printf("   Related Journal Entries: %d\n", len(details.RelatedJournalEntries))

				// Show transaction details if available
				if len(details.RecentTransactions) > 0 {
					fmt.Printf("\n   Sample Recent Transaction:\n")
					tx := details.RecentTransactions[0]
					fmt.Printf("   - Amount: %s\n", tx.Amount.String())
					fmt.Printf("   - Type: %s\n", tx.Type)
					fmt.Printf("   - Date: %s\n", tx.Date.Format("2006-01-02"))
					fmt.Printf("   - Description: %s\n", tx.Description)
					if tx.JournalEntryNumber != "" {
						fmt.Printf("   - Journal Entry: %s\n", tx.JournalEntryNumber)
					}
				}

				// Show journal entry details if available
				if len(details.RelatedJournalEntries) > 0 {
					fmt.Printf("\n   Sample Journal Entry:\n")
					entry := details.RelatedJournalEntries[0]
					fmt.Printf("   - Entry Number: %s\n", entry.EntryNumber)
					fmt.Printf("   - Description: %s\n", entry.Description)
					fmt.Printf("   - Source Type: %s\n", entry.SourceType)
					fmt.Printf("   - Total Debit: %s\n", entry.TotalDebit.String())
					fmt.Printf("   - Total Credit: %s\n", entry.TotalCredit.String())
					fmt.Printf("   - Status: %s\n", entry.Status)
					fmt.Printf("   - Lines Count: %d\n", len(entry.Lines))
				}
			}
		} else {
			fmt.Println("⚠️  No accounts found for detailed testing")
		}
	}

	// Test 3: Test error handling
	fmt.Println("\n--- Test 3: Error Handling ---")
	_, err = integratedService.GetIntegratedAccountDetails(99999) // Non-existent account
	if err != nil {
		fmt.Printf("✅ Error handling works correctly: %v\n", err)
	} else {
		fmt.Println("⚠️  Error handling may need improvement")
	}

	// Test 4: Database query performance
	fmt.Println("\n--- Test 4: Performance Test ---")
	
	// Test multiple calls to check caching/performance
	start := time.Now()
	for i := 0; i < 5; i++ {
		_, err := integratedService.GetIntegratedSummary()
		if err != nil {
			log.Printf("Performance test iteration %d failed: %v", i+1, err)
		}
	}
	duration := time.Since(start)
	fmt.Printf("✅ 5 summary calls completed in: %v (avg: %v per call)\n", 
		duration, duration/5)

	// Test 5: Validate SSOT balance calculations
	fmt.Println("\n--- Test 5: SSOT Balance Validation ---")
	
	// Get a sample account and validate its balance calculation
	cashBankAccounts, err := cashBankService.GetCashBankAccounts()
	if err != nil || len(cashBankAccounts) == 0 {
		fmt.Println("⚠️  No cash/bank accounts found for balance validation")
	} else {
		account := cashBankAccounts[0]
		fmt.Printf("Testing balance calculation for account: %s\n", account.Name)
		
		// Manual balance calculation query to verify integrated service
		var manualBalance struct {
			CurrentBalance decimal.Decimal `gorm:"column:current_balance"`
		}
		
		err = db.Table("account_balances").
			Where("account_id = ?", account.AccountID).
			Select("current_balance").
			Scan(&manualBalance).Error
			
		if err != nil {
			fmt.Printf("⚠️  Manual balance query failed: %v\n", err)
		} else {
			// Get balance from integrated service
			details, err := integratedService.GetIntegratedAccountDetails(account.ID)
			if err != nil {
				fmt.Printf("⚠️  Integrated service balance query failed: %v\n", err)
			} else {
				fmt.Printf("   Manual Balance Query: %s\n", manualBalance.CurrentBalance.String())
				fmt.Printf("   Integrated Service: %s\n", details.SSOTBalance.String())
				
				if manualBalance.CurrentBalance.Equal(details.SSOTBalance) {
					fmt.Println("✅ Balance calculations match!")
				} else {
					fmt.Printf("⚠️  Balance calculations differ by: %s\n", 
						manualBalance.CurrentBalance.Sub(details.SSOTBalance).String())
				}
			}
		}
	}

	// Test 6: Recent activities
	fmt.Println("\n--- Test 6: Recent Activities ---")
	summary, err = integratedService.GetIntegratedSummary()
	if err != nil {
		fmt.Printf("❌ Failed to get recent activities: %v\n", err)
	} else {
		fmt.Printf("✅ Recent activities count: %d\n", len(summary.RecentActivities))
		
		if len(summary.RecentActivities) > 0 {
			activity := summary.RecentActivities[0]
			fmt.Printf("   Sample Activity:\n")
			fmt.Printf("   - Type: %s\n", activity.Type)
			fmt.Printf("   - Description: %s\n", activity.Description)
			fmt.Printf("   - Amount: %s\n", activity.Amount.String())
			fmt.Printf("   - Account: %s\n", activity.AccountName)
		}
	}

	fmt.Println("\n================================================")
	fmt.Println("✅ CashBank SSOT Integration Service Test Complete!")
	fmt.Println("\nNext Steps:")
	fmt.Println("1. Start backend server: go run cmd/main.go")
	fmt.Println("2. Test API endpoints with PowerShell script")
	fmt.Println("3. Implement frontend components")
	fmt.Println("4. Test end-to-end integration")
}
