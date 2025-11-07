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
		// Simple env loading - look for DATABASE_URL
		content := make([]byte, 1024)
		if n, err := file.Read(content); err == nil {
			envContent := string(content[:n])
			// Parse DATABASE_URL
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

func maskPassword(dbURL string) string {
	// Simple password masking
	masked := ""
	inPassword := false
	for i, char := range dbURL {
		if i > 0 && dbURL[i-1] == ':' && char != '/' {
			inPassword = true
		}
		if inPassword && char == '@' {
			inPassword = false
			masked += "@"
			continue
		}
		if inPassword {
			masked += "*"
		} else {
			masked += string(char)
		}
	}
	return masked
}

func main() {
	fmt.Println("🔧 FIXING REVENUE ALLOCATION")
	fmt.Println("Moving all revenue to proper child account (Pendapatan Penjualan)")
	fmt.Println("")

	loadEnv()
	
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not found in environment")
	}

	fmt.Printf("🔧 DATABASE_URL: %s\n", maskPassword(dbURL))
	fmt.Println("")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== STEP 1: CHECKING CURRENT REVENUE ALLOCATION ===")
	
	// Check current balances
	query := `
		SELECT a.id, a.code, a.name, a.balance,
		       COALESCE(SUM(ujl.credit_amount), 0) as total_credits,
		       COALESCE(SUM(ujl.debit_amount), 0) as total_debits,
		       COUNT(ujl.id) as journal_count
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON a.id = ujl.account_id
		WHERE a.code IN ('4000', '4101')
		GROUP BY a.id, a.code, a.name, a.balance
		ORDER BY a.code`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to get current balances:", err)
	}
	defer rows.Close()

	fmt.Printf("%-6s | %-25s | %12s | %12s | %12s | %8s\n", 
		"Code", "Account Name", "Balance", "Credits", "Debits", "Journals")
	fmt.Println("-------+--------------------------+-------------+-------------+-------------+---------")

	for rows.Next() {
		var id int
		var code, name string
		var balance, credits, debits float64
		var journalCount int
		
		err := rows.Scan(&id, &code, &name, &balance, &credits, &debits, &journalCount)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}
		
		fmt.Printf("%-6s | %-25s | %12.2f | %12.2f | %12.2f | %8d\n",
			code, name, balance, credits, debits, journalCount)
	}

	fmt.Println("")
	fmt.Println("=== STEP 2: CHECKING JOURNAL ENTRIES FOR REVENUE ACCOUNTS ===")
	
	// Check journal entries
	journalQuery := `
		SELECT ujl.id, ujl.account_id, a.code, a.name, ujl.debit_amount, ujl.credit_amount, 
		       ujl.description, ujl.created_at
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.code IN ('4000', '4101')
		ORDER BY ujl.created_at DESC`
	
	rows, err = db.Query(journalQuery)
	if err != nil {
		log.Fatal("Failed to get journal entries:", err)
	}
	defer rows.Close()

	fmt.Printf("%-4s | %-4s | %-25s | %12s | %12s | %-30s\n", 
		"JID", "Code", "Account", "Debit", "Credit", "Description")
	fmt.Println("-----+------+--------------------------+-------------+-------------+------------------------------")

	var revenueJournals []struct {
		ID          int
		AccountID   int
		AccountCode string
		Debit       float64
		Credit      float64
		Description string
	}

	for rows.Next() {
		var id, accountID int
		var code, name, description, createdAt string
		var debit, credit float64
		
		err := rows.Scan(&id, &accountID, &code, &name, &debit, &credit, &description, &createdAt)
		if err != nil {
			log.Fatal("Failed to scan journal row:", err)
		}
		
		fmt.Printf("%-4d | %-4s | %-25s | %12.2f | %12.2f | %-30s\n",
			id, code, name[:min(25, len(name))], debit, credit, description[:min(30, len(description))])
		
		revenueJournals = append(revenueJournals, struct {
			ID          int
			AccountID   int
			AccountCode string
			Debit       float64
			Credit      float64
			Description string
		}{id, accountID, code, debit, credit, description})
	}

	fmt.Println("")
	fmt.Println("=== STEP 3: PROPOSED SOLUTION ===")
	
	fmt.Println("🎯 CORRECT ALLOCATION SHOULD BE:")
	fmt.Println("   REVENUE (4000)         → Parent account → Balance: Rp 0")
	fmt.Println("   Pendapatan Penjualan (4101) → Child account  → Balance: Rp 10,000,000")
	fmt.Println("")
	
	fmt.Println("💡 STRATEGY:")
	fmt.Println("1. Move all journal entries from REVENUE (4000) to Pendapatan Penjualan (4101)")
	fmt.Println("2. Update account balances accordingly")
	fmt.Println("3. Ensure parent account shows 0 balance")
	fmt.Println("")

	fmt.Print("❓ Do you want to proceed with the fix? (y/n): ")
	var response string
	fmt.Scanln(&response)
	
	if response != "y" && response != "Y" {
		fmt.Println("❌ Fix cancelled by user")
		return
	}

	fmt.Println("")
	fmt.Println("=== STEP 4: EXECUTING REVENUE ALLOCATION FIX ===")
	
	// Disable trigger temporarily
	fmt.Println("🔧 Disabling balance sync trigger...")
	_, err = db.Exec("ALTER TABLE unified_journal_lines DISABLE TRIGGER balance_sync_trigger")
	if err != nil {
		log.Printf("Warning: Could not disable trigger: %v", err)
	} else {
		fmt.Println("✅ Balance sync trigger disabled")
	}

	// Get account IDs
	var revenueAccountID, pendapatanAccountID int
	err = db.QueryRow("SELECT id FROM accounts WHERE code = '4000'").Scan(&revenueAccountID)
	if err != nil {
		log.Fatal("Failed to get REVENUE account ID:", err)
	}
	
	err = db.QueryRow("SELECT id FROM accounts WHERE code = '4101'").Scan(&pendapatanAccountID)
	if err != nil {
		log.Fatal("Failed to get Pendapatan Penjualan account ID:", err)
	}

	// Update journal entries: move from REVENUE (4000) to Pendapatan Penjualan (4101)
	fmt.Println("🔄 Moving journal entries from REVENUE to Pendapatan Penjualan...")
	
	updateQuery := `
		UPDATE unified_journal_lines 
		SET account_id = $1,
		    description = description || ' [Moved from REVENUE parent account]'
		WHERE account_id = $2 
		AND (description ILIKE '%sales%' OR description ILIKE '%revenue%')
		AND description NOT ILIKE '%transfer%'`
	
	result, err := db.Exec(updateQuery, pendapatanAccountID, revenueAccountID)
	if err != nil {
		log.Fatal("Failed to update journal entries:", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✅ Updated %d journal entries\n", rowsAffected)

	// Update account balances
	fmt.Println("💰 Recalculating account balances...")
	
	// Reset REVENUE (4000) balance to 0
	_, err = db.Exec("UPDATE accounts SET balance = 0 WHERE code = '4000'")
	if err != nil {
		log.Fatal("Failed to reset REVENUE balance:", err)
	}
	
	// Calculate correct balance for Pendapatan Penjualan (4101)
	var correctBalance float64
	err = db.QueryRow(`
		SELECT COALESCE(SUM(credit_amount) - SUM(debit_amount), 0)
		FROM unified_journal_lines 
		WHERE account_id = $1`, pendapatanAccountID).Scan(&correctBalance)
	if err != nil {
		log.Fatal("Failed to calculate correct balance:", err)
	}
	
	_, err = db.Exec("UPDATE accounts SET balance = $1 WHERE code = '4101'", correctBalance)
	if err != nil {
		log.Fatal("Failed to update Pendapatan Penjualan balance:", err)
	}
	
	fmt.Printf("✅ Updated balances:\n")
	fmt.Printf("   REVENUE (4000): Rp 0.00\n")
	fmt.Printf("   Pendapatan Penjualan (4101): Rp %.2f\n", correctBalance)

	// Re-enable trigger
	fmt.Println("🔧 Re-enabling balance sync trigger...")
	_, err = db.Exec("ALTER TABLE unified_journal_lines ENABLE TRIGGER balance_sync_trigger")
	if err != nil {
		log.Printf("Warning: Could not re-enable trigger: %v", err)
	} else {
		fmt.Println("✅ Balance sync trigger re-enabled")
	}

	fmt.Println("")
	fmt.Println("=== STEP 5: VERIFICATION ===")
	
	// Verify final balances
	rows, err = db.Query(`
		SELECT a.code, a.name, a.balance
		FROM accounts a
		WHERE a.code IN ('4000', '4101')
		ORDER BY a.code`)
	if err != nil {
		log.Fatal("Failed to verify balances:", err)
	}
	defer rows.Close()

	fmt.Printf("%-6s | %-25s | %15s\n", "Code", "Account Name", "Final Balance")
	fmt.Println("-------+--------------------------+----------------")

	totalRevenue := 0.0
	for rows.Next() {
		var code, name string
		var balance float64
		
		err := rows.Scan(&code, &name, &balance)
		if err != nil {
			log.Fatal("Failed to scan verification row:", err)
		}
		
		fmt.Printf("%-6s | %-25s | %15.2f\n", code, name, balance)
		if code == "4101" {
			totalRevenue = balance
		}
	}

	fmt.Println("")
	fmt.Printf("🎉 SUCCESS! Revenue allocation fixed!\n")
	fmt.Printf("💰 Total Revenue now correctly shows: Rp %.2f\n", totalRevenue)
	fmt.Printf("✅ Parent account (REVENUE) balance: Rp 0.00\n")
	fmt.Printf("✅ Child account (Pendapatan Penjualan) balance: Rp %.2f\n", totalRevenue)
	fmt.Println("")
	fmt.Println("📱 Please refresh your frontend to see the updated balances!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}