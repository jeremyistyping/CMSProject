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
	fmt.Println("🧹 CLEANING EXTRA JOURNAL ENTRIES")
	fmt.Println("Problem: 3 journal entries (Rp 15M) but should be 2 (Rp 10M)")
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

	fmt.Println("=== STEP 1: ANALYZING PROBLEM JOURNALS ===")
	
	// Get all journal entries for Pendapatan Penjualan
	query := `
		SELECT ujl.id, ujl.credit_amount, ujl.description, ujl.created_at
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.code = '4101' AND ujl.credit_amount > 0
		ORDER BY ujl.created_at ASC`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to get journal entries:", err)
	}
	defer rows.Close()

	fmt.Printf("%-4s | %12s | %-40s | %-15s\n", 
		"JID", "Credit", "Description", "Created")
	fmt.Println("-----+-------------+------------------------------------------+----------------")

	var journalEntries []struct {
		ID          int
		Credit      float64
		Description string
		CreatedAt   string
	}

	for rows.Next() {
		var id int
		var credit float64
		var description, createdAt string
		
		err := rows.Scan(&id, &credit, &description, &createdAt)
		if err != nil {
			log.Fatal("Failed to scan journal row:", err)
		}
		
		fmt.Printf("%-4d | %12.2f | %-40s | %-15s\n",
			id, credit, truncate(description, 40), createdAt[:16])
		
		journalEntries = append(journalEntries, struct {
			ID          int
			Credit      float64
			Description string
			CreatedAt   string
		}{id, credit, description, createdAt})
	}

	fmt.Printf("\n📊 Total Journal Entries: %d\n", len(journalEntries))
	
	totalCredit := 0.0
	for _, j := range journalEntries {
		totalCredit += j.Credit
	}
	fmt.Printf("💰 Total Credit Amount: Rp %.2f\n", totalCredit)
	fmt.Printf("🎯 Target Credit Amount: Rp 10,000,000.00\n")
	fmt.Printf("📉 Excess Amount: Rp %.2f\n", totalCredit - 10000000)
	fmt.Println("")

	// Also check debit entry in REVENUE account
	fmt.Println("=== STEP 2: CHECKING REVENUE ACCOUNT DEBIT ===")
	
	var revenueDebitID int
	var revenueDebit float64
	var revenueDesc string
	err = db.QueryRow(`
		SELECT ujl.id, ujl.debit_amount, ujl.description
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.code = '4000' AND ujl.debit_amount > 0
		LIMIT 1`).Scan(&revenueDebitID, &revenueDebit, &revenueDesc)
	
	if err == nil {
		fmt.Printf("Found debit entry in REVENUE account:\n")
		fmt.Printf("  JID %d: Rp %.2f - %s\n", revenueDebitID, revenueDebit, truncate(revenueDesc, 50))
		fmt.Println("")
	} else if err != sql.ErrNoRows {
		log.Fatal("Error checking REVENUE debit:", err)
	}

	fmt.Println("=== STEP 3: PROPOSED SOLUTION ===")
	fmt.Println("🎯 Keep only 2 legitimate sales journal entries (Rp 5M each)")
	fmt.Println("🗑️  Remove the extra/transfer journal entry")
	fmt.Println("🗑️  Remove the debit entry from REVENUE account")
	fmt.Println("")

	if len(journalEntries) < 2 {
		fmt.Println("❌ Not enough journal entries found. Cannot proceed.")
		return
	}

	// Find which entry to remove (likely the one that's moved/transferred)
	var removeID = -1
	for _, j := range journalEntries {
		if containsTransferWords(j.Description) {
			removeID = j.ID
			fmt.Printf("🗑️  Will remove transfer entry: JID %d - %s\n", j.ID, truncate(j.Description, 50))
			break
		}
	}

	if removeID == -1 && len(journalEntries) > 2 {
		// Remove the middle one (chronologically)
		removeID = journalEntries[1].ID
		fmt.Printf("🗑️  Will remove middle entry: JID %d - %s\n", removeID, truncate(journalEntries[1].Description, 50))
	}

	fmt.Print("\n❓ Do you want to proceed with cleaning? (y/n): ")
	var response string
	fmt.Scanln(&response)
	
	if response != "y" && response != "Y" {
		fmt.Println("❌ Cleaning cancelled by user")
		return
	}

	fmt.Println("")
	fmt.Println("=== STEP 4: EXECUTING CLEANUP ===")

	// Disable trigger temporarily
	fmt.Println("🔧 Disabling balance sync trigger...")
	_, err = db.Exec("ALTER TABLE unified_journal_lines DISABLE TRIGGER balance_sync_trigger")
	if err != nil {
		log.Printf("Warning: Could not disable trigger: %v", err)
	}

	// Remove extra journal entry
	if removeID != -1 {
		fmt.Printf("🗑️  Removing journal entry %d...\n", removeID)
		_, err = db.Exec("DELETE FROM unified_journal_lines WHERE id = $1", removeID)
		if err != nil {
			log.Fatal("Failed to remove extra journal entry:", err)
		}
		fmt.Println("✅ Extra journal entry removed")
	}

	// Remove debit entry from REVENUE account if exists
	if revenueDebitID != 0 {
		fmt.Printf("🗑️  Removing REVENUE debit entry %d...\n", revenueDebitID)
		_, err = db.Exec("DELETE FROM unified_journal_lines WHERE id = $1", revenueDebitID)
		if err != nil {
			log.Fatal("Failed to remove REVENUE debit entry:", err)
		}
		fmt.Println("✅ REVENUE debit entry removed")
	}

	// Re-enable trigger
	fmt.Println("🔧 Re-enabling balance sync trigger...")
	_, err = db.Exec("ALTER TABLE unified_journal_lines ENABLE TRIGGER balance_sync_trigger")
	if err != nil {
		log.Printf("Warning: Could not re-enable trigger: %v", err)
	}

	fmt.Println("")
	fmt.Println("=== STEP 5: VERIFICATION ===")
	
	// Check final journal count
	var finalCount int
	var finalTotal float64
	err = db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(ujl.credit_amount), 0)
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.code = '4101' AND ujl.credit_amount > 0`).Scan(&finalCount, &finalTotal)
	if err != nil {
		log.Fatal("Failed to verify final state:", err)
	}

	fmt.Printf("✅ Final journal entries: %d\n", finalCount)
	fmt.Printf("✅ Final total credits: Rp %.2f\n", finalTotal)

	// Check account balances
	var revenueBalance, pendapatanBalance float64
	db.QueryRow("SELECT balance FROM accounts WHERE code = '4000'").Scan(&revenueBalance)
	db.QueryRow("SELECT balance FROM accounts WHERE code = '4101'").Scan(&pendapatanBalance)

	fmt.Printf("✅ REVENUE balance: Rp %.2f\n", revenueBalance)
	fmt.Printf("✅ Pendapatan Penjualan balance: Rp %.2f\n", pendapatanBalance)

	if finalCount == 2 && finalTotal == 10000000.00 && revenueBalance == 0.00 && pendapatanBalance == 10000000.00 {
		fmt.Println("")
		fmt.Println("🎉 PERFECT! Journal entries and balances are now correct!")
		fmt.Println("✅ 2 journal entries totaling Rp 10,000,000")
		fmt.Println("✅ REVENUE: Rp 0 (parent)")
		fmt.Println("✅ Pendapatan Penjualan: Rp 10,000,000 (child)")
		fmt.Println("")
		fmt.Println("📱 Please refresh your frontend to see the correct balances!")
	} else {
		fmt.Println("")
		fmt.Println("⚠️ Some discrepancies may still exist")
		fmt.Printf("   Expected: 2 entries, Rp 10M total\n")
		fmt.Printf("   Actual: %d entries, Rp %.0f total\n", finalCount, finalTotal)
	}
}

func containsTransferWords(desc string) bool {
	transferWords := []string{"transfer", "moved", "Transfer", "Moved"}
	for _, word := range transferWords {
		if len(desc) >= len(word) {
			for i := 0; i <= len(desc)-len(word); i++ {
				if desc[i:i+len(word)] == word {
					return true
				}
			}
		}
	}
	return false
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}