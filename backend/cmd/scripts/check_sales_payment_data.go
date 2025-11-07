package main

import (
	"fmt"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("==========================================")
	fmt.Println("🔍 DETAILED SALES & PAYMENT DATA CHECK")
	fmt.Println("==========================================")

	db := database.ConnectDB()

	// Check all sales data
	fmt.Println("\n📊 SALES DATA:")
	var salesData []struct {
		ID                uint    `gorm:"column:id"`
		Code              string  `gorm:"column:code"`
		CustomerID        uint    `gorm:"column:customer_id"`
		CustomerName      string  `gorm:"column:customer_name"`
		TotalAmount       float64 `gorm:"column:total_amount"`
		OutstandingAmount float64 `gorm:"column:outstanding_amount"`
		Status            string  `gorm:"column:status"`
		Date              string  `gorm:"column:date"`
	}

	db.Raw(`
		SELECT 
			s.id,
			s.code,
			s.customer_id,
			c.name as customer_name,
			s.total_amount,
			s.outstanding_amount,
			s.status,
			s.date::text as date
		FROM sales s
		LEFT JOIN contacts c ON s.customer_id = c.id
		ORDER BY s.id DESC
		LIMIT 5
	`).Scan(&salesData)

	for _, sale := range salesData {
		fmt.Printf("  Sale ID: %d | Code: %s | Customer: %s\n", sale.ID, sale.Code, sale.CustomerName)
		fmt.Printf("    Total: Rp %.2f | Outstanding: Rp %.2f | Status: %s | Date: %s\n\n", 
			sale.TotalAmount, sale.OutstandingAmount, sale.Status, sale.Date)
	}

	// Check all payments
	fmt.Println("💰 PAYMENT DATA:")
	var paymentData []struct {
		ID            uint    `gorm:"column:id"`
		Code          string  `gorm:"column:code"`
		ContactID     uint    `gorm:"column:contact_id"`
		ContactName   string  `gorm:"column:contact_name"`
		Amount        float64 `gorm:"column:amount"`
		Status        string  `gorm:"column:status"`
		Method        string  `gorm:"column:method"`
		Date          string  `gorm:"column:date"`
	}

	db.Raw(`
		SELECT 
			p.id,
			p.code,
			p.contact_id,
			c.name as contact_name,
			p.amount,
			p.status,
			p.method,
			p.date::text as date
		FROM payments p
		LEFT JOIN contacts c ON p.contact_id = c.id
		ORDER BY p.id DESC
		LIMIT 5
	`).Scan(&paymentData)

	for _, payment := range paymentData {
		fmt.Printf("  Payment ID: %d | Code: %s | Contact: %s\n", payment.ID, payment.Code, payment.ContactName)
		fmt.Printf("    Amount: Rp %.2f | Status: %s | Method: %s | Date: %s\n\n", 
			payment.Amount, payment.Status, payment.Method, payment.Date)
	}

	// Check payment allocations
	fmt.Println("🔗 PAYMENT ALLOCATIONS:")
	var allocationData []struct {
		ID              uint     `gorm:"column:id"`
		PaymentID       uint     `gorm:"column:payment_id"`
		PaymentCode     string   `gorm:"column:payment_code"`
		InvoiceID       *uint    `gorm:"column:invoice_id"`
		BillID          *uint    `gorm:"column:bill_id"`
		SaleCode        *string  `gorm:"column:sale_code"`
		AllocatedAmount float64  `gorm:"column:allocated_amount"`
	}

	db.Raw(`
		SELECT 
			pa.id,
			pa.payment_id,
			p.code as payment_code,
			pa.invoice_id,
			pa.bill_id,
			s.code as sale_code,
			pa.allocated_amount
		FROM payment_allocations pa
		LEFT JOIN payments p ON pa.payment_id = p.id
		LEFT JOIN sales s ON pa.invoice_id = s.id
		ORDER BY pa.id DESC
		LIMIT 10
	`).Scan(&allocationData)

	if len(allocationData) == 0 {
		fmt.Println("  ❌ NO PAYMENT ALLOCATIONS FOUND!")
	} else {
		for _, alloc := range allocationData {
			fmt.Printf("  Allocation ID: %d | Payment: %s (ID: %d)\n", alloc.ID, alloc.PaymentCode, alloc.PaymentID)
			if alloc.InvoiceID != nil {
				fmt.Printf("    → Invoice ID: %d (%s) | Amount: Rp %.2f\n\n", *alloc.InvoiceID, *alloc.SaleCode, alloc.AllocatedAmount)
			} else if alloc.BillID != nil {
				fmt.Printf("    → Bill ID: %d | Amount: Rp %.2f\n\n", *alloc.BillID, alloc.AllocatedAmount)
			} else {
				fmt.Printf("    → Generic allocation | Amount: Rp %.2f\n\n", alloc.AllocatedAmount)
			}
		}
	}

	// Check specific sale from the screenshot
	fmt.Println("🎯 SPECIFIC SALE ANALYSIS (INV-2025-3037):")
	var specificSale []struct {
		SaleID            uint    `gorm:"column:sale_id"`
		SaleCode          string  `gorm:"column:sale_code"`
		CustomerName      string  `gorm:"column:customer_name"`
		TotalAmount       float64 `gorm:"column:total_amount"`
		OutstandingAmount float64 `gorm:"column:outstanding_amount"`
		Status            string  `gorm:"column:status"`
		PaymentCount      int     `gorm:"column:payment_count"`
		TotalAllocated    float64 `gorm:"column:total_allocated"`
	}

	db.Raw(`
		SELECT 
			s.id as sale_id,
			s.code as sale_code,
			c.name as customer_name,
			s.total_amount,
			s.outstanding_amount,
			s.status,
			COUNT(pa.id) as payment_count,
			COALESCE(SUM(pa.allocated_amount), 0) as total_allocated
		FROM sales s
		LEFT JOIN contacts c ON s.customer_id = c.id
		LEFT JOIN payment_allocations pa ON pa.invoice_id = s.id
		WHERE s.code LIKE '%3037%' OR s.code LIKE '%2025%'
		GROUP BY s.id, s.code, c.name, s.total_amount, s.outstanding_amount, s.status
		ORDER BY s.id DESC
		LIMIT 3
	`).Scan(&specificSale)

	if len(specificSale) > 0 {
		for _, sale := range specificSale {
			fmt.Printf("  Sale: %s (%s)\n", sale.SaleCode, sale.CustomerName)
			fmt.Printf("    Total: Rp %.2f | Outstanding: Rp %.2f | Status: %s\n", 
				sale.TotalAmount, sale.OutstandingAmount, sale.Status)
			fmt.Printf("    Payment Count: %d | Total Allocated: Rp %.2f\n", 
				sale.PaymentCount, sale.TotalAllocated)
			
			difference := sale.TotalAmount - sale.TotalAllocated
			fmt.Printf("    Expected Outstanding: Rp %.2f\n", difference)
			
			if sale.OutstandingAmount != difference {
				fmt.Printf("    ❌ MISMATCH: Outstanding shows %.2f but should be %.2f\n", 
					sale.OutstandingAmount, difference)
			} else {
				fmt.Printf("    ✅ Outstanding amount is correct\n")
			}
			fmt.Println()
		}
	} else {
		fmt.Println("  No matching sales found")
	}

	// Check for recent payments to this customer
	fmt.Println("💳 RECENT PAYMENTS TO PT Global Tech:")
	var recentPayments []struct {
		PaymentID       uint    `gorm:"column:payment_id"`
		PaymentCode     string  `gorm:"column:payment_code"`
		Amount          float64 `gorm:"column:amount"`
		Status          string  `gorm:"column:status"`
		Date            string  `gorm:"column:date"`
		HasAllocation   bool    `gorm:"column:has_allocation"`
		AllocatedAmount *float64 `gorm:"column:allocated_amount"`
	}

	db.Raw(`
		SELECT 
			p.id as payment_id,
			p.code as payment_code,
			p.amount,
			p.status,
			p.date::text as date,
			CASE WHEN pa.id IS NOT NULL THEN true ELSE false END as has_allocation,
			pa.allocated_amount
		FROM payments p
		LEFT JOIN contacts c ON p.contact_id = c.id
		LEFT JOIN payment_allocations pa ON pa.payment_id = p.id
		WHERE c.name ILIKE '%Global Tech%'
		ORDER BY p.id DESC
		LIMIT 5
	`).Scan(&recentPayments)

	for _, payment := range recentPayments {
		fmt.Printf("  Payment: %s | Amount: Rp %.2f | Status: %s | Date: %s\n", 
			payment.PaymentCode, payment.Amount, payment.Status, payment.Date)
		if payment.HasAllocation {
			fmt.Printf("    ✅ Has allocation: Rp %.2f\n", *payment.AllocatedAmount)
		} else {
			fmt.Printf("    ❌ NO ALLOCATION\n")
		}
		fmt.Println()
	}
}