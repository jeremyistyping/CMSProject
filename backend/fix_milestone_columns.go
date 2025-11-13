package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	
	_ "github.com/lib/pq"
)

func main() {
	// Read database connection from environment or use defaults
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "admin")
	dbPassword := getEnv("DB_PASSWORD", "admin123")
	dbName := getEnv("DB_NAME", "construction_db")
	
	// Connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)
	
	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	
	log.Println("Connected to database successfully")
	
	// Check if columns already exist
	var workAreaExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'milestones' AND column_name = 'work_area'
		)
	`).Scan(&workAreaExists)
	if err != nil {
		log.Fatalf("Failed to check work_area column: %v", err)
	}
	
	if workAreaExists {
		log.Println("✅ work_area column already exists")
	} else {
		log.Println("Adding work_area column...")
		_, err = db.Exec("ALTER TABLE milestones ADD COLUMN work_area VARCHAR(100)")
		if err != nil {
			log.Fatalf("Failed to add work_area column: %v", err)
		}
		log.Println("✅ work_area column added successfully")
	}
	
	// Check priority column
	var priorityExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'milestones' AND column_name = 'priority'
		)
	`).Scan(&priorityExists)
	if err != nil {
		log.Fatalf("Failed to check priority column: %v", err)
	}
	
	if priorityExists {
		log.Println("✅ priority column already exists")
	} else {
		log.Println("Adding priority column...")
		_, err = db.Exec("ALTER TABLE milestones ADD COLUMN priority VARCHAR(20) DEFAULT 'medium'")
		if err != nil {
			log.Fatalf("Failed to add priority column: %v", err)
		}
		log.Println("✅ priority column added successfully")
	}
	
	// Check assigned_team column
	var assignedTeamExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'milestones' AND column_name = 'assigned_team'
		)
	`).Scan(&assignedTeamExists)
	if err != nil {
		log.Fatalf("Failed to check assigned_team column: %v", err)
	}
	
	if assignedTeamExists {
		log.Println("✅ assigned_team column already exists")
	} else {
		log.Println("Adding assigned_team column...")
		_, err = db.Exec("ALTER TABLE milestones ADD COLUMN assigned_team VARCHAR(200)")
		if err != nil {
			log.Fatalf("Failed to add assigned_team column: %v", err)
		}
		log.Println("✅ assigned_team column added successfully")
	}
	
	// Also fix the completion_date column name mismatch if needed
	var completionDateExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'milestones' AND column_name = 'completion_date'
		)
	`).Scan(&completionDateExists)
	if err != nil {
		log.Fatalf("Failed to check completion_date column: %v", err)
	}
	
	var actualCompletionDateExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'milestones' AND column_name = 'actual_completion_date'
		)
	`).Scan(&actualCompletionDateExists)
	if err != nil {
		log.Fatalf("Failed to check actual_completion_date column: %v", err)
	}
	
	if !completionDateExists && actualCompletionDateExists {
		log.Println("Renaming actual_completion_date to completion_date...")
		_, err = db.Exec("ALTER TABLE milestones RENAME COLUMN actual_completion_date TO completion_date")
		if err != nil {
			log.Fatalf("Failed to rename column: %v", err)
		}
		log.Println("✅ Column renamed successfully")
	} else if completionDateExists {
		log.Println("✅ completion_date column exists")
	}
	
	// Create indexes
	log.Println("Creating indexes...")
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_milestones_work_area ON milestones(work_area) WHERE deleted_at IS NULL")
	if err != nil {
		log.Printf("⚠️  Warning creating work_area index: %v", err)
	}
	
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_milestones_priority ON milestones(priority) WHERE deleted_at IS NULL")
	if err != nil {
		log.Printf("⚠️  Warning creating priority index: %v", err)
	}
	
	log.Println("✅ All milestone table fixes completed successfully!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

