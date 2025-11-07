package services

import (
	"fmt"
	"time"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

type InvoiceNumberService struct {
	db *gorm.DB
}

func NewInvoiceNumberService(db *gorm.DB) *InvoiceNumberService {
	return &InvoiceNumberService{db: db}
}

// GenerateInvoiceNumber generates a new invoice number with format: {4 digit no}/{kode tipe}/{bulan romawi}-{4 digit tahun}
// Example: 0120/STA-C/IX-2025
func (s *InvoiceNumberService) GenerateInvoiceNumber(invoiceTypeID uint, date time.Time) (*models.InvoiceNumberResponse, error) {
	// Get invoice type
	var invoiceType models.InvoiceType
	if err := s.db.Where("id = ? AND is_active = ?", invoiceTypeID, true).First(&invoiceType).Error; err != nil {
		return nil, fmt.Errorf("invoice type not found or inactive: %v", err)
	}

	year := date.Year()
	month := int(date.Month())
	
	// Get or create counter for this type and year
	counter, err := s.getOrCreateCounter(invoiceTypeID, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get counter: %v", err)
	}

	// Increment counter
	counter.Counter++
	if err := s.db.Save(counter).Error; err != nil {
		return nil, fmt.Errorf("failed to update counter: %v", err)
	}

	// Format invoice number: {4 digit no}/{kode tipe}/{bulan romawi}-{4 digit tahun}
	romanMonth := models.GetRomanMonth(month)
	invoiceNumber := fmt.Sprintf("%04d/%s/%s-%04d", counter.Counter, invoiceType.Code, romanMonth, year)

	response := &models.InvoiceNumberResponse{
		InvoiceNumber: invoiceNumber,
		Counter:       counter.Counter,
		Year:          year,
		Month:         romanMonth,
		TypeCode:      invoiceType.Code,
	}

	return response, nil
}

// GetCurrentCounter returns the current counter value for a specific type and year (without incrementing)
func (s *InvoiceNumberService) GetCurrentCounter(invoiceTypeID uint, year int) (*models.InvoiceCounter, error) {
	var counter models.InvoiceCounter
	err := s.db.Where("invoice_type_id = ? AND year = ?", invoiceTypeID, year).First(&counter).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return a new counter with 0 count
			return &models.InvoiceCounter{
				InvoiceTypeID: invoiceTypeID,
				Year:          year,
				Counter:       0,
			}, nil
		}
		return nil, err
	}
	return &counter, nil
}

// PreviewInvoiceNumber generates a preview of what the next invoice number would be (without incrementing)
func (s *InvoiceNumberService) PreviewInvoiceNumber(invoiceTypeID uint, date time.Time) (*models.InvoiceNumberResponse, error) {
	// Get invoice type
	var invoiceType models.InvoiceType
	if err := s.db.Where("id = ? AND is_active = ?", invoiceTypeID, true).First(&invoiceType).Error; err != nil {
		return nil, fmt.Errorf("invoice type not found or inactive: %v", err)
	}

	year := date.Year()
	month := int(date.Month())
	
	// Get current counter (without creating or incrementing)
	counter, err := s.GetCurrentCounter(invoiceTypeID, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get counter: %v", err)
	}

	// Preview the next number (counter + 1)
	nextCounter := counter.Counter + 1
	romanMonth := models.GetRomanMonth(month)
	invoiceNumber := fmt.Sprintf("%04d/%s/%s-%04d", nextCounter, invoiceType.Code, romanMonth, year)

	response := &models.InvoiceNumberResponse{
		InvoiceNumber: invoiceNumber,
		Counter:       nextCounter,
		Year:          year,
		Month:         romanMonth,
		TypeCode:      invoiceType.Code,
	}

	return response, nil
}

// ResetCounterForYear resets the counter for a specific type and year (admin function)
func (s *InvoiceNumberService) ResetCounterForYear(invoiceTypeID uint, year int, newValue int) error {
	counter, err := s.getOrCreateCounter(invoiceTypeID, year)
	if err != nil {
		return fmt.Errorf("failed to get counter: %v", err)
	}

	counter.Counter = newValue
	if err := s.db.Save(counter).Error; err != nil {
		return fmt.Errorf("failed to reset counter: %v", err)
	}

	return nil
}

// GetCounterHistory returns counter history for a specific invoice type
func (s *InvoiceNumberService) GetCounterHistory(invoiceTypeID uint) ([]models.InvoiceCounter, error) {
	var counters []models.InvoiceCounter
	err := s.db.Where("invoice_type_id = ?", invoiceTypeID).
		Preload("InvoiceType").
		Order("year DESC").
		Find(&counters).Error
	return counters, err
}

// Private helper function to get or create counter
func (s *InvoiceNumberService) getOrCreateCounter(invoiceTypeID uint, year int) (*models.InvoiceCounter, error) {
	var counter models.InvoiceCounter
	
	// Try to find existing counter
	err := s.db.Where("invoice_type_id = ? AND year = ?", invoiceTypeID, year).First(&counter).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new counter for this type and year
			counter = models.InvoiceCounter{
				InvoiceTypeID: invoiceTypeID,
				Year:          year,
				Counter:       0,
			}
			if err := s.db.Create(&counter).Error; err != nil {
				return nil, fmt.Errorf("failed to create new counter: %v", err)
			}
		} else {
			return nil, err
		}
	}
	
	return &counter, nil
}

// ValidateInvoiceType checks if an invoice type is valid and active
func (s *InvoiceNumberService) ValidateInvoiceType(invoiceTypeID uint) error {
	var count int64
	err := s.db.Model(&models.InvoiceType{}).
		Where("id = ? AND is_active = ?", invoiceTypeID, true).
		Count(&count).Error
	
	if err != nil {
		return fmt.Errorf("database error: %v", err)
	}
	
	if count == 0 {
		return fmt.Errorf("invoice type not found or inactive")
	}
	
	return nil
}