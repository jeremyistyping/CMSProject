package repositories

import (
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// SSOTJournalRepository handles SSOT journal database operations
type SSOTJournalRepository struct {
	db *gorm.DB
}

// NewSSOTJournalRepository creates a new instance
func NewSSOTJournalRepository(db *gorm.DB) *SSOTJournalRepository {
	return &SSOTJournalRepository{db: db}
}

// Create creates a new SSOT journal entry
func (r *SSOTJournalRepository) Create(db *gorm.DB, journal *models.SSOTJournal) error {
	return db.Create(journal).Error
}

// GetByID retrieves a journal by ID
func (r *SSOTJournalRepository) GetByID(id uint) (*models.SSOTJournal, error) {
	var journal models.SSOTJournal
	err := r.db.Preload("Items").First(&journal, id).Error
	return &journal, err
}

// GetByTransactionTypeAndID retrieves journal by transaction type and ID
func (r *SSOTJournalRepository) GetByTransactionTypeAndID(transactionType string, transactionID uint) (*models.SSOTJournal, error) {
	var journal models.SSOTJournal
	err := r.db.Where("transaction_type = ? AND transaction_id = ?", transactionType, transactionID).
		Preload("Items").First(&journal).Error
	return &journal, err
}

// Update updates an existing journal
func (r *SSOTJournalRepository) Update(journal *models.SSOTJournal) error {
	return r.db.Save(journal).Error
}

// Delete deletes a journal entry
func (r *SSOTJournalRepository) Delete(id uint) error {
	return r.db.Delete(&models.SSOTJournal{}, id).Error
}