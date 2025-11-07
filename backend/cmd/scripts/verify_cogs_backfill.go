package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("VERIFY COGS BACKFILL RESULTS")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()

	// Check all COGS journals
	fmt.Println("📊 All COGS Journal Entries:")
	fmt.Println(string(make([]byte, 80)))

	type COGSJournal struct {
		ID          uint
		EntryNumber string
		Reference   string
		SourceType  string
		SourceID    uint
		SourceCode  string
		EntryDate   string
		Description string
		Notes       string
		TotalDebit  float64
		TotalCredit float64
		Status      string
		CreatedAt   string
	}

	var journals []COGSJournal
	db.Raw(`
		SELECT 
			id, entry_number, reference, source_type, source_id, source_code,
			entry_date::text, description, notes, 
			total_debit, total_credit, status,
			created_at::text
		FROM unified_journal_ledger
		WHERE notes = 'COGS' OR source_type = 'COGS' OR reference LIKE 'COGS-%'
		ORDER BY id
	`).Scan(&journals)

	fmt.Printf("Found %d COGS journal entries\n\n", len(journals))

	for _, j := range journals {
		fmt.Printf("ID: %d | Entry: %s | Ref: %s\n", j.ID, j.EntryNumber, j.Reference)
		fmt.Printf("  Source: %s #%d (%s)\n", j.SourceType, j.SourceID, j.SourceCode)
		fmt.Printf("  Date: %s | Status: %s\n", j.EntryDate, j.Status)
		fmt.Printf("  Description: %s\n", j.Description)
		fmt.Printf("  Notes: %s\n", j.Notes)
		fmt.Printf("  Amount: Dr %.2f / Cr %.2f\n", j.TotalDebit, j.TotalCredit)
		fmt.Printf("  Created: %s\n", j.CreatedAt)
		fmt.Println()
	}

	// Check sales without COGS
	fmt.Println("📊 Sales Without COGS Entries:")
	fmt.Println(string(make([]byte, 80)))

	type SaleWithoutCOGS struct {
		ID            uint
		InvoiceNumber string
		Date          string
		Status        string
		TotalAmount   float64
	}

	var salesWithoutCOGS []SaleWithoutCOGS
	db.Raw(`
		SELECT 
			s.id, s.invoice_number, s.date::text, s.status, s.total_amount
		FROM sales s
		WHERE s.status IN ('INVOICED', 'PAID')
		  AND NOT EXISTS (
			SELECT 1 FROM unified_journal_ledger ujl
			WHERE ujl.source_type = 'SALE' 
			  AND ujl.source_id = s.id 
			  AND ujl.notes = 'COGS'
		  )
		ORDER BY s.id
	`).Scan(&salesWithoutCOGS)

	fmt.Printf("Found %d sales without COGS\n\n", len(salesWithoutCOGS))

	for _, s := range salesWithoutCOGS {
		fmt.Printf("Sale #%d | Invoice: %s | Date: %s | Status: %s | Amount: %.2f\n",
			s.ID, s.InvoiceNumber, s.Date, s.Status, s.TotalAmount)
	}
	fmt.Println()

	// Check all sales
	fmt.Println("📊 All Sales (INVOICED/PAID):")
	fmt.Println(string(make([]byte, 80)))

	type Sale struct {
		ID            uint
		InvoiceNumber string
		Date          string
		Status        string
		TotalAmount   float64
		HasCOGS       bool
	}

	var allSales []Sale
	db.Raw(`
		SELECT 
			s.id, 
			s.invoice_number, 
			s.date::text, 
			s.status, 
			s.total_amount,
			EXISTS (
				SELECT 1 FROM unified_journal_ledger ujl
				WHERE ujl.source_type = 'SALE' 
				  AND ujl.source_id = s.id 
				  AND ujl.notes = 'COGS'
			) as has_cogs
		FROM sales s
		WHERE s.status IN ('INVOICED', 'PAID')
		ORDER BY s.id
	`).Scan(&allSales)

	fmt.Printf("Found %d total sales\n\n", len(allSales))

	for _, s := range allSales {
		cogsStatus := "❌ NO COGS"
		if s.HasCOGS {
			cogsStatus = "✅ HAS COGS"
		}
		fmt.Printf("Sale #%d | %s | %s | Status: %s | Amount: %.2f | %s\n",
			s.ID, s.InvoiceNumber, s.Date, s.Status, s.TotalAmount, cogsStatus)
	}
	fmt.Println()

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("✅ VERIFICATION COMPLETE")
	fmt.Println("=" + string(make([]byte, 80)))
}

