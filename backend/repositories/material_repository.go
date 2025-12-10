package repositories

import (
	"app-sistem-akuntansi/models"
	"errors"

	"gorm.io/gorm"
)

type MaterialRepository interface {
	// Material CRUD
	GetAll(filter map[string]interface{}) ([]models.Material, error)
	GetByID(id uint) (*models.Material, error)
	GetByCode(code string) (*models.Material, error)
	Create(material *models.Material) error
	Update(material *models.Material) error
	Delete(id uint) error
	GetSummary() (*models.MaterialSummary, error)

	// Material Category CRUD
	GetAllCategories() ([]models.MaterialCategory, error)
	GetCategoryByID(id uint) (*models.MaterialCategory, error)
	CreateCategory(category *models.MaterialCategory) error
	UpdateCategory(category *models.MaterialCategory) error
	DeleteCategory(id uint) error
	GetCategoryTree() ([]models.MaterialCategory, error)

	// Unit of Measure CRUD
	GetAllUoM() ([]models.UnitOfMeasure, error)
	GetUoMByID(id uint) (*models.UnitOfMeasure, error)
	CreateUoM(uom *models.UnitOfMeasure) error
	UpdateUoM(uom *models.UnitOfMeasure) error
	DeleteUoM(id uint) error
}

type materialRepository struct {
	db *gorm.DB
}

func NewMaterialRepository(db *gorm.DB) MaterialRepository {
	return &materialRepository{db: db}
}

// Material CRUD
func (r *materialRepository) GetAll(filter map[string]interface{}) ([]models.Material, error) {
	var materials []models.Material
	query := r.db.Where("deleted_at IS NULL")

	if categoryID, ok := filter["category_id"]; ok {
		query = query.Where("category_id = ?", categoryID)
	}
	if isActive, ok := filter["is_active"]; ok {
		query = query.Where("is_active = ?", isActive)
	}
	if search, ok := filter["search"]; ok && search != "" {
		searchStr := "%" + search.(string) + "%"
		query = query.Where("code ILIKE ? OR name ILIKE ?", searchStr, searchStr)
	}
	if lowStock, ok := filter["low_stock"]; ok && lowStock == true {
		query = query.Where("current_stock <= min_stock")
	}

	err := query.
		Preload("Category").
		Preload("COAAccount").
		Order("code ASC").
		Find(&materials).Error
	return materials, err
}

func (r *materialRepository) GetByID(id uint) (*models.Material, error) {
	var material models.Material
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).
		Preload("Category").
		Preload("COAAccount").
		First(&material).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("material not found")
		}
		return nil, err
	}
	return &material, nil
}

func (r *materialRepository) GetByCode(code string) (*models.Material, error) {
	var material models.Material
	err := r.db.Where("code = ? AND deleted_at IS NULL", code).
		Preload("Category").
		Preload("COAAccount").
		First(&material).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("material not found")
		}
		return nil, err
	}
	return &material, nil
}

func (r *materialRepository) Create(material *models.Material) error {
	var count int64
	r.db.Model(&models.Material{}).Where("code = ? AND deleted_at IS NULL", material.Code).Count(&count)
	if count > 0 {
		return errors.New("material code already exists")
	}
	return r.db.Create(material).Error
}

func (r *materialRepository) Update(material *models.Material) error {
	var count int64
	r.db.Model(&models.Material{}).
		Where("code = ? AND id != ? AND deleted_at IS NULL", material.Code, material.ID).
		Count(&count)
	if count > 0 {
		return errors.New("material code already exists")
	}
	return r.db.Save(material).Error
}

func (r *materialRepository) Delete(id uint) error {
	// Check if used in purchase request items
	var prItemCount int64
	r.db.Table("purchase_request_items").Where("product_id = ? AND deleted_at IS NULL", id).Count(&prItemCount)
	if prItemCount > 0 {
		return errors.New("cannot delete material used in purchase requests")
	}

	return r.db.Model(&models.Material{}).Where("id = ?", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (r *materialRepository) GetSummary() (*models.MaterialSummary, error) {
	summary := &models.MaterialSummary{}

	r.db.Model(&models.Material{}).Where("deleted_at IS NULL").Count(&summary.TotalMaterials)
	r.db.Model(&models.Material{}).Where("deleted_at IS NULL AND is_active = true").Count(&summary.ActiveMaterials)
	r.db.Model(&models.Material{}).Where("deleted_at IS NULL AND current_stock <= min_stock").Count(&summary.LowStockCount)
	r.db.Model(&models.MaterialCategory{}).Where("deleted_at IS NULL").Count(&summary.TotalCategories)

	r.db.Model(&models.Material{}).
		Where("deleted_at IS NULL").
		Select("COALESCE(SUM(current_stock * unit_price), 0)").
		Scan(&summary.TotalStockValue)

	return summary, nil
}

// Material Category CRUD
func (r *materialRepository) GetAllCategories() ([]models.MaterialCategory, error) {
	var categories []models.MaterialCategory
	err := r.db.Where("deleted_at IS NULL").Order("code ASC").Find(&categories).Error
	return categories, err
}

func (r *materialRepository) GetCategoryByID(id uint) (*models.MaterialCategory, error) {
	var category models.MaterialCategory
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).
		Preload("Parent").
		First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("material category not found")
		}
		return nil, err
	}
	return &category, nil
}

func (r *materialRepository) CreateCategory(category *models.MaterialCategory) error {
	var count int64
	r.db.Model(&models.MaterialCategory{}).Where("code = ? AND deleted_at IS NULL", category.Code).Count(&count)
	if count > 0 {
		return errors.New("category code already exists")
	}
	return r.db.Create(category).Error
}

func (r *materialRepository) UpdateCategory(category *models.MaterialCategory) error {
	var count int64
	r.db.Model(&models.MaterialCategory{}).
		Where("code = ? AND id != ? AND deleted_at IS NULL", category.Code, category.ID).
		Count(&count)
	if count > 0 {
		return errors.New("category code already exists")
	}
	return r.db.Save(category).Error
}

func (r *materialRepository) DeleteCategory(id uint) error {
	// Check if has children
	var childCount int64
	r.db.Model(&models.MaterialCategory{}).Where("parent_id = ? AND deleted_at IS NULL", id).Count(&childCount)
	if childCount > 0 {
		return errors.New("cannot delete category with children")
	}

	// Check if used by materials
	var materialCount int64
	r.db.Model(&models.Material{}).Where("category_id = ? AND deleted_at IS NULL", id).Count(&materialCount)
	if materialCount > 0 {
		return errors.New("cannot delete category used by materials")
	}

	return r.db.Model(&models.MaterialCategory{}).Where("id = ?", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (r *materialRepository) GetCategoryTree() ([]models.MaterialCategory, error) {
	var categories []models.MaterialCategory
	err := r.db.Where("deleted_at IS NULL").Order("code ASC").Find(&categories).Error
	if err != nil {
		return nil, err
	}

	// Build tree
	categoryMap := make(map[uint]*models.MaterialCategory)
	for i := range categories {
		categories[i].Children = []models.MaterialCategory{}
		categoryMap[categories[i].ID] = &categories[i]
	}

	var rootCategories []models.MaterialCategory
	for i := range categories {
		if categories[i].ParentID == nil {
			rootCategories = append(rootCategories, categories[i])
		} else {
			if parent, ok := categoryMap[*categories[i].ParentID]; ok {
				parent.Children = append(parent.Children, categories[i])
			}
		}
	}

	return rootCategories, nil
}

// Unit of Measure CRUD
func (r *materialRepository) GetAllUoM() ([]models.UnitOfMeasure, error) {
	var uoms []models.UnitOfMeasure
	err := r.db.Where("deleted_at IS NULL").Order("category ASC, code ASC").Find(&uoms).Error
	return uoms, err
}

func (r *materialRepository) GetUoMByID(id uint) (*models.UnitOfMeasure, error) {
	var uom models.UnitOfMeasure
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&uom).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("unit of measure not found")
		}
		return nil, err
	}
	return &uom, nil
}

func (r *materialRepository) CreateUoM(uom *models.UnitOfMeasure) error {
	var count int64
	r.db.Model(&models.UnitOfMeasure{}).Where("code = ? AND deleted_at IS NULL", uom.Code).Count(&count)
	if count > 0 {
		return errors.New("UoM code already exists")
	}
	return r.db.Create(uom).Error
}

func (r *materialRepository) UpdateUoM(uom *models.UnitOfMeasure) error {
	return r.db.Save(uom).Error
}

func (r *materialRepository) DeleteUoM(id uint) error {
	return r.db.Model(&models.UnitOfMeasure{}).Where("id = ?", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}
