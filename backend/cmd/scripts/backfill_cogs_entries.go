package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"app-sistem-akuntansi/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/*
BACKFILL COGS ENTRIES SCRIPT

Purpose:
  This script backfills COGS (Cost of Goods Sold) journal entries for existing sales transactions
  that don't have COGS entries yet. This is essential for accurate Profit & Loss reporting.

Problem:
  - Sales transactions create revenue journal entries (CREDIT 4101)
  - Purchase transactions create inventory journal entries (DEBIT 1301)
  - BUT, when items are sold, COGS journal entries are NOT created automatically
  - Result: P&L shows revenue but NO expenses, giving incorrect profit

Solution:
  This script creates COGS journal entries:
    DEBIT:  5101 Harga Pokok Penjualan (COGS - Expense)
    CREDIT: 1301 Persediaan Barang (Inventory - Asset)
  
  For each sale item, COGS = Quantity * Product Cost Price

Usage:
  go run backend/cmd/scripts/backfill_cogs_entries.go -start=2025-01-01 -end=2025-12-31

Flags:
  -start       Start date (YYYY-MM-DD) - required
  -end         End date (YYYY-MM-DD) - required
  -dry-run     Preview without making changes (default: false)

Example:
  # Dry run to preview
  go run backend/cmd/scripts/backfill_cogs_entries.go -start=2025-01-01 -end=2025-12-31 -dry-run

  # Actually execute
  go run backend/cmd/scripts/backfill_cogs_entries.go -start=2025-01-01 -end=2025-12-31
*/

func main() {
	// Command line flags
	startDateStr := flag.String("start", "", "Start date (YYYY-MM-DD) - required")
	endDateStr := flag.String("end", "", "End date (YYYY-MM-DD) - required")
	dryRun := flag.Bool("dry-run", false, "Preview without making changes")
	
	flag.Parse()

	// Validate required flags
	if *startDateStr == "" || *endDateStr == "" {
		log.Fatal("Error: -start and -end flags are required\n\nUsage: go run backfill_cogs_entries.go -start=2025-01-01 -end=2025-12-31")
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", *startDateStr)
	if err != nil {
		log.Fatalf("Error: Invalid start date format. Use YYYY-MM-DD: %v", err)
	}

	endDate, err := time.Parse("2006-01-02", *endDateStr)
	if err != nil {
		log.Fatalf("Error: Invalid end date format. Use YYYY-MM-DD: %v", err)
	}

	// Validate date range
	if endDate.Before(startDate) {
		log.Fatal("Error: End date must be after start date")
	}

	// Print header
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("            BACKFILL COGS JOURNAL ENTRIES SCRIPT              ")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("Start Date: %s\n", startDate.Format("2006-01-02"))
	fmt.Printf("End Date:   %s\n", endDate.Format("2006-01-02"))
	fmt.Printf("Mode:       %s\n", func() string {
		if *dryRun {
			return "DRY RUN (Preview Only)"
		}
		return "EXECUTE (Will Make Changes)"
	}())
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println()

	// Initialize database
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize services
	coaService := services.NewCOAService(db)
	cogsService := services.NewInventoryCOGSService(db, coaService)

	// Step 1: Get summary of sales in period
	fmt.Println("📊 STEP 1: Analyzing Sales Transactions")
	fmt.Println("───────────────────────────────────────────────────────────────")
	
	type SalesSummary struct {
		TotalSales      int64
		InvoicedSales   int64
		PaidSales       int64
		TotalRevenue    float64
		SalesWithCOGS   int64
		SalesWithoutCOGS int64
	}

	var summary SalesSummary

	// Total sales in period
	db.Table("sales").
		Where("date >= ? AND date <= ?", startDate, endDate).
		Where("status IN ('INVOICED', 'PAID')").
		Count(&summary.TotalSales)

	// Sales with COGS entries
	db.Table("sales").
		Where("date >= ? AND date <= ?", startDate, endDate).
		Where("status IN ('INVOICED', 'PAID')").
		Where("EXISTS (SELECT 1 FROM unified_journal_ledger WHERE source_type = 'SALE' AND source_id = sales.id AND notes = 'COGS')").
		Count(&summary.SalesWithCOGS)

	// Sales without COGS entries
	summary.SalesWithoutCOGS = summary.TotalSales - summary.SalesWithCOGS

	// Total revenue
	db.Table("sales").
		Select("COALESCE(SUM(total_amount), 0)").
		Where("date >= ? AND date <= ?", startDate, endDate).
		Where("status IN ('INVOICED', 'PAID')").
		Scan(&summary.TotalRevenue)

	fmt.Printf("Total Sales:            %d\n", summary.TotalSales)
	fmt.Printf("Sales WITH COGS:        %d ✅\n", summary.SalesWithCOGS)
	fmt.Printf("Sales WITHOUT COGS:     %d ⚠️ (Need Backfill)\n", summary.SalesWithoutCOGS)
	fmt.Printf("Total Revenue:          Rp %.2f\n", summary.TotalRevenue)
	fmt.Println()

	if summary.SalesWithoutCOGS == 0 {
		fmt.Println("✅ All sales already have COGS entries. Nothing to backfill.")
		return
	}

	// Step 2: Preview COGS that will be created
	fmt.Println("📊 STEP 2: Preview COGS Calculation")
	fmt.Println("───────────────────────────────────────────────────────────────")

	type PreviewResult struct {
		SaleID          uint
		InvoiceNumber   string
		SaleDate        time.Time
		TotalAmount     float64
		EstimatedCOGS   float64
		COGSPercentage  float64
	}

	var previews []PreviewResult

	rows, err := db.Raw(`
		SELECT 
			s.id as sale_id,
			s.invoice_number,
			s.date as sale_date,
			s.total_amount,
			COALESCE(SUM(si.quantity * p.cost_price), 0) as estimated_cogs
		FROM sales s
		LEFT JOIN sale_items si ON si.sale_id = s.id
		LEFT JOIN products p ON p.id = si.product_id
		WHERE s.date >= ? AND s.date <= ?
		  AND s.status IN ('INVOICED', 'PAID')
		  AND NOT EXISTS (
			SELECT 1 FROM unified_journal_ledger 
			WHERE source_type = 'SALE' AND source_id = s.id AND notes = 'COGS'
		  )
		GROUP BY s.id, s.invoice_number, s.date, s.total_amount
		ORDER BY s.date ASC
		LIMIT 10
	`, startDate, endDate).Rows()

	if err != nil {
		log.Fatalf("Failed to preview COGS: %v", err)
	}
	defer rows.Close()

	var totalEstimatedCOGS float64

	fmt.Printf("%-12s %-20s %-12s %15s %15s %10s\n", 
		"Sale ID", "Invoice Number", "Date", "Revenue", "Est. COGS", "COGS %")
	fmt.Println("───────────────────────────────────────────────────────────────")

	for rows.Next() {
		var p PreviewResult
		rows.Scan(&p.SaleID, &p.InvoiceNumber, &p.SaleDate, &p.TotalAmount, &p.EstimatedCOGS)
		
		if p.TotalAmount > 0 {
			p.COGSPercentage = (p.EstimatedCOGS / p.TotalAmount) * 100
		}
		
		totalEstimatedCOGS += p.EstimatedCOGS
		previews = append(previews, p)

		fmt.Printf("%-12d %-20s %-12s %15.2f %15.2f %9.2f%%\n",
			p.SaleID, p.InvoiceNumber, p.SaleDate.Format("2006-01-02"),
			p.TotalAmount, p.EstimatedCOGS, p.COGSPercentage)
	}

	if len(previews) < int(summary.SalesWithoutCOGS) {
		fmt.Printf("... and %d more sales ...\n", summary.SalesWithoutCOGS - int64(len(previews)))
	}

	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Printf("Total Estimated COGS: Rp %.2f\n", totalEstimatedCOGS)
	fmt.Println()

	// Step 3: Execute or show dry run message
	if *dryRun {
		fmt.Println("⚠️  DRY RUN MODE - No changes were made")
		fmt.Println()
		fmt.Println("To execute, run without -dry-run flag:")
		fmt.Printf("  go run backfill_cogs_entries.go -start=%s -end=%s\n", *startDateStr, *endDateStr)
		return
	}

	// Confirm execution
	fmt.Println("⚠️  WARNING: This will create COGS journal entries for all sales without them.")
	fmt.Println("   This operation cannot be easily undone.")
	fmt.Println()
	fmt.Print("Do you want to continue? (yes/no): ")
	
	var confirm string
	fmt.Scanln(&confirm)
	
	if confirm != "yes" {
		fmt.Println("❌ Operation cancelled by user")
		return
	}

	// Step 4: Execute backfill
	fmt.Println()
	fmt.Println("🔄 STEP 3: Executing COGS Backfill")
	fmt.Println("───────────────────────────────────────────────────────────────")

	ctx := context.Background()
	successCount, err := cogsService.BackfillCOGSForExistingSales(ctx, startDate, endDate)
	if err != nil {
		log.Fatalf("Failed to backfill COGS: %v", err)
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                    BACKFILL COMPLETE                          ")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("Successfully processed: %d sales\n", successCount)
	fmt.Printf("Estimated COGS added:   Rp %.2f\n", totalEstimatedCOGS)
	fmt.Println()

	// Step 5: Verify results
	fmt.Println("📊 STEP 4: Verification")
	fmt.Println("───────────────────────────────────────────────────────────────")

	cogsSummary, err := cogsService.GetCOGSSummary(startDate, endDate)
	if err != nil {
		log.Printf("Warning: Failed to get COGS summary: %v", err)
	} else {
		fmt.Printf("Total COGS:         Rp %.2f\n", cogsSummary["total_cogs"])
		fmt.Printf("Total Revenue:      Rp %.2f\n", cogsSummary["total_revenue"])
		fmt.Printf("COGS Percentage:    %.2f%%\n", cogsSummary["cogs_percentage"])
		fmt.Printf("Average COGS:       Rp %.2f\n", cogsSummary["avg_cogs"])
	}

	fmt.Println()
	fmt.Println("✅ Done! You can now generate P&L reports with accurate COGS.")
	fmt.Println()
	fmt.Println("Next Steps:")
	fmt.Println("  1. Review the P&L report to verify COGS is now showing")
	fmt.Println("  2. Compare Gross Profit margins to ensure they are realistic")
	fmt.Println("  3. If needed, adjust product cost prices in the system")
	fmt.Println()
}

