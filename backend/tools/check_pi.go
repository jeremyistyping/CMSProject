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
    
    rows, err := db.Query("SELECT column_name FROM information_schema.columns WHERE table_name = 'purchase_items' ORDER BY ordinal_position")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    fmt.Println("purchase_items columns:")
    for rows.Next() {
        var col string
        rows.Scan(&col)
        fmt.Println("-", col)
    }
}