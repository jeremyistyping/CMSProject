package services

import (
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// COAService handles Chart of Accounts operations
type COAService struct {
	db *gorm.DB
}

// NewCOAService creates a new instance
func NewCOAService(db *gorm.DB) *COAService {
	return &COAService{db: db}
}

// GetByID retrieves a COA by ID
func (s *COAService) GetByID(id uint) (*models.COA, error) {
	var coa models.COA
	err := s.db.First(&coa, id).Error
	return &coa, err
}

// GetByCode retrieves a COA by code
func (s *COAService) GetByCode(code string) (*models.COA, error) {
	var coa models.COA
	err := s.db.Where("code = ?", code).First(&coa).Error
	return &coa, err
}

// UpdateBalance updates the balance of a COA
func (s *COAService) UpdateBalance(id uint, amount float64) error {
	return s.db.Model(&models.COA{}).Where("id = ?", id).Update("balance", gorm.Expr("balance + ?", amount)).Error
}

// GetAll retrieves all COAs
func (s *COAService) GetAll() ([]models.COA, error) {
	var coas []models.COA
	err := s.db.Find(&coas).Error
	return coas, err
}

// GetAccountsWithFilter retrieves accounts with filters
func (s *COAService) GetAccountsWithFilter(filter map[string]interface{}) ([]models.Account, error) {
	var accounts []models.Account
	query := s.db.Where("deleted_at IS NULL")
	
	for key, value := range filter {
		query = query.Where(key+" = ?", value)
	}
	
	err := query.Order("code ASC").Find(&accounts).Error
	return accounts, err
}
