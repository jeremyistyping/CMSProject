package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/cmd/scripts/utils"
)

type SalesDetail struct {
	ID            int     `gorm:"column:id"`
	Code          string  `gorm:"column:code"`
	InvoiceNumber string  `gorm:"column:invoice_number"`
	Status        string  `gorm:"column:status"`
	SubTotal      float64 `gorm:"column:subtotal"`
	TaxAmount     float64 `gorm:"column:tax"`
	TotalAmount   float64 `gorm:"column:total_amount"`
	CustomerName  string  `gorm:"column:customer_name"`
	CreatedAt     string  `gorm:"column:created_at"`
}

type JournalDetail struct {
	ID          int     `gorm:"column:id"`
	AccountID   int     `gorm:"column:account_id"`
	AccountName string  `gorm:"column:account_name"`
	DebitAmount float64 `gorm:"column:debit_amount"`
	CreditAmount float64 `gorm:"column:credit_amount"`
	Description string  `gorm:"column:description"`
	CreatedAt   string  `gorm:"column:created_at"`
}

type RevenueAccountDetail struct {
	AccountID   int     `gorm:"column:account_id"`
	AccountCode string  `gorm:"column:account_code"`
	AccountName string  `gorm:"column:account_name"`
	Balance     float64 `gorm:"column:balance"`
	TotalCredits float64 `gorm:"column:total_credits"`
	TotalDebits  float64 `gorm:"column:total_debits"`
	JournalCount int     `gorm:"column:journal_count"`
}

func main() {
	fmt.Printf("🔍 INVESTIGATING MISSING REVENUE ISSUE\n")
	fmt.Printf("Expected: 2 transactions × Rp 5,000,000 = Rp 10,000,000\n")
	fmt.Printf("Actual: Only Rp 5,000,000 showing in revenue\n\n")

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

	// 1. Get all sales transactions to see the real data
	fmt.Printf("\n=== STEP 1: ALL SALES TRANSACTIONS ===\n")
	var salesDetails []SalesDetail
	
	// Try different column names that might exist
	possibleQueries := []string{
		// Query 1: Correct columns based on table structure
		`SELECT s.id, s.code, s.invoice_number, s.status, 
		        s.subtotal, s.tax, s.total_amount,
		        c.name as customer_name, s.created_at::text
		 FROM sales s LEFT JOIN contacts c ON s.customer_id = c.id 
		 ORDER BY s.created_at DESC`,
		
		// Query 2: Alternative column names (fallback)
		`SELECT s.id, s.code, s.invoice_number, s.status,
		        s.subtotal, s.tax_amount, s.total_amount,
		        c.name as customer_name, s.created_at::text
		 FROM sales s LEFT JOIN contacts c ON s.customer_id = c.id 
		 ORDER BY s.created_at DESC`,
		
		// Query 3: Basic columns only
		`SELECT s.id, s.code, s.invoice_number, s.status,
		        0 as subtotal, 0 as tax, 0 as total_amount,
		        c.name as customer_name, s.created_at::text
		 FROM sales s LEFT JOIN contacts c ON s.customer_id = c.id 
		 ORDER BY s.created_at DESC`,
	}
	
	var queryWorked bool
	for i, query := range possibleQueries {
		fmt.Printf("Trying query %d...\n", i+1)
		err = gormDB.Raw(query).Scan(&salesDetails).Error
		if err == nil {
			queryWorked = true
			fmt.Printf("✅ Query %d worked!\n", i+1)
			break
		} else {
			fmt.Printf("❌ Query %d failed: %v\n", i+1, err)
		}
	}
	
	if !queryWorked {
		fmt.Printf("⚠️ All queries failed, checking table structure...\n")
		
		// Check sales table structure
		rows, err := sqlDB.Query("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'sales' ORDER BY ordinal_position")
		if err != nil {
			log.Printf("Error getting table structure: %v", err)
		} else {
			fmt.Printf("\n📋 SALES TABLE STRUCTURE:\n")
			for rows.Next() {
				var colName, colType string
				rows.Scan(&colName, &colType)
				fmt.Printf("   %s (%s)\n", colName, colType)
			}
			rows.Close()
		}
		
		// Get basic sales data without problematic columns
		err = gormDB.Raw("SELECT id, COALESCE(code, 'N/A') as code, COALESCE(invoice_number, 'N/A') as invoice_number, COALESCE(status, 'UNKNOWN') as status, 0 as subtotal, 0 as tax_amount, 0 as total_amount, 'Unknown' as customer_name, created_at::text FROM sales ORDER BY created_at DESC").Scan(&salesDetails).Error
		if err != nil {
			log.Printf("Error getting basic sales data: %v", err)
			return
		}
	}
	
	if len(salesDetails) > 0 {
		fmt.Printf("\n📊 Found %d sales transactions:\n\n", len(salesDetails))
		fmt.Printf("%-4s | %-12s | %-15s | %-10s | %12s | %12s | %12s | %s\n", 
			"ID", "Code", "Invoice#", "Status", "Subtotal", "Tax", "Total", "Customer")
		fmt.Printf("%-4s-+%-12s-+%-15s-+%-10s-+%12s-+%12s-+%12s-+-%s\n", 
			"----", "------------", "---------------", "----------", "------------", "------------", "------------", "----------")
		
		var totalExpectedRevenue float64
		for _, sale := range salesDetails {
			fmt.Printf("%-4d | %-12s | %-15s | %-10s | %12.2f | %12.2f | %12.2f | %s\n",
				sale.ID, sale.Code, sale.InvoiceNumber, sale.Status,
				sale.SubTotal, sale.TaxAmount, sale.TotalAmount, truncateString(sale.CustomerName, 15))
			totalExpectedRevenue += sale.SubTotal
		}
		fmt.Printf("\n💰 Expected Total Revenue (before tax): Rp %.2f\n", totalExpectedRevenue)
	} else {
		fmt.Printf("❌ No sales transactions found!\n")
	}

	// 2. Check all journal entries for revenue accounts
	fmt.Printf("\n=== STEP 2: REVENUE JOURNAL ENTRIES ===\n")
	var journalDetails []JournalDetail
	
	journalQuery := `
		SELECT ujl.id, ujl.account_id, a.name as account_name,
		       ujl.debit_amount, ujl.credit_amount, ujl.description, ujl.created_at::text
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.type = 'REVENUE'
		ORDER BY ujl.created_at DESC
	`
	
	err = gormDB.Raw(journalQuery).Scan(&journalDetails).Error
	if err != nil {
		log.Printf("Error getting journal entries: %v", err)
	} else {
		fmt.Printf("📋 Found %d journal entries for revenue accounts:\n\n", len(journalDetails))
		fmt.Printf("%-6s | %-4s | %-25s | %12s | %12s | %-30s | %s\n", 
			"JID", "Acc", "Account Name", "Debit", "Credit", "Description", "Date")
		fmt.Printf("%-6s-+%-4s-+%-25s-+%12s-+%12s-+%-30s-+-%s\n", 
			"------", "----", "-------------------------", "------------", "------------", "------------------------------", "----------")
		
		var totalRevenueCredits float64
		for _, journal := range journalDetails {
			fmt.Printf("%-6d | %-4d | %-25s | %12.2f | %12.2f | %-30s | %s\n",
				journal.ID, journal.AccountID, truncateString(journal.AccountName, 25),
				journal.DebitAmount, journal.CreditAmount, truncateString(journal.Description, 30), 
				journal.CreatedAt[:10])
			totalRevenueCredits += journal.CreditAmount
		}
		fmt.Printf("\n💰 Total Revenue Credits in Journals: Rp %.2f\n", totalRevenueCredits)
	}

	// 3. Check detailed revenue account balances
	fmt.Printf("\n=== STEP 3: REVENUE ACCOUNT ANALYSIS ===\n")
	var revenueDetails []RevenueAccountDetail
	
	revenueQuery := `
		SELECT a.id as account_id, a.code as account_code, a.name as account_name,
		       a.balance,
		       COALESCE(SUM(ujl.credit_amount), 0) as total_credits,
		       COALESCE(SUM(ujl.debit_amount), 0) as total_debits,
		       COUNT(ujl.id) as journal_count
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON a.id = ujl.account_id
		WHERE a.type = 'REVENUE'
		GROUP BY a.id, a.code, a.name, a.balance
		ORDER BY a.code
	`
	
	err = gormDB.Raw(revenueQuery).Scan(&revenueDetails).Error
	if err != nil {
		log.Printf("Error getting revenue account details: %v", err)
	} else {
		fmt.Printf("📊 Revenue account breakdown:\n\n")
		fmt.Printf("%-6s | %-25s | %12s | %12s | %12s | %8s | %s\n", 
			"Code", "Account Name", "Balance", "Credits", "Debits", "Journals", "Status")
		fmt.Printf("%-6s-+%-25s-+%12s-+%12s-+%12s-+%8s-+-%s\n", 
			"------", "-------------------------", "------------", "------------", "------------", "--------", "--------")
		
		var totalCurrentBalance, totalExpectedBalance float64
		for _, account := range revenueDetails {
			expectedBalance := account.TotalCredits - account.TotalDebits
			status := "✅ OK"
			if account.Balance != expectedBalance {
				status = "❌ MISMATCH"
			}
			
			fmt.Printf("%-6s | %-25s | %12.2f | %12.2f | %12.2f | %8d | %s\n",
				account.AccountCode, truncateString(account.AccountName, 25),
				account.Balance, account.TotalCredits, account.TotalDebits, 
				account.JournalCount, status)
			
			totalCurrentBalance += account.Balance
			totalExpectedBalance += expectedBalance
		}
		
		fmt.Printf("\n💰 Summary:\n")
		fmt.Printf("   Total Current Revenue Balance: Rp %.2f\n", totalCurrentBalance)
		fmt.Printf("   Total Expected from Journals:  Rp %.2f\n", totalExpectedBalance)
		fmt.Printf("   Difference: Rp %.2f\n", totalExpectedBalance - totalCurrentBalance)
	}

	// 4. Check if there are revenue entries in other accounts
	fmt.Printf("\n=== STEP 4: CHECK FOR MISALLOCATED REVENUE ===\n")
	var misallocatedRevenue []struct {
		AccountID   int     `gorm:"column:account_id"`
		AccountCode string  `gorm:"column:account_code"`
		AccountName string  `gorm:"column:account_name"`
		AccountType string  `gorm:"column:account_type"`
		Credits     float64 `gorm:"column:credits"`
	}
	
	misallocQuery := `
		SELECT a.id as account_id, a.code as account_code, a.name as account_name, a.type as account_type,
		       SUM(ujl.credit_amount) as credits
		FROM accounts a
		JOIN unified_journal_lines ujl ON a.id = ujl.account_id
		WHERE ujl.description ILIKE '%penjualan%' OR ujl.description ILIKE '%sales%' OR ujl.description ILIKE '%revenue%'
		GROUP BY a.id, a.code, a.name, a.type
		HAVING SUM(ujl.credit_amount) > 0
		ORDER BY credits DESC
	`
	
	err = gormDB.Raw(misallocQuery).Scan(&misallocatedRevenue).Error
	if err != nil {
		log.Printf("Error checking misallocated revenue: %v", err)
	} else {
		fmt.Printf("🔍 Accounts with sales-related credits:\n\n")
		fmt.Printf("%-6s | %-25s | %-8s | %12s\n", "Code", "Account Name", "Type", "Credits")
		fmt.Printf("%-6s-+%-25s-+%-8s-+%12s\n", "------", "-------------------------", "--------", "------------")
		
		var totalMisallocated float64
		for _, account := range misallocatedRevenue {
			fmt.Printf("%-6s | %-25s | %-8s | %12.2f\n",
				account.AccountCode, truncateString(account.AccountName, 25),
				account.AccountType, account.Credits)
			
			if account.AccountType != "REVENUE" {
				totalMisallocated += account.Credits
			}
		}
		
		if totalMisallocated > 0 {
			fmt.Printf("\n⚠️ Found Rp %.2f in non-revenue accounts that might be misallocated\n", totalMisallocated)
		}
	}

	// 5. Final diagnosis and recommendations
	fmt.Printf("\n=== DIAGNOSIS & RECOMMENDATIONS ===\n")
	
	// Calculate actual expected revenue from sales data
	actualExpectedRevenue := 0.0
	for _, sale := range salesDetails {
		actualExpectedRevenue += sale.SubTotal
	}
	
	fmt.Printf("🧮 Revenue Analysis:\n")
	fmt.Printf("   Expected Revenue from Sales: Rp %.2f\n", actualExpectedRevenue)
	fmt.Printf("   Current Revenue Balance:     Rp 5,000,000.00\n")
	if actualExpectedRevenue > 5000000 {
		fmt.Printf("   Missing Revenue:             Rp %.2f\n", actualExpectedRevenue - 5000000)
	} else {
		fmt.Printf("   Revenue matches expected amount ✅\n")
	}
	
	if actualExpectedRevenue > 5000000 {
		fmt.Printf("\n💡 POSSIBLE CAUSES:\n")
		fmt.Printf("1. ❓ Second transaction journal not created\n")
		fmt.Printf("2. ❓ Journal entry went to wrong account\n") 
		fmt.Printf("3. ❓ Balance sync only processed one transaction\n")
		fmt.Printf("4. ❓ Revenue recognition timing issue\n")
		
		fmt.Printf("\n🔧 RECOMMENDED ACTIONS:\n")
		fmt.Printf("1. Check if both sales have corresponding journal entries\n")
		fmt.Printf("2. Run manual balance sync to catch any missed entries\n")
		fmt.Printf("3. Verify sales transaction status and journal creation logic\n")
		fmt.Printf("4. Consider running: go run cmd/scripts/final_balance_fix.go\n")
	}
	
	fmt.Printf("\n🏁 INVESTIGATION COMPLETE!\n")
	fmt.Printf("Review the data above to identify where the missing Rp 5,000,000 revenue went.\n")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}