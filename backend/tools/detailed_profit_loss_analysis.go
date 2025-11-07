package main

import (
	"fmt"
	"log"
	"strings"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DetailedTransaction represents individual transactions contributing to P&L
type DetailedTransaction struct {
	JournalID       uint      `json:"journal_id"`
	EntryNumber     string    `json:"entry_number"`
	EntryDate       time.Time `json:"entry_date"`
	Description     string    `json:"description"`
	SourceType      string    `json:"source_type"`
	AccountID       uint      `json:"account_id"`
	AccountCode     string    `json:"account_code"`
	AccountName     string    `json:"account_name"`
	DebitAmount     float64   `json:"debit_amount"`
	CreditAmount    float64   `json:"credit_amount"`
	NetAmount       float64   `json:"net_amount"`
}

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("📊 DETAILED PROFIT & LOSS ANALYSIS FROM SSOT JOURNAL SYSTEM")
	fmt.Println(strings.Repeat("=", 80))

	// Set period
	currentYear := time.Now().Year()
	periodStart := fmt.Sprintf("%d-01-01", currentYear)
	periodEnd := fmt.Sprintf("%d-12-31", currentYear)
	
	fmt.Printf("📅 Analysis Period: %s to %s\n\n", periodStart, periodEnd)

	// Get detailed transactions for P&L accounts
	fmt.Println("🔍 ANALYZING REVENUE TRANSACTIONS (4xxx accounts)")
	revenueTransactions := getDetailedTransactions(db, "4", periodStart, periodEnd)
	displayTransactionDetails("💰 REVENUE DETAILS", revenueTransactions)

	fmt.Println("🔍 ANALYZING COST OF GOODS SOLD TRANSACTIONS (51xx accounts)")
	cogsTransactions := getDetailedTransactions(db, "51", periodStart, periodEnd)
	displayTransactionDetails("📦 COST OF GOODS SOLD DETAILS", cogsTransactions)

	fmt.Println("🔍 ANALYZING OPERATING EXPENSE TRANSACTIONS (52xx-59xx accounts)")
	expenseTransactions := getOperatingExpenseTransactions(db, periodStart, periodEnd)
	displayTransactionDetails("💸 OPERATING EXPENSES DETAILS", expenseTransactions)

	fmt.Println("🔍 ANALYZING OTHER INCOME/EXPENSE TRANSACTIONS (6xxx-7xxx accounts)")
	otherTransactions := getOtherIncomeExpenseTransactions(db, periodStart, periodEnd)
	displayTransactionDetails("🔄 OTHER INCOME/EXPENSES DETAILS", otherTransactions)

	// Summary by source type
	fmt.Println("📋 TRANSACTION SUMMARY BY SOURCE TYPE")
	summarizeBySourceType(db, periodStart, periodEnd)

	// Monthly breakdown
	fmt.Println("📅 MONTHLY REVENUE BREAKDOWN")
	monthlyRevenue(db, currentYear)
}

func getDetailedTransactions(db *gorm.DB, accountPrefix, startDate, endDate string) []DetailedTransaction {
	var transactions []DetailedTransaction
	
	query := `
		SELECT 
			uje.id as journal_id,
			uje.entry_number,
			uje.entry_date,
			uje.description,
			uje.source_type,
			ujl.account_id,
			a.code as account_code,
			a.name as account_name,
			ujl.debit_amount,
			ujl.credit_amount,
			CASE 
				WHEN a.type IN ('REVENUE') THEN ujl.credit_amount - ujl.debit_amount
				WHEN a.type IN ('EXPENSE') THEN ujl.debit_amount - ujl.credit_amount
				ELSE ujl.debit_amount - ujl.credit_amount
			END as net_amount
		FROM unified_journal_ledger uje
		JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
		JOIN accounts a ON a.id = ujl.account_id
		WHERE uje.status = 'POSTED'
			AND uje.entry_date >= ?
			AND uje.entry_date <= ?
			AND a.code LIKE ?
			AND (ujl.debit_amount > 0 OR ujl.credit_amount > 0)
		ORDER BY uje.entry_date DESC, uje.entry_number, ujl.line_number
	`
	
	if err := db.Raw(query, startDate, endDate, accountPrefix+"%").Scan(&transactions).Error; err != nil {
		log.Printf("Error getting detailed transactions: %v", err)
	}

	return transactions
}

func getOperatingExpenseTransactions(db *gorm.DB, startDate, endDate string) []DetailedTransaction {
	var transactions []DetailedTransaction
	
	query := `
		SELECT 
			uje.id as journal_id,
			uje.entry_number,
			uje.entry_date,
			uje.description,
			uje.source_type,
			ujl.account_id,
			a.code as account_code,
			a.name as account_name,
			ujl.debit_amount,
			ujl.credit_amount,
			ujl.debit_amount - ujl.credit_amount as net_amount
		FROM unified_journal_ledger uje
		JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
		JOIN accounts a ON a.id = ujl.account_id
		WHERE uje.status = 'POSTED'
			AND uje.entry_date >= ?
			AND uje.entry_date <= ?
			AND (a.code LIKE '52%' OR a.code LIKE '53%' OR a.code LIKE '54%' 
				OR a.code LIKE '55%' OR a.code LIKE '56%' OR a.code LIKE '57%' 
				OR a.code LIKE '58%' OR a.code LIKE '59%')
			AND (ujl.debit_amount > 0 OR ujl.credit_amount > 0)
		ORDER BY uje.entry_date DESC, uje.entry_number, ujl.line_number
	`
	
	if err := db.Raw(query, startDate, endDate).Scan(&transactions).Error; err != nil {
		log.Printf("Error getting operating expense transactions: %v", err)
	}

	return transactions
}

func getOtherIncomeExpenseTransactions(db *gorm.DB, startDate, endDate string) []DetailedTransaction {
	var transactions []DetailedTransaction
	
	query := `
		SELECT 
			uje.id as journal_id,
			uje.entry_number,
			uje.entry_date,
			uje.description,
			uje.source_type,
			ujl.account_id,
			a.code as account_code,
			a.name as account_name,
			ujl.debit_amount,
			ujl.credit_amount,
			CASE 
				WHEN a.code LIKE '6%' THEN ujl.debit_amount - ujl.credit_amount  -- Expenses
				WHEN a.code LIKE '7%' THEN ujl.credit_amount - ujl.debit_amount -- Income
				ELSE ujl.debit_amount - ujl.credit_amount
			END as net_amount
		FROM unified_journal_ledger uje
		JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
		JOIN accounts a ON a.id = ujl.account_id
		WHERE uje.status = 'POSTED'
			AND uje.entry_date >= ?
			AND uje.entry_date <= ?
			AND (a.code LIKE '6%' OR a.code LIKE '7%')
			AND (ujl.debit_amount > 0 OR ujl.credit_amount > 0)
		ORDER BY uje.entry_date DESC, uje.entry_number, ujl.line_number
	`
	
	if err := db.Raw(query, startDate, endDate).Scan(&transactions).Error; err != nil {
		log.Printf("Error getting other income/expense transactions: %v", err)
	}

	return transactions
}

func displayTransactionDetails(title string, transactions []DetailedTransaction) {
	fmt.Println("\n" + title)
	fmt.Println(strings.Repeat("-", 80))
	
	if len(transactions) == 0 {
		fmt.Println("❌ No transactions found in this category")
		return
	}

	var totalAmount float64
	for _, tx := range transactions {
		totalAmount += tx.NetAmount
		
		fmt.Printf("📝 %s | %s | %s (%s)\n", 
			tx.EntryDate.Format("2006-01-02"), 
			tx.EntryNumber, 
			tx.AccountCode, 
			tx.AccountName)
		fmt.Printf("   %s | Source: %s | Amount: %s\n", 
			tx.Description, 
			tx.SourceType, 
			formatCurrency(tx.NetAmount))
		fmt.Printf("   Dr: %s, Cr: %s\n\n", 
			formatCurrency(tx.DebitAmount), 
			formatCurrency(tx.CreditAmount))
	}
	
	fmt.Printf("💰 TOTAL for %s: %s\n", title, formatCurrency(totalAmount))
	fmt.Printf("📊 Number of transactions: %d\n", len(transactions))
}

func summarizeBySourceType(db *gorm.DB, startDate, endDate string) {
	fmt.Println(strings.Repeat("-", 80))

	var summary []struct {
		SourceType string  `json:"source_type"`
		Count      int64   `json:"count"`
		TotalDebit float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
	}
	
	query := `
		SELECT 
			COALESCE(uje.source_type, 'MANUAL') as source_type,
			COUNT(*) as count,
			SUM(ujl.debit_amount) as total_debit,
			SUM(ujl.credit_amount) as total_credit
		FROM unified_journal_ledger uje
		JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
		WHERE uje.status = 'POSTED'
			AND uje.entry_date >= ?
			AND uje.entry_date <= ?
		GROUP BY uje.source_type
		ORDER BY SUM(ujl.debit_amount) + SUM(ujl.credit_amount) DESC
	`
	
	if err := db.Raw(query, startDate, endDate).Scan(&summary).Error; err != nil {
		log.Printf("Error getting summary by source type: %v", err)
		return
	}

	for _, s := range summary {
		fmt.Printf("📋 %s: %d entries | Dr: %s | Cr: %s | Net: %s\n",
			s.SourceType, 
			s.Count,
			formatCurrency(s.TotalDebit),
			formatCurrency(s.TotalCredit),
			formatCurrency(s.TotalDebit - s.TotalCredit))
	}
}

func monthlyRevenue(db *gorm.DB, year int) {
	fmt.Println(strings.Repeat("-", 80))

	var monthlyData []struct {
		Month   int     `json:"month"`
		Revenue float64 `json:"revenue"`
		Count   int64   `json:"count"`
	}
	
	query := `
		SELECT 
			EXTRACT(MONTH FROM uje.entry_date) as month,
			SUM(ujl.credit_amount - ujl.debit_amount) as revenue,
			COUNT(DISTINCT uje.id) as count
		FROM unified_journal_ledger uje
		JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
		JOIN accounts a ON a.id = ujl.account_id
		WHERE uje.status = 'POSTED'
			AND EXTRACT(YEAR FROM uje.entry_date) = ?
			AND a.code LIKE '4%'
		GROUP BY EXTRACT(MONTH FROM uje.entry_date)
		ORDER BY month
	`
	
	if err := db.Raw(query, year).Scan(&monthlyData).Error; err != nil {
		log.Printf("Error getting monthly revenue: %v", err)
		return
	}

	monthNames := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", 
					    "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	var totalRevenue float64
	for _, m := range monthlyData {
		totalRevenue += m.Revenue
		fmt.Printf("📅 %s %d: %s (%d entries)\n", 
			monthNames[m.Month], 
			year,
			formatCurrency(m.Revenue),
			m.Count)
	}
	
	if len(monthlyData) > 0 {
		avgMonthly := totalRevenue / float64(len(monthlyData))
		fmt.Printf("💰 Total Revenue: %s\n", formatCurrency(totalRevenue))
		fmt.Printf("📊 Average Monthly: %s\n", formatCurrency(avgMonthly))
	}
}

func formatCurrency(amount float64) string {
	if amount >= 0 {
		return formatNumber(amount)
	} else {
		return fmt.Sprintf("(%s)", formatNumber(-amount))
	}
}

func formatNumber(num float64) string {
	// Convert to string with 2 decimal places
	str := fmt.Sprintf("%.2f", num)
	
	// Split integer and decimal parts
	parts := strings.Split(str, ".")
	intPart := parts[0]
	decPart := parts[1]
	
	// Add thousand separators
	if len(intPart) > 3 {
		// Reverse the string to add commas from right
		runes := []rune(intPart)
		for i := len(runes) - 3; i > 0; i -= 3 {
			runes = append(runes[:i], append([]rune{','}, runes[i:]...)...)
		}
		intPart = string(runes)
	}
	
	return intPart + "." + decPart
}