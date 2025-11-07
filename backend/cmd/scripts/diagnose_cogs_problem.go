package main

import (
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	log.Println("🔍 DIAGNOSING COGS PROBLEM")
	log.Println("========================================")

	cfg := config.LoadConfig()
	log.Printf("Environment: %s\n", cfg.Environment)

	db := database.ConnectDB()
	log.Println("✅ Database connected\n")

	// Check 1: How many sales exist?
	var totalSales int64
	db.Raw("SELECT COUNT(*) FROM sales WHERE status IN ('INVOICED', 'PAID') AND deleted_at IS NULL").Scan(&totalSales)
	log.Printf("📊 Total INVOICED/PAID Sales: %d\n", totalSales)

	// Check 2: How many have COGS journal entries (account 5101)?
	var salesWithCOGS int64
	db.Raw(`
		SELECT COUNT(DISTINCT uje.source_id)
		FROM unified_journal_ledger uje
		JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
		JOIN accounts a ON a.id = ujl.account_id
		WHERE uje.source_type = 'SALE'
		  AND a.code = '5101'
		  AND uje.deleted_at IS NULL
	`).Scan(&salesWithCOGS)
	log.Printf("📊 Sales with COGS entries (account 5101): %d\n", salesWithCOGS)

	// Check 3: Total COGS amount in journals
	var totalCOGSAmount float64
	db.Raw(`
		SELECT COALESCE(SUM(ujl.debit_amount), 0)
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id AND uje.status = 'POSTED' AND uje.deleted_at IS NULL
		JOIN accounts a ON a.id = ujl.account_id
		WHERE a.code = '5101'
	`).Scan(&totalCOGSAmount)
	log.Printf("💰 Total COGS Amount: Rp %.2f\n", totalCOGSAmount)

	// Check 4: Sample COGS entries
	type COGSEntry struct {
		SaleID      uint
		InvoiceNum  string
		JournalID   uint
		COGSAmount  float64
		Description string
	}

	var cogsEntries []COGSEntry
	db.Raw(`
		SELECT 
			uje.source_id as sale_id,
			s.invoice_number as invoice_num,
			uje.id as journal_id,
			ujl.debit_amount as cogs_amount,
			ujl.description
		FROM unified_journal_ledger uje
		JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
		JOIN accounts a ON a.id = ujl.account_id
		LEFT JOIN sales s ON s.id = uje.source_id
		WHERE uje.source_type = 'SALE'
		  AND a.code = '5101'
		  AND uje.deleted_at IS NULL
		ORDER BY uje.id DESC
		LIMIT 10
	`).Scan(&cogsEntries)

	log.Println("\n📋 Sample COGS Entries (last 10):")
	log.Println("   Sale ID | Invoice         | Journal ID | COGS Amount      | Description")
	log.Println("   --------|-----------------|------------|------------------|------------------")
	for _, entry := range cogsEntries {
		log.Printf("   %-7d | %-15s | %-10d | Rp %12.2f | %s",
			entry.SaleID, entry.InvoiceNum, entry.JournalID, entry.COGSAmount, entry.Description)
	}

	// Check 5: Products without cost price that might cause zero COGS
	var productsNoCost int64
	db.Raw(`
		SELECT COUNT(*)
		FROM products
		WHERE (cost_price = 0 OR cost_price IS NULL)
		  AND stock > 0
		  AND deleted_at IS NULL
	`).Scan(&productsNoCost)
	log.Printf("\n⚠️  Products with zero cost price: %d\n", productsNoCost)

	// Check 6: Sale items with zero cost price
	type SaleItemNoCost struct {
		SaleID      uint
		InvoiceNum  string
		ProductName string
		Quantity    int
		CostPrice   float64
	}

	var itemsNoCost []SaleItemNoCost
	db.Raw(`
		SELECT 
			s.id as sale_id,
			s.invoice_number as invoice_num,
			p.name as product_name,
			si.quantity,
			p.cost_price
		FROM sales s
		JOIN sale_items si ON si.sale_id = s.id
		JOIN products p ON p.id = si.product_id
		WHERE s.status IN ('INVOICED', 'PAID')
		  AND (p.cost_price = 0 OR p.cost_price IS NULL)
		  AND s.deleted_at IS NULL
		ORDER BY s.id DESC
		LIMIT 10
	`).Scan(&itemsNoCost)

	if len(itemsNoCost) > 0 {
		log.Println("\n📋 Sale Items with Zero Cost Price (last 10):")
		log.Println("   Sale ID | Invoice         | Product Name          | Qty | Cost Price")
		log.Println("   --------|-----------------|----------------------|-----|------------")
		for _, item := range itemsNoCost {
			log.Printf("   %-7d | %-15s | %-20s | %3d | Rp %.2f",
				item.SaleID, item.InvoiceNum, item.ProductName, item.Quantity, item.CostPrice)
		}
	}

	// Summary
	log.Println("\n========================================")
	log.Println("🎯 DIAGNOSIS SUMMARY:")
	log.Println("========================================")
	log.Printf("Total Sales:              %d\n", totalSales)
	log.Printf("Sales with COGS:          %d (%.1f%%)\n", salesWithCOGS, float64(salesWithCOGS)/float64(totalSales)*100)
	log.Printf("Total COGS Amount:        Rp %.2f\n", totalCOGSAmount)
	log.Printf("Products w/ Zero Cost:    %d\n", productsNoCost)
	
	if totalCOGSAmount < 1000000 {
		log.Println("\n❌ PROBLEM IDENTIFIED:")
		log.Println("   COGS amount is extremely low!")
		log.Println("   Root causes:")
		if productsNoCost > 0 {
			log.Printf("   1. %d products have zero cost price\n", productsNoCost)
			log.Println("   2. COGS entries are created but with zero value")
		}
		if salesWithCOGS < totalSales {
			log.Println("   3. Some sales don't have COGS entries")
		}
		log.Println("\n💡 SOLUTION:")
		log.Println("   1. Run: go run cmd/scripts/fix_product_cost_prices.go")
		log.Println("   2. Run: go run cmd/scripts/backfill_missing_cogs.go")
		log.Println("   3. Rebuild backend to activate COGS recording for new sales")
	}
	log.Println("========================================")
}
