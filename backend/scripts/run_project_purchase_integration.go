package main

import (
	"app-sistem-akuntansi/database"
	"log"
)

func main() {
	db := database.ConnectDB()

	log.Println("🔄 Running Project-Purchase Integration migration...")

	stmts := []string{
		// Add project_id to purchases
		`ALTER TABLE purchases ADD COLUMN IF NOT EXISTS project_id INTEGER REFERENCES projects(id)`,
		`CREATE INDEX IF NOT EXISTS idx_purchases_project_id ON purchases(project_id)`,
		
		// Add cost tracking fields to projects
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS actual_cost DECIMAL(20,2) DEFAULT 0`,
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS material_cost DECIMAL(20,2) DEFAULT 0`,
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS labor_cost DECIMAL(20,2) DEFAULT 0`,
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS equipment_cost DECIMAL(20,2) DEFAULT 0`,
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS overhead_cost DECIMAL(20,2) DEFAULT 0`,
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS variance DECIMAL(20,2) DEFAULT 0`,
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS variance_percent DECIMAL(5,2) DEFAULT 0`,
		
		// Update variance for existing projects
		`UPDATE projects SET variance = budget - actual_cost, variance_percent = CASE WHEN budget > 0 THEN ((budget - actual_cost) / budget) * 100 ELSE 0 END WHERE budget > 0`,
	}

	for i, stmt := range stmts {
		log.Printf("[%d/%d] Executing...", i+1, len(stmts))
		if err := db.Exec(stmt).Error; err != nil {
			log.Printf("⚠️  Warning: %v", err)
		} else {
			log.Printf("✅ Completed")
		}
	}

	log.Println("✅ Project-Purchase Integration migration completed!")
	
	// Verify
	var colExists bool
	db.Raw(`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'purchases' AND column_name = 'project_id')`).Scan(&colExists)
	if colExists {
		log.Println("✅ Verified: project_id column exists in purchases")
	}
	
	db.Raw(`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'projects' AND column_name = 'actual_cost')`).Scan(&colExists)
	if colExists {
		log.Println("✅ Verified: actual_cost column exists in projects")
	}
	
	log.Println("📊 Integration complete!")
}
