package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=Moon dbname=CMSNew port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("🔗 Connected to database")

	// SQL to create weekly_reports table
	createTableSQL := `
-- Create weekly_reports table
CREATE TABLE IF NOT EXISTS weekly_reports (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    week INTEGER NOT NULL CHECK (week >= 1 AND week <= 53),
    year INTEGER NOT NULL CHECK (year >= 2000),
    project_manager VARCHAR(200),
    total_work_days INTEGER DEFAULT 0,
    weather_delays INTEGER DEFAULT 0,
    team_size INTEGER DEFAULT 0,
    accomplishments TEXT,
    challenges TEXT,
    next_week_priorities TEXT,
    generated_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    CONSTRAINT fk_weekly_reports_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,
    
    -- Ensure only one report per project per week/year
    CONSTRAINT unique_project_week_year
        UNIQUE (project_id, week, year)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_weekly_reports_project_id ON weekly_reports(project_id);
CREATE INDEX IF NOT EXISTS idx_weekly_reports_year ON weekly_reports(year);
CREATE INDEX IF NOT EXISTS idx_weekly_reports_deleted_at ON weekly_reports(deleted_at);
CREATE INDEX IF NOT EXISTS idx_weekly_reports_project_year ON weekly_reports(project_id, year);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_weekly_reports_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_weekly_reports_updated_at
    BEFORE UPDATE ON weekly_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_weekly_reports_updated_at();
`

	fmt.Println("📝 Creating weekly_reports table...")

	// Execute the SQL
	if err := db.Exec(createTableSQL).Error; err != nil {
		log.Fatalf("❌ Failed to create table: %v", err)
	}

	fmt.Println("✅ weekly_reports table created successfully!")

	// Verify the table exists
	var exists bool
	checkSQL := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'weekly_reports'
		);
	`
	
	if err := db.Raw(checkSQL).Scan(&exists).Error; err != nil {
		log.Fatalf("❌ Failed to verify table: %v", err)
	}

	if exists {
		fmt.Println("✅ Verified: weekly_reports table exists in database")
	} else {
		fmt.Println("❌ Warning: weekly_reports table not found after creation")
	}

	// Show table structure
	fmt.Println("\n📊 Table structure:")
	var columns []struct {
		ColumnName string
		DataType   string
		IsNullable string
	}

	if err := db.Raw(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'weekly_reports' 
		ORDER BY ordinal_position
	`).Scan(&columns).Error; err != nil {
		log.Printf("⚠️  Could not fetch table structure: %v", err)
	} else {
		for _, col := range columns {
			fmt.Printf("  - %s (%s) %s\n", col.ColumnName, col.DataType, col.IsNullable)
		}
	}

	fmt.Println("\n🎉 Done! You can now restart your backend server.")
}

