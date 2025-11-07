package main

import (
	"fmt"
	"strings"
	"app-sistem-akuntansi/database"
)

func main() {
	db := database.ConnectDB()
	
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("DEEP ANALYSIS - ALL DATABASE DATA (INCLUDING SOFT-DELETED)")
	fmt.Println(strings.Repeat("=", 80) + "\n")
	
	// 1. Check TOTAL journal entries (including soft-deleted)
	var totalJournals, activeJournals, deletedJournals int64
	db.Raw("SELECT COUNT(*) FROM journal_entries").Scan(&totalJournals)
	db.Raw("SELECT COUNT(*) FROM journal_entries WHERE deleted_at IS NULL").Scan(&activeJournals)
	db.Raw("SELECT COUNT(*) FROM journal_entries WHERE deleted_at IS NOT NULL").Scan(&deletedJournals)
	
	fmt.Println("📊 JOURNAL ENTRIES SUMMARY:")
	fmt.Printf("  Total (including deleted):  %d\n", totalJournals)
	fmt.Printf("  Active (not deleted):       %d\n", activeJournals)
	fmt.Printf("  Soft-Deleted:               %d ⚠️\n", deletedJournals)
	fmt.Println()
	
	// 2. Check TOTAL journal lines (including soft-deleted)
	var totalLines, activeLines, deletedLines int64
	db.Raw("SELECT COUNT(*) FROM journal_lines").Scan(&totalLines)
	db.Raw("SELECT COUNT(*) FROM journal_lines WHERE deleted_at IS NULL").Scan(&activeLines)
	db.Raw("SELECT COUNT(*) FROM journal_lines WHERE deleted_at IS NOT NULL").Scan(&deletedLines)
	
	fmt.Println("📋 JOURNAL LINES SUMMARY:")
	fmt.Printf("  Total (including deleted):  %d\n", totalLines)
	fmt.Printf("  Active (not deleted):       %d\n", activeLines)
	fmt.Printf("  Soft-Deleted:               %d ⚠️\n", deletedLines)
	fmt.Println()
	
	// 3. Show ALL journal entries including soft-deleted
	type JournalEntry struct {
		ID            uint
		Code          string
		EntryDate     string
		Description   string
		ReferenceType string
		Status        string
		TotalDebit    float64
		TotalCredit   float64
		DeletedAt     *string
		LineCount     int
	}
	
	var allJournals []JournalEntry
	db.Raw(`
		SELECT 
			je.id,
			je.code,
			je.entry_date::text,
			SUBSTRING(je.description, 1, 60) as description,
			je.reference_type,
			je.status,
			je.total_debit,
			je.total_credit,
			je.deleted_at::text,
			COUNT(jl.id) as line_count
		FROM journal_entries je
		LEFT JOIN journal_lines jl ON je.id = jl.journal_entry_id AND jl.deleted_at IS NULL
		GROUP BY je.id, je.code, je.entry_date, je.description, je.reference_type, je.status, je.total_debit, je.total_credit, je.deleted_at
		ORDER BY je.id
		LIMIT 50
	`).Scan(&allJournals)
	
	fmt.Printf("📖 ALL JOURNAL ENTRIES (First 50):\n\n")
	fmt.Println("ID   | Code              | Date       | Status | RefType     | Debit      | Credit     | Lines | Deleted?")
	fmt.Println("-----+-------------------+------------+--------+-------------+------------+------------+-------+---------")
	
	for _, j := range allJournals {
		deletedStatus := ""
		if j.DeletedAt != nil {
			deletedStatus = "DELETED"
		}
		fmt.Printf("%-4d | %-17s | %-10s | %-6s | %-11s | %10.0f | %10.0f | %-5d | %s\n",
			j.ID, j.Code, j.EntryDate[:10], j.Status, j.ReferenceType, 
			j.TotalDebit, j.TotalCredit, j.LineCount, deletedStatus)
	}
	
	// 4. Check sales and purchase tables
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TRANSACTION TABLES:")
	fmt.Println(strings.Repeat("=", 80) + "\n")
	
	var salesCount, purchaseCount int64
	db.Raw("SELECT COUNT(*) FROM sales WHERE deleted_at IS NULL").Scan(&salesCount)
	db.Raw("SELECT COUNT(*) FROM purchases WHERE deleted_at IS NULL").Scan(&purchaseCount)
	
	fmt.Printf("💰 SALES:     %d records\n", salesCount)
	fmt.Printf("🛒 PURCHASES: %d records\n", purchaseCount)
	
	// 5. Check if there are journal_entry_id references in sales/purchases
	type SalesJournalRef struct {
		SalesID        uint
		InvoiceNumber  string
		TotalAmount    float64
		JournalEntryID *uint
		JournalCode    *string
		JournalDeleted *string
	}
	
	var salesWithJournals []SalesJournalRef
	db.Raw(`
		SELECT 
			s.id as sales_id,
			s.invoice_number,
			s.total_amount,
			s.journal_entry_id,
			je.code as journal_code,
			je.deleted_at::text as journal_deleted
		FROM sales s
		LEFT JOIN journal_entries je ON s.journal_entry_id = je.id
		WHERE s.deleted_at IS NULL
		ORDER BY s.id
		LIMIT 10
	`).Scan(&salesWithJournals)
	
	fmt.Println("\n📊 SALES JOURNAL REFERENCES (First 10):")
	fmt.Println("Sales ID | Invoice        | Amount       | Journal ID | Journal Code      | Status")
	fmt.Println("---------+----------------+--------------+------------+-------------------+--------")
	
	for _, s := range salesWithJournals {
		status := "NO JOURNAL"
		journalIDStr := "-"
		journalCode := "-"
		
		if s.JournalEntryID != nil {
			journalIDStr = fmt.Sprintf("%d", *s.JournalEntryID)
			if s.JournalCode != nil {
				journalCode = *s.JournalCode
			}
			if s.JournalDeleted != nil {
				status = "DELETED"
			} else {
				status = "OK"
			}
		}
		
		fmt.Printf("%-8d | %-14s | %12.2f | %-10s | %-17s | %s\n",
			s.SalesID, s.InvoiceNumber, s.TotalAmount, journalIDStr, journalCode, status)
	}
	
	// 6. Similar for purchases
	type PurchaseJournalRef struct {
		PurchaseID     uint
		PurchaseNumber string
		TotalAmount    float64
		JournalEntryID *uint
		JournalCode    *string
		JournalDeleted *string
	}
	
	var purchasesWithJournals []PurchaseJournalRef
	db.Raw(`
		SELECT 
			p.id as purchase_id,
			p.purchase_number,
			p.total_amount,
			p.journal_entry_id,
			je.code as journal_code,
			je.deleted_at::text as journal_deleted
		FROM purchases p
		LEFT JOIN journal_entries je ON p.journal_entry_id = je.id
		WHERE p.deleted_at IS NULL
		ORDER BY p.id
		LIMIT 10
	`).Scan(&purchasesWithJournals)
	
	fmt.Println("\n🛒 PURCHASE JOURNAL REFERENCES (First 10):")
	fmt.Println("Purch ID | Number         | Amount       | Journal ID | Journal Code      | Status")
	fmt.Println("---------+----------------+--------------+------------+-------------------+--------")
	
	for _, p := range purchasesWithJournals {
		status := "NO JOURNAL"
		journalIDStr := "-"
		journalCode := "-"
		
		if p.JournalEntryID != nil {
			journalIDStr = fmt.Sprintf("%d", *p.JournalEntryID)
			if p.JournalCode != nil {
				journalCode = *p.JournalCode
			}
			if p.JournalDeleted != nil {
				status = "DELETED"
			} else {
				status = "OK"
			}
		}
		
		fmt.Printf("%-8d | %-14s | %12.2f | %-10s | %-17s | %s\n",
			p.PurchaseID, p.PurchaseNumber, p.TotalAmount, journalIDStr, journalCode, status)
	}
	
	// 7. Check WHY journal entries showing as 2 with deleted_at IS NULL
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🔍 INVESTIGATING WHY ONLY 2 ACTIVE JOURNALS:")
	fmt.Println(strings.Repeat("=", 80) + "\n")
	
	// Check if there's a condition in previous queries
	var journalsWithoutDeletedAt int64
	db.Raw("SELECT COUNT(*) FROM journal_entries WHERE deleted_at IS NULL").Scan(&journalsWithoutDeletedAt)
	fmt.Printf("Journals with deleted_at IS NULL: %d\n", journalsWithoutDeletedAt)
	
	var journalsPosted int64
	db.Raw("SELECT COUNT(*) FROM journal_entries WHERE status = 'POSTED' AND deleted_at IS NULL").Scan(&journalsPosted)
	fmt.Printf("Posted journals (not deleted):    %d\n", journalsPosted)
	
	var journalsDraft int64
	db.Raw("SELECT COUNT(*) FROM journal_entries WHERE status = 'DRAFT' AND deleted_at IS NULL").Scan(&journalsDraft)
	fmt.Printf("Draft journals (not deleted):     %d\n", journalsDraft)
	
	// 8. Show sample of soft-deleted journals
	fmt.Println("\n🗑️  SAMPLE OF SOFT-DELETED JOURNAL ENTRIES:")
	
	var deletedJournalsSample []JournalEntry
	db.Raw(`
		SELECT 
			je.id,
			je.code,
			je.entry_date::text,
			SUBSTRING(je.description, 1, 50) as description,
			je.reference_type,
			je.status,
			je.total_debit,
			je.total_credit,
			je.deleted_at::text
		FROM journal_entries je
		WHERE je.deleted_at IS NOT NULL
		ORDER BY je.id
		LIMIT 20
	`).Scan(&deletedJournalsSample)
	
	if len(deletedJournalsSample) > 0 {
		fmt.Println("\nID   | Code              | Date       | Status | Debit      | Credit     | Deleted At")
		fmt.Println("-----+-------------------+------------+--------+------------+------------+-------------------------")
		for _, j := range deletedJournalsSample {
			deletedTime := ""
			if j.DeletedAt != nil {
				deletedTime = (*j.DeletedAt)[:19]
			}
			fmt.Printf("%-4d | %-17s | %-10s | %-6s | %10.0f | %10.0f | %s\n",
				j.ID, j.Code, j.EntryDate[:10], j.Status, j.TotalDebit, j.TotalCredit, deletedTime)
		}
	} else {
		fmt.Println("  (No soft-deleted journals found)")
	}
	
	// 9. CRITICAL: Recalculate account balances from ALL journal lines (including from deleted journals)
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("⚠️  CRITICAL CHECK: Account Balances vs Journal Lines")
	fmt.Println(strings.Repeat("=", 80) + "\n")
	
	type BalanceCheck struct {
		AccountCode       string
		AccountName       string
		AccountType       string
		CurrentBalance    float64
		CalcFromActive    float64
		CalcFromAll       float64
		DiffFromActive    float64
		DiffFromAll       float64
	}
	
	var balanceChecks []BalanceCheck
	db.Raw(`
		SELECT 
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			a.balance as current_balance,
			-- Calculate from ACTIVE journals only
			CASE 
				WHEN a.type IN ('ASSET', 'EXPENSE') 
				THEN COALESCE(SUM(CASE WHEN je.deleted_at IS NULL THEN jl.debit_amount ELSE 0 END), 0) - 
				     COALESCE(SUM(CASE WHEN je.deleted_at IS NULL THEN jl.credit_amount ELSE 0 END), 0)
				ELSE COALESCE(SUM(CASE WHEN je.deleted_at IS NULL THEN jl.credit_amount ELSE 0 END), 0) - 
				     COALESCE(SUM(CASE WHEN je.deleted_at IS NULL THEN jl.debit_amount ELSE 0 END), 0)
			END as calc_from_active,
			-- Calculate from ALL journals (including deleted)
			CASE 
				WHEN a.type IN ('ASSET', 'EXPENSE') 
				THEN COALESCE(SUM(jl.debit_amount), 0) - COALESCE(SUM(jl.credit_amount), 0)
				ELSE COALESCE(SUM(jl.credit_amount), 0) - COALESCE(SUM(jl.debit_amount), 0)
			END as calc_from_all
		FROM accounts a
		LEFT JOIN journal_lines jl ON a.id = jl.account_id AND jl.deleted_at IS NULL
		LEFT JOIN journal_entries je ON jl.journal_entry_id = je.id
		WHERE a.is_active = true AND COALESCE(a.is_header, false) = false AND a.balance != 0
		GROUP BY a.id, a.code, a.name, a.type, a.balance
		ORDER BY a.code
		LIMIT 20
	`).Scan(&balanceChecks)
	
	fmt.Println("Code | Account Name              | Type     | Current  | Calc(Active) | Calc(All) | Diff")
	fmt.Println("-----+---------------------------+----------+----------+--------------+-----------+----------")
	
	for _, bc := range balanceChecks {
		bc.DiffFromActive = bc.CurrentBalance - bc.CalcFromActive
		bc.DiffFromAll = bc.CurrentBalance - bc.CalcFromAll
		
		fmt.Printf("%-4s | %-25s | %-8s | %8.0f | %12.0f | %9.0f | %8.0f\n",
			bc.AccountCode, bc.AccountName, bc.AccountType, 
			bc.CurrentBalance, bc.CalcFromActive, bc.CalcFromAll, bc.DiffFromActive)
	}
	
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("END OF DEEP ANALYSIS")
	fmt.Println(strings.Repeat("=", 80) + "\n")
}
