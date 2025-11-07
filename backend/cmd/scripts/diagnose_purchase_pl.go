package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PurchaseRecord struct {
	ID              uint      `gorm:"column:id"`
	PurchaseNumber  string    `gorm:"column:purchase_number"`
	VendorName      string    `gorm:"column:vendor_name"`
	TotalAmount     float64   `gorm:"column:total_amount"`
	PPNAmount       float64   `gorm:"column:ppn_amount"`
	GrandTotal      float64   `gorm:"column:grand_total"`
	Status          string    `gorm:"column:status"`
	ApprovalStatus  string    `gorm:"column:approval_status"`
	TransactionDate time.Time `gorm:"column:transaction_date"`
}

type JournalEntry struct {
	JournalID       uint    `gorm:"column:journal_id"`
	SourceType      string  `gorm:"column:source_type"`
	SourceReference string  `gorm:"column:source_reference"`
	EntryDate       string  `gorm:"column:entry_date"`
	Description     string  `gorm:"column:description"`
	AccountID       uint    `gorm:"column:account_id"`
	AccountCode     string  `gorm:"column:account_code"`
	AccountName     string  `gorm:"column:account_name"`
	AccountType     string  `gorm:"column:account_type"`
	DebitAmount     float64 `gorm:"column:debit_amount"`
	CreditAmount    float64 `gorm:"column:credit_amount"`
}

type AccountBalance struct {
	AccountID   uint    `gorm:"column:account_id"`
	Code        string  `gorm:"column:code"`
	Name        string  `gorm:"column:name"`
	Type        string  `gorm:"column:type"`
	Balance     float64 `gorm:"column:balance"`
	TxCount     int64   `gorm:"column:tx_count"`
	TotalDebit  float64 `gorm:"column:total_debit"`
	TotalCredit float64 `gorm:"column:total_credit"`
}

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("DIAGNOSTIC REPORT: Purchase & P&L Analysis")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// 1. Check purchase transaction
	fmt.Println("1. PURCHASE TRANSACTION (PO/2025/10/0003)")
	fmt.Println(strings.Repeat("-", 80))
	var purchases []PurchaseRecord
	db.Table("purchases").
		Where("purchase_number = ?", "PO/2025/10/0003").
		Find(&purchases)

	if len(purchases) == 0 {
		fmt.Println("❌ Purchase transaction NOT FOUND in database")
	} else {
		for _, p := range purchases {
			fmt.Printf("✅ Found Purchase ID: %d\n", p.ID)
			fmt.Printf("   Purchase Number: %s\n", p.PurchaseNumber)
			fmt.Printf("   Vendor: %s\n", p.VendorName)
			fmt.Printf("   Total Amount: Rp %.2f\n", p.TotalAmount)
			fmt.Printf("   PPN Amount: Rp %.2f\n", p.PPNAmount)
			fmt.Printf("   Grand Total: Rp %.2f\n", p.GrandTotal)
			fmt.Printf("   Status: %s\n", p.Status)
			fmt.Printf("   Approval Status: %s\n", p.ApprovalStatus)
			fmt.Printf("   Date: %s\n", p.TransactionDate.Format("2006-01-02"))
		}
	}
	fmt.Println()

	// 2. Check journal entries for this purchase
	fmt.Println("2. JOURNAL ENTRIES FOR PO/2025/10/0003")
	fmt.Println(strings.Repeat("-", 80))
	var journals []JournalEntry
	db.Table("unified_journal_ledger j").
		Select(`j.id as journal_id, j.source_type, j.source_reference, j.entry_date::text, 
				j.description, ji.account_id, a.code as account_code, a.name as account_name, 
				a.type as account_type, ji.debit_amount, ji.credit_amount`).
		Joins("LEFT JOIN unified_journal_lines ji ON j.id = ji.journal_id").
		Joins("LEFT JOIN accounts a ON ji.account_id = a.id").
		Where("j.source_reference LIKE ?", "%0003%").
		Where("j.deleted_at IS NULL").
		Order("j.entry_date DESC, ji.debit_amount DESC").
		Find(&journals)

	if len(journals) == 0 {
		fmt.Println("❌ NO JOURNAL ENTRIES found for this purchase!")
		fmt.Println("   This is the ROOT CAUSE - Purchase tidak ter-record ke journal system")
	} else {
		fmt.Printf("✅ Found %d journal line items:\n\n", len(journals))
		var totalDebit, totalCredit float64
		for _, j := range journals {
			fmt.Printf("   Account: %s - %s (%s)\n", j.AccountCode, j.AccountName, j.AccountType)
			fmt.Printf("   Debit:  Rp %12.2f\n", j.DebitAmount)
			fmt.Printf("   Credit: Rp %12.2f\n", j.CreditAmount)
			fmt.Printf("   Description: %s\n", j.Description)
			fmt.Println()
			totalDebit += j.DebitAmount
			totalCredit += j.CreditAmount
		}
		fmt.Printf("   Total Debit:  Rp %.2f\n", totalDebit)
		fmt.Printf("   Total Credit: Rp %.2f\n", totalCredit)
		fmt.Printf("   Balanced: %v\n", totalDebit == totalCredit)
	}
	fmt.Println()

	// 3. Check all accounts starting with 6 (expense accounts)
	fmt.Println("3. EXPENSE ACCOUNTS (6xxx) - BALANCES & TRANSACTIONS")
	fmt.Println(strings.Repeat("-", 80))
	var expenseAccounts []AccountBalance
	db.Table("accounts a").
		Select(`a.id as account_id, a.code, a.name, a.type, a.balance, 
				COUNT(ji.id) as tx_count,
				COALESCE(SUM(ji.debit_amount), 0) as total_debit,
				COALESCE(SUM(ji.credit_amount), 0) as total_credit`).
		Joins("LEFT JOIN unified_journal_lines ji ON ji.account_id = a.id AND ji.deleted_at IS NULL").
		Where("a.code LIKE ?", "6%").
		Where("a.deleted_at IS NULL").
		Group("a.id, a.code, a.name, a.type, a.balance").
		Order("a.code").
		Find(&expenseAccounts)

	if len(expenseAccounts) == 0 {
		fmt.Println("❌ No expense accounts (6xxx) found")
	} else {
		fmt.Printf("Found %d expense accounts:\n\n", len(expenseAccounts))
		for _, acc := range expenseAccounts {
			fmt.Printf("   %s - %s\n", acc.Code, acc.Name)
			fmt.Printf("   Type: %s | Balance: Rp %.2f\n", acc.Type, acc.Balance)
			fmt.Printf("   Transactions: %d | Debit: Rp %.2f | Credit: Rp %.2f\n", 
				acc.TxCount, acc.TotalDebit, acc.TotalCredit)
			if acc.TxCount > 0 {
				fmt.Println("   ✅ HAS TRANSACTIONS")
			} else {
				fmt.Println("   ⚠️  NO TRANSACTIONS")
			}
			fmt.Println()
		}
	}

	// 4. Check all accounts starting with 5 (COGS accounts)
	fmt.Println("4. COGS ACCOUNTS (5xxx) - BALANCES & TRANSACTIONS")
	fmt.Println(strings.Repeat("-", 80))
	var cogsAccounts []AccountBalance
	db.Table("accounts a").
		Select(`a.id as account_id, a.code, a.name, a.type, a.balance, 
				COUNT(ji.id) as tx_count,
				COALESCE(SUM(ji.debit_amount), 0) as total_debit,
				COALESCE(SUM(ji.credit_amount), 0) as total_credit`).
		Joins("LEFT JOIN unified_journal_lines ji ON ji.account_id = a.id AND ji.deleted_at IS NULL").
		Where("a.code LIKE ?", "5%").
		Where("a.deleted_at IS NULL").
		Group("a.id, a.code, a.name, a.type, a.balance").
		Order("a.code").
		Find(&cogsAccounts)

	if len(cogsAccounts) == 0 {
		fmt.Println("❌ No COGS accounts (5xxx) found")
	} else {
		fmt.Printf("Found %d COGS accounts:\n\n", len(cogsAccounts))
		for _, acc := range cogsAccounts {
			fmt.Printf("   %s - %s\n", acc.Code, acc.Name)
			fmt.Printf("   Type: %s | Balance: Rp %.2f\n", acc.Type, acc.Balance)
			fmt.Printf("   Transactions: %d | Debit: Rp %.2f | Credit: Rp %.2f\n", 
				acc.TxCount, acc.TotalDebit, acc.TotalCredit)
			if acc.TxCount > 0 {
				fmt.Println("   ✅ HAS TRANSACTIONS")
			} else {
				fmt.Println("   ⚠️  NO TRANSACTIONS")
			}
			fmt.Println()
		}
	}

	// 5. Summary and recommendations
	fmt.Println("5. DIAGNOSIS SUMMARY & RECOMMENDATIONS")
	fmt.Println(strings.Repeat("=", 80))
	
	if len(journals) == 0 {
		fmt.Println("❌ ROOT CAUSE IDENTIFIED:")
		fmt.Println("   Purchase transaction exists but NO JOURNAL ENTRIES created!")
		fmt.Println()
		fmt.Println("📋 RECOMMENDED ACTIONS:")
		fmt.Println("   1. Check purchase_journal_service to ensure it creates journal entries")
		fmt.Println("   2. Manually create journal entry for this purchase")
		fmt.Println("   3. Re-generate P&L report after fixing")
	} else {
		hasExpenseDebit := false
		hasCOGSDebit := false
		
		for _, j := range journals {
			if len(j.AccountCode) > 0 && j.AccountCode[0] == '6' && j.DebitAmount > 0 {
				hasExpenseDebit = true
			}
			if len(j.AccountCode) > 0 && j.AccountCode[0] == '5' && j.DebitAmount > 0 {
				hasCOGSDebit = true
			}
		}
		
		if hasExpenseDebit {
			fmt.Println("✅ Purchase recorded to Expense account (6xxx)")
			fmt.Println("⚠️  Issue: P&L service may not categorize 6xxx correctly")
			fmt.Println()
			fmt.Println("📋 RECOMMENDED ACTIONS:")
			fmt.Println("   1. Fix P&L categorization for account 6xxx")
			fmt.Println("   2. Include 'Other Expenses' in Total Expenses calculation")
		} else if hasCOGSDebit {
			fmt.Println("✅ Purchase recorded to COGS account (5xxx)")
			fmt.Println("   This is correct for P&L reporting")
		} else {
			fmt.Println("⚠️  Purchase may be recorded to Inventory (1xxx)")
			fmt.Println()
			fmt.Println("📋 RECOMMENDED ACTIONS:")
			fmt.Println("   1. Create COGS journal when goods are sold")
			fmt.Println("   2. Or change purchase recording to expense directly")
		}
	}
	
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("End of Diagnostic Report")
	fmt.Println(strings.Repeat("=", 80))
}

