package services

import (
	"errors"
	"fmt"
	"time"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

type InvoiceService struct {
	db              *gorm.DB
	settingsService *SettingsService
}

func NewInvoiceService(db *gorm.DB) *InvoiceService {
	settingsService := NewSettingsService(db)
	return &InvoiceService{
		db:              db,
		settingsService: settingsService,
	}
}

// GenerateInvoiceNumber generates next invoice number using monthly sequence (compatibility path)
func (s *InvoiceService) GenerateInvoiceNumber() (string, error) {
	_, err := s.settingsService.GetSettings()
	if err != nil { return "", fmt.Errorf("failed to get settings: %v", err) }

	var invoiceNumber string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var settingsForUpdate models.Settings
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&settingsForUpdate).Error; err != nil { return err }

		// Use the same sequence table as full service
		if err := tx.AutoMigrate(&InvoiceCodeSequence{}); err != nil { return err }
		year := time.Now().Year(); month := int(time.Now().Month())
		var seq InvoiceCodeSequence
		res := tx.Set("gorm:query_option", "FOR UPDATE").Where("year = ? AND month = ?", year, month).First(&seq)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				seq = InvoiceCodeSequence{Year: year, Month: month, LastNumber: 0}
				if err := tx.Create(&seq).Error; err != nil { return err }
			} else { return res.Error }
		}
		seq.LastNumber++
		invoiceNumber = fmt.Sprintf("%s/%04d/%02d/%04d", settingsForUpdate.InvoicePrefix, year, month, seq.LastNumber)
		if err := tx.Save(&seq).Error; err != nil { return err }
		return nil
	})
	if err != nil { return "", fmt.Errorf("failed to generate invoice number: %v", err) }
	return invoiceNumber, nil
}

// GenerateQuoteNumber generates next quote number using monthly sequence (compatibility path)
func (s *InvoiceService) GenerateQuoteNumber() (string, error) {
	_, err := s.settingsService.GetSettings()
	if err != nil { return "", fmt.Errorf("failed to get settings: %v", err) }

	var quoteNumber string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var settingsForUpdate models.Settings
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&settingsForUpdate).Error; err != nil { return err }
		if err := tx.AutoMigrate(&QuoteCodeSequence{}); err != nil { return err }
		year := time.Now().Year(); month := int(time.Now().Month())
		var seq QuoteCodeSequence
		res := tx.Set("gorm:query_option", "FOR UPDATE").Where("year = ? AND month = ?", year, month).First(&seq)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				seq = QuoteCodeSequence{Year: year, Month: month, LastNumber: 0}
				if err := tx.Create(&seq).Error; err != nil { return err }
			} else { return res.Error }
		}
		seq.LastNumber++
		quoteNumber = fmt.Sprintf("%s/%04d/%02d/%04d", settingsForUpdate.QuotePrefix, year, month, seq.LastNumber)
		if err := tx.Save(&seq).Error; err != nil { return err }
		return nil
	})
	if err != nil { return "", fmt.Errorf("failed to generate quote number: %v", err) }
	return quoteNumber, nil
}

// FormatCurrency formats amount according to system settings
func (s *InvoiceService) FormatCurrency(amount float64) (string, error) {
	settings, err := s.settingsService.GetSettings()
	if err != nil {
		return "", err
	}

	// Format with decimal places from settings
	formatStr := fmt.Sprintf("%%.%df", settings.DecimalPlaces)
	formatted := fmt.Sprintf(formatStr, amount)
	
	// Apply thousand separator (simplified implementation)
	// In production, you might want to use a more sophisticated number formatting library
	
	return fmt.Sprintf("%s %s", settings.Currency, formatted), nil
}

// FormatDate formats date according to system settings
func (s *InvoiceService) FormatDate(date time.Time) (string, error) {
	settings, err := s.settingsService.GetSettings()
	if err != nil {
		return "", err
	}

	switch settings.DateFormat {
	case "DD/MM/YYYY":
		return date.Format("02/01/2006"), nil
	case "MM/DD/YYYY":
		return date.Format("01/02/2006"), nil
	case "DD-MM-YYYY":
		return date.Format("02-01-2006"), nil
	case "YYYY-MM-DD":
		return date.Format("2006-01-02"), nil
	default:
		return date.Format("02/01/2006"), nil // Default to DD/MM/YYYY
	}
}