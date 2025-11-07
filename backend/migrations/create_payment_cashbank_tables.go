package migrations

import (
	"gorm.io/gorm"
)

// CreatePaymentCashBankTables creates tables for payment and cash/bank modules
func CreatePaymentCashBankTables(db *gorm.DB) error {
	// Create Payment table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(20) UNIQUE NOT NULL,
			contact_id BIGINT UNSIGNED NOT NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			date DATE NOT NULL,
			amount DECIMAL(15,2) DEFAULT 0,
			method VARCHAR(20),
			reference VARCHAR(50),
			status VARCHAR(20),
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			INDEX idx_contact_id (contact_id),
			INDEX idx_user_id (user_id),
			INDEX idx_date (date),
			INDEX idx_status (status),
			INDEX idx_deleted_at (deleted_at),
			FOREIGN KEY (contact_id) REFERENCES contacts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`).Error; err != nil {
		return err
	}

	// Create Payment Allocations table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS payment_allocations (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			payment_id BIGINT UNSIGNED NOT NULL,
			invoice_id BIGINT UNSIGNED NULL,
			bill_id BIGINT UNSIGNED NULL,
			allocated_amount DECIMAL(15,2) DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_payment_id (payment_id),
			INDEX idx_invoice_id (invoice_id),
			INDEX idx_bill_id (bill_id),
			FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE CASCADE,
			FOREIGN KEY (invoice_id) REFERENCES sales(id),
			FOREIGN KEY (bill_id) REFERENCES purchases(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`).Error; err != nil {
		return err
	}

	// Create CashBank table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cash_banks (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(20) UNIQUE NOT NULL,
			name VARCHAR(100) NOT NULL,
			type VARCHAR(20) NOT NULL,
			account_id BIGINT UNSIGNED,
			bank_name VARCHAR(100),
			account_no VARCHAR(50),
			currency VARCHAR(5) DEFAULT 'IDR',
			balance DECIMAL(15,2) DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			INDEX idx_type (type),
			INDEX idx_account_id (account_id),
			INDEX idx_is_active (is_active),
			INDEX idx_deleted_at (deleted_at),
			FOREIGN KEY (account_id) REFERENCES accounts(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`).Error; err != nil {
		return err
	}

	// Create CashBank Transactions table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cash_bank_transactions (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			cash_bank_id BIGINT UNSIGNED NOT NULL,
			reference_type VARCHAR(50),
			reference_id BIGINT UNSIGNED,
			amount DECIMAL(20,2) DEFAULT 0,
			balance_after DECIMAL(20,2) DEFAULT 0,
			transaction_date DATETIME NOT NULL,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			INDEX idx_cash_bank_id (cash_bank_id),
			INDEX idx_reference (reference_type, reference_id),
			INDEX idx_transaction_date (transaction_date),
			INDEX idx_deleted_at (deleted_at),
			FOREIGN KEY (cash_bank_id) REFERENCES cash_banks(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`).Error; err != nil {
		return err
	}

	// Create CashBank Transfers table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cash_bank_transfers (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			transfer_number VARCHAR(50) UNIQUE NOT NULL,
			from_account_id BIGINT UNSIGNED NOT NULL,
			to_account_id BIGINT UNSIGNED NOT NULL,
			date DATE NOT NULL,
			amount DECIMAL(15,2) DEFAULT 0,
			exchange_rate DECIMAL(12,6) DEFAULT 1,
			converted_amount DECIMAL(15,2) DEFAULT 0,
			reference VARCHAR(100),
			notes TEXT,
			status VARCHAR(20),
			user_id BIGINT UNSIGNED NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_from_account (from_account_id),
			INDEX idx_to_account (to_account_id),
			INDEX idx_date (date),
			INDEX idx_user_id (user_id),
			FOREIGN KEY (from_account_id) REFERENCES cash_banks(id),
			FOREIGN KEY (to_account_id) REFERENCES cash_banks(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`).Error; err != nil {
		return err
	}

	// Create Bank Reconciliations table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS bank_reconciliations (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			cash_bank_id BIGINT UNSIGNED NOT NULL,
			reconcile_date DATE NOT NULL,
			statement_balance DECIMAL(15,2) DEFAULT 0,
			system_balance DECIMAL(15,2) DEFAULT 0,
			difference DECIMAL(15,2) DEFAULT 0,
			status VARCHAR(20),
			user_id BIGINT UNSIGNED NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_cash_bank_id (cash_bank_id),
			INDEX idx_reconcile_date (reconcile_date),
			INDEX idx_user_id (user_id),
			FOREIGN KEY (cash_bank_id) REFERENCES cash_banks(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`).Error; err != nil {
		return err
	}

	// Create Reconciliation Items table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS reconciliation_items (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			reconciliation_id BIGINT UNSIGNED NOT NULL,
			transaction_id BIGINT UNSIGNED NOT NULL,
			is_cleared BOOLEAN DEFAULT FALSE,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_reconciliation_id (reconciliation_id),
			INDEX idx_transaction_id (transaction_id),
			FOREIGN KEY (reconciliation_id) REFERENCES bank_reconciliations(id) ON DELETE CASCADE,
			FOREIGN KEY (transaction_id) REFERENCES cash_bank_transactions(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`).Error; err != nil {
		return err
	}

	// Add indexes for better performance
	db.Exec("CREATE INDEX idx_payment_date_status ON payments(date, status);")
	db.Exec("CREATE INDEX idx_cashbank_type_active ON cash_banks(type, is_active);")
	db.Exec("CREATE INDEX idx_cashbank_tx_date_type ON cash_bank_transactions(transaction_date, reference_type);")

	return nil
}

// RollbackPaymentCashBankTables drops the payment and cash/bank tables
func RollbackPaymentCashBankTables(db *gorm.DB) error {
	tables := []string{
		"reconciliation_items",
		"bank_reconciliations",
		"cash_bank_transfers",
		"cash_bank_transactions",
		"cash_banks",
		"payment_allocations",
		"payments",
	}
	
	for _, table := range tables {
		if err := db.Exec("DROP TABLE IF EXISTS " + table).Error; err != nil {
			return err
		}
	}
	
	return nil
}

// SeedPaymentCashBankData seeds initial data for payment and cash/bank modules
func SeedPaymentCashBankData(db *gorm.DB) error {
	// Create default cash account
	if err := db.Exec(`
		INSERT INTO cash_banks (code, name, type, currency, balance, is_active, description)
		VALUES 
		('CSH-001', 'Petty Cash - Main Office', 'CASH', 'IDR', 0, TRUE, 'Main office petty cash'),
		('BNK-001', 'Bank BCA - Operations', 'BANK', 'IDR', 0, TRUE, 'Main operational bank account')
		ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;
	`).Error; err != nil {
		return err
	}
	
	return nil
}
