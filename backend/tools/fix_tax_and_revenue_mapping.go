package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	_ "github.com/lib/pq"
)

func main() {
	// Prefer DATABASE_URL if provided, else build DSN from individual vars
	databaseURL := os.Getenv("DATABASE_URL")
	var dsn string
	if databaseURL != "" {
		dsn = databaseURL
	} else {
		host := getenv("DB_HOST", "localhost")
		port := getenv("DB_PORT", "5432")
		user := getenv("DB_USER", "accounting_user")
		pass := getenv("DB_PASSWORD", "accounting_password")
		dbname := getenv("DB_NAME", "accounting_db")
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil { log.Fatalf("DB open error: %v", err) }
	defer db.Close()
	if err := db.Ping(); err != nil { log.Fatalf("DB ping error: %v", err) }

	log.Println("🔧 Fixing tax and revenue account mapping...")

	// 1) Ensure account 2103 is PPN Keluaran (Liability)
	fix2103 := `
	UPDATE accounts
	SET name = 'PPN Keluaran', type = 'LIABILITY', is_active = true
	WHERE code = '2103';`
	if _, err := db.Exec(fix2103); err != nil { log.Fatalf("Failed to fix 2103: %v", err) }
	log.Println("✅ Updated account 2103 -> PPN Keluaran (Liability)")

	// 2) Ensure 2102 is PPN Masukan (Asset) and active
	fix2102 := `
	UPDATE accounts
	SET name = 'PPN Masukan', type = 'ASSET', is_active = true
	WHERE code = '2102';`
	if _, err := db.Exec(fix2102); err != nil { log.Fatalf("Failed to fix 2102: %v", err) }
	log.Println("✅ Ensured account 2102 -> PPN Masukan (Asset)")

	// 3) Ensure 4101 is Pendapatan Penjualan (Revenue)
	fix4101 := `
	UPDATE accounts
	SET name = 'Pendapatan Penjualan', type = 'REVENUE', is_active = true
	WHERE code = '4101';`
	if _, err := db.Exec(fix4101); err != nil { log.Fatalf("Failed to fix 4101: %v", err) }
	log.Println("✅ Ensured account 4101 -> Pendapatan Penjualan (Revenue)")

	// 4) Optional: zero-out any negative asset balance in 2103 by moving to liability side is handled by correct type;
	// balances are derived from journal lines, so refresh MV next.

	// 5) Refresh MV if exists
	if _, err := db.Exec("REFRESH MATERIALIZED VIEW IF EXISTS account_balances"); err != nil {
		log.Printf("⚠️ Could not refresh materialized view: %v", err)
	} else {
		log.Println("✅ account_balances refreshed")
	}

	log.Println("🎉 Mapping fix complete. Please reload the frontend (Ctrl+F5).")
}

func getenv(k, def string) string { v := os.Getenv(k); if v == "" { return def }; return v }
