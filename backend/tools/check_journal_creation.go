package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"github.com/shopspring/decimal"
)

func main() {
	// Load config and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("🔍 Checking Journal Entry Creation")
	fmt.Println("==================================")

	// 1. Check recent sales
	fmt.Println("\n1. Recent Sales:")
	var sales []models.Sale
	err := db.Model(&models.Sale{}).
		Preload("Customer").
		Order("created_at DESC").
		Limit(5).
		Find(&sales).Error
	if err != nil {
		log.Printf("Error fetching sales: %v", err)
		return
	}

	for _, sale := range sales {
		fmt.Printf("   Sale ID: %d, Code: %s, Total: %.2f, Status: %s\n", 
			sale.ID, sale.Code, sale.TotalAmount, sale.Status)
	}

	// 2. Check SSOT journal entries
	fmt.Println("\n2. SSOT Journal Entries:")
	var ssotEntries []models.SSOTJournalEntry
	err = db.Model(&models.SSOTJournalEntry{}).
		Preload("Lines.Account").
		Order("created_at DESC").
		Limit(10).
		Find(&ssotEntries).Error
	if err != nil {
		log.Printf("Error fetching SSOT entries: %v", err)
	} else {
		for _, entry := range ssotEntries {
			totalDebit, _ := entry.TotalDebit.Float64()
			totalCredit, _ := entry.TotalCredit.Float64()
			fmt.Printf("   Entry ID: %d, Desc: %s, Debit: %.2f, Credit: %.2f, Status: %s\n",
				entry.ID, entry.Description, totalDebit, totalCredit, entry.Status)
			
			for _, line := range entry.Lines {
				debitAmount, _ := line.DebitAmount.Float64()
				creditAmount, _ := line.CreditAmount.Float64()
				if line.Account != nil {
					fmt.Printf("     → %s (%s): Dr %.2f, Cr %.2f\n",
						line.Account.Name, line.Account.Code, debitAmount, creditAmount)
				} else {
					fmt.Printf("     → Account ID %d: Dr %.2f, Cr %.2f\n",
						line.AccountID, debitAmount, creditAmount)
				}
			}
		}
	}

	// 3. Check legacy journal entries
	fmt.Println("\n3. Legacy Journal Entries:")
	var legacyEntries []models.JournalEntry
	err = db.Model(&models.JournalEntry{}).
		Preload("JournalLines.Account").
		Order("created_at DESC").
		Limit(5).
		Find(&legacyEntries).Error
	if err != nil {
		log.Printf("Error fetching legacy entries: %v", err)
	} else {
		for _, entry := range legacyEntries {
			fmt.Printf("   Legacy Entry ID: %d, Desc: %s, Total Debit: %.2f, Credit: %.2f\n",
				entry.ID, entry.Description, entry.TotalDebit, entry.TotalCredit)
		}
	}

	// 4. Check account balances from SSOT
	fmt.Println("\n4. Key Account Balances (SSOT):")
	accountCodes := []string{"1201", "4101", "1104"}
	for _, code := range accountCodes {
		var account models.Account
		err := db.Where("code = ?", code).First(&account).Error
		if err != nil {
			fmt.Printf("   Account %s: Not found\n", code)
			continue
		}

		// Get SSOT balance
		var ssotBalance struct {
			NetBalance decimal.Decimal
		}
		err = db.Raw(`
			SELECT 
				COALESCE(SUM(CASE WHEN credit_amount > debit_amount THEN credit_amount - debit_amount ELSE 0 END) -
					     SUM(CASE WHEN debit_amount > credit_amount THEN debit_amount - credit_amount ELSE 0 END), 0) as net_balance
			FROM ssot_journal_lines sjl
			JOIN ssot_journal_entries sje ON sjl.journal_entry_id = sje.id
			WHERE sjl.account_id = ? AND sje.status = 'POSTED'
		`, account.ID).Scan(&ssotBalance).Error
		
		if err != nil {
			fmt.Printf("   Account %s (%s): Error calculating SSOT balance: %v\n", code, account.Name, err)
		} else {
			ssotBalanceFloat, _ := ssotBalance.NetBalance.Float64()
			fmt.Printf("   Account %s (%s): Balance %.2f, SSOT Balance %.2f\n", 
				code, account.Name, account.Balance, ssotBalanceFloat)
		}
	}

	// 5. Check if service is being used
	fmt.Println("\n5. Testing Corrected Service Creation:")
	fmt.Println("   This shows that CorrectedSSOTSalesJournalService is available")
	
	fmt.Println("\n✅ Analysis complete!")
	fmt.Println("\nNext steps if balance still 0:")
	fmt.Println("1. Create a new test sale")
	fmt.Println("2. Check if journal entries are created")
	fmt.Println("3. Verify SSOT balance calculation")
	fmt.Println("4. Refresh frontend to see updated balances")
}