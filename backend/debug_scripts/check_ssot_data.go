package main

import (
	"log"
	"fmt"

	"app-sistem-akuntansi/database"
)

func main() {
	// Initialize database
	db := database.ConnectDB()

	// Check journal entries count
	var journalCount int64
	var err error
	err = db.Table("journal_entries").Count(&journalCount).Error
	if err != nil {
		log.Printf("Error counting journal entries: %v", err)
	} else {
		fmt.Printf("Total journal entries: %d\n", journalCount)
	}

	// Check sales count
	var salesCount int64
	err = db.Table("sales").Count(&salesCount).Error
	if err != nil {
		log.Printf("Error counting sales: %v", err)
	} else {
		fmt.Printf("Total sales: %d\n", salesCount)
	}

	// Check purchases count  
	var purchaseCount int64
	err = db.Table("purchases").Count(&purchaseCount).Error
	if err != nil {
		log.Printf("Error counting purchases: %v", err)
	} else {
		fmt.Printf("Total purchases: %d\n", purchaseCount)
	}

	// Check payments count
	var paymentCount int64
	err = db.Table("payments").Count(&paymentCount).Error
	if err != nil {
		log.Printf("Error counting payments: %v", err)
	} else {
		fmt.Printf("Total payments: %d\n", paymentCount)
	}

	// Check if account_balances view exists
	var viewExists bool
	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'account_balances')").Scan(&viewExists).Error
	if err != nil {
		log.Printf("Error checking account_balances view: %v", err)
	} else {
		fmt.Printf("Account balances view exists: %t\n", viewExists)
	}

	// Check account balances if table exists
	if viewExists {
		var balanceCount int64
		err = db.Table("account_balances").Count(&balanceCount).Error
		if err != nil {
			log.Printf("Error counting account balances: %v", err)
		} else {
			fmt.Printf("Total account balances: %d\n", balanceCount)
		}
	}

	// Check some journal entries details
	type JournalEntry struct {
		ID      uint   `json:"id"`
		Code    string `json:"code"`
		Date    string `json:"date"`
		Status  string `json:"status"`
		Amount  float64 `json:"total_amount"`
	}
	
	var journals []JournalEntry
	err = db.Table("journal_entries").
		Select("id, code, entry_date, status, total_debit").
		Order("created_at DESC").
		Limit(5).
		Find(&journals).Error
	
	if err != nil {
		log.Printf("Error getting journal entries: %v", err)
	} else {
		fmt.Printf("\nRecent journal entries:\n")
		for _, j := range journals {
			fmt.Printf("  ID: %d, Code: %s, Date: %s, Status: %s, Amount: %.2f\n", 
				j.ID, j.Code, j.Date, j.Status, j.Amount)
		}
	}
}