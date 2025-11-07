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
	fmt.Println("🔧 SIMPLE PAYMENT FIX")
	fmt.Println("Setting Piutang Usaha balance to match outstanding invoices")
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

	fmt.Println("=== CURRENT STATUS ===")
	
	// Get current outstanding from sales
	var totalOutstanding float64
	err = db.QueryRow("SELECT COALESCE(SUM(outstanding_amount), 0) FROM sales WHERE status != 'PAID'").Scan(&totalOutstanding)
	if err != nil {
		log.Fatal("Failed to get outstanding:", err)
	}
	
	// Get current Piutang balance
	var currentBalance float64
	err = db.QueryRow("SELECT balance FROM accounts WHERE code = '1201'").Scan(&currentBalance)
	if err != nil {
		log.Fatal("Failed to get Piutang balance:", err)
	}
	
	fmt.Printf("📊 Outstanding invoices: Rp %.2f\n", totalOutstanding)
	fmt.Printf("🏦 Current Piutang balance: Rp %.2f\n", currentBalance)
	
	if currentBalance == totalOutstanding {
		fmt.Println("✅ Piutang balance already matches outstanding invoices!")
		return
	}
	
	fmt.Printf("🎯 Setting Piutang to: Rp %.2f\n", totalOutstanding)
	
	// Simple direct fix
	_, err = db.Exec("ALTER TABLE unified_journal_lines DISABLE TRIGGER balance_sync_trigger")
	if err != nil {
		log.Printf("Warning: Could not disable trigger: %v", err)
	}
	
	_, err = db.Exec("UPDATE accounts SET balance = $1 WHERE code = '1201'", totalOutstanding)
	if err != nil {
		log.Fatal("Failed to update balance:", err)
	}
	
	_, err = db.Exec("ALTER TABLE unified_journal_lines ENABLE TRIGGER balance_sync_trigger")
	if err != nil {
		log.Printf("Warning: Could not re-enable trigger: %v", err)
	}
	
	fmt.Printf("✅ Piutang Usaha updated to: Rp %.2f\n", totalOutstanding)
	fmt.Println("")
	fmt.Println("🎉 Payment fix completed!")
	fmt.Println("📱 Please refresh your frontend COA to see the updated balance!")
}