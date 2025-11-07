package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	fmt.Println("🔍 DEBUG: Analyzing Revenue and PPN Balance Issues")
	fmt.Println("================================================")

	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()

	// 1. Check account balances for critical accounts
	fmt.Println("\n1. Checking Account Balances for Critical Accounts:")
	fmt.Println("--------------------------------------------------")
	
	var accounts []models.Account
	err := db.Where("code IN (?)", []string{"1201", "2103", "4101"}).Order("code").Find(&accounts).Error
	if err != nil {
		log.Fatalf("Error querying accounts: %v", err)
	}

	for _, account := range accounts {
		fmt.Printf("Account %s (%s): Balance = Rp %.2f, Type = %s\n", 
			account.Code, account.Name, account.Balance, account.Type)
	}

	// 2. Check all sales data
	fmt.Println("\n2. Checking All Sales Data:")
	fmt.Println("---------------------------")
	
	var sales []models.Sale
	err = db.Find(&sales).Error
	if err != nil {
		log.Fatalf("Error querying sales: %v", err)
	}

	var totalSalesAmount float64
	var totalPPNAmount float64
	
	for _, sale := range sales {
		fmt.Printf("Sale ID %d (%s): Total=Rp %.2f, PPN=Rp %.2f, Status=%s\n", 
			sale.ID, sale.Code, sale.TotalAmount, sale.PPN, sale.Status)
		
		if sale.Status == models.SaleStatusInvoiced {
			totalSalesAmount += sale.TotalAmount
			totalPPNAmount += sale.PPN
		}
	}

	fmt.Printf("\nTOTAL INVOICED SALES: Rp %.2f", totalSalesAmount)
	fmt.Printf("\nTOTAL INVOICED PPN: Rp %.2f", totalPPNAmount)
	fmt.Printf("\nTOTAL REVENUE (Sales - PPN): Rp %.2f", totalSalesAmount - totalPPNAmount)

	// 3. Check journal entries for sales
	fmt.Println("\n\n3. Checking Journal Entries for Sales:")
	fmt.Println("--------------------------------------")
	
	var journalEntries []models.SSOTJournalEntry
	err = db.Where("source_type = ?", "SALES").Preload("Lines").Find(&journalEntries).Error
	if err != nil {
		log.Fatalf("Error querying journal entries: %v", err)
	}

	fmt.Printf("Found %d journal entries for sales\n", len(journalEntries))
	
	var totalDebit, totalCredit float64
	var revenueCredit, ppnCredit, arDebit float64
	
	for _, entry := range journalEntries {
		fmt.Printf("Journal Entry ID %d: Source Code %s, Status %s\n", 
			entry.ID, entry.SourceCode, entry.Status)
		
		for _, line := range entry.Lines {
			var account models.Account
			db.First(&account, line.AccountID)
			
			debit := line.DebitAmount.InexactFloat64()
			credit := line.CreditAmount.InexactFloat64()
			
			fmt.Printf("  - Account %s (%s): Debit=%.2f, Credit=%.2f\n", 
				account.Code, account.Name, debit, credit)
			
			totalDebit += debit
			totalCredit += credit
			
			// Track specific account impacts
			switch account.Code {
			case "1201": // Accounts Receivable
				arDebit += debit
			case "4101": // Sales Revenue
				revenueCredit += credit
			case "2103": // PPN Keluaran
				ppnCredit += credit
			}
		}
		fmt.Println()
	}

	fmt.Printf("JOURNAL TOTALS:\n")
	fmt.Printf("  Total Debits: Rp %.2f\n", totalDebit)
	fmt.Printf("  Total Credits: Rp %.2f\n", totalCredit)
	fmt.Printf("  AR Debits: Rp %.2f\n", arDebit)
	fmt.Printf("  Revenue Credits: Rp %.2f\n", revenueCredit)
	fmt.Printf("  PPN Credits: Rp %.2f\n", ppnCredit)

	// 4. Check PPN balance validation issue
	fmt.Println("\n\n4. Checking PPN Balance Issue:")
	fmt.Println("------------------------------")
	
	var ppnAccount models.Account
	err = db.Where("code = ?", "2103").First(&ppnAccount).Error
	if err != nil {
		log.Fatalf("Error finding PPN account: %v", err)
	}

	fmt.Printf("PPN Keluaran Account (2103) Balance: Rp %.2f\n", ppnAccount.Balance)
	
	if ppnAccount.Balance < 0 {
		fmt.Printf("❌ PROBLEM IDENTIFIED: PPN Keluaran has NEGATIVE balance!\n")
		fmt.Printf("   This explains why the PPN validation is failing.\n")
		fmt.Printf("   PPN Keluaran (liability account) should have POSITIVE or ZERO balance.\n")
	} else {
		fmt.Printf("✅ PPN balance is non-negative\n")
	}

	// 5. Suggest fix
	fmt.Println("\n\n5. ANALYSIS & RECOMMENDATIONS:")
	fmt.Println("===============================")
	
	fmt.Printf("Frontend shows Revenue: Rp 31,000,000\n")
	fmt.Printf("Actual Account 4101 Balance: Rp %.2f\n", accounts[2].Balance) // accounts[2] should be 4101
	fmt.Printf("Calculated Revenue from Sales: Rp %.2f\n", totalSalesAmount - totalPPNAmount)
	fmt.Printf("Journal Revenue Credits: Rp %.2f\n", revenueCredit)

	if ppnAccount.Balance < 0 {
		fmt.Printf("\n❌ PRIMARY ISSUE: PPN Keluaran has negative balance (%.2f)\n", ppnAccount.Balance)
		fmt.Printf("   This is preventing new sales from being confirmed.\n")
		fmt.Printf("   The validation service correctly blocks negative PPN balances.\n")
		
		// Calculate what the balance should be
		expectedPPNBalance := ppnCredit // Should equal total PPN credits
		fmt.Printf("\n💡 EXPECTED PPN Balance: Rp %.2f (from journal credits)\n", expectedPPNBalance)
		fmt.Printf("   ACTUAL PPN Balance: Rp %.2f\n", ppnAccount.Balance)
		fmt.Printf("   DIFFERENCE: Rp %.2f\n", expectedPPNBalance - ppnAccount.Balance)
		
		fmt.Printf("\n🛠️ RECOMMENDED FIX:\n")
		fmt.Printf("   1. Reset PPN account balance to match journal credits\n")
		fmt.Printf("   2. Verify all account balances match their journal line totals\n")
		fmt.Printf("   3. Check for any manual balance adjustments that weren't journaled\n")
	}

	if accounts[2].Balance != revenueCredit {
		fmt.Printf("\n❌ REVENUE BALANCE MISMATCH:\n")
		fmt.Printf("   Account 4101 Balance: Rp %.2f\n", accounts[2].Balance)
		fmt.Printf("   Journal Credits: Rp %.2f\n", revenueCredit)
		fmt.Printf("   Difference: Rp %.2f\n", accounts[2].Balance - revenueCredit)
	} else {
		fmt.Printf("\n✅ Revenue balance matches journal credits\n")
	}

	fmt.Println("\n🔍 DEBUG ANALYSIS COMPLETE")
}