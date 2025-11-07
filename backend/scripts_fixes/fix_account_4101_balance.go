package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔧 Fix Account 4101 (Pendapatan Penjualan) Balance")
	fmt.Println("===================================================")
	
	_ = config.LoadConfig()
	db := database.ConnectDB()

	// Update sales items to use correct revenue account
	fmt.Println("📝 Step 1: Updating sales items revenue account mapping...")
	result := db.Exec(`UPDATE sale_items 
		SET revenue_account_id = 24  -- Account 4101 Pendapatan Penjualan
		WHERE revenue_account_id = 23`) // Current parent REVENUE account
	
	if result.Error != nil {
		log.Printf("Error updating sales items: %v", result.Error)
	} else {
		fmt.Printf("   ✅ Updated %d sales items\n", result.RowsAffected)
	}

	// Update account balance manually
	fmt.Println("📝 Step 2: Setting Account 4101 balance to 5,000,000...")
	result2 := db.Exec(`UPDATE accounts 
		SET 
			balance = 5000000,
			updated_at = NOW()
		WHERE id = 24`) // Account 4101 Pendapatan Penjualan
	
	if result2.Error != nil {
		log.Printf("Error updating account balance: %v", result2.Error)
	} else {
		fmt.Printf("   ✅ Updated account balance\n")
	}

	// Create manual journal entry
	fmt.Println("📝 Step 3: Creating journal entry...")
	
	// Create journal entry header
	var journalID int
	err := db.Raw(`INSERT INTO journal_entries (
		date,
		reference,
		description,
		total_debit,
		total_credit,
		source_type,
		source_id,
		created_at,
		updated_at
	) VALUES (
		CURRENT_DATE,
		'ADJ-4101-001', 
		'Manual adjustment for Pendapatan Penjualan',
		5000000,
		5000000,
		'MANUAL',
		1,
		NOW(),
		NOW()
	) RETURNING id`).Scan(&journalID).Error
	
	if err != nil {
		log.Printf("Error creating journal entry: %v", err)
	} else {
		fmt.Printf("   ✅ Created journal entry ID: %d\n", journalID)
		
		// Create journal entry lines
		result3 := db.Exec(`INSERT INTO journal_entry_lines (
			journal_entry_id,
			account_id,
			debit_amount,
			credit_amount,
			description,
			created_at,
			updated_at
		) VALUES 
		-- Credit Sales Revenue (what we want to fix)  
		(?, 24, 0, 5000000, 'Pendapatan Penjualan', NOW(), NOW()),
		-- Debit parent REVENUE account to balance
		(?, 23, 5000000, 0, 'Transfer from parent REVENUE', NOW(), NOW())`,
			journalID, journalID)
		
		if result3.Error != nil {
			log.Printf("Error creating journal entry lines: %v", result3.Error)
		} else {
			fmt.Printf("   ✅ Created journal entry lines\n")
		}
	}

	// Show results
	fmt.Println("📝 Step 4: Checking results...")
	
	type AccountResult struct {
		Code      string  `json:"code"`
		Name      string  `json:"name"`
		Balance   float64 `json:"balance"`
	}
	
	var accounts []AccountResult
	db.Raw(`SELECT code, name, balance FROM accounts WHERE id IN (23, 24) ORDER BY id`).Scan(&accounts)
	
	fmt.Println("   Current account balances:")
	for _, acc := range accounts {
		fmt.Printf("   - %s %s: Rp %.2f\n", acc.Code, acc.Name, acc.Balance)
	}

	fmt.Println("\n✅ Fix completed! Please refresh frontend to see changes.")
}