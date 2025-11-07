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
	fmt.Println("🧹 FIXING DUPLICATE PAYMENT JOURNAL ENTRIES")
	fmt.Println("Problem: Multiple payment records for same invoice causing balance errors")
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

	fmt.Println("=== STEP 1: ANALYZING PAYMENT RECORDS ===")
	
	// Check payment records
	query := `
		SELECT id, code, amount, reference, status, notes, created_at
		FROM payments 
		ORDER BY created_at DESC`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to get payment records:", err)
	}
	defer rows.Close()

	fmt.Printf("%-2s | %-15s | %12s | %-15s | %-10s | %-30s\n", 
		"ID", "Code", "Amount", "Reference", "Status", "Notes")
	fmt.Println("---+----------------+-------------+----------------+-----------+-------------------------------")

	for rows.Next() {
		var id int
		var code, reference, status, notes, createdAt string
		var amount float64
		
		err := rows.Scan(&id, &code, &amount, &reference, &status, &notes, &createdAt)
		if err != nil {
			log.Fatal("Failed to scan payment:", err)
		}
		
		fmt.Printf("%-2d | %-15s | %12.2f | %-15s | %-10s | %-30s\n",
			id, code, amount, reference, status, truncate(notes, 30))
	}

	fmt.Println("")
	fmt.Println("=== STEP 2: ANALYZING PIUTANG JOURNALS ===")
	
	// Check Piutang journal entries with payment references
	journalQuery := `
		SELECT ujl.id, ujl.debit_amount, ujl.credit_amount, ujl.description, ujl.created_at
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.code = '1201' 
		  AND (ujl.description ILIKE '%RCV%' OR ujl.description ILIKE '%reduction%')
		ORDER BY ujl.created_at DESC`
	
	rows, err = db.Query(journalQuery)
	if err != nil {
		log.Fatal("Failed to get payment journals:", err)
	}
	defer rows.Close()

	fmt.Printf("📋 Payment-related Piutang journal entries:\n\n")
	fmt.Printf("%-4s | %12s | %12s | %-40s | %s\n", 
		"JID", "Debit", "Credit", "Description", "Date")
	fmt.Println("-----+-------------+-------------+------------------------------------------+----------")

	type PaymentJournal struct {
		ID          int
		Debit       float64
		Credit      float64
		Description string
		CreatedAt   string
	}

	var paymentJournals []PaymentJournal
	for rows.Next() {
		var j PaymentJournal
		err := rows.Scan(&j.ID, &j.Debit, &j.Credit, &j.Description, &j.CreatedAt)
		if err != nil {
			log.Fatal("Failed to scan journal:", err)
		}
		
		fmt.Printf("%-4d | %12.2f | %12.2f | %-40s | %s\n",
			j.ID, j.Debit, j.Credit, truncate(j.Description, 40), j.CreatedAt[:10])
		
		paymentJournals = append(paymentJournals, j)
	}

	fmt.Println("")
	fmt.Printf("📊 Found %d payment-related journal entries\n", len(paymentJournals))

	// Look for duplicate payments (same amount)
	duplicates := findDuplicatePayments(paymentJournals)
	
	if len(duplicates) > 0 {
		fmt.Printf("🔍 Found %d potential duplicate payment entries:\n\n", len(duplicates))
		
		for i, dup := range duplicates {
			fmt.Printf("Duplicate Group %d (Amount: Rp %.2f):\n", i+1, dup[0].Credit)
			for _, j := range dup {
				fmt.Printf("  - JID %d: %s [%s]\n", j.ID, truncate(j.Description, 50), j.CreatedAt[:16])
			}
			fmt.Println()
		}
		
		fmt.Print("❓ Do you want to remove duplicate payment entries? (y/n): ")
		var response string
		fmt.Scanln(&response)
		
		if response == "y" || response == "Y" {
			removeDuplicates(db, duplicates)
		} else {
			fmt.Println("❌ Cleanup cancelled by user")
		}
	} else {
		fmt.Println("✅ No obvious duplicate payment entries found")
		
		// Check for single payment that should be removed
		fmt.Println("")
		fmt.Println("🔍 Checking payment logic...")
		
		// Count payments for each invoice
		invoicePayments := make(map[string][]PaymentJournal)
		for _, j := range paymentJournals {
			// Extract invoice reference from description
			invoiceRef := extractInvoiceRef(j.Description)
			if invoiceRef != "" {
				invoicePayments[invoiceRef] = append(invoicePayments[invoiceRef], j)
			}
		}
		
		fmt.Printf("📊 Payments per invoice:\n")
		for invoice, payments := range invoicePayments {
			fmt.Printf("   %s: %d payments\n", invoice, len(payments))
			if len(payments) > 1 {
				fmt.Printf("     ⚠️ Multiple payments detected for %s\n", invoice)
			}
		}
	}
	
	fmt.Println("")
	fmt.Println("🏁 ANALYSIS COMPLETE!")
}

func findDuplicatePayments(journals []PaymentJournal) [][]PaymentJournal {
	var duplicates [][]PaymentJournal
	processed := make(map[int]bool)
	
	for i := 0; i < len(journals); i++ {
		if processed[i] || journals[i].Credit == 0 {
			continue
		}
		
		group := []PaymentJournal{journals[i]}
		processed[i] = true
		
		for j := i + 1; j < len(journals); j++ {
			if processed[j] {
				continue
			}
			
			if journals[j].Credit == journals[i].Credit {
				group = append(group, journals[j])
				processed[j] = true
			}
		}
		
		if len(group) > 1 {
			duplicates = append(duplicates, group)
		}
	}
	
	return duplicates
}

func extractInvoiceRef(description string) string {
	// Simple extraction of invoice reference from description
	if len(description) > 10 {
		// Look for pattern like "RCV/2025/09/0005" or "INV/2025/09/0002"
		for i := 0; i < len(description)-10; i++ {
			if description[i:i+4] == "RCV/" || description[i:i+4] == "INV/" {
				end := i + 4
				for end < len(description) && (description[end] >= '0' && description[end] <= '9' || description[end] == '/') {
					end++
				}
				if end > i+10 {
					return description[i:end]
				}
			}
		}
	}
	return ""
}

func removeDuplicates(db *sql.DB, duplicates [][]PaymentJournal) {
	fmt.Println("")
	fmt.Println("=== REMOVING DUPLICATE PAYMENTS ===")
	
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
				fmt.Printf("🗑️  Removing duplicate payment JID %d (Rp %.2f)\n", 
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
	
	// Recalculate Piutang balance
	var newBalance float64
	err = db.QueryRow(`
		SELECT COALESCE(SUM(debit_amount) - SUM(credit_amount), 0)
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.code = '1201'`).Scan(&newBalance)
	if err != nil {
		log.Fatal("Failed to calculate new balance:", err)
	}
	
	// Update account balance
	_, err = db.Exec("UPDATE accounts SET balance = $1 WHERE code = '1201'", newBalance)
	if err != nil {
		log.Fatal("Failed to update account balance:", err)
	}
	
	// Re-enable trigger
	_, err = db.Exec("ALTER TABLE unified_journal_lines ENABLE TRIGGER balance_sync_trigger")
	if err != nil {
		log.Printf("Warning: Could not re-enable trigger: %v", err)
	}
	
	fmt.Printf("✅ Removed Rp %.2f in duplicate payments\n", totalRemoved)
	fmt.Printf("✅ New Piutang Usaha balance: Rp %.2f\n", newBalance)
	
	fmt.Println("")
	fmt.Println("🎉 Duplicate payment cleanup completed!")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}