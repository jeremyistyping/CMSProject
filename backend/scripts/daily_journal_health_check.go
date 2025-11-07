package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

// DailyJournalHealthCheck performs automated journal integrity checks
// Run this daily via cron job or scheduler
func main() {
	fmt.Println("===========================================")
	fmt.Println("🏥 DAILY JOURNAL HEALTH CHECK")
	fmt.Println("===========================================")
	fmt.Printf("Run Time: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Initialize database
	db, err := database.InitDatabase()
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Run all health checks
	issues := 0
	
	issues += checkDuplicateJournals(db)
	issues += checkUnbalancedJournals(db)
	issues += checkMissingSalesJournals(db)
	issues += checkMissingPurchaseJournals(db)
	issues += checkJournalsWithoutLines(db)
	issues += checkHeaderLineMismatch(db)
	
	// Summary
	fmt.Println("\n===========================================")
	fmt.Println("📊 HEALTH CHECK SUMMARY")
	fmt.Println("===========================================")
	
	if issues == 0 {
		fmt.Println("✅ ALL CHECKS PASSED - System Healthy!")
		os.Exit(0)
	} else {
		fmt.Printf("⚠️  FOUND %d ISSUE(S) - Review Required!\n", issues)
		os.Exit(1) // Non-zero exit code for monitoring systems
	}
}

func checkDuplicateJournals(db *gorm.DB) int {
	fmt.Println("1️⃣  Checking for duplicate journal entries...")
	
	type DuplicateResult struct {
		SourceType   string
		SourceID     uint64
		JournalCount int
	}
	
	var results []DuplicateResult
	err := db.Raw(`
		SELECT 
			source_type,
			source_id,
			COUNT(*) as journal_count
		FROM unified_journal_ledger
		WHERE deleted_at IS NULL
		  AND status = 'POSTED'
		GROUP BY source_type, source_id
		HAVING COUNT(*) > 1
		ORDER BY journal_count DESC
		LIMIT 10
	`).Scan(&results).Error
	
	if err != nil {
		fmt.Printf("   ❌ Error checking duplicates: %v\n", err)
		return 1
	}
	
	if len(results) == 0 {
		fmt.Println("   ✅ No duplicate journals found")
		return 0
	}
	
	fmt.Printf("   ❌ CRITICAL: Found %d duplicate journal entries:\n", len(results))
	for _, r := range results {
		fmt.Printf("      - %s ID %d has %d journals\n", r.SourceType, r.SourceID, r.JournalCount)
	}
	return len(results)
}

func checkUnbalancedJournals(db *gorm.DB) int {
	fmt.Println("\n2️⃣  Checking for unbalanced journal entries...")
	
	type UnbalancedResult struct {
		ID          uint64
		EntryNumber string
		SourceType  string
		TotalDebit  float64
		TotalCredit float64
		Difference  float64
	}
	
	var results []UnbalancedResult
	err := db.Raw(`
		SELECT 
			id,
			entry_number,
			source_type,
			total_debit,
			total_credit,
			(total_debit - total_credit) as difference
		FROM unified_journal_ledger
		WHERE (is_balanced = false 
		   OR ABS(total_debit - total_credit) > 0.01)
		  AND deleted_at IS NULL
		ORDER BY id DESC
		LIMIT 10
	`).Scan(&results).Error
	
	if err != nil {
		fmt.Printf("   ❌ Error checking balance: %v\n", err)
		return 1
	}
	
	if len(results) == 0 {
		fmt.Println("   ✅ All journals are balanced")
		return 0
	}
	
	fmt.Printf("   ❌ CRITICAL: Found %d unbalanced journals:\n", len(results))
	for _, r := range results {
		fmt.Printf("      - Journal %s (%s): Debit=%.2f Credit=%.2f Diff=%.2f\n", 
			r.EntryNumber, r.SourceType, r.TotalDebit, r.TotalCredit, r.Difference)
	}
	return len(results)
}

func checkMissingSalesJournals(db *gorm.DB) int {
	fmt.Println("\n3️⃣  Checking for missing sales journals...")
	
	type MissingResult struct {
		SaleID      uint
		SaleCode    string
		Status      string
		TotalAmount float64
		CreatedAt   time.Time
	}
	
	var results []MissingResult
	err := db.Raw(`
		SELECT 
			s.id as sale_id,
			s.code as sale_code,
			s.status,
			s.total_amount,
			s.created_at
		FROM sales s
		LEFT JOIN unified_journal_ledger ujl 
			ON ujl.source_type = 'SALE' 
			AND ujl.source_id = s.id 
			AND ujl.deleted_at IS NULL
		WHERE s.status IN ('INVOICED', 'PAID')
		  AND s.deleted_at IS NULL
		  AND ujl.id IS NULL
		ORDER BY s.created_at DESC
		LIMIT 10
	`).Scan(&results).Error
	
	if err != nil {
		fmt.Printf("   ❌ Error checking sales: %v\n", err)
		return 1
	}
	
	if len(results) == 0 {
		fmt.Println("   ✅ All sales have journal entries")
		return 0
	}
	
	fmt.Printf("   ⚠️  WARNING: Found %d sales without journals:\n", len(results))
	for _, r := range results {
		fmt.Printf("      - Sale %s (%s) Amount: %.2f Date: %s\n", 
			r.SaleCode, r.Status, r.TotalAmount, r.CreatedAt.Format("2006-01-02"))
	}
	return len(results)
}

func checkMissingPurchaseJournals(db *gorm.DB) int {
	fmt.Println("\n4️⃣  Checking for missing purchase journals...")
	
	type MissingResult struct {
		PurchaseID  uint
		PurchaseCode string
		Status      string
		TotalAmount float64
		CreatedAt   time.Time
	}
	
	var results []MissingResult
	err := db.Raw(`
		SELECT 
			p.id as purchase_id,
			p.code as purchase_code,
			p.status,
			p.total_amount,
			p.created_at
		FROM purchases p
		LEFT JOIN unified_journal_ledger ujl 
			ON ujl.source_type = 'PURCHASE' 
			AND ujl.source_id = p.id 
			AND ujl.deleted_at IS NULL
		WHERE p.status = 'APPROVED'
		  AND p.deleted_at IS NULL
		  AND ujl.id IS NULL
		ORDER BY p.created_at DESC
		LIMIT 10
	`).Scan(&results).Error
	
	if err != nil {
		fmt.Printf("   ❌ Error checking purchases: %v\n", err)
		return 1
	}
	
	if len(results) == 0 {
		fmt.Println("   ✅ All purchases have journal entries")
		return 0
	}
	
	fmt.Printf("   ⚠️  WARNING: Found %d purchases without journals:\n", len(results))
	for _, r := range results {
		fmt.Printf("      - Purchase %s (%s) Amount: %.2f Date: %s\n", 
			r.PurchaseCode, r.Status, r.TotalAmount, r.CreatedAt.Format("2006-01-02"))
	}
	return len(results)
}

func checkJournalsWithoutLines(db *gorm.DB) int {
	fmt.Println("\n5️⃣  Checking for journals without lines...")
	
	type OrphanResult struct {
		ID          uint64
		EntryNumber string
		SourceType  string
		Status      string
	}
	
	var results []OrphanResult
	err := db.Raw(`
		SELECT 
			ujl.id,
			ujl.entry_number,
			ujl.source_type,
			ujl.status
		FROM unified_journal_ledger ujl
		LEFT JOIN unified_journal_lines ujll ON ujll.journal_id = ujl.id
		WHERE ujl.deleted_at IS NULL
		  AND ujl.status = 'POSTED'
		GROUP BY ujl.id, ujl.entry_number, ujl.source_type, ujl.status
		HAVING COUNT(ujll.id) = 0
		ORDER BY ujl.id DESC
		LIMIT 10
	`).Scan(&results).Error
	
	if err != nil {
		fmt.Printf("   ❌ Error checking orphans: %v\n", err)
		return 1
	}
	
	if len(results) == 0 {
		fmt.Println("   ✅ All journals have lines")
		return 0
	}
	
	fmt.Printf("   ⚠️  WARNING: Found %d orphaned journals:\n", len(results))
	for _, r := range results {
		fmt.Printf("      - Journal %s (%s) has no lines\n", r.EntryNumber, r.SourceType)
	}
	return len(results)
}

func checkHeaderLineMismatch(db *gorm.DB) int {
	fmt.Println("\n6️⃣  Checking for header-line total mismatches...")
	
	type MismatchResult struct {
		JournalID     uint64
		EntryNumber   string
		HeaderDebit   float64
		LinesDebit    float64
		HeaderCredit  float64
		LinesCredit   float64
		DebitDiff     float64
		CreditDiff    float64
	}
	
	var results []MismatchResult
	err := db.Raw(`
		WITH line_totals AS (
			SELECT 
				journal_id,
				SUM(debit_amount) as total_debit_lines,
				SUM(credit_amount) as total_credit_lines
			FROM unified_journal_lines
			GROUP BY journal_id
		)
		SELECT 
			ujl.id as journal_id,
			ujl.entry_number,
			ujl.total_debit as header_debit,
			lt.total_debit_lines as lines_debit,
			ujl.total_credit as header_credit,
			lt.total_credit_lines as lines_credit,
			(ujl.total_debit - lt.total_debit_lines) as debit_diff,
			(ujl.total_credit - lt.total_credit_lines) as credit_diff
		FROM unified_journal_ledger ujl
		JOIN line_totals lt ON lt.journal_id = ujl.id
		WHERE (ABS(ujl.total_debit - lt.total_debit_lines) > 0.01
		   OR ABS(ujl.total_credit - lt.total_credit_lines) > 0.01)
		  AND ujl.deleted_at IS NULL
		ORDER BY ujl.id DESC
		LIMIT 10
	`).Scan(&results).Error
	
	if err != nil {
		fmt.Printf("   ❌ Error checking mismatches: %v\n", err)
		return 1
	}
	
	if len(results) == 0 {
		fmt.Println("   ✅ All header totals match line totals")
		return 0
	}
	
	fmt.Printf("   ❌ CRITICAL: Found %d mismatched journals:\n", len(results))
	for _, r := range results {
		fmt.Printf("      - Journal %s: Debit diff=%.2f Credit diff=%.2f\n", 
			r.EntryNumber, r.DebitDiff, r.CreditDiff)
	}
	return len(results)
}
