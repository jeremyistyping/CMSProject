package main

import (
	"database/sql"
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("  FIX VAT/PPN ACCOUNTS DISPLAY ISSUE")
	fmt.Println("==================================================")
	fmt.Println()

	// Initialize database using existing ConnectDB
	db := database.ConnectDB()

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get SQL DB: %v", err)
	}

	fmt.Println("✅ Database connected")
	fmt.Println()

	// Step 1: Check if accounts exist
	fmt.Println("STEP 1: Checking VAT/PPN accounts...")
	fmt.Println("------------------------------------")
	
	var account1240ID, account2103ID int
	var account1240Name, account2103Name string
	
	// Check account 1240
	err = sqlDB.QueryRow(`
		SELECT id, name FROM accounts 
		WHERE code = '1240' AND deleted_at IS NULL 
		LIMIT 1
	`).Scan(&account1240ID, &account1240Name)
	
	if err == sql.ErrNoRows {
		log.Fatal("❌ CRITICAL: Account 1240 (PPN Masukan) NOT FOUND!")
	} else if err != nil {
		log.Fatalf("❌ Error checking account 1240: %v", err)
	}
	
	fmt.Printf("✅ Account 1240 found: %s (ID: %d)\n", account1240Name, account1240ID)
	
	// Check account 2103
	err = sqlDB.QueryRow(`
		SELECT id, name FROM accounts 
		WHERE code = '2103' AND deleted_at IS NULL 
		LIMIT 1
	`).Scan(&account2103ID, &account2103Name)
	
	if err == sql.ErrNoRows {
		log.Fatal("❌ CRITICAL: Account 2103 (PPN Keluaran) NOT FOUND!")
	} else if err != nil {
		log.Fatalf("❌ Error checking account 2103: %v", err)
	}
	
	fmt.Printf("✅ Account 2103 found: %s (ID: %d)\n", account2103Name, account2103ID)
	fmt.Println()

	// Step 2: Check tax_account_settings
	fmt.Println("STEP 2: Checking tax_account_settings...")
	fmt.Println("------------------------------------")
	
	var settingsCount int
	err = sqlDB.QueryRow(`
		SELECT COUNT(*) FROM tax_account_settings WHERE deleted_at IS NULL
	`).Scan(&settingsCount)
	
	if err != nil {
		log.Fatalf("❌ Error checking tax_account_settings: %v", err)
	}
	
	fmt.Printf("ℹ️  Found %d tax_account_settings record(s)\n", settingsCount)
	fmt.Println()

	// Step 3: Update or Create settings
	fmt.Println("STEP 3: Updating tax_account_settings...")
	fmt.Println("------------------------------------")
	
	if settingsCount > 0 {
		// Update existing records
		result, err := sqlDB.Exec(`
			UPDATE tax_account_settings
			SET 
				purchase_input_vat_account_id = $1,
				sales_output_vat_account_id = $2,
				updated_at = NOW()
			WHERE deleted_at IS NULL
		`, account1240ID, account2103ID)
		
		if err != nil {
			log.Fatalf("❌ Error updating tax_account_settings: %v", err)
		}
		
		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("✅ Updated %d tax_account_settings record(s)\n", rowsAffected)
	} else {
		// Create default settings
		fmt.Println("⚠️  No tax_account_settings found, creating default...")
		
		// Get other required account IDs
		var (
			salesReceivableID, salesCashID, salesBankID, salesRevenueID int
			purchasePayableID, purchaseCashID, purchaseBankID, purchaseExpenseID int
		)
		
		// Get account IDs with defaults
		sqlDB.QueryRow("SELECT COALESCE((SELECT id FROM accounts WHERE code = '1201' AND deleted_at IS NULL LIMIT 1), 1)").Scan(&salesReceivableID)
		sqlDB.QueryRow("SELECT COALESCE((SELECT id FROM accounts WHERE code = '1101' AND deleted_at IS NULL LIMIT 1), 1)").Scan(&salesCashID)
		sqlDB.QueryRow("SELECT COALESCE((SELECT id FROM accounts WHERE code = '1102' AND deleted_at IS NULL LIMIT 1), 1)").Scan(&salesBankID)
		sqlDB.QueryRow("SELECT COALESCE((SELECT id FROM accounts WHERE code = '4101' AND deleted_at IS NULL LIMIT 1), 1)").Scan(&salesRevenueID)
		sqlDB.QueryRow("SELECT COALESCE((SELECT id FROM accounts WHERE code = '2001' AND deleted_at IS NULL LIMIT 1), 1)").Scan(&purchasePayableID)
		sqlDB.QueryRow("SELECT COALESCE((SELECT id FROM accounts WHERE code = '1101' AND deleted_at IS NULL LIMIT 1), 1)").Scan(&purchaseCashID)
		sqlDB.QueryRow("SELECT COALESCE((SELECT id FROM accounts WHERE code = '1102' AND deleted_at IS NULL LIMIT 1), 1)").Scan(&purchaseBankID)
		sqlDB.QueryRow("SELECT COALESCE((SELECT id FROM accounts WHERE code = '6001' AND deleted_at IS NULL LIMIT 1), 1)").Scan(&purchaseExpenseID)
		
		_, err = sqlDB.Exec(`
			INSERT INTO tax_account_settings (
				sales_receivable_account_id,
				sales_cash_account_id,
				sales_bank_account_id,
				sales_revenue_account_id,
				sales_output_vat_account_id,
				purchase_payable_account_id,
				purchase_cash_account_id,
				purchase_bank_account_id,
				purchase_input_vat_account_id,
				purchase_expense_account_id,
				is_active,
				apply_to_all_companies,
				updated_by,
				created_at,
				updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true, true, 1, NOW(), NOW()
			)
		`, salesReceivableID, salesCashID, salesBankID, salesRevenueID, account2103ID,
		   purchasePayableID, purchaseCashID, purchaseBankID, account1240ID, purchaseExpenseID)
		
		if err != nil {
			log.Fatalf("❌ Error creating tax_account_settings: %v", err)
		}
		
		fmt.Println("✅ Created default tax_account_settings with VAT accounts")
	}
	fmt.Println()

	// Step 4: Verification
	fmt.Println("STEP 4: Verification...")
	fmt.Println("------------------------------------")
	
	rows, err := sqlDB.Query(`
		SELECT 
			ts.id,
			ts.is_active,
			ts.purchase_input_vat_account_id,
			ts.sales_output_vat_account_id,
			a1.code as ppn_masukan_code,
			a1.name as ppn_masukan_name,
			a2.code as ppn_keluaran_code,
			a2.name as ppn_keluaran_name
		FROM tax_account_settings ts
		LEFT JOIN accounts a1 ON ts.purchase_input_vat_account_id = a1.id
		LEFT JOIN accounts a2 ON ts.sales_output_vat_account_id = a2.id
		WHERE ts.deleted_at IS NULL
		ORDER BY ts.is_active DESC, ts.updated_at DESC
	`)
	
	if err != nil {
		log.Fatalf("❌ Error verifying settings: %v", err)
	}
	defer rows.Close()
	
	verifiedCount := 0
	for rows.Next() {
		var (
			id, inputVATID, outputVATID int
			isActive bool
			inputCode, inputName, outputCode, outputName sql.NullString
		)
		
		err = rows.Scan(&id, &isActive, &inputVATID, &outputVATID, &inputCode, &inputName, &outputCode, &outputName)
		if err != nil {
			log.Printf("⚠️  Error scanning row: %v", err)
			continue
		}
		
		fmt.Printf("\nSettings ID: %d (Active: %v)\n", id, isActive)
		fmt.Println("  PPN Masukan (Input VAT):")
		if inputCode.Valid {
			fmt.Printf("    ✅ %s - %s (ID: %d)\n", inputCode.String, inputName.String, inputVATID)
		} else {
			fmt.Println("    ❌ NOT CONFIGURED")
		}
		
		fmt.Println("  PPN Keluaran (Output VAT):")
		if outputCode.Valid {
			fmt.Printf("    ✅ %s - %s (ID: %d)\n", outputCode.String, outputName.String, outputVATID)
		} else {
			fmt.Println("    ❌ NOT CONFIGURED")
		}
		
		verifiedCount++
	}
	
	fmt.Println()
	fmt.Println("==================================================")
	fmt.Printf("✅ Verified %d tax_account_settings record(s)\n", verifiedCount)
	fmt.Println("==================================================")
	fmt.Println()
	fmt.Println("📝 NEXT STEPS:")
	fmt.Println("   1. Restart your backend server (if running)")
	fmt.Println("   2. Clear browser cache or hard refresh (Ctrl+Shift+R)")
	fmt.Println("   3. Go to Settings > Tax Accounts page")
	fmt.Println("   4. VAT/PPN accounts should now appear!")
	fmt.Println()
	fmt.Println("✅ Fix completed successfully!")
	fmt.Println("==================================================")
}
