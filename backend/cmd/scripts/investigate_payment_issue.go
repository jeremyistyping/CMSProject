package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func loadEnv() {
	envFile := ".env"
	if file, err := os.Open(envFile); err == nil {
		defer file.Close()
		content := make([]byte, 1024)
		if n, err := file.Read(content); err == nil {
			envContent := string(content[:n])
			lines := []string{}
			current := ""
			for _, char := range envContent {
				if char == '\n' || char == '\r' {
					if current != "" {
						lines = append(lines, current)
						current = ""
					}
				} else {
					current += string(char)
				}
			}
			if current != "" {
				lines = append(lines, current)
			}
			
			for _, line := range lines {
				if len(line) > 13 && line[:13] == "DATABASE_URL=" {
					os.Setenv("DATABASE_URL", line[13:])
					break
				}
			}
		}
	}
}

func main() {
	fmt.Println("🔍 INVESTIGATING PAYMENT RECORDING ISSUE")
	fmt.Println("Problem: Payment recorded but Piutang Usaha & Bank balance not updated")
	fmt.Println("")

	loadEnv()
	
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not found in environment")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== STEP 1: CHECKING SALES STATUS ===")
	
	// Check sales transactions and their status
	query := `
		SELECT id, code, invoice_number, status, subtotal, total_amount, 
		       paid_amount, outstanding_amount, created_at
		FROM sales 
		ORDER BY created_at DESC`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to get sales data:", err)
	}
	defer rows.Close()

	fmt.Printf("%-2s | %-12s | %-15s | %-8s | %12s | %12s | %12s | %12s\n", 
		"ID", "Code", "Invoice", "Status", "Subtotal", "Total", "Paid", "Outstanding")
	fmt.Println("---+-------------+----------------+---------+-------------+-------------+-------------+-------------")

	for rows.Next() {
		var id int
		var code, invoice, status string
		var subtotal, totalAmount, paidAmount, outstandingAmount float64
		var createdAt string
		
		err := rows.Scan(&id, &code, &invoice, &status, &subtotal, &totalAmount, 
			&paidAmount, &outstandingAmount, &createdAt)
		if err != nil {
			log.Fatal("Failed to scan sales row:", err)
		}
		
		fmt.Printf("%-2d | %-12s | %-15s | %-8s | %12.2f | %12.2f | %12.2f | %12.2f\n",
			id, code, invoice, status, subtotal, totalAmount, paidAmount, outstandingAmount)
	}

	fmt.Println("")
	fmt.Println("=== STEP 2: CHECKING PAYMENT RECORDS ===")
	
	// Check if payments table exists and has records
	paymentTables := []string{"payments", "payment_records", "sales_payments", "invoice_payments"}
	
	for _, table := range paymentTables {
		var count int
		err = db.QueryRow(fmt.Sprintf(`
			SELECT COUNT(*) FROM information_schema.tables 
			WHERE table_name = '%s'`, table)).Scan(&count)
		
		if err == nil && count > 0 {
			fmt.Printf("🔍 Found table: %s\n", table)
			
			// Get records from this table
			var recordCount int
			err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&recordCount)
			if err == nil {
				fmt.Printf("   Records in %s: %d\n", table, recordCount)
				
				if recordCount > 0 {
					// Show recent records
					paymentQuery := fmt.Sprintf(`
						SELECT * FROM %s 
						ORDER BY created_at DESC LIMIT 3`, table)
					
					paymentRows, err := db.Query(paymentQuery)
					if err == nil {
						fmt.Printf("   Recent records:\n")
						cols, _ := paymentRows.Columns()
						fmt.Printf("   Columns: %v\n", cols)
						
						for paymentRows.Next() && paymentRows.Next() {
							values := make([]interface{}, len(cols))
							valuePtrs := make([]interface{}, len(cols))
							for i := range cols {
								valuePtrs[i] = &values[i]
							}
							
							paymentRows.Scan(valuePtrs...)
							
							fmt.Printf("   Row: ")
							for i, col := range cols {
								val := values[i]
								if val == nil {
									fmt.Printf("%s=NULL ", col)
								} else {
									switch v := val.(type) {
									case []byte:
										fmt.Printf("%s=%s ", col, string(v))
									default:
										fmt.Printf("%s=%v ", col, v)
									}
								}
							}
							fmt.Println("")
							break // Only show first record
						}
						paymentRows.Close()
					}
				}
			}
			fmt.Println("")
		}
	}

	fmt.Println("=== STEP 3: CHECKING JOURNAL ENTRIES FOR PAYMENTS ===")
	
	// Check for payment-related journal entries
	paymentJournalQuery := `
		SELECT ujl.id, ujl.account_id, a.code, a.name, ujl.debit_amount, ujl.credit_amount, 
		       ujl.description, ujl.created_at
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE (ujl.description ILIKE '%payment%' OR ujl.description ILIKE '%bayar%')
		   OR (a.code IN ('1104', '1201')) -- BANK UOB, Piutang Usaha
		ORDER BY ujl.created_at DESC
		LIMIT 10`
	
	rows, err = db.Query(paymentJournalQuery)
	if err != nil {
		log.Printf("Failed to get payment journals: %v", err)
	} else {
		fmt.Printf("📋 Recent payment-related journal entries:\n\n")
		fmt.Printf("%-4s | %-4s | %-25s | %12s | %12s | %-30s | %s\n", 
			"JID", "Code", "Account", "Debit", "Credit", "Description", "Date")
		fmt.Println("-----+------+--------------------------+-------------+-------------+------------------------------+---------")

		hasPaymentJournals := false
		for rows.Next() {
			hasPaymentJournals = true
			var id, accountID int
			var code, name, description, createdAt string
			var debit, credit float64
			
			err := rows.Scan(&id, &accountID, &code, &name, &debit, &credit, &description, &createdAt)
			if err != nil {
				log.Printf("Error scanning payment journal: %v", err)
				continue
			}
			
			fmt.Printf("%-4d | %-4s | %-25s | %12.2f | %12.2f | %-30s | %s\n",
				id, code, truncate(name, 25), debit, credit, 
				truncate(description, 30), createdAt[:10])
		}
		rows.Close()

		if !hasPaymentJournals {
			fmt.Println("❌ No payment-related journal entries found!")
		}
	}

	fmt.Println("")
	fmt.Println("=== STEP 4: ACCOUNT BALANCE ANALYSIS ===")
	
	// Check specific account balances
	accountCodes := []string{"1104", "1201"} // BANK UOB, Piutang Usaha
	accountNames := []string{"BANK UOB", "Piutang Usaha"}
	
	for i, code := range accountCodes {
		var balance float64
		var totalDebits, totalCredits float64
		var journalCount int
		
		err = db.QueryRow(`
			SELECT a.balance,
			       COALESCE(SUM(ujl.debit_amount), 0) as total_debits,
			       COALESCE(SUM(ujl.credit_amount), 0) as total_credits,
			       COUNT(ujl.id) as journal_count
			FROM accounts a
			LEFT JOIN unified_journal_lines ujl ON a.id = ujl.account_id
			WHERE a.code = $1
			GROUP BY a.id, a.balance`, code).Scan(&balance, &totalDebits, &totalCredits, &journalCount)
		
		if err != nil {
			log.Printf("Error checking account %s: %v", code, err)
			continue
		}
		
		expectedBalance := totalDebits - totalCredits
		if code == "1201" { // Piutang Usaha (Asset account)
			expectedBalance = totalDebits - totalCredits
		}
		
		status := "✅ OK"
		if balance != expectedBalance {
			status = "❌ MISMATCH"
		}
		
		fmt.Printf("🏦 %s (%s):\n", accountNames[i], code)
		fmt.Printf("   Current Balance: Rp %.2f\n", balance)
		fmt.Printf("   Total Debits: Rp %.2f\n", totalDebits)
		fmt.Printf("   Total Credits: Rp %.2f\n", totalCredits)
		fmt.Printf("   Expected Balance: Rp %.2f\n", expectedBalance)
		fmt.Printf("   Journal Entries: %d\n", journalCount)
		fmt.Printf("   Status: %s\n", status)
		fmt.Println("")
	}

	fmt.Println("=== STEP 5: DIAGNOSIS & RECOMMENDATIONS ===")
	fmt.Println("")
	fmt.Println("💡 POSSIBLE CAUSES:")
	fmt.Println("1. ❓ Payment journal entries not created properly")
	fmt.Println("2. ❓ Balance sync not triggered after payment")
	fmt.Println("3. ❓ Payment recorded in different table but not processed")
	fmt.Println("4. ❓ Frontend-backend sync issue")
	fmt.Println("")
	fmt.Println("🔧 RECOMMENDED ACTIONS:")
	fmt.Println("1. Check if payment was actually saved to database")
	fmt.Println("2. Verify payment processing logic creates journal entries") 
	fmt.Println("3. Run balance sync to update account balances")
	fmt.Println("4. Manual payment journal entry if needed")
	fmt.Println("")
	fmt.Println("🏁 INVESTIGATION COMPLETE!")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}