package main

import (
	"database/sql"
	"log"
	
	_ "github.com/lib/pq"
)

func main() {
	// Connection string
	connStr := "host=localhost port=5432 user=postgres password=Moon dbname=CMSNew sslmode=disable"
	
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
	log.Println("========================================")
	
	// Check all milestones
	rows, err := db.Query(`
		SELECT id, project_id, title, description, work_area, priority, 
		       assigned_team, target_date, completion_date, status, progress,
		       created_at, updated_at, deleted_at
		FROM milestones
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Fatalf("Failed to query milestones: %v", err)
	}
	defer rows.Close()
	
	log.Println("All Milestones in Database:")
	log.Println("========================================")
	
	count := 0
	for rows.Next() {
		var id, projectID int
		var title, description, workArea, priority, assignedTeam, status string
		var targetDate, createdAt, updatedAt string
		var completionDate, deletedAt sql.NullString
		var progress float64
		
		err := rows.Scan(&id, &projectID, &title, &description, &workArea, &priority,
			&assignedTeam, &targetDate, &completionDate, &status, &progress,
			&createdAt, &updatedAt, &deletedAt)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		
		count++
		log.Printf("\nMilestone #%d:", count)
		log.Printf("  ID: %d", id)
		log.Printf("  Project ID: %d", projectID)
		log.Printf("  Title: %s", title)
		log.Printf("  Description: %s", description)
		log.Printf("  Work Area: %s", workArea)
		log.Printf("  Priority: %s", priority)
		log.Printf("  Assigned Team: %s", assignedTeam)
		log.Printf("  Target Date: %s", targetDate)
		if completionDate.Valid {
			log.Printf("  Completion Date: %s", completionDate.String)
		} else {
			log.Printf("  Completion Date: NULL")
		}
		log.Printf("  Status: %s", status)
		log.Printf("  Progress: %.2f%%", progress)
		log.Printf("  Created At: %s", createdAt)
		log.Printf("  Updated At: %s", updatedAt)
		if deletedAt.Valid {
			log.Printf("  Deleted At: %s (SOFT DELETED)", deletedAt.String)
		} else {
			log.Printf("  Deleted At: NULL (ACTIVE)")
		}
	}
	
	log.Println("\n========================================")
	log.Printf("Total milestones found: %d", count)
	
	// Check milestones count by project
	log.Println("\n========================================")
	log.Println("Milestones count by project:")
	
	rows2, err := db.Query(`
		SELECT project_id, COUNT(*) as count, 
		       COUNT(*) FILTER (WHERE deleted_at IS NULL) as active_count
		FROM milestones
		GROUP BY project_id
		ORDER BY project_id
	`)
	if err != nil {
		log.Printf("Error querying counts: %v", err)
		return
	}
	defer rows2.Close()
	
	for rows2.Next() {
		var projectID, total, active int
		if err := rows2.Scan(&projectID, &total, &active); err != nil {
			log.Printf("Error scanning count: %v", err)
			continue
		}
		log.Printf("  Project ID %d: %d total (%d active, %d deleted)", projectID, total, active, total-active)
	}
}

