package main

import (
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	log.Println("🔍 CHECKING SALES DETAILS")
	log.Println("========================================")

	_ = config.LoadConfig()
	db := database.ConnectDB()
	log.Println("✅ Database connected\n")

	// Get sales details with items and products
	type SaleDetail struct {
		SaleID       uint
		InvoiceNum   string
		SaleDate     string
		TotalAmount  float64
		ItemID       uint
		ProductID    uint
		ProductName  string
		Quantity     int
		UnitPrice    float64
		TotalPrice   float64
		CostPrice    float64
		ItemCOGS     float64
	}

	var sales []SaleDetail
	db.Raw(`
		SELECT 
			s.id as sale_id,
			s.invoice_number as invoice_num,
			s.date as sale_date,
			s.total_amount,
			si.id as item_id,
			p.id as product_id,
			p.name as product_name,
			si.quantity,
			si.unit_price,
			si.total_price,
			p.cost_price,
			(si.quantity * p.cost_price) as item_cogs
		FROM sales s
		JOIN sale_items si ON si.sale_id = s.id
		JOIN products p ON p.id = si.product_id
		WHERE s.status IN ('INVOICED', 'PAID')
		  AND s.deleted_at IS NULL
		ORDER BY s.id, si.id
	`).Scan(&sales)

	log.Println("📋 SALES DETAILS:")
	log.Println("=====================================================================================================")
	
	currentSaleID := uint(0)
	saleTotalCOGS := 0.0
	
	for _, sale := range sales {
		if sale.SaleID != currentSaleID {
			if currentSaleID != 0 {
				log.Printf("\n   → Total COGS for Sale #%d: Rp %.2f\n", currentSaleID, saleTotalCOGS)
				log.Println("   ------------------------------------------------------------------------------------")
			}
			currentSaleID = sale.SaleID
			saleTotalCOGS = 0
			
			log.Printf("\n📦 SALE #%d - %s (Date: %s)", sale.SaleID, sale.InvoiceNum, sale.SaleDate)
			log.Printf("\n   Total Amount: Rp %.2f\n", sale.TotalAmount)
			log.Println("   Product Name         | Qty | Unit Price     | Total Price    | Cost Price    | Item COGS")
			log.Println("   ---------------------|-----|----------------|----------------|---------------|-------------")
		}
		
		log.Printf("   %-20s | %3d | Rp %11.2f | Rp %11.2f | Rp %10.2f | Rp %10.2f",
			truncate(sale.ProductName, 20), sale.Quantity, sale.UnitPrice, 
			sale.TotalPrice, sale.CostPrice, sale.ItemCOGS)
		
		saleTotalCOGS += sale.ItemCOGS
	}
	
	if currentSaleID != 0 {
		log.Printf("\n   → Total COGS for Sale #%d: Rp %.2f\n", currentSaleID, saleTotalCOGS)
	}

	log.Println("\n=====================================================================================================")

	// Calculate expected vs actual COGS
	var expectedCOGS float64
	for _, sale := range sales {
		expectedCOGS += sale.ItemCOGS
	}

	var actualCOGS float64
	db.Raw(`
		SELECT COALESCE(SUM(ujl.debit_amount), 0)
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id AND uje.status = 'POSTED' AND uje.deleted_at IS NULL
		JOIN accounts a ON a.id = ujl.account_id
		WHERE a.code = '5101'
	`).Scan(&actualCOGS)

	log.Println("\n🎯 COGS COMPARISON:")
	log.Println("========================================")
	log.Printf("Expected COGS (from products): Rp %.2f\n", expectedCOGS)
	log.Printf("Actual COGS (in journals):     Rp %.2f\n", actualCOGS)
	log.Printf("Difference:                    Rp %.2f\n", expectedCOGS-actualCOGS)
	
	if expectedCOGS != actualCOGS {
		log.Println("\n❌ MISMATCH DETECTED!")
		if expectedCOGS > actualCOGS {
			log.Println("   → Actual COGS is LOWER than expected")
			log.Println("   → Need to backfill or recreate COGS entries")
		} else {
			log.Println("   → Actual COGS is HIGHER than expected")
			log.Println("   → May have duplicate entries or wrong calculation")
		}
	} else {
		log.Println("\n✅ COGS MATCH!")
	}
	
	log.Println("========================================")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
