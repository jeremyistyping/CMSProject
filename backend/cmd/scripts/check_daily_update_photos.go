package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Build connection string
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if sslmode == "" {
		sslmode = "disable"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("✅ Connected to database successfully")

	// Query latest daily updates with photos
	query := `
		SELECT id, project_id, date, photos, work_description, created_at
		FROM daily_updates
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 3
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to query daily updates:", err)
	}
	defer rows.Close()

	log.Println("\n📸 Latest Daily Updates with Photos:")
	log.Println("=====================================")

	for rows.Next() {
		var id, projectID int
		var date, workDescription, createdAt string
		var photos pq.StringArray

		err := rows.Scan(&id, &projectID, &date, &photos, &workDescription, &createdAt)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		log.Printf("\n📋 Daily Update ID: %d", id)
		log.Printf("   Project ID: %d", projectID)
		log.Printf("   Date: %s", date)
		log.Printf("   Description: %s", workDescription)
		log.Printf("   Created: %s", createdAt)
		log.Printf("   Number of photos: %d", len(photos))
		
		if len(photos) > 0 {
			log.Println("   Photos:")
			for i, photo := range photos {
				log.Printf("     %d. %s", i+1, photo)
			}
		} else {
			log.Println("   ⚠️ No photos")
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal("Error iterating rows:", err)
	}
}

