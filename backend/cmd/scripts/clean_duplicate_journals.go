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

func maskPassword(dbURL string) string {
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

type JournalEntry struct {
	ID            int
	Debit         float64
	Credit        float64
	Description   string
	CreatedAt     string
	TransactionID sql.NullInt64
}

func main() {
	fmt.Println("🔍 IDENTIFYING & CLEANING DUPLICATE JOURNAL ENTRIES")
	fmt.Println("Expected: 2 sales = 2 journal entries = Rp 10,000,000")
	fmt.Println("Current: 3 journal entries = Rp 15,000,000")
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

	fmt.Println("=== STEP 1: DETAILED JOURNAL ANALYSIS ===")
	
	// Check all journal entries for Pendapatan Penjualan account
	query := `
		SELECT ujl.id, ujl.debit_amount, ujl.credit_amount, ujl.description, 
		       ujl.created_at, ujl.transaction_id
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.code = '4101'
		ORDER BY ujl.created_at DESC`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to get journal entries:", err)
	}
	defer rows.Close()

	fmt.Printf("%-4s | %12s | %12s | %-40s | %-15s | %-8s\n", 
		"JID", "Debit", "Credit", "Description", "Created", "TxID")
	fmt.Println("-----+-------------+-------------+------------------------------------------+----------------+---------")

	var journals []JournalEntry
	for rows.Next() {
		var j JournalEntry
		err := rows.Scan(&j.ID, &j.Debit, &j.Credit, &j.Description, &j.CreatedAt, &j.TransactionID)
		if err != nil {
			log.Fatal("Failed to scan journal row:", err)
		}
		
		txID := "NULL"
		if j.TransactionID.Valid {
			txID = fmt.Sprintf("%d", j.TransactionID.Int64)
		}
		
		fmt.Printf("%-4d | %12.2f | %12.2f | %-40s | %-15s | %-8s\n",
			j.ID, j.Debit, j.Credit, 
			truncate(j.Description, 40), 
			j.CreatedAt[:16], txID)
		
		journals = append(journals, j)
	}

	fmt.Println("")
	fmt.Printf("📊 Total Journal Entries Found: %d\n", len(journals))
	fmt.Printf("💰 Total Credit Amount: Rp %.2f\n", calculateTotalCredits(journals))
	
	fmt.Println("")
	fmt.Println("=== STEP 2: CHECKING SALES TRANSACTIONS ===")
	
	// Check sales transactions
	salesQuery := `
		SELECT id, code, invoice_number, subtotal, status, created_at
		FROM sales 
		ORDER BY created_at DESC`
	
	rows, err = db.Query(salesQuery)
	if err != nil {
		log.Fatal("Failed to get sales:", err)
	}
	defer rows.Close()

	fmt.Printf("%-4s | %-12s | %-15s | %12s | %-10s | %-15s\n", 
		"ID", "Code", "Invoice", "Subtotal", "Status", "Created")
	fmt.Println("-----+-------------+----------------+-------------+-----------+----------------")

	var totalSalesRevenue float64
	salesCount := 0
	for rows.Next() {
		var id int
		var code, invoice, status, createdAt string
		var subtotal float64
		
		err := rows.Scan(&id, &code, &invoice, &subtotal, &status, &createdAt)
		if err != nil {
			log.Fatal("Failed to scan sales row:", err)
		}
		
		fmt.Printf("%-4d | %-12s | %-15s | %12.2f | %-10s | %-15s\n",
			id, code, invoice, subtotal, status, createdAt[:16])
		
		totalSalesRevenue += subtotal
		salesCount++
	}

	fmt.Println("")
	fmt.Printf("📊 Total Sales: %d\n", salesCount)
	fmt.Printf("💰 Expected Revenue from Sales: Rp %.2f\n", totalSalesRevenue)
	
	fmt.Println("")
	fmt.Println("=== STEP 3: IDENTIFYING DUPLICATES ===")
	
	// Look for potential duplicates
	duplicates := findDuplicates(journals)
	
	if len(duplicates) > 0 {
		fmt.Printf("🔍 Found %d potential duplicate journal entries:\n", len(duplicates))
		fmt.Println("")
		
		for i, dup := range duplicates {
			fmt.Printf("Duplicate Group %d:\n", i+1)
			for _, j := range dup {
				fmt.Printf("  - JID %d: Rp %.2f, %s\n", j.ID, j.Credit, truncate(j.Description, 50))
			}
			fmt.Println("")
		}
		
		fmt.Print("❓ Do you want to clean the duplicate entries? (y/n): ")
		var response string
		fmt.Scanln(&response)
		
		if response == "y" || response == "Y" {
			cleanDuplicates(db, duplicates, totalSalesRevenue)
		} else {
			fmt.Println("❌ Cleanup cancelled by user")
		}
	} else {
		fmt.Println("✅ No obvious duplicates found")
		fmt.Println("💡 The extra Rp 5,000,000 might be from a legitimate transaction")
		
		// Check for transfer entries that might cause confusion
		fmt.Println("")
		fmt.Println("🔍 Checking for transfer/adjustment entries...")
		
		for _, j := range journals {
			if containsTransferKeywords(j.Description) {
				fmt.Printf("⚠️  Transfer entry found: JID %d - %s (Rp %.2f)\n", 
					j.ID, truncate(j.Description, 50), j.Credit)
			}
		}
	}
	
	fmt.Println("")
	fmt.Println("🏁 ANALYSIS COMPLETE!")
}

func calculateTotalCredits(journals []JournalEntry) float64 {
	total := 0.0
	for _, j := range journals {
		total += j.Credit
	}
	return total
}

func findDuplicates(journals []JournalEntry) [][]JournalEntry {
	var duplicates [][]JournalEntry
	
	// Simple duplicate detection based on credit amount and similar description
	for i := 0; i < len(journals); i++ {
		for j := i + 1; j < len(journals); j++ {
			if journals[i].Credit == journals[j].Credit && 
			   journals[i].Credit > 0 &&
			   similarDescription(journals[i].Description, journals[j].Description) {
				duplicates = append(duplicates, []JournalEntry{journals[i], journals[j]})
			}
		}
	}
	
	return duplicates
}

func similarDescription(desc1, desc2 string) bool {
	// Simple similarity check
	return len(desc1) > 10 && len(desc2) > 10 && 
		   (desc1[:10] == desc2[:10] || 
		    containsTransferKeywords(desc1) || containsTransferKeywords(desc2))
}

func containsTransferKeywords(desc string) bool {
	keywords := []string{"transfer", "moved", "adjustment", "correction"}
	descLower := fmt.Sprintf("%s", desc)
	for _, keyword := range keywords {
		if contains(descLower, keyword) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   fmt.Sprintf("%s", s) != fmt.Sprintf("%s", s[:len(s)-len(substr)]) || 
		   fmt.Sprintf("%s", s) == substr
}

func cleanDuplicates(db *sql.DB, duplicates [][]JournalEntry, expectedTotal float64) {
	fmt.Println("")
	fmt.Println("=== CLEANING DUPLICATE ENTRIES ===")
	
	// Disable trigger
	_, err := db.Exec("ALTER TABLE unified_journal_lines DISABLE TRIGGER balance_sync_trigger")
	if err != nil {
		log.Printf("Warning: Could not disable trigger: %v", err)
	}
	
	var totalRemoved float64
	for _, dupGroup := range duplicates {
		if len(dupGroup) >= 2 {
			// Keep the first (oldest) entry, remove others
			for i := 1; i < len(dupGroup); i++ {
				fmt.Printf("🗑️  Removing duplicate JID %d (Rp %.2f)\n", 
					dupGroup[i].ID, dupGroup[i].Credit)
				
				_, err := db.Exec("DELETE FROM unified_journal_lines WHERE id = $1", dupGroup[i].ID)
				if err != nil {
					log.Printf("Error removing duplicate %d: %v", dupGroup[i].ID, err)
				} else {
					totalRemoved += dupGroup[i].Credit
				}
			}
		}
	}
	
	// Recalculate balance
	var newBalance float64
	err = db.QueryRow(`
		SELECT COALESCE(SUM(credit_amount) - SUM(debit_amount), 0)
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.code = '4101'`).Scan(&newBalance)
	if err != nil {
		log.Fatal("Failed to calculate new balance:", err)
	}
	
	// Update account balance
	_, err = db.Exec("UPDATE accounts SET balance = $1 WHERE code = '4101'", newBalance)
	if err != nil {
		log.Fatal("Failed to update account balance:", err)
	}
	
	// Re-enable trigger
	_, err = db.Exec("ALTER TABLE unified_journal_lines ENABLE TRIGGER balance_sync_trigger")
	if err != nil {
		log.Printf("Warning: Could not re-enable trigger: %v", err)
	}
	
	fmt.Printf("✅ Removed Rp %.2f in duplicates\n", totalRemoved)
	fmt.Printf("✅ New Pendapatan Penjualan balance: Rp %.2f\n", newBalance)
	
	if newBalance == expectedTotal {
		fmt.Println("🎉 Balance now matches expected sales revenue!")
	} else {
		fmt.Printf("⚠️  Balance (Rp %.2f) still differs from expected (Rp %.2f)\n", 
			newBalance, expectedTotal)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}