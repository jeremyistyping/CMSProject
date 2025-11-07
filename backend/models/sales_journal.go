package models

import (
	"time"
	"gorm.io/gorm"
)

// SimpleSSOTJournal represents a simplified SSOT journal for sales
type SimpleSSOTJournal struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	EntryNumber       string         `json:"entry_number" gorm:"size:50"`
	TransactionType   string         `json:"transaction_type" gorm:"size:20;index"`
	TransactionID     uint           `json:"transaction_id" gorm:"index"`
	TransactionNumber string         `json:"transaction_number" gorm:"size:50"`
	Date              time.Time      `json:"date"`
	Description       string         `json:"description" gorm:"type:text"`
	TotalAmount       float64        `json:"total_amount" gorm:"type:decimal(15,2)"`
	TotalDebit        float64        `json:"total_debit" gorm:"type:decimal(15,2)"`
	TotalCredit       float64        `json:"total_credit" gorm:"type:decimal(15,2)"`
	Status            string         `json:"status" gorm:"size:20;default:'DRAFT'"`
	CreatedBy         uint           `json:"created_by"`
	PostedBy          *uint          `json:"posted_by"`
	PostedAt          *time.Time     `json:"posted_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	Items []SimpleSSOTJournalItem `json:"items" gorm:"foreignKey:JournalID"`
}

// SimpleSSOTJournalItem represents a journal line item
type SimpleSSOTJournalItem struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	JournalID   uint           `json:"journal_id" gorm:"index"`
	AccountID   uint           `json:"account_id" gorm:"index"`
	AccountCode string         `json:"account_code" gorm:"size:20"`
	AccountName string         `json:"account_name" gorm:"size:100"`
	Debit       float64        `json:"debit" gorm:"type:decimal(15,2)"`
	Credit      float64        `json:"credit" gorm:"type:decimal(15,2)"`
	Description string         `json:"description" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	Journal SimpleSSOTJournal `json:"-" gorm:"foreignKey:JournalID"`
	Account Account           `json:"account" gorm:"foreignKey:AccountID"`
}

func (SimpleSSOTJournal) TableName() string {
	return "simple_ssot_journals"
}

func (SimpleSSOTJournalItem) TableName() string {
	return "simple_ssot_journal_items"
}