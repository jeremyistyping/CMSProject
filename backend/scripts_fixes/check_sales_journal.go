package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Initialize database connection
	db := database.ConnectDB()

	fmt.Println("=== Sales and Journal Entry Analysis ===")
	fmt.Println()

	// Check recent sales transactions
	fmt.Println("1. Recent Sales Transactions:")
	var sales []models.Sale
	err := db.Preload("Customer").Where("created_at > ?", time.Now().Add(-24*time.Hour)).Order("created_at desc").Limit(10).Find(&sales).Error
	if err != nil {
		log.Printf("Error fetching sales: %v", err)
	} else {
		for _, sale := range sales {
			customerName := "Unknown"
			if sale.Customer.Name != "" {
				customerName = sale.Customer.Name
			}
			fmt.Printf("  Sale ID: %d, Invoice: %s, Customer: %s, Total: %.2f, Status: %s, Date: %s\n", 
				sale.ID, sale.InvoiceNumber, customerName, sale.TotalAmount, sale.Status, sale.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Println()

	// Check recent journal entries (SSOT)
	fmt.Println("2. Recent SSOT Journal Entries:")
	type SSOTJournal struct {
		ID          uint      `json:"id"`
		EntryNumber string    `json:"entry_number"`
		SourceType  string    `json:"source_type"`
		SourceID    *uint     `json:"source_id"`
		TotalDebit  float64   `json:"total_debit"`
		TotalCredit float64   `json:"total_credit"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
	}

	var ssotJournals []SSOTJournal
	err = db.Raw(`
		SELECT id, entry_number, source_type, source_id, total_debit, total_credit, status, created_at
		FROM unified_journal_ledger 
		WHERE created_at > ? 
		ORDER BY created_at DESC 
		LIMIT 10
	`, time.Now().Add(-24*time.Hour)).Scan(&ssotJournals).Error
	if err != nil {
		log.Printf("Error fetching SSOT journals: %v", err)
	} else {
		for _, journal := range ssotJournals {
			sourceID := "N/A"
			if journal.SourceID != nil {
				sourceID = fmt.Sprintf("%d", *journal.SourceID)
			}
			fmt.Printf("  SSOT Journal ID: %d, Entry: %s, Source: %s(%s), Debit: %.2f, Credit: %.2f, Status: %s, Date: %s\n", 
				journal.ID, journal.EntryNumber, journal.SourceType, sourceID, journal.TotalDebit, journal.TotalCredit, journal.Status, journal.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Println()

	// Check recent SSOT journal lines
	fmt.Println("3. Recent SSOT Journal Lines:")
	type SSOTJournalLine struct {
		ID          uint    `json:"id"`
		JournalID   uint    `json:"journal_id"`
		AccountID   uint    `json:"account_id"`
		Description string  `json:"description"`
		DebitAmount float64 `json:"debit_amount"`
		CreditAmount float64 `json:"credit_amount"`
	}

	var ssotLines []SSOTJournalLine
	err = db.Raw(`
		SELECT ujl.id, ujl.journal_id, ujl.account_id, ujl.description, ujl.debit_amount, ujl.credit_amount
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
		WHERE uj.created_at > ?
		ORDER BY ujl.id DESC
		LIMIT 20
	`, time.Now().Add(-24*time.Hour)).Scan(&ssotLines).Error
	if err != nil {
		log.Printf("Error fetching SSOT journal lines: %v", err)
	} else {
		for _, line := range ssotLines {
			fmt.Printf("  SSOT Line ID: %d, Journal ID: %d, Account: %d, Debit: %.2f, Credit: %.2f, Desc: %s\n", 
				line.ID, line.JournalID, line.AccountID, line.DebitAmount, line.CreditAmount, line.Description)
		}
	}
	fmt.Println()

	// Check account balances for key accounts
	fmt.Println("4. Key Account Balances:")
	var accounts []models.Account
	keyCodes := []string{"1201", "4101", "2103", "2102"} // AR, Revenue, PPN Payable
	err = db.Where("code IN ?", keyCodes).Find(&accounts).Error
	if err != nil {
		log.Printf("Error fetching accounts: %v", err)
	} else {
		for _, account := range accounts {
			fmt.Printf("  Account %s (%s): Balance = %.2f\n", account.Code, account.Name, account.Balance)
		}
	}
	fmt.Println()

	// Check for sales invoice INV/2025/09/0002
	fmt.Println("5. Checking specific invoice INV/2025/09/0002:")
	var specificSale models.Sale
	err = db.Preload("Customer").Where("invoice_number = ?", "INV/2025/09/0002").First(&specificSale).Error
	if err != nil {
		log.Printf("Error finding specific sale: %v", err)
	} else {
		fmt.Printf("  Found sale: ID=%d, Total=%.2f, Status=%s\n", specificSale.ID, specificSale.TotalAmount, specificSale.Status)
		
		// Check SSOT journals for this sale
		var relatedSSotJournals []SSOTJournal
		err = db.Raw(`
			SELECT id, entry_number, source_type, source_id, total_debit, total_credit, status, created_at
			FROM unified_journal_ledger 
			WHERE source_type = ? AND source_id = ?
		`, "sale", specificSale.ID).Scan(&relatedSSotJournals).Error
		if err != nil {
			log.Printf("  Error finding related SSOT journals: %v", err)
		} else {
			fmt.Printf("  Found %d related SSOT journals:\n", len(relatedSSotJournals))
			for _, journal := range relatedSSotJournals {
				fmt.Printf("    SSOT Journal ID: %d, Entry: %s, Debit: %.2f, Credit: %.2f, Status: %s\n", 
					journal.ID, journal.EntryNumber, journal.TotalDebit, journal.TotalCredit, journal.Status)
				
				// Get journal lines for this journal
				var lines []SSOTJournalLine
				err = db.Raw(`
					SELECT id, journal_id, account_id, description, debit_amount, credit_amount
					FROM unified_journal_lines 
					WHERE journal_id = ?
				`, journal.ID).Scan(&lines).Error
				if err != nil {
					log.Printf("    Error getting SSOT journal lines: %v", err)
				} else {
					for _, line := range lines {
						fmt.Printf("      SSOT Line: Account %d, Debit: %.2f, Credit: %.2f, Desc: %s\n", 
							line.AccountID, line.DebitAmount, line.CreditAmount, line.Description)
					}
				}
			}
		}
		
		// Also check old journal system if it exists
		fmt.Println("  Checking old journal system:")
		var legacyJournals []models.Journal
		err = db.Where("reference_type = ? AND reference_id = ?", "SALE", specificSale.ID).Find(&legacyJournals).Error
		if err != nil {
			log.Printf("  Error finding legacy journals: %v", err)
		} else {
			fmt.Printf("  Found %d legacy journals\n", len(legacyJournals))
			for _, journal := range legacyJournals {
				fmt.Printf("    Legacy Journal ID: %d, Status: %s\n", journal.ID, journal.Status)
			}
		}
	}

	fmt.Println("\n=== Analysis Complete ===")
}
