package models

import (
	"time"
	"gorm.io/gorm"
)

// Transaction represents a general transaction record
// This serves as a base for various transaction types in the system
type Transaction struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Type            string         `json:"type" gorm:"not null;size:50"` // CASH_BANK, JOURNAL, PAYMENT, etc.
	ReferenceType   string         `json:"reference_type" gorm:"size:50"`
	ReferenceID     uint           `json:"reference_id" gorm:"index"`
	AccountID       uint           `json:"account_id" gorm:"not null;index"`
	Amount          float64        `json:"amount" gorm:"type:decimal(20,2)"`
	Description     string         `json:"description" gorm:"type:text"`
	TransactionDate time.Time      `json:"transaction_date"`
	UserID          uint           `json:"user_id" gorm:"not null;index"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Account Account `json:"account" gorm:"foreignKey:AccountID"`
	User    User    `json:"user" gorm:"foreignKey:UserID"`
}

// Transaction Types Constants
const (
	TransactionTypeCashBank = "CASH_BANK"
	TransactionTypeJournal  = "JOURNAL"
	TransactionTypePayment  = "PAYMENT"
	TransactionTypeSales    = "SALES"
	TransactionTypePurchase = "PURCHASE"
	TransactionTypeAsset    = "ASSET"
	TransactionTypeExpense  = "EXPENSE"
)

// Transaction Reference Types Constants
const (
	TransactionRefCashBankTransaction = "CASH_BANK_TRANSACTION"
	TransactionRefJournalEntry        = "JOURNAL_ENTRY"
	TransactionRefPaymentRecord       = "PAYMENT_RECORD"
	TransactionRefSalesInvoice        = "SALES_INVOICE"
	TransactionRefPurchaseInvoice     = "PURCHASE_INVOICE"
	TransactionRefAssetTransaction    = "ASSET_TRANSACTION"
	TransactionRefExpenseRecord       = "EXPENSE_RECORD"
)
