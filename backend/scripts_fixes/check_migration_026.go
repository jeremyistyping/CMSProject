package main

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=sistem_akuntans_test sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    var exists bool
    err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM migration_logs WHERE migration_name = '026_purchase_balance_minimal')").Scan(&exists)
    if err != nil {
        log.Fatal(err)
    }
    
    if exists {
        fmt.Println("✅ Migration 026_purchase_balance_minimal HAS BEEN EXECUTED")
    } else {
        fmt.Println("❌ Migration 026_purchase_balance_minimal NOT EXECUTED")
    }
    
    // Check if account 2101 exists
    err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM accounts WHERE code = '2101')").Scan(&exists)
    if err != nil {
        log.Fatal(err)
    }
    
    if exists {
        fmt.Println("✅ Account 2101 (Hutang Usaha) EXISTS")
    } else {
        fmt.Println("❌ Account 2101 (Hutang Usaha) NOT FOUND")
    }
    
    // Check total migration count
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM migration_logs").Scan(&count)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("📊 Total migrations executed: %d\n", count)
}