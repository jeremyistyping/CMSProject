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
    
    fmt.Println("ALL ACCOUNTS WITH CODE 2102:")
    rows, err := db.Query("SELECT id, code, name, type, balance, created_at FROM accounts WHERE code = '2102' ORDER BY id")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    for rows.Next() {
        var id int
        var code, name, accType string
        var balance float64
        var createdAt string
        err := rows.Scan(&id, &code, &name, &accType, &balance, &createdAt)
        if err != nil {
            continue
        }
        fmt.Printf("ID:%d - %s %s (%s): %.0f [Created: %s]\n", 
            id, code, name, accType, balance, createdAt)
    }

    fmt.Println("\nJOURNAL LINES USING 2102 ACCOUNTS:")
    rows, err = db.Query(`
        SELECT l.source_type, l.source_code, ujl.account_id, a.name,
               ujl.debit_amount, ujl.credit_amount
        FROM unified_journal_ledger l
        JOIN unified_journal_lines ujl ON ujl.journal_id = l.id
        JOIN accounts a ON ujl.account_id = a.id
        WHERE a.code = '2102'
        ORDER BY l.source_type, ujl.account_id
    `)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    for rows.Next() {
        var sourceType, sourceCode, accountName string
        var accountId int
        var debitAmount, creditAmount float64
        err := rows.Scan(&sourceType, &sourceCode, &accountId, &accountName, 
            &debitAmount, &creditAmount)
        if err != nil {
            continue
        }
        
        if debitAmount > 0 {
            fmt.Printf("%s %s: Dr. Account ID:%d (%s) %.0f\n", 
                sourceType, sourceCode, accountId, accountName, debitAmount)
        }
        if creditAmount > 0 {
            fmt.Printf("%s %s: Cr. Account ID:%d (%s) %.0f\n", 
                sourceType, sourceCode, accountId, accountName, creditAmount)
        }
    }
}