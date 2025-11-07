package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
)

func main() {
	log.Println("🔍 SSOT Journal Inspection by Account Code (read-only)")

	db := database.ConnectDB()

	targetCodes := []string{"3101", "3000", "1104"} // Modal Pemilik, Equity header, Bank UOB
	log.Printf("Inspecting account codes: %v\n", targetCodes)

	// 1) Summary per account (posted only)
	type Summary struct {
		AccountID    uint
		AccountCode  string
		AccountName  string
		Lines        int64
		TotalDebit   float64
		TotalCredit  float64
	}

	var summaries []Summary
	err := db.Raw(`
		SELECT a.id as account_id,
		       a.code as account_code,
		       a.name as account_name,
		       COUNT(ujl.id) as lines,
		       COALESCE(SUM(ujl.debit_amount),0) as total_debit,
		       COALESCE(SUM(ujl.credit_amount),0) as total_credit
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE a.code IN ? AND (uje.status = 'POSTED' OR uje.status IS NULL)
		GROUP BY a.id, a.code, a.name
		ORDER BY a.code
	`, targetCodes).Scan(&summaries).Error
	if err != nil {
		log.Fatalf("failed to query summaries: %v", err)
	}

	fmt.Println("\n📊 Summary (POSTED only):")
	for _, s := range summaries {
		fmt.Printf("  %s %s → lines=%d, Dr=%.2f, Cr=%.2f, Net=%.2f\n",
			s.AccountCode, s.AccountName, s.Lines, s.TotalDebit, s.TotalCredit, s.TotalDebit-s.TotalCredit)
	}

	// 2) Last 20 posted journal lines for 3101 (Modal Pemilik)
	type Line struct {
		JournalID    uint
		EntryNumber  string
		SourceType   string
		SourceID     *uint
		EntryDate    time.Time
		Debit        float64
		Credit       float64
		Description  string
	}
	var lines []Line
	err = db.Raw(`
		SELECT uje.id as journal_id,
		       uje.entry_number,
		       uje.source_type,
		       uje.source_id,
		       uje.entry_date,
		       ujl.debit_amount as debit,
		       ujl.credit_amount as credit,
		       COALESCE(ujl.description, uje.description) as description
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		JOIN accounts a ON a.id = ujl.account_id
		WHERE a.code = '3101' AND uje.status = 'POSTED'
		ORDER BY uje.entry_date DESC, uje.id DESC
		LIMIT 20
	`).Scan(&lines).Error
	if err != nil {
		log.Fatalf("failed to query lines: %v", err)
	}

	fmt.Println("\n🧾 Last 20 posted journal lines for 3101 (Modal Pemilik):")
	if len(lines) == 0 {
		fmt.Println("  (no lines)")
	}
	for _, l := range lines {
		ref := "-"
		if l.SourceID != nil {
			ref = fmt.Sprintf("%d", *l.SourceID)
		}
		fmt.Printf("  #%s %s ref=%s date=%s Dr=%.2f Cr=%.2f | %s\n",
			l.EntryNumber, l.SourceType, ref, l.EntryDate.Format("2006-01-02"), l.Debit, l.Credit, l.Description)
	}

	// 3) Direct accounts table balances for comparison
	type AccRow struct {
		Code     string
		Name     string
		Balance  float64
		IsHeader bool
	}
	var accRows []AccRow
	err = db.Raw(`
		SELECT code, name, COALESCE(balance,0) as balance, COALESCE(is_header,false) as is_header
		FROM accounts
		WHERE code IN ?
		ORDER BY code
	`, targetCodes).Scan(&accRows).Error
	if err != nil {
		log.Fatalf("failed to query accounts table: %v", err)
	}
	fmt.Println("\n📚 Accounts table (direct balance field):")
	for _, r := range accRows {
		fmt.Printf("  %s %s → balance=%.2f header=%t\n", r.Code, r.Name, r.Balance, r.IsHeader)
	}

	fmt.Println("\n✅ Read-only inspection completed")
}
