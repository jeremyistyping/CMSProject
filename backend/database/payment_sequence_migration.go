package database

import (
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// CreatePaymentSequenceTables creates the payment code sequence table
func CreatePaymentSequenceTables(db *gorm.DB) error {
	// Auto-migrate the PaymentCodeSequence model
	if err := db.AutoMigrate(&models.PaymentCodeSequence{}); err != nil {
		return err
	}

	// Add unique index for prefix, year, month combination
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_sequence_unique 
		ON payment_code_sequences (prefix, year, month)
	`).Error; err != nil {
		return err
	}

	return nil
}

// InitializePaymentSequenceFromExistingData initializes sequence numbers based on existing payment data
func InitializePaymentSequenceFromExistingData(db *gorm.DB) error {
	// Get distinct prefix/year/month combinations from existing payments
	var results []struct {
		Prefix string
		Year   int
		Month  int
		Count  int
	}

	query := `
		SELECT 
			CASE 
				WHEN code LIKE 'RCV/%' THEN 'RCV'
				WHEN code LIKE 'PAY/%' THEN 'PAY'
				ELSE 'UNKNOWN'
			END as prefix,
			EXTRACT(YEAR FROM created_at) as year,
			EXTRACT(MONTH FROM created_at) as month,
			COUNT(*) as count
		FROM payments 
		WHERE deleted_at IS NULL
		GROUP BY 
			CASE 
				WHEN code LIKE 'RCV/%' THEN 'RCV'
				WHEN code LIKE 'PAY/%' THEN 'PAY'
				ELSE 'UNKNOWN'
			END,
			EXTRACT(YEAR FROM created_at),
			EXTRACT(MONTH FROM created_at)
		HAVING 
			CASE 
				WHEN code LIKE 'RCV/%' THEN 'RCV'
				WHEN code LIKE 'PAY/%' THEN 'PAY'
				ELSE 'UNKNOWN'
			END != 'UNKNOWN'
	`

	if err := db.Raw(query).Scan(&results).Error; err != nil {
		return err
	}

	// Insert sequence records for each combination
	for _, result := range results {
		sequence := models.PaymentCodeSequence{
			Prefix:         result.Prefix,
			Year:           result.Year,
			Month:          result.Month,
			SequenceNumber: result.Count, // Start from current count
		}

		// Use ON DUPLICATE KEY UPDATE equivalent (UPSERT)
		if err := db.Where("prefix = ? AND year = ? AND month = ?", 
			result.Prefix, result.Year, result.Month).
			Assign(&sequence).FirstOrCreate(&sequence).Error; err != nil {
			return err
		}
	}

	return nil
}
