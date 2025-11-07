package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔍 Sales Accounting Analysis")
	fmt.Println("============================")

	// Database connection
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = ""
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sistem_akuntansi"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 1. Analyze current COA structure
	fmt.Println("📊 Current Chart of Accounts Analysis:")
	analyzeCOA(db)

	// 2. Check sales transactions
	fmt.Println("\n📋 Sales Transactions Analysis:")
	analyzeSalesTransactions(db)

	// 3. Check journal entries for sales
	fmt.Println("\n📝 Journal Entries for Sales:")
	analyzeJournalEntries(db)

	// 4. Expected vs Actual accounting logic
	fmt.Println("\n🎯 Expected Accounting Logic Analysis:")
	analyzeAccountingLogic(db)
}

func analyzeCOA(db *gorm.DB) {
	var accounts []struct {
		Code    string
		Name    string
		Type    string
		Balance float64
		Status  string
	}

	query := `
		SELECT code, name, type, balance, status 
		FROM accounts 
		WHERE deleted_at IS NULL 
		ORDER BY CAST(code AS UNSIGNED)
	`

	err := db.Raw(query).Scan(&accounts).Error
	if err != nil {
		log.Printf("Failed to fetch accounts: %v", err)
		return
	}

	// Group accounts by type
	accountsByType := make(map[string][]struct {
		Code    string
		Name    string
		Type    string
		Balance float64
		Status  string
	})

	for _, account := range accounts {
		accountsByType[account.Type] = append(accountsByType[account.Type], account)
	}

	// Check key accounts for sales transactions
	fmt.Println("\n🔍 Key Accounts for Sales Process:")
	
	keyAccounts := map[string]string{
		"1201": "Piutang Usaha (Accounts Receivable)",
		"4101": "Pendapatan Penjualan (Sales Revenue)", 
		"2102": "Utang Pajak (Tax Payable - for PPN)",
		"1106": "PPN Masukan (Input VAT)",
		"5101": "Harga Pokok Penjualan (COGS)",
		"1301": "Persediaan Barang Dagangan (Inventory)",
		"1101": "Kas (Cash)",
		"1102": "Bank BCA",
		"1103": "Bank Mandiri",
	}

	for code, description := range keyAccounts {
		found := false
		for _, account := range accounts {
			if account.Code == code {
				fmt.Printf("✅ %s: %s - Balance: Rp %.2f (%s)\n", 
					code, account.Name, account.Balance, account.Status)
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("❌ Missing: %s (%s)\n", code, description)
		}
	}
}

func analyzeSalesTransactions(db *gorm.DB) {
	var salesData []struct {
		ID               uint
		Code             string
		CustomerName     string
		TotalAmount      float64
		PaidAmount       float64
		OutstandingAmount float64
		Status           string
		CreatedAt        string
	}

	query := `
		SELECT 
			s.id,
			s.code,
			c.name as customer_name,
			s.total_amount,
			s.paid_amount,
			s.outstanding_amount,
			s.status,
			s.created_at
		FROM sales s
		LEFT JOIN contacts c ON s.customer_id = c.id
		WHERE s.deleted_at IS NULL
		ORDER BY s.created_at DESC
		LIMIT 10
	`

	err := db.Raw(query).Scan(&salesData).Error
	if err != nil {
		log.Printf("Failed to fetch sales data: %v", err)
		return
	}

	fmt.Printf("Found %d sales transactions:\n", len(salesData))
	fmt.Printf("%-6s %-15s %-20s %-12s %-12s %-12s %-10s\n", 
		"ID", "CODE", "CUSTOMER", "TOTAL", "PAID", "OUTSTANDING", "STATUS")
	fmt.Println("================================================================================")

	totalSales := 0.0
	totalOutstanding := 0.0

	for _, sale := range salesData {
		fmt.Printf("%-6d %-15s %-20s %-12.0f %-12.0f %-12.0f %-10s\n",
			sale.ID, sale.Code, sale.CustomerName, 
			sale.TotalAmount, sale.PaidAmount, sale.OutstandingAmount, sale.Status)
		
		totalSales += sale.TotalAmount
		totalOutstanding += sale.OutstandingAmount
	}

	fmt.Println("================================================================================")
	fmt.Printf("Total Sales: Rp %.2f\n", totalSales)
	fmt.Printf("Total Outstanding: Rp %.2f\n", totalOutstanding)
}

func analyzeJournalEntries(db *gorm.DB) {
	var journalEntries []struct {
		EntryID     uint
		Reference   string
		Description string
		AccountCode string
		AccountName string
		DebitAmount float64
		CreditAmount float64
		EntryDate   string
	}

	query := `
		SELECT 
			je.id as entry_id,
			je.reference,
			je.description,
			a.code as account_code,
			a.name as account_name,
			jl.debit_amount,
			jl.credit_amount,
			je.entry_date
		FROM journal_entries je
		JOIN journal_lines jl ON je.id = jl.journal_entry_id
		JOIN accounts a ON jl.account_id = a.id
		WHERE je.reference_type = 'SALE' 
		AND je.deleted_at IS NULL
		ORDER BY je.entry_date DESC, je.id DESC, jl.line_number
		LIMIT 50
	`

	err := db.Raw(query).Scan(&journalEntries).Error
	if err != nil {
		log.Printf("Failed to fetch journal entries: %v", err)
		return
	}

	if len(journalEntries) == 0 {
		fmt.Println("❌ No journal entries found for sales transactions!")
		fmt.Println("This indicates that sales are not being properly journalized.")
		return
	}

	fmt.Printf("Found %d journal lines for sales:\n", len(journalEntries))
	fmt.Printf("%-8s %-15s %-10s %-25s %-12s %-12s\n", 
		"ENTRY_ID", "REFERENCE", "ACC_CODE", "ACCOUNT_NAME", "DEBIT", "CREDIT")
	fmt.Println("==================================================================================")

	currentEntryID := uint(0)
	entryDebit := 0.0
	entryCredit := 0.0

	for _, entry := range journalEntries {
		if currentEntryID != 0 && currentEntryID != entry.EntryID {
			// Print totals for previous entry
			fmt.Printf("%-8s %-15s %-10s %-25s %-12.2f %-12.2f\n", 
				"", "TOTALS:", "", "", entryDebit, entryCredit)
			fmt.Println("----------------------------------")
			entryDebit = 0.0
			entryCredit = 0.0
		}

		fmt.Printf("%-8d %-15s %-10s %-25s %-12.2f %-12.2f\n",
			entry.EntryID, entry.Reference, entry.AccountCode, entry.AccountName,
			entry.DebitAmount, entry.CreditAmount)

		currentEntryID = entry.EntryID
		entryDebit += entry.DebitAmount
		entryCredit += entry.CreditAmount
	}

	// Print final totals
	if currentEntryID != 0 {
		fmt.Printf("%-8s %-15s %-10s %-25s %-12.2f %-12.2f\n", 
			"", "TOTALS:", "", "", entryDebit, entryCredit)
	}
}

func analyzeAccountingLogic(db *gorm.DB) {
	fmt.Println("\n📚 Expected Accounting Logic for Sales:")
	fmt.Println("=====================================")
	
	fmt.Println("1. INVOICE CREATION (Invoiced Status):")
	fmt.Println("   Debit:  1201 Piutang Usaha (Accounts Receivable)")
	fmt.Println("   Credit: 4101 Pendapatan Penjualan (Sales Revenue)")
	fmt.Println("   Credit: 2102 Utang Pajak/PPN Keluaran (Output VAT) [if applicable]")
	
	fmt.Println("\n2. PAYMENT RECEIVED (Paid Status):")
	fmt.Println("   Debit:  1101 Kas/Bank (Cash/Bank)")
	fmt.Println("   Credit: 1201 Piutang Usaha (Accounts Receivable)")

	fmt.Println("\n3. INVENTORY MOVEMENT (if using perpetual inventory):")
	fmt.Println("   Debit:  5101 Harga Pokok Penjualan (COGS)")
	fmt.Println("   Credit: 1301 Persediaan Barang Dagangan (Inventory)")

	// Check current account balances against expected logic
	fmt.Println("\n🔍 Current Balance Analysis:")
	
	var accountBalances []struct {
		Code    string
		Name    string
		Balance float64
		Type    string
	}

	db.Raw(`
		SELECT code, name, balance, type 
		FROM accounts 
		WHERE code IN ('1201', '4101', '2102', '5101', '1301', '1101', '1102', '1103')
		AND deleted_at IS NULL
		ORDER BY code
	`).Scan(&accountBalances)

	for _, account := range accountBalances {
		var analysis string
		
		switch account.Code {
		case "1201": // Accounts Receivable
			if account.Balance > 0 {
				analysis = "✅ Positive balance indicates outstanding invoices"
			} else {
				analysis = "⚠️  Zero/negative balance - check if all invoices are paid"
			}
		case "4101": // Sales Revenue  
			if account.Balance < 0 {
				analysis = "✅ Negative balance is correct for revenue accounts"
			} else {
				analysis = "❌ Revenue should have negative balance"
			}
		case "2102": // Tax Payable
			if account.Balance < 0 {
				analysis = "✅ Negative balance indicates tax liability"
			} else {
				analysis = "⚠️  Check tax calculation and reporting"
			}
		case "5101": // COGS
			if account.Balance >= 0 {
				analysis = "✅ Positive balance is correct for expense accounts"
			} else {
				analysis = "❌ COGS should have positive balance"
			}
		case "1301": // Inventory
			analysis = "Asset account - should reflect current inventory value"
		default:
			analysis = "Bank/Cash account"
		}
		
		fmt.Printf("%-6s %-30s %15.2f %s\n", 
			account.Code, account.Name, account.Balance, analysis)
	}

	// Final recommendations
	fmt.Println("\n🎯 Recommendations:")
	
	// Check if sales revenue and accounts receivable make sense together
	var arBalance, salesRevenue float64
	db.Raw("SELECT balance FROM accounts WHERE code = '1201'").Scan(&arBalance)
	db.Raw("SELECT balance FROM accounts WHERE code = '4101'").Scan(&salesRevenue)
	
	fmt.Printf("Current Outstanding (AR): Rp %.2f\n", arBalance)
	fmt.Printf("Total Sales Revenue: Rp %.2f\n", -salesRevenue)
	
	if arBalance > 0 && salesRevenue < 0 {
		fmt.Println("✅ Basic accounting relationship looks correct")
		fmt.Println("   - Positive AR indicates unpaid invoices")
		fmt.Println("   - Negative Sales Revenue is correct")
	} else {
		fmt.Println("⚠️  Review accounting entries:")
		if arBalance <= 0 {
			fmt.Println("   - AR should be positive if there are unpaid invoices")
		}
		if salesRevenue >= 0 {
			fmt.Println("   - Sales Revenue should be negative (credit balance)")
		}
	}
}