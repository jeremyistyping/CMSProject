package models

import (
	"time"
	"gorm.io/gorm"
)

// These models are used specifically for auto-migration of additional structures
// that are created in the enhanced cash bank migration

// CashBankTransfer represents transfer between cash/bank accounts (for GORM migration)
type CashBankTransferMigration struct {
	ID              uint      `gorm:"primaryKey"`
	TransferNumber  string    `gorm:"unique;not null;size:50"`
	FromAccountID   uint      `gorm:"not null;index"`
	ToAccountID     uint      `gorm:"not null;index"`
	Date            time.Time
	Amount          float64   `gorm:"type:decimal(15,2)"`
	ExchangeRate    float64   `gorm:"type:decimal(12,6);default:1"`
	ConvertedAmount float64   `gorm:"type:decimal(15,2)"`
	Reference       string    `gorm:"size:100"`
	Notes           string    `gorm:"type:text"`
	Status          string    `gorm:"size:20;default:PENDING"`
	UserID          uint      `gorm:"not null;index"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (CashBankTransferMigration) TableName() string {
	return "cash_bank_transfers"
}

// BankReconciliation represents bank reconciliation records (for GORM migration)
type BankReconciliationMigration struct {
	ID               uint      `gorm:"primaryKey"`
	CashBankID       uint      `gorm:"not null;index"`
	ReconcileDate    time.Time
	StatementBalance float64   `gorm:"type:decimal(15,2)"`
	SystemBalance    float64   `gorm:"type:decimal(15,2)"`
	Difference       float64   `gorm:"type:decimal(15,2)"`
	Status           string    `gorm:"size:20;default:PENDING"`
	UserID           uint      `gorm:"not null;index"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

func (BankReconciliationMigration) TableName() string {
	return "bank_reconciliations"
}

// ReconciliationItem represents items in bank reconciliation (for GORM migration)
type ReconciliationItemMigration struct {
	ID               uint   `gorm:"primaryKey"`
	ReconciliationID uint   `gorm:"not null;index"`
	TransactionID    uint   `gorm:"not null;index"`
	IsCleared        bool   `gorm:"default:false"`
	Notes            string `gorm:"type:text"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

func (ReconciliationItemMigration) TableName() string {
	return "reconciliation_items"
}

// Function to migrate additional cash bank tables using GORM
// MigrationRecord tracks applied migrations to prevent re-running
type MigrationRecord struct {
	ID          uint      `gorm:"primaryKey"`
	MigrationID string    `gorm:"unique;not null;size:100"` // Unique identifier for migration
	Description string    `gorm:"type:text"`                  // Description of what was migrated
	Version     string    `gorm:"size:20"`                    // Version when migration was applied
	AppliedAt   time.Time `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func MigrateCashBankTables(db *gorm.DB) error {
	// Migrate additional tables using GORM AutoMigrate
	err := db.AutoMigrate(
		&CashBankTransferMigration{},
		&BankReconciliationMigration{},
		&ReconciliationItemMigration{},
	)
	
	if err != nil {
		return err
	}
	
	// Add foreign key constraints if they don't exist
	// Note: GORM will handle most constraints, but we can add custom ones here
	
	return nil
}
