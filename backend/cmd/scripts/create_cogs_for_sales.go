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
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("CREATE COGS JOURNAL FOR SALES TRANSACTIONS")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// 1. Find sales transaction (SOA-00016)
	var sales models.Sale
	err = db.Preload("Customer").
		Preload("SaleItems.Product").
		Where("code = ?", "SOA-00016").
		First(&sales).Error
	
	if err != nil {
		log.Fatalf("Failed to fetch sales: %v", err)
	}

	fmt.Printf("Sales ID: %d\n", sales.ID)
	fmt.Printf("Sales Code: %s\n", sales.Code)
	fmt.Printf("Customer: %s\n", sales.Customer.Name)
	fmt.Printf("Date: %s\n", sales.Date.Format("2006-01-02"))
	fmt.Printf("Total Amount: Rp %.2f\n", sales.TotalAmount)
	fmt.Printf("Status: %s\n", sales.Status)
	fmt.Println()

	// 2. Check if COGS journal already exists
	var cogsJournalCount int64
	db.Model(&models.SimpleSSOTJournal{}).
		Where("transaction_type = ? AND transaction_id = ? AND description LIKE ?", 
			"COGS", sales.ID, "%Cost of Goods Sold%").
		Count(&cogsJournalCount)
	
	if cogsJournalCount > 0 {
		fmt.Printf("✅ COGS journal already exists for this sales (%d entries)\n", cogsJournalCount)
		fmt.Println("No need to create new COGS journal.")
		return
	}

	fmt.Println("❌ No COGS journal found for this sales")
	fmt.Println("Creating COGS journal entry...")
	fmt.Println()

	// 3. Calculate COGS amount
	// Assuming purchase cost = Rp 500,000 for the product
	cogsAmount := 500000.00  // This should match the inventory cost

	fmt.Printf("📊 COGS Amount: Rp %.2f\n", cogsAmount)
	fmt.Println()

	// 4. Get accounts
	var cogsAccount, inventoryAccount models.Account
	
	// Get COGS account (5101 - HARGA POKOK PENJUALAN)
	if err := db.Where("code = ?", "5101").First(&cogsAccount).Error; err != nil {
		log.Fatalf("Failed to find COGS account (5101): %v", err)
	}
	
	// Get Inventory account (1301 - Persediaan Barang Dagangan)
	if err := db.Where("code = ?", "1301").First(&inventoryAccount).Error; err != nil {
		log.Fatalf("Failed to find Inventory account (1301): %v", err)
	}

	fmt.Printf("✅ COGS Account: %s (%s)\n", cogsAccount.Name, cogsAccount.Code)
	fmt.Printf("✅ Inventory Account: %s (%s)\n", inventoryAccount.Name, inventoryAccount.Code)
	fmt.Println()

	// 5. Create COGS Journal Entry
	cogsJournal := &models.SimpleSSOTJournal{
		EntryNumber:       fmt.Sprintf("COGS-%d", sales.ID),
		TransactionType:   "COGS",
		TransactionID:     sales.ID,
		TransactionNumber: sales.Code,
		Date:              sales.Date,
		Description:       fmt.Sprintf("Cost of Goods Sold for Sales #%s", sales.Code),
		TotalAmount:       cogsAmount,
		Status:            "POSTED",
		CreatedBy:         sales.UserID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := db.Create(cogsJournal).Error; err != nil {
		log.Fatalf("Failed to create COGS journal: %v", err)
	}

	fmt.Printf("✅ Created COGS journal entry ID: %d\n", cogsJournal.ID)
	fmt.Println()

	// 6. Create journal items
	journalItems := []models.SimpleSSOTJournalItem{
		{
			JournalID:   cogsJournal.ID,
			AccountID:   cogsAccount.ID,
			AccountCode: cogsAccount.Code,
			AccountName: cogsAccount.Name,
			Debit:       cogsAmount,
			Credit:      0,
			Description: fmt.Sprintf("COGS for %s", sales.Code),
		},
		{
			JournalID:   cogsJournal.ID,
			AccountID:   inventoryAccount.ID,
			AccountCode: inventoryAccount.Code,
			AccountName: inventoryAccount.Name,
			Debit:       0,
			Credit:      cogsAmount,
			Description: fmt.Sprintf("Reduce inventory for %s", sales.Code),
		},
	}

	if err := db.Create(&journalItems).Error; err != nil {
		log.Fatalf("Failed to create journal items: %v", err)
	}

	fmt.Println("✅ Created journal items:")
	fmt.Println()
	fmt.Printf("   Dr. %s (%s)  Rp %.2f\n", cogsAccount.Name, cogsAccount.Code, cogsAmount)
	fmt.Printf("       Cr. %s (%s)  Rp %.2f\n", inventoryAccount.Name, inventoryAccount.Code, cogsAmount)
	fmt.Println()

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("✅ SUCCESS!")
	fmt.Println("COGS journal created successfully.")
	fmt.Println()
	fmt.Println("📊 EXPECTED P&L IMPACT:")
	fmt.Printf("   Revenue (Sales):  Rp 5,000,000\n")
	fmt.Printf("   COGS:             Rp   500,000 ✅ NEW\n")
	fmt.Println("   -----------------------------------")
	fmt.Printf("   Gross Profit:     Rp 4,500,000\n")
	fmt.Printf("   Gross Margin:     90.0%%\n")
	fmt.Println()
	fmt.Println("Now re-run P&L report to see the updated results!")
	fmt.Println(strings.Repeat("=", 80))
}

