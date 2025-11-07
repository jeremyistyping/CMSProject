package main

import (
	"fmt"
	"log"
	"strings"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ProfitLossData represents the structure of P&L statement
type ProfitLossData struct {
	CompanyName    string    `json:"company_name"`
	PeriodStart    string    `json:"period_start"`
	PeriodEnd      string    `json:"period_end"`
	GeneratedAt    time.Time `json:"generated_at"`
	
	// Revenue Section
	Revenue struct {
		SalesRevenue      float64 `json:"sales_revenue"`
		ServiceRevenue    float64 `json:"service_revenue"`
		OtherRevenue      float64 `json:"other_revenue"`
		TotalRevenue      float64 `json:"total_revenue"`
	} `json:"revenue"`
	
	// Cost of Goods Sold
	COGS struct {
		DirectMaterials   float64 `json:"direct_materials"`
		DirectLabor       float64 `json:"direct_labor"`
		Manufacturing     float64 `json:"manufacturing"`
		OtherCOGS         float64 `json:"other_cogs"`
		TotalCOGS         float64 `json:"total_cogs"`
	} `json:"cogs"`
	
	GrossProfit       float64 `json:"gross_profit"`
	GrossProfitMargin float64 `json:"gross_profit_margin"`
	
	// Operating Expenses
	OperatingExpenses struct {
		AdminExpenses     float64 `json:"admin_expenses"`
		SellingExpenses   float64 `json:"selling_expenses"`
		GeneralExpenses   float64 `json:"general_expenses"`
		TotalOpEx         float64 `json:"total_opex"`
	} `json:"operating_expenses"`
	
	OperatingIncome   float64 `json:"operating_income"`
	OperatingMargin   float64 `json:"operating_margin"`
	
	// Other Income/Expenses
	OtherIncome       float64 `json:"other_income"`
	OtherExpenses     float64 `json:"other_expenses"`
	
	// Tax and Final Results
	IncomeBeforeTax   float64 `json:"income_before_tax"`
	TaxExpense        float64 `json:"tax_expense"`
	NetIncome         float64 `json:"net_income"`
	NetIncomeMargin   float64 `json:"net_income_margin"`
}

// AccountBalance represents account balance from journal entries
type AccountBalance struct {
	AccountID   uint    `json:"account_id"`
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name"`
	AccountType string  `json:"account_type"`
	DebitTotal  float64 `json:"debit_total"`
	CreditTotal float64 `json:"credit_total"`
	NetBalance  float64 `json:"net_balance"`
}

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("📊 Generating Profit & Loss Statement from SSOT Journal System")
	fmt.Println(strings.Repeat("=", 70))

	// Set period - let's use current year or ask for specific period
	currentYear := time.Now().Year()
	periodStart := fmt.Sprintf("%d-01-01", currentYear)
	periodEnd := fmt.Sprintf("%d-12-31", currentYear)
	
	fmt.Printf("📅 Period: %s to %s\n\n", periodStart, periodEnd)

	// Get account balances from SSOT journal entries
	accountBalances := getAccountBalancesFromSSOT(db, periodStart, periodEnd)
	
	fmt.Printf("🔍 Found %d accounts with transactions in the period\n\n", len(accountBalances))

	// Generate P&L
	pl := generateProfitLoss(accountBalances, periodStart, periodEnd)
	
	// Display the P&L
	displayProfitLoss(pl)
}

func getAccountBalancesFromSSOT(db *gorm.DB, startDate, endDate string) []AccountBalance {
	var balances []AccountBalance
	
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
		WHERE uje.status = 'POSTED' 
			AND uje.entry_date >= ? 
			AND uje.entry_date <= ?
		GROUP BY a.id, a.code, a.name, a.type
		HAVING COALESCE(SUM(ujl.debit_amount), 0) > 0 OR COALESCE(SUM(ujl.credit_amount), 0) > 0
		ORDER BY a.code
	`
	
	if err := db.Raw(query, startDate, endDate).Scan(&balances).Error; err != nil {
		log.Printf("Error getting account balances: %v", err)
		return balances
	}

	// Debug: Show some account details
	fmt.Println("📋 Account Balances Summary:")
	for i, balance := range balances {
		if i < 10 { // Show first 10 accounts
			fmt.Printf("   %s - %s: %.2f\n", balance.AccountCode, balance.AccountName, balance.NetBalance)
		}
	}
	if len(balances) > 10 {
		fmt.Printf("   ... and %d more accounts\n", len(balances)-10)
	}
	fmt.Println()

	return balances
}

func generateProfitLoss(balances []AccountBalance, startDate, endDate string) ProfitLossData {
	pl := ProfitLossData{
		CompanyName: "PT. Sistem Akuntansi",
		PeriodStart: startDate,
		PeriodEnd:   endDate,
		GeneratedAt: time.Now(),
	}

	for _, balance := range balances {
		code := balance.AccountCode
		amount := balance.NetBalance
		
		// Skip if amount is zero
		if amount == 0 {
			continue
		}

		// Categorize accounts based on code ranges
		switch {
		// REVENUE ACCOUNTS (4xxx)
		case strings.HasPrefix(code, "40") || strings.HasPrefix(code, "41"):
			// Sales Revenue
			pl.Revenue.SalesRevenue += amount
			
		case strings.HasPrefix(code, "42"):
			// Service Revenue  
			pl.Revenue.ServiceRevenue += amount
			
		case strings.HasPrefix(code, "49"):
			// Other Revenue
			pl.Revenue.OtherRevenue += amount
			
		// COST OF GOODS SOLD (51xx)
		case strings.HasPrefix(code, "510"):
			// Direct materials, direct COGS
			pl.COGS.DirectMaterials += amount
			
		case strings.HasPrefix(code, "511"):
			// Direct labor
			pl.COGS.DirectLabor += amount
			
		case strings.HasPrefix(code, "512"):
			// Manufacturing overhead
			pl.COGS.Manufacturing += amount
			
		case strings.HasPrefix(code, "513") || strings.HasPrefix(code, "514") || strings.HasPrefix(code, "519"):
			// Other COGS
			pl.COGS.OtherCOGS += amount

		// OPERATING EXPENSES
		case strings.HasPrefix(code, "52"):
			// Administrative expenses (520x-529x)
			pl.OperatingExpenses.AdminExpenses += amount
			
		case strings.HasPrefix(code, "53"):
			// Selling & Marketing expenses (530x-539x)
			pl.OperatingExpenses.SellingExpenses += amount
			
		case strings.HasPrefix(code, "54") || strings.HasPrefix(code, "55") || strings.HasPrefix(code, "56") || 
			 strings.HasPrefix(code, "57") || strings.HasPrefix(code, "58") || strings.HasPrefix(code, "59"):
			// General expenses (540x-599x)
			pl.OperatingExpenses.GeneralExpenses += amount
		
		case strings.HasPrefix(code, "60") || strings.HasPrefix(code, "61"):
			// Operating expenses (60xx-61xx) - including account 6001 Beban Operasional
			pl.OperatingExpenses.GeneralExpenses += amount

		// OTHER INCOME/EXPENSES
		case strings.HasPrefix(code, "62") || strings.HasPrefix(code, "63") || strings.HasPrefix(code, "64") ||
			 strings.HasPrefix(code, "65") || strings.HasPrefix(code, "66") || strings.HasPrefix(code, "67") ||
			 strings.HasPrefix(code, "68") || strings.HasPrefix(code, "69"):
			// Other/Non-operating expenses (62xx-69xx)
			// Other expenses (6xxx)
			pl.OtherExpenses += amount
			
		case strings.HasPrefix(code, "7"):
			// Other income (7xxx)
			pl.OtherIncome += amount
		}
	}

	// Calculate totals and ratios
	pl.Revenue.TotalRevenue = pl.Revenue.SalesRevenue + pl.Revenue.ServiceRevenue + pl.Revenue.OtherRevenue
	pl.COGS.TotalCOGS = pl.COGS.DirectMaterials + pl.COGS.DirectLabor + pl.COGS.Manufacturing + pl.COGS.OtherCOGS
	pl.GrossProfit = pl.Revenue.TotalRevenue - pl.COGS.TotalCOGS
	
	if pl.Revenue.TotalRevenue > 0 {
		pl.GrossProfitMargin = (pl.GrossProfit / pl.Revenue.TotalRevenue) * 100
	}

	pl.OperatingExpenses.TotalOpEx = pl.OperatingExpenses.AdminExpenses + pl.OperatingExpenses.SellingExpenses + pl.OperatingExpenses.GeneralExpenses
	pl.OperatingIncome = pl.GrossProfit - pl.OperatingExpenses.TotalOpEx
	
	if pl.Revenue.TotalRevenue > 0 {
		pl.OperatingMargin = (pl.OperatingIncome / pl.Revenue.TotalRevenue) * 100
	}

	pl.IncomeBeforeTax = pl.OperatingIncome + pl.OtherIncome - pl.OtherExpenses
	
	// Estimate tax expense (assume 25% rate if income is positive)
	if pl.IncomeBeforeTax > 0 {
		pl.TaxExpense = pl.IncomeBeforeTax * 0.25
	}
	
	pl.NetIncome = pl.IncomeBeforeTax - pl.TaxExpense
	
	if pl.Revenue.TotalRevenue > 0 {
		pl.NetIncomeMargin = (pl.NetIncome / pl.Revenue.TotalRevenue) * 100
	}

	return pl
}

func displayProfitLoss(pl ProfitLossData) {
	fmt.Println("📊 PROFIT & LOSS STATEMENT")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Company: %s\n", pl.CompanyName)
	fmt.Printf("Period: %s to %s\n", pl.PeriodStart, pl.PeriodEnd)
	fmt.Printf("Generated: %s\n", pl.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("-", 70))

	// REVENUE SECTION
	fmt.Println("💰 REVENUE")
	fmt.Printf("   Sales Revenue              : %15s\n", formatCurrency(pl.Revenue.SalesRevenue))
	fmt.Printf("   Service Revenue            : %15s\n", formatCurrency(pl.Revenue.ServiceRevenue))
	fmt.Printf("   Other Revenue              : %15s\n", formatCurrency(pl.Revenue.OtherRevenue))
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("   TOTAL REVENUE              : %15s\n", formatCurrency(pl.Revenue.TotalRevenue))
	fmt.Println()

	// COST OF GOODS SOLD
	fmt.Println("📦 COST OF GOODS SOLD")
	fmt.Printf("   Direct Materials           : %15s\n", formatCurrency(pl.COGS.DirectMaterials))
	fmt.Printf("   Direct Labor               : %15s\n", formatCurrency(pl.COGS.DirectLabor))
	fmt.Printf("   Manufacturing Overhead     : %15s\n", formatCurrency(pl.COGS.Manufacturing))
	fmt.Printf("   Other COGS                 : %15s\n", formatCurrency(pl.COGS.OtherCOGS))
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("   TOTAL COGS                 : %15s\n", formatCurrency(pl.COGS.TotalCOGS))
	fmt.Println()

	// GROSS PROFIT
	fmt.Printf("📈 GROSS PROFIT               : %15s", formatCurrency(pl.GrossProfit))
	fmt.Printf(" (%.1f%%)\n", pl.GrossProfitMargin)
	fmt.Println()

	// OPERATING EXPENSES
	fmt.Println("💸 OPERATING EXPENSES")
	fmt.Printf("   Administrative Expenses    : %15s\n", formatCurrency(pl.OperatingExpenses.AdminExpenses))
	fmt.Printf("   Selling & Marketing        : %15s\n", formatCurrency(pl.OperatingExpenses.SellingExpenses))
	fmt.Printf("   General Expenses           : %15s\n", formatCurrency(pl.OperatingExpenses.GeneralExpenses))
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("   TOTAL OPERATING EXPENSES   : %15s\n", formatCurrency(pl.OperatingExpenses.TotalOpEx))
	fmt.Println()

	// OPERATING INCOME
	fmt.Printf("🎯 OPERATING INCOME           : %15s", formatCurrency(pl.OperatingIncome))
	fmt.Printf(" (%.1f%%)\n", pl.OperatingMargin)
	fmt.Println()

	// OTHER INCOME/EXPENSES
	if pl.OtherIncome != 0 || pl.OtherExpenses != 0 {
		fmt.Println("🔄 OTHER INCOME/EXPENSES")
		fmt.Printf("   Other Income               : %15s\n", formatCurrency(pl.OtherIncome))
		fmt.Printf("   Other Expenses             : %15s\n", formatCurrency(pl.OtherExpenses))
		fmt.Println()
	}

	// FINAL RESULTS
	fmt.Printf("💼 INCOME BEFORE TAX          : %15s\n", formatCurrency(pl.IncomeBeforeTax))
	fmt.Printf("🏛️ TAX EXPENSE (Est. 25%%)     : %15s\n", formatCurrency(pl.TaxExpense))
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("🏆 NET INCOME                 : %15s", formatCurrency(pl.NetIncome))
	fmt.Printf(" (%.1f%%)\n", pl.NetIncomeMargin)
	fmt.Println(strings.Repeat("=", 70))

	// SUMMARY RATIOS
	fmt.Println("\n📋 KEY FINANCIAL RATIOS")
	fmt.Printf("   Gross Profit Margin        : %14.1f%%\n", pl.GrossProfitMargin)
	fmt.Printf("   Operating Margin           : %14.1f%%\n", pl.OperatingMargin)
	fmt.Printf("   Net Income Margin          : %14.1f%%\n", pl.NetIncomeMargin)
	
	if pl.Revenue.TotalRevenue > 0 {
		opexRatio := (pl.OperatingExpenses.TotalOpEx / pl.Revenue.TotalRevenue) * 100
		fmt.Printf("   Operating Expense Ratio    : %14.1f%%\n", opexRatio)
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("✅ Profit & Loss Statement Generated Successfully!")
	
	// Analysis notes
	fmt.Println("\n💡 ANALYSIS NOTES:")
	if pl.Revenue.TotalRevenue == 0 {
		fmt.Println("⚠️ No revenue recorded in this period")
	}
	if pl.NetIncome > 0 {
		fmt.Println("✅ Company is profitable in this period")
	} else if pl.NetIncome < 0 {
		fmt.Println("⚠️ Company shows net loss in this period")
	}
	if pl.GrossProfitMargin < 20 {
		fmt.Println("⚠️ Gross profit margin is relatively low (<20%)")
	} else if pl.GrossProfitMargin > 50 {
		fmt.Println("✅ Strong gross profit margin (>50%)")
	}
}

func formatCurrency(amount float64) string {
	// Go doesn't support %,.2f format, so we'll format manually
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
