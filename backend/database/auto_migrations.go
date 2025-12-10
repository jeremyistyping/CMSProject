package database

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

// Global counter for idempotent skips (thread-safe)
var idempotentSkipCount int64

// RunAutoMigrations runs all pending SQL migrations automatically
func RunAutoMigrations(db *gorm.DB) error {
	log.Println("🔄 Starting auto-migrations...")

	// Create migration_logs table first if it doesn't exist
	if err := createMigrationLogsTable(db); err != nil {
		return fmt.Errorf("failed to create migration_logs table: %v", err)
	}

	// Run pre-migration fixes to ensure compatibility
	if err := runPreMigrationFixes(db); err != nil {
		log.Printf("⚠️  Pre-migration fixes failed: %v", err)
	}

	// Silently verify settings history table
	if err := ensureSettingsHistoryTable(db); err != nil {
		log.Printf("⚠️  Settings history table verification failed: %v", err)
	}

	// Get migration files
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %v", err)
	}

	// Reset idempotent skip counter
	atomic.StoreInt64(&idempotentSkipCount, 0)

	// Run each migration
	successCount := 0
	failedCount := 0
	for _, file := range migrationFiles {
		if err := runMigration(db, file); err != nil {
			log.Printf("❌ Migration failed: %s - %v", file, err)
			failedCount++
			continue
		}
		successCount++
	}

	// Print concise migration summary
	if failedCount > 0 {
		log.Printf("⚠️  %d migration(s) failed. Check logs above for details.", failedCount)
	} else if successCount > 0 {
		log.Printf("✅ Database migrations: %d completed successfully", successCount)
	}

	// Check and create Standard Purchase Approval workflow
	if err := ensureStandardPurchaseApprovalWorkflow(db); err != nil {
		log.Printf("⚠️  WORKFLOW AUTO-MIGRATION FAILED: %v", err)
	}

	// Fix audit_logs schema
	if err := AutoFixAuditLogsSchema(db); err != nil {
		log.Printf("⚠️  AUDIT_LOGS SCHEMA FIX FAILED: %v", err)
	}

	// Fix activity_logs user_id constraint
	if err := FixActivityLogsUserIDMigration(db); err != nil {
		log.Printf("⚠️  ACTIVITY_LOGS USER_ID FIX FAILED: %v", err)
	}

	// Ensure project_budgets table exists
	if err := ensureProjectBudgetsTable(db); err != nil {
		log.Printf("⚠️  PROJECT_BUDGETS TABLE CREATION FAILED: %v", err)
	}

	log.Println("✅ Auto-migrations completed")
	return nil
}

// createMigrationLogsTable creates the migration_logs table if it doesn't exist
func createMigrationLogsTable(db *gorm.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS migration_logs (
		id SERIAL PRIMARY KEY,
		migration_name VARCHAR(255) NOT NULL UNIQUE,
		status VARCHAR(20) NOT NULL DEFAULT 'SUCCESS' CHECK (status IN ('SUCCESS', 'FAILED', 'SKIPPED')),
		message TEXT,
		description TEXT,
		executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		execution_time_ms INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	if err := db.Exec(createTableSQL).Error; err != nil {
		return err
	}

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_migration_logs_name ON migration_logs(migration_name)",
		"CREATE INDEX IF NOT EXISTS idx_migration_logs_status ON migration_logs(status)",
		"CREATE INDEX IF NOT EXISTS idx_migration_logs_executed_at ON migration_logs(executed_at)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}

	return nil
}

// getMigrationFiles gets all SQL migration files sorted by name
func getMigrationFiles() ([]string, error) {
	primaryDir, err := findMigrationDir()
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(primaryDir)
	if err != nil {
		return nil, err
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	sort.Strings(migrationFiles)
	log.Printf("Using migration dir: %s", primaryDir)
	log.Printf("Found %d migration files", len(migrationFiles))

	return migrationFiles, nil
}

// findMigrationDir tries multiple locations to locate the migrations folder
func findMigrationDir() (string, error) {
	candidates := []string{}

	if envDir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR")); envDir != "" {
		candidates = append(candidates, filepath.Clean(envDir))
	}

	cwd, _ := os.Getwd()
	if cwd != "" {
		candidates = append(candidates,
			filepath.Clean(filepath.Join(cwd, "migrations")),
			filepath.Clean(filepath.Join(cwd, "backend", "migrations")),
		)
	}

	candidates = append(candidates,
		filepath.Clean("./migrations"),
		filepath.Clean("backend/migrations"),
	)

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}
	return "", fmt.Errorf("migrations directory not found")
}

// runMigration runs a single migration file
func runMigration(db *gorm.DB, filename string) error {
	var lastStatus string
	statusErr := db.Raw("SELECT status FROM migration_logs WHERE migration_name = ? ORDER BY executed_at DESC LIMIT 1", filename).Scan(&lastStatus).Error
	if statusErr == nil && strings.EqualFold(lastStatus, "SUCCESS") {
		return nil
	}

	startTime := time.Now()
	log.Printf("🔄 Running: %s", filename)

	migrationDir, dirErr := findMigrationDir()
	if dirErr != nil {
		logMigrationResult(db, filename, "FAILED", fmt.Sprintf("Failed to locate migrations dir: %v", dirErr), 0)
		return dirErr
	}
	migrationPath := filepath.Join(migrationDir, filename)
	contentBytes, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		logMigrationResult(db, filename, "FAILED", fmt.Sprintf("Failed to read file: %v", err), 0)
		return err
	}
	content := string(contentBytes)

	sqlStatements := strings.Split(content, ";")

	for _, raw := range sqlStatements {
		stmt := strings.TrimSpace(raw)
		if stmt == "" || strings.HasPrefix(stmt, "--") || strings.HasPrefix(stmt, "/*") {
			continue
		}

		tx := db.Begin()
		if tx.Error != nil {
			executionTime := int(time.Since(startTime).Milliseconds())
			logMigrationResult(db, filename, "FAILED", fmt.Sprintf("Failed to begin transaction: %v", tx.Error), executionTime)
			return tx.Error
		}

		if err := tx.Exec(stmt).Error; err != nil {
			tx.Rollback()
			if isAlreadyExistsError(err) {
				atomic.AddInt64(&idempotentSkipCount, 1)
				continue
			}
			executionTime := int(time.Since(startTime).Milliseconds())
			logMigrationResult(db, filename, "FAILED", fmt.Sprintf("SQL error: %v", err), executionTime)
			return err
		}

		if err := tx.Commit().Error; err != nil {
			executionTime := int(time.Since(startTime).Milliseconds())
			logMigrationResult(db, filename, "FAILED", fmt.Sprintf("Failed to commit transaction: %v", err), executionTime)
			return err
		}
	}

	executionTime := int(time.Since(startTime).Milliseconds())
	logMigrationResult(db, filename, "SUCCESS", "Migration completed successfully", executionTime)
	log.Printf("✅ Migration completed: %s (%dms)", filename, executionTime)
	return nil
}

// isAlreadyExistsError detects non-fatal, idempotent errors
func isAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "already exists") ||
		strings.Contains(s, "duplicate key value") ||
		strings.Contains(s, "already defined")
}

func logMigrationResult(db *gorm.DB, migrationName, status, message string, executionTimeMs int) {
	sql := `
	INSERT INTO migration_logs (migration_name, status, message, execution_time_ms)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (migration_name) DO UPDATE SET
		status = EXCLUDED.status,
		message = EXCLUDED.message,
		execution_time_ms = EXCLUDED.execution_time_ms,
		executed_at = CURRENT_TIMESTAMP
	`

	if err := db.Exec(sql, migrationName, status, message, executionTimeMs).Error; err != nil {
		log.Printf("⚠️  Failed to log migration result: %v", err)
	}
}

// MigrationLog represents a migration log entry
type MigrationLog struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	MigrationName   string    `json:"migration_name" gorm:"size:255;uniqueIndex"`
	Status          string    `json:"status" gorm:"size:20"`
	Message         string    `json:"message" gorm:"type:text"`
	ExecutedAt      time.Time `json:"executed_at"`
	ExecutionTimeMs int       `json:"execution_time_ms"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// runPreMigrationFixes runs any necessary fixes before migrations
func runPreMigrationFixes(db *gorm.DB) error {
	// Add any pre-migration fixes here if needed
	return nil
}

// ensureSettingsHistoryTable ensures the settings_history table exists
func ensureSettingsHistoryTable(db *gorm.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS settings_histories (
		id SERIAL PRIMARY KEY,
		settings_id INTEGER,
		key VARCHAR(255),
		old_value TEXT,
		new_value TEXT,
		changed_by INTEGER,
		changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	return db.Exec(createTableSQL).Error
}

// ApprovalWorkflow represents the approval_workflows table
type ApprovalWorkflow struct {
	ID              uint    `gorm:"primaryKey"`
	Name            string  `gorm:"not null;size:100"`
	Module          string  `gorm:"not null;size:50"`
	MinAmount       float64 `gorm:"type:decimal(15,2);default:0"`
	MaxAmount       float64 `gorm:"type:decimal(15,2)"`
	IsActive        bool    `gorm:"default:true"`
	RequireDirector bool    `gorm:"default:false"`
	RequireFinance  bool    `gorm:"default:false"`
}

// ApprovalStep represents the approval_steps table
type ApprovalStep struct {
	ID           uint   `gorm:"primaryKey"`
	WorkflowID   uint   `gorm:"not null;index"`
	StepOrder    int    `gorm:"not null"`
	StepName     string `gorm:"not null;size:100"`
	ApproverRole string `gorm:"not null;size:50"`
	IsOptional   bool   `gorm:"default:false"`
	TimeLimit    int    `gorm:"default:24"`
}

// ensureStandardPurchaseApprovalWorkflow checks and creates Standard Purchase Approval workflow
func ensureStandardPurchaseApprovalWorkflow(db *gorm.DB) error {
	var existingWorkflow ApprovalWorkflow
	result := db.Where("name = ? AND module = ?", "Standard Purchase Approval", "PURCHASE").First(&existingWorkflow)

	if result.Error == nil {
		return nil // Workflow already exists
	}

	if result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}

	// Create new workflow
	workflow := ApprovalWorkflow{
		Name:            "Standard Purchase Approval",
		Module:          "PURCHASE",
		MinAmount:       0,
		MaxAmount:       0,
		IsActive:        true,
		RequireDirector: true,
		RequireFinance:  true,
	}

	if err := db.Create(&workflow).Error; err != nil {
		return fmt.Errorf("failed to create workflow: %v", err)
	}

	// Create approval steps
	steps := []ApprovalStep{
		{WorkflowID: workflow.ID, StepOrder: 1, StepName: "Purchasing Review", ApproverRole: "PURCHASING", IsOptional: false, TimeLimit: 24},
		{WorkflowID: workflow.ID, StepOrder: 2, StepName: "Cost Control Review", ApproverRole: "COST_CONTROL", IsOptional: false, TimeLimit: 24},
		{WorkflowID: workflow.ID, StepOrder: 3, StepName: "GM Approval", ApproverRole: "GM", IsOptional: false, TimeLimit: 48},
		{WorkflowID: workflow.ID, StepOrder: 4, StepName: "Project Director Approval", ApproverRole: "PROJECT_DIRECTOR", IsOptional: false, TimeLimit: 48},
		{WorkflowID: workflow.ID, StepOrder: 5, StepName: "Managing Director Approval", ApproverRole: "MANAGING_DIRECTOR", IsOptional: false, TimeLimit: 72},
	}

	for _, step := range steps {
		if err := db.Create(&step).Error; err != nil {
			return fmt.Errorf("failed to create step %s: %v", step.StepName, err)
		}
	}

	log.Println("✅ Created Standard Purchase Approval workflow with 5 steps")
	return nil
}

// ensureProjectBudgetsTable ensures the project_budgets table exists with correct structure
func ensureProjectBudgetsTable(db *gorm.DB) error {
	// Check if table already exists
	var tableExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'project_budgets'
		)
	`).Scan(&tableExists).Error

	if err != nil {
		return fmt.Errorf("failed to check if project_budgets table exists: %v", err)
	}

	if tableExists {
		// Table exists, check if account_id column exists
		var accountIdExists bool
		err := db.Raw(`
			SELECT EXISTS (
				SELECT FROM information_schema.columns 
				WHERE table_schema = 'public' 
				AND table_name = 'project_budgets'
				AND column_name = 'account_id'
			)
		`).Scan(&accountIdExists).Error

		if err != nil {
			return fmt.Errorf("failed to check if account_id column exists: %v", err)
		}

		if !accountIdExists {
			log.Println("🔄 Adding account_id column to project_budgets table...")
			
			// Add account_id column
			if err := db.Exec(`
				ALTER TABLE project_budgets 
				ADD COLUMN account_id INTEGER REFERENCES coa_accounts(id) ON DELETE RESTRICT
			`).Error; err != nil {
				return fmt.Errorf("failed to add account_id column: %v", err)
			}

			// Make category column nullable for backward compatibility
			if err := db.Exec(`
				ALTER TABLE project_budgets 
				ALTER COLUMN category DROP NOT NULL
			`).Error; err != nil {
				log.Printf("⚠️  Warning: Failed to make category nullable: %v", err)
			}

			// Create index
			if err := db.Exec(`
				CREATE INDEX IF NOT EXISTS idx_project_budgets_account_id ON project_budgets(account_id)
			`).Error; err != nil {
				log.Printf("⚠️  Warning: Failed to create index on account_id: %v", err)
			}

			log.Println("✅ account_id column added to project_budgets table")
		} else {
			log.Println("✅ project_budgets table already has correct structure")
		}
		return nil
	}

	log.Println("🔄 Creating project_budgets table...")

	// Create the table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS project_budgets (
		id SERIAL PRIMARY KEY,
		project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
		account_id INTEGER NOT NULL REFERENCES coa_accounts(id) ON DELETE RESTRICT,
		estimated_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL,
		
		CONSTRAINT project_budgets_unique_project_account UNIQUE (project_id, account_id),
		CONSTRAINT project_budgets_positive_amount CHECK (estimated_amount >= 0)
	)`

	if err := db.Exec(createTableSQL).Error; err != nil {
		return fmt.Errorf("failed to create project_budgets table: %v", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_project_budgets_project_id ON project_budgets(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_budgets_account_id ON project_budgets(account_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_budgets_deleted_at ON project_budgets(deleted_at)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to create index: %v", err)
		}
	}

	// Add comments
	comments := []string{
		"COMMENT ON TABLE project_budgets IS 'Budget allocation per project per COA account'",
		"COMMENT ON COLUMN project_budgets.project_id IS 'Reference to project'",
		"COMMENT ON COLUMN project_budgets.account_id IS 'Reference to COA account'",
		"COMMENT ON COLUMN project_budgets.estimated_amount IS 'Estimated budget amount for this project-account combination'",
		"COMMENT ON COLUMN project_budgets.deleted_at IS 'Soft delete timestamp'",
	}

	for _, commentSQL := range comments {
		if err := db.Exec(commentSQL).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to add comment: %v", err)
		}
	}

	// Create trigger function
	triggerFunctionSQL := `
	CREATE OR REPLACE FUNCTION update_project_budgets_updated_at()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql`

	if err := db.Exec(triggerFunctionSQL).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to create trigger function: %v", err)
	}

	// Create trigger
	triggerSQL := `
	CREATE TRIGGER trigger_update_project_budgets_updated_at
		BEFORE UPDATE ON project_budgets
		FOR EACH ROW
		EXECUTE FUNCTION update_project_budgets_updated_at()`

	if err := db.Exec(triggerSQL).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to create trigger: %v", err)
	}

	log.Println("✅ project_budgets table created successfully")
	return nil
}
