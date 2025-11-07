package services

import (
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// JournalCoordinator provides a thin coordination layer around journal creation.
// This minimal implementation is kept to satisfy upstream dependencies while
// the system fully migrates to SSOT Journal System.
// It prevents duplicate creation by delegating idempotency to the caller for now.
type JournalCoordinator struct {
	db *gorm.DB
}

// JournalCreationRequest carries minimal info needed for coordination hooks.
type JournalCreationRequest struct {
	TransactionType string
	ReferenceID     uint
	UserID          uint
	Description     string
}

// JournalCreationResult represents the outcome of a coordinated journal creation.
type JournalCreationResult struct {
	Success        bool
	WasDuplicate   bool
	ExistingEntry  *models.JournalEntry
	JournalEntryID uint
}

func NewJournalCoordinator(db *gorm.DB) *JournalCoordinator {
	return &JournalCoordinator{db: db}
}

// CreateJournalEntryWithCoordination wraps the provided creator function.
// NOTE: This simplified implementation does not implement distributed locks
// or duplicate detection. Callers should perform their own idempotency checks
// until the SSOT migration removes this layer entirely.
func (jc *JournalCoordinator) CreateJournalEntryWithCoordination(
	req *JournalCreationRequest,
	creator func() (*models.JournalEntry, error),
) (*JournalCreationResult, error) {
	entry, err := creator()
	if err != nil {
		return &JournalCreationResult{Success: false}, err
	}
	res := &JournalCreationResult{
		Success:        true,
		WasDuplicate:   false,
		ExistingEntry:  nil,
		JournalEntryID: entry.ID,
	}
	return res, nil
}
