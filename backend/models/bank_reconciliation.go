package models

import (
	"time"
	"gorm.io/gorm"
)

// BankReconciliationSnapshot - Snapshot transaksi bank untuk periode tertentu
type BankReconciliationSnapshot struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	CashBankID        uint           `json:"cash_bank_id" gorm:"not null;index"`
	Period            string         `json:"period" gorm:"not null;size:7;index"` // Format: YYYY-MM
	SnapshotDate      time.Time      `json:"snapshot_date" gorm:"not null;index"`
	GeneratedBy       uint           `json:"generated_by" gorm:"not null"`
	
	// Balance Information
	OpeningBalance    float64        `json:"opening_balance" gorm:"type:decimal(20,2);default:0"`
	ClosingBalance    float64        `json:"closing_balance" gorm:"type:decimal(20,2);default:0"`
	TotalDebit        float64        `json:"total_debit" gorm:"type:decimal(20,2);default:0"`
	TotalCredit       float64        `json:"total_credit" gorm:"type:decimal(20,2);default:0"`
	TransactionCount  int            `json:"transaction_count" gorm:"default:0"`
	
	// Integrity & Security
	DataHash          string         `json:"data_hash" gorm:"size:64;not null"` // SHA-256 hash
	IsLocked          bool           `json:"is_locked" gorm:"default:false"`
	LockedAt          *time.Time     `json:"locked_at"`
	LockedBy          *uint          `json:"locked_by"`
	
	// Metadata
	Notes             string         `json:"notes" gorm:"type:text"`
	Status            string         `json:"status" gorm:"size:20;default:'ACTIVE'"` // ACTIVE, ARCHIVED, SUPERSEDED
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	CashBank          CashBank                           `json:"cash_bank" gorm:"foreignKey:CashBankID"`
	GeneratedByUser   User                               `json:"generated_by_user" gorm:"foreignKey:GeneratedBy"`
	Transactions      []ReconciliationTransactionSnapshot `json:"transactions" gorm:"foreignKey:SnapshotID"`
	LockedByUser      *User                              `json:"locked_by_user,omitempty" gorm:"foreignKey:LockedBy"`
}

// ReconciliationTransactionSnapshot - Detail transaksi dalam snapshot
type ReconciliationTransactionSnapshot struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	SnapshotID        uint           `json:"snapshot_id" gorm:"not null;index"`
	TransactionID     uint           `json:"transaction_id" gorm:"not null;index"`
	
	// Transaction Data (frozen at snapshot time)
	TransactionDate   time.Time      `json:"transaction_date" gorm:"not null"`
	ReferenceType     string         `json:"reference_type" gorm:"size:50"`
	ReferenceID       uint           `json:"reference_id"`
	ReferenceNumber   string         `json:"reference_number" gorm:"size:100"`
	Amount            float64        `json:"amount" gorm:"type:decimal(20,2)"`
	DebitAmount       float64        `json:"debit_amount" gorm:"type:decimal(20,2);default:0"`
	CreditAmount      float64        `json:"credit_amount" gorm:"type:decimal(20,2);default:0"`
	BalanceAfter      float64        `json:"balance_after" gorm:"type:decimal(20,2)"`
	Description       string         `json:"description" gorm:"type:text"`
	Notes             string         `json:"notes" gorm:"type:text"`
	
	// Audit Info
	CreatedBy         uint           `json:"created_by"`
	CreatedAt         time.Time      `json:"created_at"`
	
	// Relations
	Snapshot          BankReconciliationSnapshot `json:"-" gorm:"foreignKey:SnapshotID"`
}

// BankReconciliation - Record proses rekonsiliasi
type BankReconciliation struct {
	ID                     uint           `json:"id" gorm:"primaryKey"`
	ReconciliationNumber   string         `json:"reconciliation_number" gorm:"unique;not null;size:50;index"`
	CashBankID             uint           `json:"cash_bank_id" gorm:"not null;index"`
	Period                 string         `json:"period" gorm:"not null;size:7;index"` // Format: YYYY-MM
	
	// Snapshot References
	BaseSnapshotID         uint           `json:"base_snapshot_id" gorm:"not null;index"`     // PDF lama
	ComparisonSnapshotID   *uint          `json:"comparison_snapshot_id" gorm:"index"`        // PDF baru (optional)
	
	// Reconciliation Info
	ReconciliationDate     time.Time      `json:"reconciliation_date" gorm:"not null"`
	ReconciliationBy       uint           `json:"reconciliation_by" gorm:"not null"`
	
	// Balance Comparison
	BaseBalance            float64        `json:"base_balance" gorm:"type:decimal(20,2)"`
	CurrentBalance         float64        `json:"current_balance" gorm:"type:decimal(20,2)"`
	Variance               float64        `json:"variance" gorm:"type:decimal(20,2)"`
	
	// Transaction Comparison
	BaseTransactionCount   int            `json:"base_transaction_count"`
	CurrentTransactionCount int           `json:"current_transaction_count"`
	TransactionVariance    int            `json:"transaction_variance"`
	
	// Differences Found
	MissingTransactions    int            `json:"missing_transactions" gorm:"default:0"`
	AddedTransactions      int            `json:"added_transactions" gorm:"default:0"`
	ModifiedTransactions   int            `json:"modified_transactions" gorm:"default:0"`
	
	// Status & Approval
	Status                 string         `json:"status" gorm:"size:20;default:'PENDING'"` // PENDING, APPROVED, REJECTED, NEEDS_REVIEW
	ReviewedBy             *uint          `json:"reviewed_by"`
	ReviewedAt             *time.Time     `json:"reviewed_at"`
	ReviewNotes            string         `json:"review_notes" gorm:"type:text"`
	
	// Result
	IsBalanced             bool           `json:"is_balanced" gorm:"default:false"`
	BalanceConfirmed       bool           `json:"balance_confirmed" gorm:"default:false"`
	
	Notes                  string         `json:"notes" gorm:"type:text"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	CashBank               CashBank                      `json:"cash_bank" gorm:"foreignKey:CashBankID"`
	BaseSnapshot           BankReconciliationSnapshot    `json:"base_snapshot" gorm:"foreignKey:BaseSnapshotID"`
	ComparisonSnapshot     *BankReconciliationSnapshot   `json:"comparison_snapshot,omitempty" gorm:"foreignKey:ComparisonSnapshotID"`
	ReconciliationByUser   User                          `json:"reconciliation_by_user" gorm:"foreignKey:ReconciliationBy"`
	ReviewedByUser         *User                         `json:"reviewed_by_user,omitempty" gorm:"foreignKey:ReviewedBy"`
	Differences            []ReconciliationDifference    `json:"differences" gorm:"foreignKey:ReconciliationID"`
}

// ReconciliationDifference - Detail perbedaan yang ditemukan
type ReconciliationDifference struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	ReconciliationID  uint           `json:"reconciliation_id" gorm:"not null;index"`
	
	DifferenceType    string         `json:"difference_type" gorm:"size:50;not null"` // MISSING, ADDED, MODIFIED, AMOUNT_CHANGE, DATE_CHANGE
	Severity          string         `json:"severity" gorm:"size:20;default:'MEDIUM'"` // LOW, MEDIUM, HIGH, CRITICAL
	
	// Transaction References
	BaseTransactionID      *uint          `json:"base_transaction_id" gorm:"index"`
	CurrentTransactionID   *uint          `json:"current_transaction_id" gorm:"index"`
	
	// Difference Details
	Field                  string         `json:"field" gorm:"size:50"`           // amount, date, description, etc
	OldValue               string         `json:"old_value" gorm:"type:text"`
	NewValue               string         `json:"new_value" gorm:"type:text"`
	AmountDifference       float64        `json:"amount_difference" gorm:"type:decimal(20,2)"`
	
	// Resolution
	Status                 string         `json:"status" gorm:"size:20;default:'PENDING'"` // PENDING, RESOLVED, IGNORED, ESCALATED
	ResolutionNotes        string         `json:"resolution_notes" gorm:"type:text"`
	ResolvedBy             *uint          `json:"resolved_by"`
	ResolvedAt             *time.Time     `json:"resolved_at"`
	
	Description            string         `json:"description" gorm:"type:text"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	
	// Relations
	Reconciliation         BankReconciliation `json:"-" gorm:"foreignKey:ReconciliationID"`
	ResolvedByUser         *User              `json:"resolved_by_user,omitempty" gorm:"foreignKey:ResolvedBy"`
}

// CashBankAuditTrail - Audit trail untuk semua perubahan cash bank
type CashBankAuditTrail struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	CashBankID        uint           `json:"cash_bank_id" gorm:"not null;index"`
	TransactionID     *uint          `json:"transaction_id" gorm:"index"`
	
	Action            string         `json:"action" gorm:"size:50;not null"` // CREATE, UPDATE, DELETE, VOID, RESTORE
	EntityType        string         `json:"entity_type" gorm:"size:50;not null"` // CASH_BANK, TRANSACTION, TRANSFER, DEPOSIT, WITHDRAWAL
	EntityID          uint           `json:"entity_id" gorm:"not null"`
	
	// Change Details
	FieldChanged      string         `json:"field_changed" gorm:"size:100"`
	OldValue          string         `json:"old_value" gorm:"type:text"`
	NewValue          string         `json:"new_value" gorm:"type:text"`
	
	// Context
	Reason            string         `json:"reason" gorm:"type:text"`
	IPAddress         string         `json:"ip_address" gorm:"size:45"`
	UserAgent         string         `json:"user_agent" gorm:"type:text"`
	
	// Approval (for backdated or sensitive changes)
	RequiresApproval  bool           `json:"requires_approval" gorm:"default:false"`
	ApprovedBy        *uint          `json:"approved_by"`
	ApprovedAt        *time.Time     `json:"approved_at"`
	ApprovalStatus    string         `json:"approval_status" gorm:"size:20"` // PENDING, APPROVED, REJECTED
	
	UserID            uint           `json:"user_id" gorm:"not null"`
	CreatedAt         time.Time      `json:"created_at"`
	
	// Relations
	CashBank          CashBank       `json:"cash_bank" gorm:"foreignKey:CashBankID"`
	User              User           `json:"user" gorm:"foreignKey:UserID"`
	ApprovedByUser    *User          `json:"approved_by_user,omitempty" gorm:"foreignKey:ApprovedBy"`
}

// TableName overrides
func (BankReconciliationSnapshot) TableName() string {
	return "bank_reconciliation_snapshots"
}

func (ReconciliationTransactionSnapshot) TableName() string {
	return "reconciliation_transaction_snapshots"
}

func (BankReconciliation) TableName() string {
	return "bank_reconciliations"
}

func (ReconciliationDifference) TableName() string {
	return "reconciliation_differences"
}

func (CashBankAuditTrail) TableName() string {
	return "cash_bank_audit_trail"
}
