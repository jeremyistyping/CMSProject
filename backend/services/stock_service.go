package services

import (
	"fmt"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

// StockService handles stock/inventory operations
type StockService struct {
	db *gorm.DB
}

// NewStockService creates a new instance
func NewStockService(db *gorm.DB) *StockService {
	return &StockService{db: db}
}

// ReduceStock reduces stock for a product
func (s *StockService) ReduceStock(productID uint, quantity int, tx *gorm.DB) error {
	dbToUse := s.db
	if tx != nil {
		dbToUse = tx
	}

	// Get product
	var product models.Product
	if err := dbToUse.First(&product, productID).Error; err != nil {
		return fmt.Errorf("product not found: %v", err)
	}

	// Skip stock management for service products
	if product.IsService {
		return nil // No stock management needed for services
	}

	// Check available stock
	if product.Stock < quantity {
		return fmt.Errorf("insufficient stock: available %d, requested %d", product.Stock, quantity)
	}

	// Reduce stock
	product.Stock -= quantity
	return dbToUse.Save(&product).Error
}

// RestoreStock restores stock for a product
func (s *StockService) RestoreStock(productID uint, quantity int, tx *gorm.DB) error {
	dbToUse := s.db
	if tx != nil {
		dbToUse = tx
	}

	// Get product
	var product models.Product
	if err := dbToUse.First(&product, productID).Error; err != nil {
		return fmt.Errorf("product not found: %v", err)
	}

	// Skip stock management for service products
	if product.IsService {
		return nil // No stock management needed for services
	}

	// Restore stock
	product.Stock += quantity
	return dbToUse.Save(&product).Error
}

// GetStock gets current stock for a product
func (s *StockService) GetStock(productID uint) (int, error) {
	var product models.Product
	if err := s.db.First(&product, productID).Error; err != nil {
		return 0, fmt.Errorf("product not found: %v", err)
	}
	return product.Stock, nil
}

// CheckStock checks if sufficient stock is available
func (s *StockService) CheckStock(productID uint, quantity int) (bool, error) {
	stock, err := s.GetStock(productID)
	if err != nil {
		return false, err
	}
	return stock >= quantity, nil
}