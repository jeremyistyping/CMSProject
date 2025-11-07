package main

import (
	"log"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("🔍 Checking Deposit/Cash Issue")
	log.Println("===============================")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	
	log.Printf("Database connected successfully")

	// Check cash bank entries
	log.Println("\n1. CASH BANK ENTRIES:")
	var cashbankCount int64
	db.Table("cashbank").Count(&cashbankCount)
	log.Printf("Total cashbank entries: %d", cashbankCount)

	// Get recent cashbank entries
	var cashbankEntries []map[string]interface{}
	err = db.Table("cashbank").
		Select("id, date, account_id, type, amount, description").
		Order("created_at DESC").
		Limit(10).
		Find(&cashbankEntries).Error
	if err != nil {
		log.Printf("Error getting cashbank entries: %v", err)
	} else {
		log.Printf("Recent cashbank entries (%d):", len(cashbankEntries))
		for i, entry := range cashbankEntries {
			if i < 5 {
				log.Printf("  - ID: %v, Type: %v, Amount: %v, Account: %v, Description: %v", 
					entry["id"], entry["type"], entry["amount"], entry["account_id"], entry["description"])
			}
		}
	}

	// Check account balances
	log.Println("\n2. CASH ACCOUNT BALANCES:")
	var cashAccounts []map[string]interface{}
	err = db.Table("accounts").
		Select("id, code, name, balance").
		Where("code LIKE '110%' OR code LIKE '11%' OR name ILIKE '%kas%' OR name ILIKE '%bank%'").
		Order("code").
		Find(&cashAccounts).Error
	if err != nil {
		log.Printf("Error getting cash accounts: %v", err)
	} else {
		log.Printf("Cash accounts (%d):", len(cashAccounts))
		for _, account := range cashAccounts {
			log.Printf("  - %s: %s = %v", account["code"], account["name"], account["balance"])
		}
	}

	// Check SSOT journal entries
	log.Println("\n3. SSOT JOURNAL ENTRIES:")
	var journalCount int64
	db.Table("unified_journal_ledger").Count(&journalCount)
	log.Printf("Total SSOT journal entries: %d", journalCount)

	if journalCount > 0 {
		var journals []map[string]interface{}
		err = db.Table("unified_journal_ledger").
			Select("id, entry_number, source_type, description, total_debit, total_credit, status").
			Order("created_at DESC").
			Find(&journals).Error
		if err != nil {
			log.Printf("Error getting journal entries: %v", err)
		} else {
			log.Printf("SSOT journal entries (%d):", len(journals))
			for _, journal := range journals {
				log.Printf("  - %s: %s (Dr: %v, Cr: %v) - %v", 
					journal["entry_number"], journal["description"], 
					journal["total_debit"], journal["total_credit"], journal["status"])
				
				// Get lines for this journal
				var lines []map[string]interface{}
				db.Table("unified_journal_lines ujl").
					Select("ujl.debit_amount, ujl.credit_amount, ujl.description, a.code, a.name").
					Joins("LEFT JOIN accounts a ON a.id = ujl.account_id").
					Where("ujl.journal_id = ?", journal["id"]).
					Find(&lines)
				
				for _, line := range lines {
					log.Printf("    -> %s (%s): Dr %v, Cr %v - %s", 
						line["code"], line["name"], line["debit_amount"], line["credit_amount"], line["description"])
				}
			}
		}
	}

	// Check if there are any deposits
	log.Println("\n4. CHECKING FOR DEPOSITS:")
	var deposits []map[string]interface{}
	err = db.Table("payments").
		Select("id, payment_date, amount, payment_type, status, description").
		Where("payment_type = ? OR description ILIKE '%deposit%'", "DEPOSIT").
		Order("created_at DESC").
		Limit(5).
		Find(&deposits).Error
	if err != nil {
		log.Printf("Error getting deposits: %v", err)
	} else {
		log.Printf("Deposit entries (%d):", len(deposits))
		for _, deposit := range deposits {
			log.Printf("  - ID: %v, Date: %v, Amount: %v, Type: %v, Status: %v", 
				deposit["id"], deposit["payment_date"], deposit["amount"], 
				deposit["payment_type"], deposit["status"])
		}
	}

	log.Println("\n✅ Check completed!")
}