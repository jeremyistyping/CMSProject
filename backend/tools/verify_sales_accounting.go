package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"

	_ "github.com/lib/pq"
)

type Sale struct {
	ID             int     `db:"id"`
	Code           string  `db:"code"`
	CustomerName   string  `db:"customer_name"`
	TotalAmount    float64 `db:"total_amount"`
	PaidAmount     float64 `db:"paid_amount"`
	Outstanding    float64 `db:"outstanding"`
	Status         string  `db:"status"`
	PPNRate        float64 `db:"ppn_rate"`
	PPNAmount      float64 `db:"ppn_amount"`
	SubtotalAmount float64 `db:"subtotal_amount"`
}

type SaleItem struct {
	ProductName string  `db:"product_name"`
	Quantity    float64 `db:"quantity"`
	UnitPrice   float64 `db:"unit_price"`
	Discount    float64 `db:"discount"`
	TotalPrice  float64 `db:"total_price"`
}

func main() {
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== SALES ACCOUNTING VERIFICATION ===\n")

	// 1. Cek Sales Transactions
	fmt.Println("1. SALES TRANSACTIONS:")
	rows, err := db.Query(`
		SELECT s.id, s.code, c.name as customer_name, s.total_amount,
			   COALESCE(s.paid_amount, 0) as paid_amount,
			   COALESCE(s.outstanding_amount, s.total_amount - COALESCE(s.paid_amount, 0)) as outstanding,
			   s.status, COALESCE(s.ppn_rate, 0) as ppn_rate,
			   COALESCE(s.ppn_amount, 0) as ppn_amount,
			   COALESCE(s.subtotal, s.total_amount - COALESCE(s.ppn_amount, 0)) as subtotal_amount
		FROM sales s
		JOIN contacts c ON s.customer_id = c.id
		ORDER BY s.code
	`)
	if err != nil {
		fmt.Printf("Error fetching sales: %v\n", err)
		return
	}
	defer rows.Close()

	var sales []Sale
	for rows.Next() {
		var s Sale
		err := rows.Scan(&s.ID, &s.Code, &s.CustomerName, &s.TotalAmount,
			&s.PaidAmount, &s.Outstanding, &s.Status, &s.PPNRate,
			&s.PPNAmount, &s.SubtotalAmount)
		if err != nil {
			fmt.Printf("Error scanning sale: %v\n", err)
			continue
		}
		sales = append(sales, s)
		fmt.Printf("- %s: Customer=%s, Subtotal=%.0f, PPN(%.0f%%)=%.0f, Total=%.0f\n",
			s.Code, s.CustomerName, s.SubtotalAmount, s.PPNRate, s.PPNAmount, s.TotalAmount)
		fmt.Printf("  Status=%s, Paid=%.0f, Outstanding=%.0f\n\n", 
			s.Status, s.PaidAmount, s.Outstanding)
	}

	// 2. Cek Sale Items untuk setiap sale
	fmt.Println("2. SALES ITEMS DETAILS:")
	for _, sale := range sales {
		fmt.Printf("Items for %s:\n", sale.Code)
		
		rows, err := db.Query(`
			SELECT p.name as product_name, si.quantity, si.unit_price,
				   COALESCE(si.discount, 0) as discount, si.total_price
			FROM sale_items si
			JOIN products p ON si.product_id = p.id
			WHERE si.sale_id = $1
			ORDER BY si.id
		`, sale.ID)
		if err != nil {
			fmt.Printf("  Error fetching items: %v\n", err)
			continue
		}
		defer rows.Close()

		subtotalCalculated := 0.0
		for rows.Next() {
			var item SaleItem
			err := rows.Scan(&item.ProductName, &item.Quantity, &item.UnitPrice,
				&item.Discount, &item.TotalPrice)
			if err != nil {
				continue
			}
			
			lineSubtotal := (item.Quantity * item.UnitPrice) - item.Discount
			subtotalCalculated += lineSubtotal
			
			fmt.Printf("  - %s: Qty=%.0f x %.0f - Disc=%.0f = %.0f (Recorded: %.0f)\n",
				item.ProductName, item.Quantity, item.UnitPrice, 
				item.Discount, lineSubtotal, item.TotalPrice)
		}
		
		fmt.Printf("  Calculated Subtotal: %.0f\n", subtotalCalculated)
		
		// Verify PPN calculation
		expectedPPN := subtotalCalculated * (sale.PPNRate / 100)
		expectedTotal := subtotalCalculated + expectedPPN
		
		fmt.Printf("  Expected PPN (%.0f%%): %.0f (Recorded: %.0f)\n", 
			sale.PPNRate, expectedPPN, sale.PPNAmount)
		fmt.Printf("  Expected Total: %.0f (Recorded: %.0f)\n", 
			expectedTotal, sale.TotalAmount)
		
		if abs(expectedTotal-sale.TotalAmount) <= 1 {
			fmt.Printf("  ✅ SALES CALCULATION CORRECT\n\n")
		} else {
			fmt.Printf("  ❌ SALES CALCULATION ERROR! Diff: %.0f\n\n", 
				expectedTotal-sale.TotalAmount)
		}
	}

	// 3. Cek Journal Entries untuk Sales (dari SSOT Journal)
	fmt.Println("3. SALES JOURNAL ENTRIES (SSOT):")
	rows, err = db.Query(`
		SELECT l.id, l.source_code, l.description,
			   ujl.line_number, a.code as account_code, a.name as account_name,
			   ujl.debit_amount, ujl.credit_amount, ujl.description as line_desc
		FROM unified_journal_ledger l
		JOIN unified_journal_lines ujl ON ujl.journal_id = l.id
		JOIN accounts a ON ujl.account_id = a.id
		WHERE l.source_type = 'SALE' OR l.source_code LIKE 'INV%'
		ORDER BY l.id, ujl.line_number
	`)
	if err != nil {
		fmt.Printf("Error reading sales journals: %v\n", err)
	} else {
		defer rows.Close()
		currentLedgerId := -1
		for rows.Next() {
			var ledgerId, lineNumber int
			var sourceCode, description, accountCode, accountName, lineDesc string
			var debitAmount, creditAmount float64
			
			err := rows.Scan(&ledgerId, &sourceCode, &description,
				&lineNumber, &accountCode, &accountName,
				&debitAmount, &creditAmount, &lineDesc)
			if err != nil {
				continue
			}
			
			if currentLedgerId != ledgerId {
				fmt.Printf("\nJournal Entry ID:%d (%s):\n%s\n", 
					ledgerId, sourceCode, description)
				currentLedgerId = ledgerId
			}
			
			if debitAmount > 0 {
				fmt.Printf("  %d. Dr. %s (%s): %.0f - %s\n", 
					lineNumber, accountCode, accountName, debitAmount, lineDesc)
			}
			if creditAmount > 0 {
				fmt.Printf("  %d. Cr. %s (%s): %.0f - %s\n", 
					lineNumber, accountCode, accountName, creditAmount, lineDesc)
			}
		}
	}

	// 4. Verify Sales Accounting Logic
	fmt.Println("\n4. SALES ACCOUNTING LOGIC VERIFICATION:")
	fmt.Println("Expected Journal Entry for Sales with PPN:")
	fmt.Println("  Dr. Piutang Usaha/Kas = Total Amount (including PPN)")
	fmt.Println("  Cr. Pendapatan Penjualan = Subtotal Amount (before PPN)")  
	fmt.Println("  Cr. PPN Keluaran = PPN Amount")
	fmt.Println("")

	// Check if PPN Keluaran account exists (should be different from PPN Masukan)
	fmt.Println("5. PPN ACCOUNTS CHECK:")
	rows, err = db.Query(`
		SELECT code, name, type, balance
		FROM accounts 
		WHERE code LIKE '2102%' OR name LIKE '%PPN%' OR name LIKE '%VAT%'
		ORDER BY code
	`)
	if err != nil {
		fmt.Printf("Error checking PPN accounts: %v\n", err)
	} else {
		defer rows.Close()
		fmt.Println("PPN-related accounts:")
		for rows.Next() {
			var code, name, accType string
			var balance float64
			err := rows.Scan(&code, &name, &accType, &balance)
			if err != nil {
				continue
			}
			fmt.Printf("  %s - %s (%s): %.0f\n", code, name, accType, balance)
		}
	}

	// 6. Payment Analysis
	fmt.Println("\n6. SALES PAYMENT ANALYSIS:")
	rows, err = db.Query(`
		SELECT l.id, l.source_code, l.description,
			   ujl.line_number, a.code as account_code, a.name as account_name,
			   ujl.debit_amount, ujl.credit_amount
		FROM unified_journal_ledger l
		JOIN unified_journal_lines ujl ON ujl.journal_id = l.id
		JOIN accounts a ON ujl.account_id = a.id
		WHERE l.source_type = 'PAYMENT' AND l.source_code LIKE 'RCV%'
		ORDER BY l.id, ujl.line_number
	`)
	if err != nil {
		fmt.Printf("Error reading payment journals: %v\n", err)
	} else {
		defer rows.Close()
		currentLedgerId := -1
		for rows.Next() {
			var ledgerId, lineNumber int
			var sourceCode, description, accountCode, accountName string
			var debitAmount, creditAmount float64
			
			err := rows.Scan(&ledgerId, &sourceCode, &description,
				&lineNumber, &accountCode, &accountName,
				&debitAmount, &creditAmount)
			if err != nil {
				continue
			}
			
			if currentLedgerId != ledgerId {
				fmt.Printf("\nPayment Journal ID:%d (%s):\n%s\n", 
					ledgerId, sourceCode, description)
				currentLedgerId = ledgerId
			}
			
			if debitAmount > 0 {
				fmt.Printf("  %d. Dr. %s (%s): %.0f\n", 
					lineNumber, accountCode, accountName, debitAmount)
			}
			if creditAmount > 0 {
				fmt.Printf("  %d. Cr. %s (%s): %.0f\n", 
					lineNumber, accountCode, accountName, creditAmount)
			}
		}
	}

	fmt.Println("\n=== SALES ACCOUNTING VERIFICATION SELESAI ===")
}

func abs(x float64) float64 {
	return math.Abs(x)
}