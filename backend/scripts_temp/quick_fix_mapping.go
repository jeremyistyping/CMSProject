package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	
	fmt.Println("Fixing account mapping...")
	
	// Fix wrong account mapping (Piutang Usaha → Utang Usaha)
	result := db.Exec(`
		UPDATE unified_journal_lines 
		SET account_id = (SELECT id FROM accounts WHERE code = '2101') 
		WHERE account_id = (SELECT id FROM accounts WHERE code = '1201')
	`)
	fmt.Printf("Updated %d journal lines\n", result.RowsAffected)
	
	// Fix balances
	db.Exec("UPDATE accounts SET balance = 0 WHERE code = '1201'")
	db.Exec("UPDATE accounts SET balance = 5550000 WHERE code = '2101'")
	
	fmt.Println("Fixed account balances")
}