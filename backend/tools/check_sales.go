package main

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/lib/pq"
)

func main() {
    dsn := "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    fmt.Println("SALES TABLE COLUMNS:")
    rows, err := db.Query("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'sales' ORDER BY ordinal_position")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    for rows.Next() {
        var col, dataType string
        rows.Scan(&col, &dataType)
        fmt.Printf("- %s (%s)\n", col, dataType)
    }

    fmt.Println("\nSALE_ITEMS TABLE COLUMNS:")
    rows, err = db.Query("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'sale_items' ORDER BY ordinal_position")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    for rows.Next() {
        var col, dataType string
        rows.Scan(&col, &dataType)
        fmt.Printf("- %s (%s)\n", col, dataType)
    }

    fmt.Println("\nSAMPLE SALES DATA:")
    rows, err = db.Query("SELECT id, code, total_amount, status FROM sales ORDER BY id LIMIT 5")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        defer rows.Close()
        for rows.Next() {
            var id int
            var code, status string
            var total float64
            rows.Scan(&id, &code, &total, &status)
            fmt.Printf("ID:%d - %s: %.0f (%s)\n", id, code, total, status)
        }
    }
}