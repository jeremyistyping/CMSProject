package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("❌ DATABASE_URL not found in environment variables")
	}

	log.Println("🔄 Starting database reset process...")
	log.Println("⚠️  WARNING: This will DELETE ALL DATA in the database!")
	log.Println("📍 Database:", maskPassword(databaseURL))

	// Ask for confirmation
	fmt.Print("\n❓ Are you sure you want to reset the database? Type 'YES' to confirm: ")
	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != "YES" {
		log.Println("❌ Database reset cancelled")
		return
	}

	// Connect to database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to database successfully")

	// Step 1: Drop all tables
	log.Println("\n🗑️  Step 1: Dropping all tables...")
	if err := dropAllTables(db); err != nil {
		log.Fatalf("❌ Failed to drop tables: %v", err)
	}
	log.Println("✅ All tables dropped successfully")

	// Step 2: Drop all sequences
	log.Println("\n🗑️  Step 2: Dropping all sequences...")
	if err := dropAllSequences(db); err != nil {
		log.Fatalf("❌ Failed to drop sequences: %v", err)
	}
	log.Println("✅ All sequences dropped successfully")

	// Step 3: Drop all types (enums)
	log.Println("\n🗑️  Step 3: Dropping all custom types...")
	if err := dropAllTypes(db); err != nil {
		log.Fatalf("❌ Failed to drop types: %v", err)
	}
	log.Println("✅ All custom types dropped successfully")

	log.Println("\n✅ Database reset completed successfully!")
	log.Println("📝 Next steps:")
	log.Println("   1. Run your application to apply migrations")
	log.Println("   2. Or run migrations manually using golang-migrate")
	log.Println("\n💡 To start fresh, run: go run main.go")
}

// dropAllTables drops all tables in the public schema
func dropAllTables(db *sql.DB) error {
	// Get all table names
	query := `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public'
		ORDER BY tablename;
	`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if len(tables) == 0 {
		log.Println("   ℹ️  No tables found to drop")
		return nil
	}

	log.Printf("   📋 Found %d tables to drop", len(tables))

	// Drop all tables with CASCADE
	for _, table := range tables {
		dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", table)
		if _, err := db.Exec(dropQuery); err != nil {
			log.Printf("   ⚠️  Warning: Failed to drop table %s: %v", table, err)
		} else {
			log.Printf("   ✓ Dropped table: %s", table)
		}
	}

	return nil
}

// dropAllSequences drops all sequences in the public schema
func dropAllSequences(db *sql.DB) error {
	query := `
		SELECT sequence_name 
		FROM information_schema.sequences 
		WHERE sequence_schema = 'public'
		ORDER BY sequence_name;
	`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query sequences: %w", err)
	}
	defer rows.Close()

	var sequences []string
	for rows.Next() {
		var seqName string
		if err := rows.Scan(&seqName); err != nil {
			return fmt.Errorf("failed to scan sequence name: %w", err)
		}
		sequences = append(sequences, seqName)
	}

	if len(sequences) == 0 {
		log.Println("   ℹ️  No sequences found to drop")
		return nil
	}

	log.Printf("   📋 Found %d sequences to drop", len(sequences))

	for _, seq := range sequences {
		dropQuery := fmt.Sprintf("DROP SEQUENCE IF EXISTS %s CASCADE;", seq)
		if _, err := db.Exec(dropQuery); err != nil {
			log.Printf("   ⚠️  Warning: Failed to drop sequence %s: %v", seq, err)
		} else {
			log.Printf("   ✓ Dropped sequence: %s", seq)
		}
	}

	return nil
}

// dropAllTypes drops all custom types (enums) in the public schema
func dropAllTypes(db *sql.DB) error {
	query := `
		SELECT t.typname
		FROM pg_type t
		JOIN pg_namespace n ON n.oid = t.typnamespace
		WHERE n.nspname = 'public'
		AND t.typtype = 'e'
		ORDER BY t.typname;
	`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query types: %w", err)
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var typeName string
		if err := rows.Scan(&typeName); err != nil {
			return fmt.Errorf("failed to scan type name: %w", err)
		}
		types = append(types, typeName)
	}

	if len(types) == 0 {
		log.Println("   ℹ️  No custom types found to drop")
		return nil
	}

	log.Printf("   📋 Found %d custom types to drop", len(types))

	for _, typeName := range types {
		dropQuery := fmt.Sprintf("DROP TYPE IF EXISTS %s CASCADE;", typeName)
		if _, err := db.Exec(dropQuery); err != nil {
			log.Printf("   ⚠️  Warning: Failed to drop type %s: %v", typeName, err)
		} else {
			log.Printf("   ✓ Dropped type: %s", typeName)
		}
	}

	return nil
}

// maskPassword masks the password in database URL for logging
func maskPassword(dbURL string) string {
	// Format: postgres://username:password@host:port/database
	if strings.Contains(dbURL, "@") {
		parts := strings.Split(dbURL, "@")
		if len(parts) == 2 {
			userPart := parts[0]
			if strings.Contains(userPart, ":") {
				userParts := strings.Split(userPart, ":")
				if len(userParts) >= 2 {
					return userParts[0] + ":****@" + parts[1]
				}
			}
		}
	}
	return dbURL
}
