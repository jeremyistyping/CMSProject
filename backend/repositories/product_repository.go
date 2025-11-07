package repositories

import (
	"context"
	"fmt"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

// FindAll retrieves all products with preloaded relationships
func (pr *ProductRepository) FindAll(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	err := pr.db.WithContext(ctx).
		Preload("Category").
		Find(&products).Error
	return products, err
}

// FindByIDWithContext retrieves a product by ID with preloaded relationships using context
func (pr *ProductRepository) FindByIDWithContext(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	err := pr.db.WithContext(ctx).
		Preload("Category").
		First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// FindByID retrieves a product by ID - matches existing repository pattern
func (pr *ProductRepository) FindByID(id uint) (*models.Product, error) {
	return pr.FindByIDWithContext(context.Background(), id)
}

// FindByCode retrieves a product by code
func (pr *ProductRepository) FindByCode(ctx context.Context, code string) (*models.Product, error) {
	var product models.Product
	err := pr.db.WithContext(ctx).
		Preload("Category").
		Where("code = ?", code).
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// Create creates a new product
func (pr *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	return pr.db.WithContext(ctx).Create(product).Error
}

// Update updates an existing product
func (pr *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	return pr.db.WithContext(ctx).Save(product).Error
}

// Delete soft deletes a product
func (pr *ProductRepository) Delete(ctx context.Context, id uint) error {
	return pr.db.WithContext(ctx).Delete(&models.Product{}, id).Error
}

// UpdateStock updates product stock
func (pr *ProductRepository) UpdateStock(ctx context.Context, productID uint, newStock int) error {
	return pr.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("id = ?", productID).
		Update("stock", newStock).Error
}

// DecrementStock decrements product stock by the specified quantity
func (pr *ProductRepository) DecrementStock(ctx context.Context, productID uint, quantity int) error {
	result := pr.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("id = ? AND stock >= ?", productID, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity))
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("insufficient stock or product not found")
	}
	
	return nil
}

// IncrementStock increments product stock by the specified quantity
func (pr *ProductRepository) IncrementStock(ctx context.Context, productID uint, quantity int) error {
	return pr.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("id = ?", productID).
		Update("stock", gorm.Expr("stock + ?", quantity)).Error
}

// FindLowStockProducts retrieves products with stock below minimum threshold
func (pr *ProductRepository) FindLowStockProducts(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	err := pr.db.WithContext(ctx).
		Preload("Category").
		Where("stock <= min_stock").
		Find(&products).Error
	return products, err
}

// FindByCategory retrieves products by category ID
func (pr *ProductRepository) FindByCategory(ctx context.Context, categoryID uint) ([]models.Product, error) {
	var products []models.Product
	err := pr.db.WithContext(ctx).
		Preload("Category").
		Where("category_id = ?", categoryID).
		Find(&products).Error
	return products, err
}

// Search products by name or code
func (pr *ProductRepository) Search(ctx context.Context, query string) ([]models.Product, error) {
	var products []models.Product
	searchTerm := "%" + query + "%"
	err := pr.db.WithContext(ctx).
		Preload("Category").
		Where("name LIKE ? OR code LIKE ?", searchTerm, searchTerm).
		Find(&products).Error
	return products, err
}

// GetStockValuation calculates total stock valuation
func (pr *ProductRepository) GetStockValuation(ctx context.Context) (float64, error) {
	var totalValue float64
	err := pr.db.WithContext(ctx).
		Model(&models.Product{}).
		Select("COALESCE(SUM(stock * purchase_price), 0)").
		Scan(&totalValue).Error
	return totalValue, err
}

// BulkUpdatePrices updates prices for multiple products
func (pr *ProductRepository) BulkUpdatePrices(ctx context.Context, updates []models.ProductPriceUpdate) error {
	tx := pr.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, update := range updates {
		err := tx.Model(&models.Product{}).
			Where("id = ?", update.ProductID).
			Updates(map[string]interface{}{
				"purchase_price": update.PurchasePrice,
				"sale_price":     update.SalePrice,
			}).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetProductsByIDs retrieves multiple products by their IDs
func (pr *ProductRepository) GetProductsByIDs(ctx context.Context, ids []uint) ([]models.Product, error) {
	var products []models.Product
	err := pr.db.WithContext(ctx).
		Preload("Category").
		Where("id IN ?", ids).
		Find(&products).Error
	return products, err
}

// CheckCodeExists checks if a product code already exists
func (pr *ProductRepository) CheckCodeExists(ctx context.Context, code string, excludeID uint) (bool, error) {
	var count int64
	query := pr.db.WithContext(ctx).Model(&models.Product{}).Where("code = ?", code)
	
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	
	err := query.Count(&count).Error
	return count > 0, err
}
