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
		// Simple env loading - look for DATABASE_URL
		content := make([]byte, 1024)
		if n, err := file.Read(content); err == nil {
			envContent := string(content[:n])
			// Parse DATABASE_URL
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
	fmt.Println("🔍 CHECKING SALES TABLE STRUCTURE & DATA")
	fmt.Println("")

	loadEnv()
	
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not found in environment")
	}

	fmt.Printf("🔧 DATABASE_URL: %s\n", maskPassword(dbURL))
	fmt.Println("")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== STEP 1: SALES TABLE STRUCTURE ===")
	
	// Check table structure
	query := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_name = 'sales' 
		ORDER BY ordinal_position`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to get table structure:", err)
	}
	defer rows.Close()

	fmt.Printf("%-20s | %-15s | %-10s | %-20s\n", "Column", "Data Type", "Nullable", "Default")
	fmt.Println("---------------------+-----------------+------------+---------------------")

	for rows.Next() {
		var colName, dataType, nullable, defaultVal sql.NullString
		err := rows.Scan(&colName, &dataType, &nullable, &defaultVal)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}
		
		def := "NULL"
		if defaultVal.Valid {
			def = defaultVal.String
		}
		
		fmt.Printf("%-20s | %-15s | %-10s | %-20s\n", 
			colName.String, dataType.String, nullable.String, def)
	}

	fmt.Println("")
	fmt.Println("=== STEP 2: ACTUAL SALES DATA ===")
	
	// Try different combinations to get sales data
	queries := []string{
		"SELECT id, code, invoice_number, status, amount, created_at FROM sales ORDER BY created_at DESC LIMIT 5",
		"SELECT id, code, invoice_number, status, total, created_at FROM sales ORDER BY created_at DESC LIMIT 5", 
		"SELECT id, code, invoice_number, status, grand_total, created_at FROM sales ORDER BY created_at DESC LIMIT 5",
		"SELECT id, code, invoice_number, status, subtotal, tax, total_amount, created_at FROM sales ORDER BY created_at DESC LIMIT 5",
		"SELECT * FROM sales ORDER BY created_at DESC LIMIT 2", // Get all columns
	}

	for i, query := range queries {
		fmt.Printf("Trying query %d...\n", i+1)
		rows, err := db.Query(query)
		if err != nil {
			fmt.Printf("❌ Query %d failed: %v\n", i+1, err)
			continue
		}
		
		fmt.Printf("✅ Query %d worked!\n", i+1)
		
		// Get column names
		cols, _ := rows.Columns()
		fmt.Printf("Columns: %v\n", cols)
		
		// Get data
		for rows.Next() {
			values := make([]interface{}, len(cols))
			valuePtrs := make([]interface{}, len(cols))
			for i := range cols {
				valuePtrs[i] = &values[i]
			}
			
			rows.Scan(valuePtrs...)
			
			fmt.Printf("Row: ")
			for i, col := range cols {
				val := values[i]
				if val == nil {
					fmt.Printf("%s=NULL ", col)
				} else {
					switch v := val.(type) {
					case []byte:
						fmt.Printf("%s=%s ", col, string(v))
					default:
						fmt.Printf("%s=%v ", col, v)
					}
				}
			}
			fmt.Println("")
		}
		rows.Close()
		break // Stop after first successful query
	}

	fmt.Println("")
	fmt.Println("=== STEP 3: SALES ITEMS DATA ===")
	
	// Check sales items
	query = `
		SELECT si.sales_id, si.product_name, si.quantity, si.price, si.total,
		       s.code as sales_code, s.invoice_number
		FROM sales_items si
		LEFT JOIN sales s ON si.sales_id = s.id
		ORDER BY si.sales_id, si.id`
	
	rows, err = db.Query(query)
	if err != nil {
		fmt.Printf("❌ Sales items query failed: %v\n", err)
	} else {
		fmt.Println("Sales Items Found:")
		fmt.Printf("%-8s | %-15s | %-8s | %-12s | %-12s | %-15s\n", 
			"Sale ID", "Product", "Qty", "Price", "Total", "Invoice")
		fmt.Println("---------+-----------------+----------+--------------+--------------+----------------")
		
		totalAmount := 0.0
		for rows.Next() {
			var salesID int
			var productName, salesCode, invoiceNum string
			var quantity int
			var price, total float64
			
			err := rows.Scan(&salesID, &productName, &quantity, &price, &total, &salesCode, &invoiceNum)
			if err != nil {
				log.Printf("Error scanning sales item: %v", err)
				continue
			}
			
			fmt.Printf("%-8d | %-15s | %-8d | %12.2f | %12.2f | %-15s\n", 
				salesID, productName, quantity, price, total, invoiceNum)
			
			totalAmount += total
		}
		rows.Close()
		
		fmt.Printf("\n💰 Total from Sales Items: Rp %.2f\n", totalAmount)
	}

	fmt.Println("")
	fmt.Println("🏁 TABLE STRUCTURE CHECK COMPLETE!")
}

func maskPassword(dbURL string) string {
	// Simple password masking
	masked := ""
	inPassword := false
	for i, char := range dbURL {
		if i > 0 && dbURL[i-1] == ':' && char != '/' {
			inPassword = true
		}
		if inPassword && char == '@' {
			inPassword = false
			masked += "@"
			continue
		}
		if inPassword {
			masked += "*"
		} else {
			masked += string(char)
		}
	}
	return masked
}