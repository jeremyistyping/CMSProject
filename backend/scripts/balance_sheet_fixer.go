package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// BalanceSheetAccount represents account balance for balance sheet
type BalanceSheetAccount struct {
	AccountID    uint    `json:"account_id"`
	AccountCode  string  `json:"account_code"`
	AccountName  string  `json:"account_name"`
	AccountType  string  `json:"account_type"`
	DebitTotal   float64 `json:"debit_total"`
	CreditTotal  float64 `json:"credit_total"`
	NetBalance   float64 `json:"net_balance"`
}

// BalanceSheetSummary represents balance sheet totals
type BalanceSheetSummary struct {
	TotalAssets              float64 `json:"total_assets"`
	TotalLiabilities         float64 `json:"total_liabilities"`
	TotalEquity              float64 `json:"total_equity"`
	TotalLiabilitiesAndEquity float64 `json:"total_liabilities_and_equity"`
	BalanceDifference        float64 `json:"balance_difference"`
	IsBalanced               bool    `json:"is_balanced"`
	AsOfDate                 string  `json:"as_of_date"`
}

// AccountFix represents account corrections needed
type AccountFix struct {
	AccountID   uint    `json:"account_id"`
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name"`
	CurrentType string  `json:"current_type"`
	CorrectType string  `json:"correct_type"`
	Issue       string  `json:"issue"`
	Amount      float64 `json:"amount"`
}

func main() {
	fmt.Println("=== BALANCE SHEET FIXER FOR SSOT JOURNAL SYSTEM ===")
	fmt.Println("Script untuk membuat balance sheet seimbang")
	fmt.Println("=" + strings.Repeat("=", 50))
	
	// Setup database connection
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("❌ Error connecting to database: %v", err)
	}

	// Get today's date
	asOfDate := time.Now().Format("2006-01-02")
	fmt.Printf("📅 Analyzing balance sheet as of: %s\n\n", asOfDate)

	// Step 1: Analyze current balance sheet
	fmt.Println("🔍 STEP 1: Analyzing current balance sheet...")
	summary, err := analyzeBalanceSheet(db, asOfDate)
	if err != nil {
		log.Fatalf("❌ Error analyzing balance sheet: %v", err)
	}

	printBalanceSheetSummary(summary)

	// Step 2: Check if balance sheet is already balanced
	if summary.IsBalanced {
		fmt.Println("✅ Balance sheet is already balanced! No fixes needed.")
		return
	}

	// Step 3: Identify account classification issues
	fmt.Println("\n🔧 STEP 2: Identifying account classification issues...")
	fixes, err := identifyAccountIssues(db)
	if err != nil {
		log.Fatalf("❌ Error identifying issues: %v", err)
	}

	if len(fixes) == 0 {
		fmt.Println("ℹ️  No account classification issues found.")
	} else {
		printAccountFixes(fixes)
		
		// Step 4: Apply fixes
		fmt.Println("\n🛠️  STEP 3: Applying account classification fixes...")
		err = applyAccountFixes(db, fixes)
		if err != nil {
			log.Fatalf("❌ Error applying fixes: %v", err)
		}
	}

	// Step 5: Check duplicate journal entries
	fmt.Println("\n🔍 STEP 4: Checking for duplicate journal entries...")
	duplicates, err := findDuplicateEntries(db)
	if err != nil {
		log.Fatalf("❌ Error checking duplicates: %v", err)
	}

	if len(duplicates) > 0 {
		fmt.Printf("⚠️  Found %d potential duplicate entries\n", len(duplicates))
		// Option to fix duplicates would go here
	} else {
		fmt.Println("✅ No duplicate entries found")
	}

	// Step 6: Verify balance sheet after fixes
	fmt.Println("\n🔍 STEP 5: Verifying balance sheet after fixes...")
	finalSummary, err := analyzeBalanceSheet(db, asOfDate)
	if err != nil {
		log.Fatalf("❌ Error verifying balance sheet: %v", err)
	}

	fmt.Println("\n📊 FINAL BALANCE SHEET STATUS:")
	printBalanceSheetSummary(finalSummary)

	if finalSummary.IsBalanced {
		fmt.Println("\n🎉 SUCCESS! Balance sheet is now balanced!")
	} else {
		fmt.Printf("\n⚠️  Balance sheet still shows difference of Rp %.2f\n", finalSummary.BalanceDifference)
		fmt.Println("Additional manual review may be needed.")
	}
}

func setupDatabase() (*gorm.DB, error) {
	// Database connection setup for PostgreSQL
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: time.Second,
				LogLevel:      logger.Silent,
			},
		),
	})
	
	return db, err
}

func analyzeBalanceSheet(db *gorm.DB, asOfDate string) (*BalanceSheetSummary, error) {
	var accounts []BalanceSheetAccount
	
	query := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			COALESCE(SUM(ujl.debit_amount), 0) as debit_total,
			COALESCE(SUM(ujl.credit_amount), 0) as credit_total,
			CASE 
				WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
					COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
				ELSE 
					COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
			END as net_balance
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE (uje.status = 'POSTED' AND uje.entry_date <= ?) OR uje.status IS NULL
		GROUP BY a.id, a.code, a.name, a.type
		HAVING a.type IN ('ASSET', 'LIABILITY', 'EQUITY')
		ORDER BY a.code
	`
	
	if err := db.Raw(query, asOfDate).Scan(&accounts).Error; err != nil {
		return nil, err
	}

	summary := &BalanceSheetSummary{AsOfDate: asOfDate}
	
	// Calculate totals
	for _, account := range accounts {
		switch account.AccountType {
		case "ASSET":
			summary.TotalAssets += account.NetBalance
		case "LIABILITY":
			summary.TotalLiabilities += account.NetBalance
		case "EQUITY":
			summary.TotalEquity += account.NetBalance
		}
	}

	summary.TotalLiabilitiesAndEquity = summary.TotalLiabilities + summary.TotalEquity
	summary.BalanceDifference = summary.TotalAssets - summary.TotalLiabilitiesAndEquity
	summary.IsBalanced = math.Abs(summary.BalanceDifference) <= 0.01
	
	return summary, nil
}

func identifyAccountIssues(db *gorm.DB) ([]AccountFix, error) {
	var fixes []AccountFix
	var accounts []struct {
		ID   uint   `json:"id"`
		Code string `json:"code"`
		Name string `json:"name"`
		Type string `json:"type"`
	}

	// Get all accounts
	if err := db.Table("accounts").Select("id, code, name, type").Scan(&accounts).Error; err != nil {
		return nil, err
	}

	for _, account := range accounts {
		var issue string
		var correctType string
		
		// Check account classification based on Indonesian COA
		switch {
		// Asset accounts (1xxx)
		case account.Code[0] == '1':
			if account.Type != "ASSET" {
				issue = fmt.Sprintf("Account %s should be ASSET type", account.Code)
				correctType = "ASSET"
			}
			
		// Liability accounts (2xxx) - with special handling for PPN Masukan
		case account.Code[0] == '2':
			// Special case: PPN Masukan (2102) should be ASSET
			if account.Code == "2102" || (account.Type == "LIABILITY" && 
				(account.Name == "PPN Masukan" || account.Name == "Pajak Masukan")) {
				if account.Type != "ASSET" {
					issue = "PPN Masukan should be classified as ASSET (current asset)"
					correctType = "ASSET"
				}
			} else if account.Type != "LIABILITY" {
				issue = fmt.Sprintf("Account %s should be LIABILITY type", account.Code)
				correctType = "LIABILITY"
			}
			
		// Equity accounts (3xxx)
		case account.Code[0] == '3':
			if account.Type != "EQUITY" {
				issue = fmt.Sprintf("Account %s should be EQUITY type", account.Code)
				correctType = "EQUITY"
			}
		}

		if issue != "" {
			fixes = append(fixes, AccountFix{
				AccountID:   account.ID,
				AccountCode: account.Code,
				AccountName: account.Name,
				CurrentType: account.Type,
				CorrectType: correctType,
				Issue:       issue,
			})
		}
	}

	return fixes, nil
}

func applyAccountFixes(db *gorm.DB, fixes []AccountFix) error {
	if len(fixes) == 0 {
		return nil
	}

	fmt.Printf("Applying %d account classification fixes...\n", len(fixes))
	
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, fix := range fixes {
		fmt.Printf("  📝 Fixing %s (%s): %s -> %s\n", 
			fix.AccountCode, fix.AccountName, fix.CurrentType, fix.CorrectType)
		
		result := tx.Table("accounts").
			Where("id = ?", fix.AccountID).
			Update("type", fix.CorrectType)
		
		if result.Error != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update account %s: %v", fix.AccountCode, result.Error)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit account fixes: %v", err)
	}

	fmt.Println("✅ All account fixes applied successfully!")
	return nil
}

func findDuplicateEntries(db *gorm.DB) ([]string, error) {
	var duplicates []string
	
	// Check for duplicate PPN Keluaran entries
	var duplicateCheck []struct {
		AccountCode string `json:"account_code"`
		EntryDate   string `json:"entry_date"`
		Amount      float64 `json:"amount"`
		Count       int    `json:"count"`
	}

	query := `
		SELECT 
			a.code as account_code,
			DATE(uje.entry_date) as entry_date,
			ujl.credit_amount as amount,
			COUNT(*) as count
		FROM unified_journal_lines ujl
		JOIN accounts a ON a.id = ujl.account_id
		JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE a.code = '2103' AND uje.status = 'POSTED'
		GROUP BY a.code, DATE(uje.entry_date), ujl.credit_amount
		HAVING COUNT(*) > 1
		ORDER BY entry_date DESC, amount DESC
	`

	if err := db.Raw(query).Scan(&duplicateCheck).Error; err != nil {
		return nil, err
	}

	for _, dup := range duplicateCheck {
		duplicates = append(duplicates, fmt.Sprintf(
			"Account %s on %s: Rp %.0f (appears %d times)",
			dup.AccountCode, dup.EntryDate, dup.Amount, dup.Count))
	}

	return duplicates, nil
}

func printBalanceSheetSummary(summary *BalanceSheetSummary) {
	fmt.Println("📊 Balance Sheet Summary:")
	fmt.Printf("   Total Assets:              Rp %15.0f\n", summary.TotalAssets)
	fmt.Printf("   Total Liabilities:         Rp %15.0f\n", summary.TotalLiabilities)
	fmt.Printf("   Total Equity:              Rp %15.0f\n", summary.TotalEquity)
	fmt.Printf("   Total Liab + Equity:       Rp %15.0f\n", summary.TotalLiabilitiesAndEquity)
	fmt.Printf("   Balance Difference:        Rp %15.0f\n", summary.BalanceDifference)
	
	if summary.IsBalanced {
		fmt.Println("   Status:                    ✅ BALANCED")
	} else {
		fmt.Println("   Status:                    ❌ NOT BALANCED")
	}
}

func printAccountFixes(fixes []AccountFix) {
	fmt.Printf("Found %d account classification issues:\n", len(fixes))
	for i, fix := range fixes {
		fmt.Printf("  %d. %s (%s): %s\n", i+1, fix.AccountCode, fix.AccountName, fix.Issue)
		fmt.Printf("     Current Type: %s -> Correct Type: %s\n", fix.CurrentType, fix.CorrectType)
	}
}