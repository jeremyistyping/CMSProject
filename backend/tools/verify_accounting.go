package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"

	_ "github.com/lib/pq"
)

type JournalEntry struct {
	TransactionType string  `db:"transaction_type"`
	ReferenceID     string  `db:"reference_id"`
	AccountCode     string  `db:"account_code"`
	AccountName     string  `db:"account_name"`
	DebitAmount     float64 `db:"debit_amount"`
	CreditAmount    float64 `db:"credit_amount"`
	Description     string  `db:"description"`
}

type Account struct {
	Code    string  `db:"code"`
	Name    string  `db:"name"`
	Type    string  `db:"account_type"`
	Balance float64 `db:"balance"`
}

type Purchase struct {
	ID            int     `db:"id"`
	PurchaseCode  string  `db:"purchase_code"`
	VendorName    string  `db:"vendor_name"`
	TotalAmount   float64 `db:"total_amount"`
	PaidAmount    float64 `db:"paid_amount"`
	Outstanding   float64 `db:"outstanding"`
	Status        string  `db:"status"`
	ApprovalStatus string `db:"approval_status"`
}

func main() {
	// Database connection
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== VERIFIKASI SISTEM AKUNTANSI ===\n")

	// 1. Cek semua Purchase Transactions
	fmt.Println("1. PURCHASE TRANSACTIONS:")
	purchases := []Purchase{}
	rows, err := db.Query(`
		SELECT p.id, p.code as purchase_code, c.name as vendor_name, p.total_amount, 
			   COALESCE(p.paid_amount, 0) as paid_amount,
			   COALESCE(p.outstanding_amount, p.total_amount - COALESCE(p.paid_amount, 0)) as outstanding,
			   p.status, COALESCE(p.approval_status, 'PENDING') as approval_status
		FROM purchases p 
		JOIN contacts c ON p.vendor_id = c.id
		ORDER BY p.code
	`)
	if err != nil {
		log.Fatal("Error fetching purchases:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p Purchase
		err := rows.Scan(&p.ID, &p.PurchaseCode, &p.VendorName, &p.TotalAmount, 
			&p.PaidAmount, &p.Outstanding, &p.Status, &p.ApprovalStatus)
		if err != nil {
			log.Fatal("Error scanning purchase:", err)
		}
		purchases = append(purchases, p)
		fmt.Printf("- %s: Total=%.0f, Paid=%.0f, Outstanding=%.0f, Status=%s, Approval=%s\n",
			p.PurchaseCode, p.TotalAmount, p.PaidAmount, p.Outstanding, p.Status, p.ApprovalStatus)
	}

	// 2. Cek Journal Entries untuk Purchase
	fmt.Println("\n2. JOURNAL ENTRIES FOR PURCHASES:")
	journalEntries := []JournalEntry{}
	rows, err = db.Query(`
		SELECT 'purchase' as transaction_type, je.reference as reference_id, 
			   je.code as account_code, a.name as account_name,
			   COALESCE(jl.debit_amount, 0) as debit_amount, 
			   COALESCE(jl.credit_amount, 0) as credit_amount, 
			   je.description
		FROM journal_entries je 
		JOIN journal_lines jl ON je.id = jl.journal_entry_id
		JOIN accounts a ON jl.account_id = a.id
		WHERE je.reference_type = 'purchase'
		ORDER BY je.reference, jl.account_id
	`)
	if err != nil {
		log.Fatal("Error fetching journal entries:", err)
	}
	defer rows.Close()

	currentRef := ""
	for rows.Next() {
		var je JournalEntry
		err := rows.Scan(&je.TransactionType, &je.ReferenceID, &je.AccountCode, 
			&je.AccountName, &je.DebitAmount, &je.CreditAmount, &je.Description)
		if err != nil {
			log.Fatal("Error scanning journal entry:", err)
		}
		journalEntries = append(journalEntries, je)
		
		if currentRef != je.ReferenceID {
			fmt.Printf("\nJournal for %s:\n", je.ReferenceID)
			currentRef = je.ReferenceID
		}
		
		if je.DebitAmount > 0 {
			fmt.Printf("  Dr. %s (%s): %.0f - %s\n", 
				je.AccountCode, je.AccountName, je.DebitAmount, je.Description)
		}
		if je.CreditAmount > 0 {
			fmt.Printf("  Cr. %s (%s): %.0f - %s\n", 
				je.AccountCode, je.AccountName, je.CreditAmount, je.Description)
		}
	}

	// 3. Analisis Balance per Journal Entry
	fmt.Println("\n3. BALANCE VERIFICATION PER JOURNAL:")
	refMap := make(map[string][]JournalEntry)
	for _, je := range journalEntries {
		refMap[je.ReferenceID] = append(refMap[je.ReferenceID], je)
	}

	for ref, entries := range refMap {
		totalDebit := 0.0
		totalCredit := 0.0
		
		for _, entry := range entries {
			totalDebit += entry.DebitAmount
			totalCredit += entry.CreditAmount
		}
		
		fmt.Printf("%s: Debit=%.0f, Credit=%.0f, Balance=%.0f\n", 
			ref, totalDebit, totalCredit, totalDebit-totalCredit)
		
		if totalDebit != totalCredit {
			fmt.Printf("  ❌ TIDAK BALANCE! Selisih: %.0f\n", totalDebit-totalCredit)
		} else {
			fmt.Printf("  ✅ BALANCE\n")
		}
	}

	// 4. Cek Current Account Balances
	fmt.Println("\n4. CURRENT ACCOUNT BALANCES:")
	accounts := []Account{}
	rows, err = db.Query(`
		SELECT code, name, type as account_type, COALESCE(balance, 0) as balance 
		FROM accounts 
		WHERE COALESCE(balance, 0) != 0 OR code IN ('1301', '2101', '2102')
		ORDER BY code
	`)
	if err != nil {
		log.Fatal("Error fetching accounts:", err)
	}
	defer rows.Close()

	totalAssets := 0.0
	totalLiabilities := 0.0
	totalEquity := 0.0

	for rows.Next() {
		var acc Account
		err := rows.Scan(&acc.Code, &acc.Name, &acc.Type, &acc.Balance)
		if err != nil {
			log.Fatal("Error scanning account:", err)
		}
		accounts = append(accounts, acc)
		fmt.Printf("%s - %s (%s): %.0f\n", acc.Code, acc.Name, acc.Type, acc.Balance)
		
		switch acc.Type {
		case "Asset":
			totalAssets += acc.Balance
		case "Liability":
			totalLiabilities += acc.Balance
		case "Equity":
			totalEquity += acc.Balance
		case "Revenue":
			totalEquity += acc.Balance // Revenue increases equity
		case "Expense":
			totalEquity -= acc.Balance // Expense decreases equity
		}
	}

	// 5. Accounting Equation Verification
	fmt.Println("\n5. ACCOUNTING EQUATION VERIFICATION:")
	fmt.Printf("Total Assets: %.0f\n", totalAssets)
	fmt.Printf("Total Liabilities: %.0f\n", totalLiabilities)
	fmt.Printf("Total Equity (including Revenue/Expense): %.0f\n", totalEquity)
	fmt.Printf("Assets = Liabilities + Equity\n")
	fmt.Printf("%.0f = %.0f + %.0f\n", totalAssets, totalLiabilities, totalEquity)
	
	difference := totalAssets - (totalLiabilities + totalEquity)
	if difference == 0 {
		fmt.Printf("✅ ACCOUNTING EQUATION BALANCED\n")
	} else {
		fmt.Printf("❌ NOT BALANCED! Difference: %.0f\n", difference)
	}

	// 6. Cek PPN Calculation
	fmt.Println("\n6. PPN VALIDATION:")
	ppnRows, err := db.Query(`
		SELECT pi.quantity, pi.unit_price, COALESCE(pi.discount, 0) as discount_amount,
			   p.total_amount, p.code as purchase_code, COALESCE(p.ppn_rate, 0) as ppn_rate,
			   COALESCE(p.ppn_amount, 0) as ppn_amount
		FROM purchase_items pi
		JOIN purchases p ON pi.purchase_id = p.id
		WHERE p.status = 'APPROVED'
		ORDER BY p.code
	`)
	if err != nil {
		log.Fatal("Error fetching PPN data:", err)
	}
	defer ppnRows.Close()

	for ppnRows.Next() {
		var qty, unitPrice, discountAmount, totalAmount, ppnRate, ppnAmount float64
		var purchaseCode string
		
		err := ppnRows.Scan(&qty, &unitPrice, &discountAmount, &totalAmount, &purchaseCode, &ppnRate, &ppnAmount)
		if err != nil {
			log.Fatal("Error scanning PPN data:", err)
		}
		
		subtotal := (qty * unitPrice) - discountAmount
		expectedPPN := subtotal * (ppnRate / 100)
		expectedTotal := subtotal + expectedPPN
		
		fmt.Printf("%s: Qty=%.0f, Unit=%.0f, Discount=%.0f, Subtotal=%.0f\n", 
			purchaseCode, qty, unitPrice, discountAmount, subtotal)
		fmt.Printf("  PPN Rate=%.0f%%, PPN Amount=%.0f (Expected=%.0f)\n", 
			ppnRate, ppnAmount, expectedPPN)
		fmt.Printf("  Total=%.0f (Expected=%.0f)\n", totalAmount, expectedTotal)
		
		if abs(expectedTotal-totalAmount) > 1 { // Allow small rounding differences
			fmt.Printf("  ❌ PPN CALCULATION ERROR! Difference: %.0f\n", expectedTotal-totalAmount)
		} else {
			fmt.Printf("  ✅ PPN CALCULATION CORRECT\n")
		}
		fmt.Println()
	}

	fmt.Println("\n=== ANALISIS SELESAI ===")
}

func abs(x float64) float64 {
	return math.Abs(x)
}
