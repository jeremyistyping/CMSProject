package main

import (
	"app-sistem-akuntansi/database"
	"log"
)

func main() {
	// Connect to database
	db := database.ConnectDB()

	log.Println("🔄 Running Cost Control approval migration...")

	// Execute migration statements one by one
	stmts := []string{
		// Add Cost Control fields
		`ALTER TABLE purchases ADD COLUMN IF NOT EXISTS cost_control_approved_by INTEGER REFERENCES users(id)`,
		`ALTER TABLE purchases ADD COLUMN IF NOT EXISTS cost_control_approved_at TIMESTAMPTZ`,
		`ALTER TABLE purchases ADD COLUMN IF NOT EXISTS cost_control_comments TEXT`,
		
		// Add GM fields
		`ALTER TABLE purchases ADD COLUMN IF NOT EXISTS gm_approved_by INTEGER REFERENCES users(id)`,
		`ALTER TABLE purchases ADD COLUMN IF NOT EXISTS gm_approved_at TIMESTAMPTZ`,
		`ALTER TABLE purchases ADD COLUMN IF NOT EXISTS gm_comments TEXT`,
		
		// Add current step tracker
		`ALTER TABLE purchases ADD COLUMN IF NOT EXISTS current_approval_step VARCHAR(20) DEFAULT 'NONE'`,
		
		// Update approval_status length
		`ALTER TABLE purchases ALTER COLUMN approval_status TYPE VARCHAR(30)`,
		
		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_purchases_cost_control_approved_by ON purchases(cost_control_approved_by)`,
		`CREATE INDEX IF NOT EXISTS idx_purchases_gm_approved_by ON purchases(gm_approved_by)`,
		`CREATE INDEX IF NOT EXISTS idx_purchases_current_approval_step ON purchases(current_approval_step)`,
		`CREATE INDEX IF NOT EXISTS idx_purchases_approval_status ON purchases(approval_status)`,
		
		// Update existing data
		`UPDATE purchases SET current_approval_step = CASE WHEN approval_status = 'APPROVED' THEN 'COMPLETED' WHEN approval_status IN ('PENDING', 'PENDING_APPROVAL') THEN 'COST_CONTROL' ELSE 'NONE' END WHERE current_approval_step IS NULL OR current_approval_step = ''`,
	}

	for i, stmt := range stmts {
		log.Printf("[%d/%d] Executing statement...", i+1, len(stmts))
		if err := db.Exec(stmt).Error; err != nil {
			log.Printf("⚠️  Warning on statement %d: %v", i+1, err)
		} else {
			log.Printf("✅ Statement %d completed", i+1)
		}
	}

	log.Println("✅ Cost Control approval migration completed!")
	
	// Verify new columns exist
	var columnExists bool
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'purchases' 
			AND column_name = 'cost_control_approved_by'
		)
	`).Scan(&columnExists)
	
	if columnExists {
		log.Println("✅ Verified: cost_control_approved_by column exists")
	} else {
		log.Println("⚠️  Warning: cost_control_approved_by column not found")
	}
	
	// Check current_approval_step column
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'purchases' 
			AND column_name = 'current_approval_step'
		)
	`).Scan(&columnExists)
	
	if columnExists {
		log.Println("✅ Verified: current_approval_step column exists")
	} else {
		log.Println("⚠️  Warning: current_approval_step column not found")
	}
	
	log.Println("📊 Migration summary complete!")
}
