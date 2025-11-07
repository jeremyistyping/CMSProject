package main

import (
	"fmt"
	"app-sistem-akuntansi/database"
)

func main() {
	db := database.ConnectDB()
	
	fmt.Println("\n=== CHECKING ALL JOURNAL ENTRIES ===\n")
	
	// 1. Count total journal entries
	var totalCount int64
	db.Raw("SELECT COUNT(*) FROM journal_entries").Scan(&totalCount)
	fmt.Printf("Total journal entries in database: %d\n\n", totalCount)
	
	// 2. Count by status
	type StatusCount struct {
		Status string
		Count  int64
	}
	
	var statusCounts []StatusCount
	db.Raw(`
		SELECT status, COUNT(*) as count 
		FROM journal_entries 
		GROUP BY status 
		ORDER BY count DESC
	`).Scan(&statusCounts)
	
	fmt.Println("Journal entries by status:")
	for _, sc := range statusCounts {
		fmt.Printf("  %s: %d\n", sc.Status, sc.Count)
	}
	
	// 3. Count journal lines
	var linesCount int64
	db.Raw("SELECT COUNT(*) FROM journal_lines").Scan(&linesCount)
	fmt.Printf("\nTotal journal lines in database: %d\n\n", linesCount)
	
	// 4. Show recent journal entries
	type JournalSummary struct {
		ID            uint
		Code          string
		EntryDate     string
		Description   string
		ReferenceType string
		Status        string
		TotalDebit    float64
		TotalCredit   float64
		LineCount     int
	}
	
	var recentJournals []JournalSummary
	db.Raw(`
		SELECT 
			je.id,
			je.code,
			je.entry_date::text as entry_date,
			SUBSTRING(je.description, 1, 50) as description,
			je.reference_type,
			je.status,
			je.total_debit,
			je.total_credit,
			COUNT(jl.id) as line_count
		FROM journal_entries je
		LEFT JOIN journal_lines jl ON je.id = jl.journal_entry_id
		GROUP BY je.id, je.code, je.entry_date, je.description, je.reference_type, je.status, je.total_debit, je.total_credit
		ORDER BY je.entry_date DESC, je.id DESC
		LIMIT 20
	`).Scan(&recentJournals)
	
	fmt.Println("Recent 20 journal entries:")
	fmt.Println("ID   | Code              | Date       | Status | Lines | Debit      | Credit     | Ref Type")
	fmt.Println("-" + "----+-------------------+------------+--------+-------+------------+------------+---------")
	
	for _, j := range recentJournals {
		fmt.Printf("%-4d | %-17s | %-10s | %-6s | %-5d | %10.2f | %10.2f | %s\n",
			j.ID, j.Code, j.EntryDate[:10], j.Status, j.LineCount, j.TotalDebit, j.TotalCredit, j.ReferenceType)
	}
	
	// 5. Check for journal entries without lines
	var orphanCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM journal_entries je
		LEFT JOIN journal_lines jl ON je.id = jl.journal_entry_id
		WHERE jl.id IS NULL
	`).Scan(&orphanCount)
	
	if orphanCount > 0 {
		fmt.Printf("\n⚠️  Warning: %d journal entries have NO journal lines!\n", orphanCount)
	}
	
	// 6. Sum of all posted journal lines by account
	fmt.Println("\n=== POSTED JOURNAL LINES SUMMARY (Top 20 Accounts) ===\n")
	
	type AccountSummary struct {
		AccountCode  string
		AccountName  string
		AccountType  string
		TotalDebit   float64
		TotalCredit  float64
		NetAmount    float64
		LineCount    int
	}
	
	var accountSummaries []AccountSummary
	db.Raw(`
		SELECT 
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			COALESCE(SUM(jl.debit_amount), 0) as total_debit,
			COALESCE(SUM(jl.credit_amount), 0) as total_credit,
			CASE 
				WHEN a.type IN ('ASSET', 'EXPENSE') 
				THEN COALESCE(SUM(jl.debit_amount), 0) - COALESCE(SUM(jl.credit_amount), 0)
				ELSE COALESCE(SUM(jl.credit_amount), 0) - COALESCE(SUM(jl.debit_amount), 0)
			END as net_amount,
			COUNT(jl.id) as line_count
		FROM accounts a
		LEFT JOIN journal_lines jl ON a.id = jl.account_id
		LEFT JOIN journal_entries je ON jl.journal_entry_id = je.id AND je.status = 'POSTED'
		WHERE a.is_active = true AND COALESCE(a.is_header, false) = false
		GROUP BY a.id, a.code, a.name, a.type
		HAVING COUNT(jl.id) > 0
		ORDER BY ABS(COALESCE(SUM(jl.debit_amount), 0) + COALESCE(SUM(jl.credit_amount), 0)) DESC
		LIMIT 20
	`).Scan(&accountSummaries)
	
	fmt.Println("Code | Account Name                  | Type     | Total Debit  | Total Credit | Net Amount   | Lines")
	fmt.Println("-----+-------------------------------+----------+--------------+--------------+--------------+------")
	
	for _, as := range accountSummaries {
		fmt.Printf("%-4s | %-29s | %-8s | %12.2f | %12.2f | %12.2f | %d\n",
			as.AccountCode, as.AccountName, as.AccountType, 
			as.TotalDebit, as.TotalCredit, as.NetAmount, as.LineCount)
	}
	
	// 7. Check for deleted journal entries/lines
	var deletedJournals int64
	var deletedLines int64
	
	db.Raw("SELECT COUNT(*) FROM journal_entries WHERE deleted_at IS NOT NULL").Scan(&deletedJournals)
	db.Raw("SELECT COUNT(*) FROM journal_lines WHERE deleted_at IS NOT NULL").Scan(&deletedLines)
	
	if deletedJournals > 0 || deletedLines > 0 {
		fmt.Printf("\n⚠️  SOFT DELETED DATA:\n")
		fmt.Printf("  - Deleted journal entries: %d\n", deletedJournals)
		fmt.Printf("  - Deleted journal lines: %d\n", deletedLines)
	}
	
	fmt.Println("\n" + "=" + "================================================================")
}
