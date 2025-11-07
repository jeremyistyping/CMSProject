package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Migration: Create Missing COGS Journals for Existing Sales
// This script auto-generates COGS journal entries for sales that don't have them
// Run once during deployment to fix existing data

func main() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("MIGRATION: Create Missing COGS Journals")
	fmt.Println("Version: 1.0")
	fmt.Println("Date: 2025-10-17")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Database connection - use environment variables in production
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Warn, // Only show warnings/errors
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Database connected")
	fmt.Println()

	// 1. Get all sales transactions (INVOICED or PAID)
	fmt.Println("📊 Step 1: Finding sales transactions...")
	var sales []models.Sale
	err = db.Preload("Customer").
		Preload("SaleItems.Product").
		Where("status IN (?)", []string{"INVOICED", "PAID", "COMPLETED"}).
		Order("date ASC").
		Find(&sales).Error
	
	if err != nil {
		log.Fatalf("❌ Failed to fetch sales: %v", err)
	}

	fmt.Printf("   Found %d sales transactions\n", len(sales))
	fmt.Println()

	if len(sales) == 0 {
		fmt.Println("✅ No sales found. Nothing to migrate.")
		return
	}

	// 2. Check which sales need COGS journals
	fmt.Println("📊 Step 2: Checking for missing COGS journals...")
	needsCOGS := []models.Sale{}
	
	for _, sale := range sales {
		var cogsCount int64
		db.Model(&models.SimpleSSOTJournal{}).
			Where("transaction_type = ? AND transaction_id = ?", "COGS", sale.ID).
			Count(&cogsCount)
		
		if cogsCount == 0 {
			needsCOGS = append(needsCOGS, sale)
			fmt.Printf("   ⚠️  Sales %s (Rp %.2f) - Missing COGS journal\n", 
				sale.Code, sale.TotalAmount)
		}
	}

	fmt.Println()
	fmt.Printf("   %d sales need COGS journals\n", len(needsCOGS))
	fmt.Println()

	if len(needsCOGS) == 0 {
		fmt.Println("✅ All sales already have COGS journals. Nothing to migrate.")
		return
	}

	// 3. Get required accounts
	fmt.Println("📊 Step 3: Loading chart of accounts...")
	
	var cogsAccount, inventoryAccount models.Account
	
	// Try to find COGS account (5101 or 5001)
	if err := db.Where("code = ?", "5101").First(&cogsAccount).Error; err != nil {
		// Try 5001 as fallback
		if err := db.Where("code = ?", "5001").First(&cogsAccount).Error; err != nil {
			log.Fatalf("❌ COGS account not found (tried 5101 and 5001). Please create it first.")
		}
	}
	
	// Find Inventory account (1301)
	if err := db.Where("code = ?", "1301").First(&inventoryAccount).Error; err != nil {
		log.Fatalf("❌ Inventory account (1301) not found. Please create it first.")
	}

	fmt.Printf("   ✅ COGS Account: %s (%s)\n", cogsAccount.Name, cogsAccount.Code)
	fmt.Printf("   ✅ Inventory Account: %s (%s)\n", inventoryAccount.Name, inventoryAccount.Code)
	fmt.Println()

	// 4. Create COGS journals
	fmt.Println("📊 Step 4: Creating COGS journals...")
	fmt.Println(strings.Repeat("-", 80))

	successCount := 0
	errorCount := 0
	skippedCount := 0

	for i, sale := range needsCOGS {
		fmt.Printf("\n[%d/%d] Processing Sale %s...\n", i+1, len(needsCOGS), sale.Code)
		
		// Calculate COGS amount
		// Strategy: Get from purchase history or use 80% of sales price as estimate
		cogsAmount, err := calculateCOGSAmount(db, &sale)
		if err != nil {
			fmt.Printf("   ⚠️  Warning: %v\n", err)
			fmt.Printf("   Using estimated COGS (80%% of sales price)\n")
			// Use 80% of net amount as COGS estimate
			netAmount := sale.TotalAmount / (1 + sale.PPNRate/100)
			cogsAmount = netAmount * 0.8
		}

		if cogsAmount <= 0 {
			fmt.Printf("   ⚠️  Skipping: COGS amount is zero\n")
			skippedCount++
			continue
		}

		fmt.Printf("   COGS Amount: Rp %.2f\n", cogsAmount)

		// Create COGS journal
		err = createCOGSJournal(db, &sale, cogsAmount, &cogsAccount, &inventoryAccount)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			errorCount++
		} else {
			fmt.Printf("   ✅ COGS journal created successfully\n")
			successCount++
		}
	}

	// 5. Summary
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("MIGRATION SUMMARY")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Total sales processed:     %d\n", len(needsCOGS))
	fmt.Printf("✅ Successfully created:   %d\n", successCount)
	fmt.Printf("⚠️  Skipped (zero amount): %d\n", skippedCount)
	fmt.Printf("❌ Errors:                 %d\n", errorCount)
	fmt.Println(strings.Repeat("=", 80))

	if errorCount > 0 {
		fmt.Println()
		fmt.Println("⚠️  Some COGS journals failed to create. Please review errors above.")
		fmt.Println("You can re-run this migration script safely (it skips existing journals).")
	} else if successCount > 0 {
		fmt.Println()
		fmt.Println("✅ MIGRATION COMPLETED SUCCESSFULLY!")
		fmt.Println()
		fmt.Println("📋 Next Steps:")
		fmt.Println("   1. Restart backend server to apply code changes")
		fmt.Println("   2. Re-generate P&L reports to see corrected values")
		fmt.Println("   3. Verify Gross Profit and Net Income calculations")
	} else {
		fmt.Println()
		fmt.Println("✅ NO MIGRATION NEEDED - All data already correct")
	}
	fmt.Println()
}

// calculateCOGSAmount tries to calculate actual COGS from purchase history
func calculateCOGSAmount(db *gorm.DB, sale *models.Sale) (float64, error) {
	var totalCOGS float64

	// Load sale items with products
	var saleItems []models.SaleItem
	if err := db.Preload("Product").Where("sale_id = ?", sale.ID).Find(&saleItems).Error; err != nil {
		return 0, fmt.Errorf("failed to load sale items: %v", err)
	}

	// For each sale item, try to find corresponding purchase
	for _, item := range saleItems {
		// Strategy 1: Check if product has recent purchase price
		var purchaseItem models.PurchaseItem
		err := db.Preload("Purchase").
			Where("product_id = ?", item.ProductID).
			Where("created_at <= ?", sale.Date).
			Order("created_at DESC").
			First(&purchaseItem).Error
		
		if err == nil && purchaseItem.Purchase.Status == "APPROVED" {
			// Found purchase - use actual cost
			costPerUnit := purchaseItem.UnitPrice
			totalCOGS += costPerUnit * float64(item.Quantity)
			continue
		}

		// Strategy 2: Use product's standard cost (if available)
		if item.Product.CostPrice > 0 {
			totalCOGS += item.Product.CostPrice * float64(item.Quantity)
			continue
		}
		
		// Strategy 2b: Use purchase price as fallback
		if item.Product.PurchasePrice > 0 {
			totalCOGS += item.Product.PurchasePrice * float64(item.Quantity)
			continue
		}

		// Strategy 3: Estimate as percentage of sale price
		// Use 80% of item total as COGS (20% margin estimate)
		totalCOGS += item.TotalPrice * 0.8
	}

	if totalCOGS <= 0 {
		// Fallback: Use 80% of total sales amount
		netAmount := sale.TotalAmount / (1 + sale.PPNRate/100)
		return netAmount * 0.8, fmt.Errorf("no purchase history found, using estimate")
	}

	return totalCOGS, nil
}

// createCOGSJournal creates a COGS journal entry
func createCOGSJournal(db *gorm.DB, sale *models.Sale, cogsAmount float64, cogsAccount, inventoryAccount *models.Account) error {
	// Create COGS journal entry
	cogsJournal := &models.SimpleSSOTJournal{
		EntryNumber:       fmt.Sprintf("COGS-%d", sale.ID),
		TransactionType:   "COGS",
		TransactionID:     sale.ID,
		TransactionNumber: sale.Code,
		Date:              sale.Date,
		Description:       fmt.Sprintf("Cost of Goods Sold for Sales #%s (Auto-migrated)", sale.Code),
		TotalAmount:       cogsAmount,
		Status:            "POSTED",
		CreatedBy:         sale.UserID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := db.Create(cogsJournal).Error; err != nil {
		return fmt.Errorf("failed to create COGS journal: %v", err)
	}

	// Create journal line items
	journalItems := []models.SimpleSSOTJournalItem{
		{
			JournalID:   cogsJournal.ID,
			AccountID:   cogsAccount.ID,
			AccountCode: cogsAccount.Code,
			AccountName: cogsAccount.Name,
			Debit:       cogsAmount,
			Credit:      0,
			Description: fmt.Sprintf("COGS for %s", sale.Code),
		},
		{
			JournalID:   cogsJournal.ID,
			AccountID:   inventoryAccount.ID,
			AccountCode: inventoryAccount.Code,
			AccountName: inventoryAccount.Name,
			Debit:       0,
			Credit:      cogsAmount,
			Description: fmt.Sprintf("Reduce inventory for %s", sale.Code),
		},
	}

	if err := db.Create(&journalItems).Error; err != nil {
		return fmt.Errorf("failed to create journal items: %v", err)
	}

	return nil
}

