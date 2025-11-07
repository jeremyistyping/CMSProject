package main

import (
	"log"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

// CheckAccountingIntegrity verifies the accounting system integrity
func main() {
	log.Println("========================================")
	log.Println("🔍 ACCOUNTING INTEGRITY CHECKER")
	log.Println("========================================")

	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Environment: %s", cfg.Environment)

	// Connect to database
	db := database.ConnectDB()
	log.Println("✅ Database connected\n")

	hasErrors := false

	// Check 1: Journal Entry Balance
	log.Println("📊 CHECK 1: Journal Entry Balance")
	log.Println("   Verifying all journal entries are balanced (Debit = Credit)...")
	
	var unbalancedJournals []struct {
		ID          uint64
		EntryNumber string
		TotalDebit  float64
		TotalCredit float64
		Difference  float64
	}

	db.Raw(`
		SELECT 
			id,
			entry_number,
			total_debit,
			total_credit,
			ABS(total_debit - total_credit) as difference
		FROM unified_journal_ledger
		WHERE ABS(total_debit - total_credit) > 0.01
		  AND deleted_at IS NULL
		ORDER BY id
	`).Scan(&unbalancedJournals)

	if len(unbalancedJournals) > 0 {
		log.Printf("   ❌ FOUND %d unbalanced journal entries:", len(unbalancedJournals))
		for _, j := range unbalancedJournals {
			log.Printf("      #%-5d %-20s | Debit: Rp %12.2f | Credit: Rp %12.2f | Diff: Rp %10.2f",
				j.ID, j.EntryNumber, j.TotalDebit, j.TotalCredit, j.Difference)
		}
		hasErrors = true
	} else {
		log.Println("   ✅ All journal entries are balanced")
	}

	// Check 2: COGS Recording
	log.Println("\n📊 CHECK 2: COGS Recording for Sales")
	log.Println("   Checking if all INVOICED/PAID sales have COGS entries...")

	var salesWithoutCOGS []struct {
		ID            uint
		InvoiceNumber string
		Date          time.Time
		TotalAmount   float64
		Status        string
	}

	db.Raw(`
		SELECT 
			s.id,
			s.invoice_number,
			s.date,
			s.total_amount,
			s.status
		FROM sales s
		WHERE s.status IN ('INVOICED', 'PAID')
		  AND s.deleted_at IS NULL
		  AND NOT EXISTS (
		      SELECT 1 
		      FROM unified_journal_ledger uje
		      JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
		      JOIN accounts a ON a.id = ujl.account_id
		      WHERE uje.source_type = 'SALE' 
		        AND uje.source_id = s.id
		        AND a.code = '5101'  -- COGS account
		        AND uje.deleted_at IS NULL
		  )
		ORDER BY s.id
	`).Scan(&salesWithoutCOGS)

	if len(salesWithoutCOGS) > 0 {
		log.Printf("   ❌ FOUND %d sales without COGS entries:", len(salesWithoutCOGS))
		log.Println("      ID   | Invoice         | Date       | Amount          | Status")
		log.Println("      -----|-----------------|------------|-----------------|--------")
		for _, s := range salesWithoutCOGS {
			log.Printf("      %-4d | %-15s | %s | Rp %12.2f | %s",
				s.ID, s.InvoiceNumber, s.Date.Format("2006-01-02"), s.TotalAmount, s.Status)
		}
		log.Println("\n   💡 Fix: Run 'go run cmd/scripts/backfill_missing_cogs.go'")
		hasErrors = true
	} else {
		log.Println("   ✅ All sales have COGS entries")
	}

	// Check 3: Products with Zero Cost Price
	log.Println("\n📊 CHECK 3: Products with Zero Cost Price")
	log.Println("   Checking products that may cause zero COGS...")

	var productsNoCost []struct {
		ID        uint
		Name      string
		Price     float64
		CostPrice float64
		Stock     int
	}

	db.Raw(`
		SELECT id, name, sale_price as price, cost_price, stock
		FROM products
		WHERE (cost_price = 0 OR cost_price IS NULL)
		  AND stock > 0
		  AND deleted_at IS NULL
		ORDER BY stock DESC
		LIMIT 20
	`).Scan(&productsNoCost)

	if len(productsNoCost) > 0 {
		log.Printf("   ⚠️  FOUND %d products with zero cost price (showing top 20):", len(productsNoCost))
		log.Println("      ID   | Name                          | Stock | Price        | Cost Price")
		log.Println("      -----|-------------------------------|-------|--------------|------------")
		for _, p := range productsNoCost {
			log.Printf("      %-4d | %-29s | %5d | Rp %9.2f | Rp %9.2f",
				p.ID, truncateString(p.Name, 29), p.Stock, p.Price, p.CostPrice)
		}
		log.Println("\n   💡 Fix: Run 'go run cmd/scripts/fix_product_cost_prices.go'")
		hasErrors = true
	} else {
		log.Println("   ✅ All products have valid cost prices")
	}

	// Check 4: Account Balance Verification
	log.Println("\n📊 CHECK 4: Account Balance Verification")
	log.Println("   Comparing account balances with journal entries...")

	var balanceMismatches []struct {
		Code             string
		Name             string
		AccountBalance   float64
		CalculatedBalance float64
		Difference       float64
	}

	db.Raw(`
		SELECT 
			a.code,
			a.name,
			a.balance as account_balance,
			COALESCE(
				CASE 
					WHEN UPPER(a.type) IN ('ASSET', 'EXPENSE') THEN 
						SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
					ELSE 
						SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
				END,
			0) as calculated_balance,
			ABS(a.balance - COALESCE(
				CASE 
					WHEN UPPER(a.type) IN ('ASSET', 'EXPENSE') THEN 
						SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
					ELSE 
						SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
				END,
			0)) as difference
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id AND uje.status = 'POSTED' AND uje.deleted_at IS NULL
		WHERE a.is_header = false AND a.deleted_at IS NULL
		GROUP BY a.id, a.code, a.name, a.balance, a.type
		HAVING ABS(a.balance - COALESCE(
			CASE 
				WHEN UPPER(a.type) IN ('ASSET', 'EXPENSE') THEN 
					SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
				ELSE 
					SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
			END,
		0)) > 0.01
		ORDER BY difference DESC
		LIMIT 10
	`).Scan(&balanceMismatches)

	if len(balanceMismatches) > 0 {
		log.Printf("   ⚠️  FOUND %d accounts with balance mismatches (showing top 10):", len(balanceMismatches))
		log.Println("      Code | Name                     | Account Bal.  | Calc. Bal.    | Difference")
		log.Println("      -----|--------------------------|---------------|---------------|------------")
		for _, b := range balanceMismatches {
			log.Printf("      %-4s | %-24s | Rp %10.2f | Rp %10.2f | Rp %9.2f",
				b.Code, truncateString(b.Name, 24), b.AccountBalance, b.CalculatedBalance, b.Difference)
		}
		log.Println("\n   💡 Note: Small differences (<Rp 1) are acceptable due to rounding")
	} else {
		log.Println("   ✅ All account balances match journal entries")
	}

	// Check 5: Profit & Loss Sanity Check
	log.Println("\n📊 CHECK 5: Profit & Loss Sanity Check")
	log.Println("   Checking for accounting anomalies in P&L...")

	var plCheck struct {
		TotalRevenue float64
		TotalCOGS    float64
		TotalExpense float64
		GrossProfit  float64
		NetIncome    float64
	}

	db.Raw(`
		SELECT 
			COALESCE(SUM(CASE WHEN a.code LIKE '4%' THEN ujl.credit_amount - ujl.debit_amount ELSE 0 END), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN a.code = '5101' THEN ujl.debit_amount - ujl.credit_amount ELSE 0 END), 0) as total_cogs,
			COALESCE(SUM(CASE WHEN a.code LIKE '5%' AND a.code != '5101' THEN ujl.debit_amount - ujl.credit_amount ELSE 0 END), 0) as total_expense,
			COALESCE(SUM(CASE WHEN a.code LIKE '4%' THEN ujl.credit_amount - ujl.debit_amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN a.code = '5101' THEN ujl.debit_amount - ujl.credit_amount ELSE 0 END), 0) as gross_profit,
			COALESCE(SUM(CASE WHEN a.code LIKE '4%' THEN ujl.credit_amount - ujl.debit_amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN a.code LIKE '5%' THEN ujl.debit_amount - ujl.credit_amount ELSE 0 END), 0) as net_income
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id AND uje.status = 'POSTED' AND uje.deleted_at IS NULL
		JOIN accounts a ON a.id = ujl.account_id
	`).Scan(&plCheck)

	log.Printf("   Revenue:      Rp %15.2f", plCheck.TotalRevenue)
	log.Printf("   COGS:         Rp %15.2f", plCheck.TotalCOGS)
	log.Printf("   Gross Profit: Rp %15.2f (%.1f%%)", 
		plCheck.GrossProfit, 
		calculatePercentage(plCheck.GrossProfit, plCheck.TotalRevenue))
	log.Printf("   Expenses:     Rp %15.2f", plCheck.TotalExpense)
	log.Printf("   Net Income:   Rp %15.2f (%.1f%%)", 
		plCheck.NetIncome,
		calculatePercentage(plCheck.NetIncome, plCheck.TotalRevenue))

	// Sanity checks
	if plCheck.TotalRevenue > 0 && plCheck.TotalCOGS == 0 {
		log.Println("\n   ⚠️  WARNING: Revenue exists but COGS is zero!")
		log.Println("       This suggests COGS journal entries are missing.")
		hasErrors = true
	}

	if plCheck.GrossProfit < 0 {
		log.Println("\n   ⚠️  WARNING: Negative gross profit!")
		log.Println("       COGS is higher than revenue. Check your pricing and cost prices.")
	}

	grossMargin := calculatePercentage(plCheck.GrossProfit, plCheck.TotalRevenue)
	if grossMargin > 90 {
		log.Println("\n   ⚠️  WARNING: Gross margin is very high (>90%)!")
		log.Println("       This might indicate missing or incorrect COGS entries.")
		hasErrors = true
	}

	// Summary
	log.Println("\n========================================")
	if hasErrors {
		log.Println("❌ INTEGRITY CHECK FAILED")
		log.Println("   Please review and fix the issues above.")
		log.Println("   Run the suggested fix scripts to resolve problems.")
	} else {
		log.Println("✅ INTEGRITY CHECK PASSED")
		log.Println("   Your accounting system is healthy!")
	}
	log.Println("========================================")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func calculatePercentage(part, whole float64) float64 {
	if whole == 0 {
		return 0
	}
	return (part / whole) * 100
}
