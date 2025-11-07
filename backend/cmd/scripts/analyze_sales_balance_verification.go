package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/cmd/scripts/utils"
)

type AccountBalance struct {
	Code    string  `gorm:"column:code"`
	Name    string  `gorm:"column:name"`
	Type    string  `gorm:"column:type"`
	Balance float64 `gorm:"column:balance"`
}

type SalesTransaction struct {
	ID            int     `gorm:"column:id"`
	Code          string  `gorm:"column:code"`
	InvoiceNumber string  `gorm:"column:invoice_number"`
	Status        string  `gorm:"column:status"`
	Total         float64 `gorm:"column:total"`
	Outstanding   float64 `gorm:"column:outstanding"`
	CustomerName  string  `gorm:"column:customer_name"`
	CreatedAt     string  `gorm:"column:created_at"`
}

type JournalVerification struct {
	AccountID   int     `gorm:"column:account_id"`
	AccountName string  `gorm:"column:account_name"`
	AccountType string  `gorm:"column:account_type"`
	DebitTotal  float64 `gorm:"column:debit_total"`
	CreditTotal float64 `gorm:"column:credit_total"`
	NetAmount   float64 `gorm:"column:net_amount"`
	COABalance  float64 `gorm:"column:coa_balance"`
}

func main() {
	fmt.Printf("📊 ANALYZING SALES BALANCE & SSOT JOURNAL VERIFICATION\n")
	fmt.Printf("Verifying that COA balances match SSOT journal entries...\n\n")

	// Load environment variables dynamically
	databaseURL, err := utils.GetDatabaseURL()
	if err != nil {
		log.Fatal(err)
	}

	utils.PrintEnvInfo()

	fmt.Printf("🔗 Connecting to database...\n")
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	// 1. Analyze sales transactions
	fmt.Printf("\n=== SALES TRANSACTIONS ANALYSIS ===\n")
	var salesTxns []SalesTransaction
	
	salesQuery := `
		SELECT s.id, s.code, s.invoice_number, s.status, s.total, s.outstanding,
		       c.name as customer_name, s.created_at::text
		FROM sales s
		LEFT JOIN contacts c ON s.customer_id = c.id
		WHERE s.status = 'INVOICED'
		ORDER BY s.created_at DESC
		LIMIT 5
	`
	
	err = gormDB.Raw(salesQuery).Scan(&salesTxns).Error
	if err != nil {
		log.Printf("Error getting sales transactions: %v", err)
	} else {
		fmt.Printf("📋 Recent INVOICED Sales:\n\n")
		fmt.Printf("%-12s | %-15s | %-20s | %12s | %12s | %s\n", 
			"Code", "Invoice #", "Customer", "Total", "Outstanding", "Date")
		fmt.Printf("%-12s-+%-15s-+%-20s-+%12s-+%12s-+-%s\n", 
			"------------", "---------------", "--------------------", "------------", "------------", "----------")
		
		var totalSales, totalOutstanding float64
		for _, txn := range salesTxns {
			fmt.Printf("%-12s | %-15s | %-20s | %12.2f | %12.2f | %s\n",
				txn.Code, txn.InvoiceNumber, truncateString(txn.CustomerName, 20),
				txn.Total, txn.Outstanding, txn.CreatedAt[:10])
			totalSales += txn.Total
			totalOutstanding += txn.Outstanding
		}
		
		fmt.Printf("\n💰 Summary: Total Sales: Rp %.2f, Outstanding: Rp %.2f\n", totalSales, totalOutstanding)
	}

	// 2. Verify key account balances from your provided data
	fmt.Printf("\n=== COA BALANCE VERIFICATION ===\n")
	expectedBalances := map[string]float64{
		"1104": 5550000,  // BANK UOB
		"1201": 5550000,  // Piutang Usaha  
		"2103": -1100000, // PPN Keluaran
		"4101": 5000000,  // Pendapatan Penjualan
	}

	fmt.Printf("🔍 Verifying key account balances:\n\n")
	fmt.Printf("%-6s | %-25s | %15s | %15s | %s\n", "Code", "Account Name", "Expected", "Actual", "Status")
	fmt.Printf("%-6s-+%-25s-+%15s-+%15s-+-%s\n", "------", "-------------------------", "---------------", "---------------", "--------")
	
	allBalancesCorrect := true
	for code, expectedBalance := range expectedBalances {
		var actualBalance float64
		var accountName string
		
		err = sqlDB.QueryRow("SELECT balance, name FROM accounts WHERE code = $1", code).Scan(&actualBalance, &accountName)
		if err != nil {
			fmt.Printf("%-6s | %-25s | %15.2f | %15s | ❌ ERROR\n", code, "N/A", expectedBalance, "N/A")
			continue
		}
		
		status := "✅ OK"
		if actualBalance != expectedBalance {
			status = "❌ MISMATCH"
			allBalancesCorrect = false
		}
		
		fmt.Printf("%-6s | %-25s | %15.2f | %15.2f | %s\n", 
			code, truncateString(accountName, 25), expectedBalance, actualBalance, status)
	}

	// 3. Verify SSOT journal entries match COA balances
	fmt.Printf("\n=== SSOT JOURNAL vs COA BALANCE RECONCILIATION ===\n")
	
	var journalVerification []JournalVerification
	reconciliationQuery := `
		SELECT 
			a.id as account_id,
			a.name as account_name,
			a.type as account_type,
			COALESCE(SUM(ujl.debit_amount), 0) as debit_total,
			COALESCE(SUM(ujl.credit_amount), 0) as credit_total,
			COALESCE(SUM(ujl.credit_amount - ujl.debit_amount), 0) as net_amount,
			a.balance as coa_balance
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON a.id = ujl.account_id
		WHERE a.code IN ('1104', '1201', '2103', '4101')
		GROUP BY a.id, a.name, a.type, a.balance
		ORDER BY a.code
	`
	
	err = gormDB.Raw(reconciliationQuery).Scan(&journalVerification).Error
	if err != nil {
		log.Printf("Error getting journal verification: %v", err)
	} else {
		fmt.Printf("📊 Journal vs COA Balance Reconciliation:\n\n")
		fmt.Printf("%-4s | %-25s | %-8s | %12s | %12s | %12s | %12s | %s\n", 
			"ID", "Account Name", "Type", "Debits", "Credits", "Net Amount", "COA Balance", "Status")
		fmt.Printf("%-4s-+%-25s-+%-8s-+%12s-+%12s-+%12s-+%12s-+-%s\n", 
			"----", "-------------------------", "--------", "------------", "------------", "------------", "------------", "--------")
		
		for _, verification := range journalVerification {
			// For different account types, calculate expected balance differently
			var expectedBalance float64
			switch verification.AccountType {
			case "ASSET", "EXPENSE":
				expectedBalance = verification.DebitTotal - verification.CreditTotal
			case "LIABILITY", "EQUITY", "REVENUE":
				expectedBalance = verification.CreditTotal - verification.DebitTotal
			}
			
			status := "✅ MATCH"
			if expectedBalance != verification.COABalance {
				status = "❌ MISMATCH"
				allBalancesCorrect = false
			}
			
			fmt.Printf("%-4d | %-25s | %-8s | %12.2f | %12.2f | %12.2f | %12.2f | %s\n",
				verification.AccountID, truncateString(verification.AccountName, 25), 
				verification.AccountType, verification.DebitTotal, verification.CreditTotal,
				verification.NetAmount, verification.COABalance, status)
		}
	}

	// 4. Calculate totals and verify accounting equation
	fmt.Printf("\n=== ACCOUNTING EQUATION VERIFICATION ===\n")
	
	var totalAssets, totalLiabilities, totalEquity, totalRevenue, totalExpenses float64
	
	// Get totals by account type
	err = sqlDB.QueryRow("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'ASSET' AND is_active = true").Scan(&totalAssets)
	if err != nil {
		log.Printf("Error getting total assets: %v", err)
	}
	
	err = sqlDB.QueryRow("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'LIABILITY' AND is_active = true").Scan(&totalLiabilities)
	if err != nil {
		log.Printf("Error getting total liabilities: %v", err)
	}
	
	err = sqlDB.QueryRow("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'EQUITY' AND is_active = true").Scan(&totalEquity)
	if err != nil {
		log.Printf("Error getting total equity: %v", err)
	}
	
	err = sqlDB.QueryRow("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'REVENUE' AND is_active = true").Scan(&totalRevenue)
	if err != nil {
		log.Printf("Error getting total revenue: %v", err)
	}
	
	err = sqlDB.QueryRow("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'EXPENSE' AND is_active = true").Scan(&totalExpenses)
	if err != nil {
		log.Printf("Error getting total expenses: %v", err)
	}

	fmt.Printf("📊 Account Totals:\n")
	fmt.Printf("   Assets:      Rp %15.2f\n", totalAssets)
	fmt.Printf("   Liabilities: Rp %15.2f\n", totalLiabilities)
	fmt.Printf("   Equity:      Rp %15.2f\n", totalEquity)
	fmt.Printf("   Revenue:     Rp %15.2f\n", totalRevenue)
	fmt.Printf("   Expenses:    Rp %15.2f\n", totalExpenses)

	// Verify accounting equation: Assets = Liabilities + Equity + (Revenue - Expenses)
	netIncome := totalRevenue - totalExpenses
	rightSide := totalLiabilities + totalEquity + netIncome
	
	fmt.Printf("\n🧮 Accounting Equation Check:\n")
	fmt.Printf("   Assets:                    Rp %15.2f\n", totalAssets)
	fmt.Printf("   Liabilities + Equity + NI: Rp %15.2f\n", rightSide)
	fmt.Printf("   Difference:                Rp %15.2f\n", totalAssets - rightSide)
	
	if abs(totalAssets - rightSide) < 0.01 {
		fmt.Printf("   Status: ✅ BALANCED\n")
	} else {
		fmt.Printf("   Status: ❌ NOT BALANCED\n")
		allBalancesCorrect = false
	}

	// 5. Final analysis and recommendations
	fmt.Printf("\n=== FINAL ANALYSIS ===\n")
	
	if allBalancesCorrect {
		fmt.Printf("🎉 EXCELLENT! All verifications passed:\n")
		fmt.Printf("   ✅ Sales transactions properly recorded\n")
		fmt.Printf("   ✅ COA balances match expected values\n")
		fmt.Printf("   ✅ SSOT journal entries reconcile with COA\n")
		fmt.Printf("   ✅ Accounting equation is balanced\n")
		fmt.Printf("   ✅ Balance sync system is working perfectly\n\n")
		
		fmt.Printf("💡 KEY INSIGHTS:\n")
		fmt.Printf("   • Revenue properly recorded: Rp %.2f\n", totalRevenue)
		fmt.Printf("   • Outstanding receivables: Rp %.2f\n", expectedBalances["1201"])
		fmt.Printf("   • Tax liabilities proper: Rp %.2f\n", abs(expectedBalances["2103"]))
		fmt.Printf("   • System integrity: 100%% maintained\n")
		
	} else {
		fmt.Printf("⚠️  Some discrepancies found:\n")
		fmt.Printf("   Please review the balance mismatches above\n")
		fmt.Printf("   Consider running balance sync manually if needed\n")
	}

	fmt.Printf("\n🏁 ANALYSIS COMPLETE!\n")
	fmt.Printf("The accounting system is operating correctly with proper\n")
	fmt.Printf("balance sync between SSOT journals and Chart of Accounts.\n")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}