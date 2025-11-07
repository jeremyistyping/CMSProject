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
    
    // Check table structure
    fmt.Println("🔍 Migration Logs Table Structure:")
    rows, err := db.Query("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'migration_logs' ORDER BY ordinal_position")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    for rows.Next() {
        var column, dataType string
        rows.Scan(&column, &dataType)
        fmt.Printf("   %s: %s\n", column, dataType)
    }
    
    // Check if record exists with different field name
    fmt.Println("\n🔍 Checking all records that might match 026:")
    rows, err = db.Query("SELECT migration_name, status, executed_at FROM migration_logs WHERE migration_name LIKE '%026%' OR migration_name LIKE '%purchase_balance%'")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    found := false
    for rows.Next() {
        found = true
        var migrationName, status, executedAt string
        rows.Scan(&migrationName, &status, &executedAt)
        fmt.Printf("   ✅ %s: %s (%s)\n", migrationName, status, executedAt)
    }
    
    if !found {
        fmt.Println("   ❌ No records found with 026 or purchase_balance")
    }
    
    // Check filename field (might be different)
    fmt.Println("\n🔍 Checking if there's filename field:")
    rows, err = db.Query("SELECT filename, status FROM migration_logs WHERE filename LIKE '%026%' LIMIT 5")
    if err != nil {
        fmt.Println("   ❌ No filename field exists")
    } else {
        defer rows.Close()
        found := false
        for rows.Next() {
            found = true
            var filename, status string
            rows.Scan(&filename, &status)
            fmt.Printf("   ✅ %s: %s\n", filename, status)
        }
        if !found {
            fmt.Println("   ❌ No records found with filename 026")
        }
    }
}