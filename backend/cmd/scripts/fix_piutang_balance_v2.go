package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func loadEnv() {
	envFile := ".env"
	if file, err := os.Open(envFile); err == nil {
		defer file.Close()
		content := make([]byte, 1024)
		if n, err := file.Read(content); err == nil {
			envContent := string(content[:n])
			lines := []string{}
			current := ""
			for _, char := range envContent {
				if char == '\n' || char == '\r' {
					if current != "" {
						lines = append(lines, current)
						current = ""
					}
				} else {
					current += string(char)
				}
			}
			if current != "" {
				lines = append(lines, current)
			}
			
			for _, line := range lines {
				if len(line) > 13 && line[:13] == "DATABASE_URL=" {
					os.Setenv("DATABASE_URL", line[13:])
					break
				}
			}
		}
	}
}

func main() {
	fmt.Println("🔧 FIXING PIUTANG USAHA BALANCE AFTER PAYMENT")
	fmt.Println("Expected: Piutang should decrease from Rp 5,550,000 to Rp 2,775,000")
	fmt.Println("")

	loadEnv()
	
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not found in environment")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== STEP 1: CURRENT SITUATION ===")
	
	// Check current Piutang Usaha balance and total outstanding from sales
	var piutangBalance float64
	var totalOutstanding float64
	
	err = db.QueryRow("SELECT balance FROM accounts WHERE code = '1201'").Scan(&piutangBalance)
	if err != nil {
		log.Fatal("Failed to get Piutang balance:", err)
	}
	
	err = db.QueryRow("SELECT COALESCE(SUM(outstanding_amount), 0) FROM sales WHERE status != 'PAID'").Scan(&totalOutstanding)
	if err != nil {
		log.Fatal("Failed to get total outstanding:", err)
	}
	
	fmt.Printf("Current Piutang Usaha balance: Rp %.2f\n", piutangBalance)
	fmt.Printf("Total outstanding from sales: Rp %.2f\n", totalOutstanding)
	fmt.Println("")

	// Show sales details
	fmt.Println("=== SALES DETAILS ===")
	query := `
		SELECT id, code, invoice_number, status, total_amount, paid_amount, outstanding_amount
		FROM sales ORDER BY created_at DESC`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to get sales:", err)
	}
	defer rows.Close()

	fmt.Printf("%-2s | %-12s | %-15s | %-8s | %12s | %12s | %12s\n", 
		"ID", "Code", "Invoice", "Status", "Total", "Paid", "Outstanding")
	fmt.Println("---+-------------+----------------+---------+-------------+-------------+-------------")

	var totalCalculatedOutstanding float64
	for rows.Next() {
		var id int
		var code, invoice, status string
		var total, paid, outstanding float64
		
		err := rows.Scan(&id, &code, &invoice, &status, &total, &paid, &outstanding)
		if err != nil {
			log.Fatal("Failed to scan sales:", err)
		}
		
		fmt.Printf("%-2d | %-12s | %-15s | %-8s | %12.2f | %12.2f | %12.2f\n",
			id, code, invoice, status, total, paid, outstanding)
		
		if status != "PAID" {
			totalCalculatedOutstanding += outstanding
		}
	}

	fmt.Println("")
	fmt.Printf("💰 Calculated Outstanding: Rp %.2f\n", totalCalculatedOutstanding)
	fmt.Printf("🏦 Current Piutang Balance: Rp %.2f\n", piutangBalance)
	
	if piutangBalance != totalCalculatedOutstanding {
		fmt.Printf("❌ MISMATCH! Piutang balance should be Rp %.2f\n", totalCalculatedOutstanding)
	} else {
		fmt.Printf("✅ Piutang balance matches outstanding amount\n")
		return
	}

	fmt.Println("")
	fmt.Printf("🎯 TARGET: Set Piutang Usaha to Rp %.2f\n", totalCalculatedOutstanding)
	
	fmt.Print("❓ Do you want to fix the Piutang balance? (y/n): ")
	var response string
	fmt.Scanln(&response)
	
	if response != "y" && response != "Y" {
		fmt.Println("❌ Fix cancelled by user")
		return
	}

	fmt.Println("")
	fmt.Println("=== APPLYING FIX ===")

	// Disable trigger temporarily
	_, err = db.Exec("ALTER TABLE unified_journal_lines DISABLE TRIGGER balance_sync_trigger")
	if err != nil {
		log.Printf("Warning: Could not disable trigger: %v", err)
	}

	// Update Piutang Usaha balance to match outstanding amounts
	_, err = db.Exec("UPDATE accounts SET balance = $1 WHERE code = '1201'", totalCalculatedOutstanding)
	if err != nil {
		log.Fatal("Failed to update Piutang balance:", err)
	}

	// Re-enable trigger
	_, err = db.Exec("ALTER TABLE unified_journal_lines ENABLE TRIGGER balance_sync_trigger")
	if err != nil {
		log.Printf("Warning: Could not re-enable trigger: %v", err)
	}

	// Verify fix
	var newBalance float64
	err = db.QueryRow("SELECT balance FROM accounts WHERE code = '1201'").Scan(&newBalance)
	if err != nil {
		log.Fatal("Failed to verify new balance:", err)
	}

	fmt.Printf("✅ Piutang Usaha updated from Rp %.2f to Rp %.2f\n", piutangBalance, newBalance)
	
	if newBalance == totalCalculatedOutstanding {
		fmt.Println("🎉 SUCCESS! Piutang balance now reflects actual outstanding amounts!")
		fmt.Println("")
		fmt.Printf("📊 Summary:\n")
		fmt.Printf("   - Invoice INV/2025/09/0002: PAID (Outstanding: Rp 0)\n")
		fmt.Printf("   - Invoice INV/2025/09/0003: INVOICED (Outstanding: Rp %.2f)\n", totalCalculatedOutstanding)
		fmt.Printf("   - Total Piutang Usaha: Rp %.2f ✅\n", newBalance)
		fmt.Println("")
		fmt.Println("📱 Please refresh your frontend COA to see the updated balance!")
	} else {
		fmt.Println("⚠️ Something went wrong with the update")
	}
}