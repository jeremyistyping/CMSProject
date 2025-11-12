package main

import (
	"log"
	"os"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Database connection string
	dsn := "host=localhost user=postgres password=Moon dbname=CMSNew port=5432 sslmode=disable"
	
	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	log.Println("Connected to database")
	
	// Create milestones table
	sqlCreate := `
-- Create milestones table
CREATE TABLE IF NOT EXISTS milestones (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    target_date TIMESTAMP NOT NULL,
    actual_completion_date TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
    order_number INTEGER DEFAULT 0,
    weight DECIMAL(5,2) DEFAULT 0 CHECK (weight >= 0 AND weight <= 100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT fk_milestone_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_milestones_project_id ON milestones(project_id);
CREATE INDEX IF NOT EXISTS idx_milestones_status ON milestones(status);
CREATE INDEX IF NOT EXISTS idx_milestones_target_date ON milestones(target_date);
CREATE INDEX IF NOT EXISTS idx_milestones_deleted_at ON milestones(deleted_at);

-- Create composite index for common queries
CREATE INDEX IF NOT EXISTS idx_milestones_project_status ON milestones(project_id, status) WHERE deleted_at IS NULL;
`
	
	// Execute SQL
	if err := db.Exec(sqlCreate).Error; err != nil {
		log.Fatalf("Failed to create milestones table: %v", err)
	}
	
	log.Println("✅ Milestones table created successfully!")
	
	// Verify table exists
	var exists bool
	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'milestones')").Scan(&exists).Error
	if err != nil {
		log.Printf("Warning: Could not verify table creation: %v", err)
	} else if exists {
		log.Println("✅ Verified: milestones table exists in database")
		
		// Count rows
		var count int64
		db.Raw("SELECT COUNT(*) FROM milestones").Scan(&count)
		log.Printf("📊 Current milestone count: %d", count)
	} else {
		log.Println("❌ Warning: Table might not have been created properly")
	}
	
	os.Exit(0)
}

