package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("CHECKING & FIXING SALES JOURNALS")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// 1. Get all sales
	var sales []models.Sale
	db.Preload("Customer").
		Where("status IN (?)", []string{"INVOICED", "PAID", "COMPLETED"}).
		Order("date").
		Find(&sales)
	
	fmt.Printf("Found %d sales transactions\n\n", len(sales))

	if len(sales) == 0 {
		fmt.Println("❌ NO SALES FOUND!")
		return
	}

	for _, sale := range sales {
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("Sale: %s | Customer: %s | Date: %s\n", 
			sale.Code, sale.Customer.Name, sale.Date.Format("2006-01-02"))
		fmt.Printf("Total: Rp %.2f | Status: %s\n", sale.TotalAmount, sale.Status)
		fmt.Println()

		// Check if sales journal exists
		var saleJournalCount int64
		db.Model(&models.SimpleSSOTJournal{}).
			Where("transaction_type = ? AND transaction_id = ?", "SALE", sale.ID).
			Count(&saleJournalCount)
		
		if saleJournalCount > 0 {
			fmt.Printf("   ✅ Sales journal EXISTS (%d entries)\n", saleJournalCount)
			
			// Show journal details
			var journals []models.SimpleSSOTJournal
			db.Where("transaction_type = ? AND transaction_id = ?", "SALE", sale.ID).
				Find(&journals)
			
			for _, j := range journals {
				fmt.Printf("      Journal ID: %d | Entry: %s | Amount: Rp %.2f\n", 
					j.ID, j.EntryNumber, j.TotalAmount)
				
				// Show items
				var items []models.SimpleSSOTJournalItem
				db.Where("journal_id = ?", j.ID).Find(&items)
				for _, item := range items {
					fmt.Printf("         %s (%s): Dr %.2f / Cr %.2f\n",
						item.AccountName, item.AccountCode, item.Debit, item.Credit)
				}
			}
		} else {
			fmt.Println("   ❌ NO Sales journal found!")
			fmt.Println("   Creating sales journal...")
			
			err := createSalesJournal(db, &sale)
			if err != nil {
				fmt.Printf("   ❌ Error: %v\n", err)
			} else {
				fmt.Println("   ✅ Sales journal created successfully!")
			}
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("DONE")
	fmt.Println(strings.Repeat("=", 80))
}

func createSalesJournal(db *gorm.DB, sale *models.Sale) error {
	// Calculate amounts
	netAmount := sale.TotalAmount / (1 + sale.PPNRate/100)
	ppnAmount := sale.TotalAmount - netAmount

	// Get accounts
	var receivableAccount, revenueAccount, ppnAccount models.Account
	
	// 1201 - Piutang Usaha
	if err := db.Where("code = ?", "1201").First(&receivableAccount).Error; err != nil {
		return fmt.Errorf("receivable account not found: %v", err)
	}
	
	// 4101 - Pendapatan Penjualan
	if err := db.Where("code = ?", "4101").First(&revenueAccount).Error; err != nil {
		return fmt.Errorf("revenue account not found: %v", err)
	}
	
	// PPN Keluaran (try 2130, 2103, or any 21xx with 'PPN KELUARAN' name)
	err := db.Where("code = ?", "2130").First(&ppnAccount).Error
	if err != nil {
		// Try 2103
		err = db.Where("code = ?", "2103").First(&ppnAccount).Error
		if err != nil {
			// Try by name
			err = db.Where("name LIKE ?", "%PPN KELUARAN%").
				Where("type = ?", "LIABILITY").
				First(&ppnAccount).Error
			if err != nil {
				return fmt.Errorf("ppn output account not found (tried 2130, 2103, and name search): %v", err)
			}
		}
	}

	// Create sales journal
	saleJournal := &models.SimpleSSOTJournal{
		EntryNumber:       fmt.Sprintf("SALE-%d", sale.ID),
		TransactionType:   "SALE",
		TransactionID:     sale.ID,
		TransactionNumber: sale.Code,
		Date:              sale.Date,
		Description:       fmt.Sprintf("Sales Invoice #%s - %s (Auto-created)", sale.Code, sale.Customer.Name),
		TotalAmount:       sale.TotalAmount,
		Status:            "POSTED",
		CreatedBy:         sale.UserID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := db.Create(saleJournal).Error; err != nil {
		return fmt.Errorf("failed to create sales journal: %v", err)
	}

	// Create journal items
	journalItems := []models.SimpleSSOTJournalItem{
		{
			JournalID:   saleJournal.ID,
			AccountID:   receivableAccount.ID,
			AccountCode: receivableAccount.Code,
			AccountName: receivableAccount.Name,
			Debit:       sale.TotalAmount,
			Credit:      0,
			Description: fmt.Sprintf("Sales receivable - %s", sale.Code),
		},
		{
			JournalID:   saleJournal.ID,
			AccountID:   revenueAccount.ID,
			AccountCode: revenueAccount.Code,
			AccountName: revenueAccount.Name,
			Debit:       0,
			Credit:      netAmount,
			Description: fmt.Sprintf("Sales revenue - %s", sale.Code),
		},
		{
			JournalID:   saleJournal.ID,
			AccountID:   ppnAccount.ID,
			AccountCode: ppnAccount.Code,
			AccountName: ppnAccount.Name,
			Debit:       0,
			Credit:      ppnAmount,
			Description: fmt.Sprintf("Output VAT - %s", sale.Code),
		},
	}

	if err := db.Create(&journalItems).Error; err != nil {
		return fmt.Errorf("failed to create journal items: %v", err)
	}

	fmt.Printf("\n   Created journal entry:\n")
	fmt.Printf("      Dr. %s (1201)      Rp %.2f\n", receivableAccount.Name, sale.TotalAmount)
	fmt.Printf("          Cr. %s (4101)  Rp %.2f\n", revenueAccount.Name, netAmount)
	fmt.Printf("          Cr. %s (2130)      Rp %.2f\n", ppnAccount.Name, ppnAmount)
	fmt.Println()

	return nil
}

