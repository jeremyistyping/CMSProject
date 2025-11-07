package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Drop old unique constraint (constraint, not just index)
	log.Println("🔄 Dropping old unique constraint...")
	_, err = db.Exec("ALTER TABLE product_categories DROP CONSTRAINT IF EXISTS uni_product_categories_code")
	if err != nil {
		log.Printf("⚠️  Warning dropping old constraint: %v", err)
	} else {
		log.Println("✅ Old constraint dropped")
	}

	// Create new partial unique index that allows duplicates for soft deleted records
	log.Println("🔄 Creating new unique index that supports soft delete...")
	_, err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_category_code_deleted ON product_categories (code, deleted_at)")
	if err != nil {
		log.Fatalf("❌ Failed to create new index: %v", err)
	}

	log.Println("✅ New unique index created successfully!")
	log.Println("🎉 Migration completed! You can now create categories with previously deleted codes.")
}
