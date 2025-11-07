package main

import (
	"fmt"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Connect to database
	db := database.ConnectDB()

	fmt.Println("=== ANALYZING CASH BANK TRANSACTIONS ===")
	
	// 1. Get all cash bank transactions
	var transactions []models.CashBankTransaction
	db.Where("deleted_at IS NULL").Find(&transactions)

	fmt.Printf("\nFound %d cash bank transactions:\n", len(transactions))
	fmt.Println("ID\tCashBank ID\tAmount\t\tBalance After\tRef Type\tRef ID\tDate\t\tNotes")
	fmt.Println("--\t-----------\t------\t\t-------------\t--------\t------\t----\t\t-----")
	
	for _, tx := range transactions {
		fmt.Printf("%d\t%d\t\t%.2f\t\t%.2f\t\t%s\t\t%v\t%s\t%s\n", 
			tx.ID, tx.CashBankID, tx.Amount, tx.BalanceAfter, tx.ReferenceType, tx.ReferenceID,
			tx.TransactionDate.Format("2006-01-02"), tx.Notes)
	}

	// 2. Check which CashBanks these transactions belong to
	fmt.Printf("\n2. CASHBANK TRANSACTION SUMMARY:\n")
	
	cashBankTxMap := make(map[uint][]models.CashBankTransaction)
	for _, tx := range transactions {
		cashBankTxMap[tx.CashBankID] = append(cashBankTxMap[tx.CashBankID], tx)
	}

	var cashBanks []models.CashBank
	db.Where("deleted_at IS NULL").Find(&cashBanks)

	for _, cb := range cashBanks {
		txs, hasTx := cashBankTxMap[cb.ID]
		if hasTx {
			fmt.Printf("\n%s (ID: %d, Balance: %.2f):\n", cb.Name, cb.ID, cb.Balance)
			var calculatedBalance float64 = 0
			for _, tx := range txs {
				calculatedBalance += tx.Amount
				fmt.Printf("  - Amount: %.2f, Balance After: %.2f, Type: %s, Date: %s\n", 
					tx.Amount, tx.BalanceAfter, tx.ReferenceType, tx.TransactionDate.Format("2006-01-02"))
			}
			fmt.Printf("  Calculated Balance: %.2f, Actual Balance: %.2f", calculatedBalance, cb.Balance)
			if calculatedBalance != cb.Balance {
				fmt.Printf(" ❌ MISMATCH!")
			} else {
				fmt.Printf(" ✅ MATCH")
			}
			fmt.Println()
		}
	}

	// 3. Check for orphaned transactions
	fmt.Printf("\n3. ORPHANED TRANSACTIONS CHECK:\n")
	orphanFound := false
	
	for _, tx := range transactions {
		found := false
		for _, cb := range cashBanks {
			if cb.ID == tx.CashBankID {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Orphaned transaction ID %d references CashBank ID %d (not found)\n", tx.ID, tx.CashBankID)
			orphanFound = true
		}
	}
	
	if !orphanFound {
		fmt.Println("No orphaned transactions found.")
	}

	// 4. Analyze the specific case - Bank BNI test123 (ID: 6)
	fmt.Printf("\n4. BANK BNI TEST123 ANALYSIS:\n")
	
	var bankBNI models.CashBank
	if err := db.Where("id = 6").First(&bankBNI).Error; err == nil {
		fmt.Printf("Bank: %s (ID: %d)\n", bankBNI.Name, bankBNI.ID)
		fmt.Printf("Current Balance: %.2f\n", bankBNI.Balance)
		fmt.Printf("Created At: %s\n", bankBNI.CreatedAt.Format("2006-01-02 15:04:05"))
		
		// Get its transactions
		var bniTransactions []models.CashBankTransaction
		db.Where("cash_bank_id = 6 AND deleted_at IS NULL").Find(&bniTransactions)
		
		fmt.Printf("Transactions: %d\n", len(bniTransactions))
		for _, tx := range bniTransactions {
			fmt.Printf("  - ID: %d, Amount: %.2f, Date: %s, Type: %s, Ref ID: %v\n", 
				tx.ID, tx.Amount, tx.TransactionDate.Format("2006-01-02 15:04:05"), tx.ReferenceType, tx.ReferenceID)
			
			// Check if this transaction has legitimate origin
			if tx.ReferenceType != "" && tx.ReferenceID != 0 {
				fmt.Printf("    ✅ Has reference: %s ID %d\n", tx.ReferenceType, tx.ReferenceID)
			} else {
				fmt.Printf("    🚨 No reference - might be from seed/manual entry\n")
			}
		}
	}

	// 5. Recommendation
	fmt.Printf("\n5. RECOMMENDATION:\n")
	fmt.Println("If any cash bank transaction has no legitimate reference (ReferenceType and ReferenceID),")
	fmt.Println("it might be a residual from seed data and should be investigated or removed.")
	fmt.Println("The correct approach is:")
	fmt.Println("1. CashBank balance should always match the sum of legitimate transactions")
	fmt.Println("2. Each transaction should have proper ReferenceType and ReferenceID")
	fmt.Println("3. COA Account balance should be synced with CashBank balance through journal entries")
}
