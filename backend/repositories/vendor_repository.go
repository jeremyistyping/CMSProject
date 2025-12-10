package repositories

import (
	"app-sistem-akuntansi/models"
	"errors"

	"gorm.io/gorm"
)

type VendorRepository interface {
	// Vendor CRUD
	GetAll(filter map[string]interface{}) ([]models.Vendor, error)
	GetByID(id uint) (*models.Vendor, error)
	GetByCode(code string) (*models.Vendor, error)
	Create(vendor *models.Vendor) error
	Update(vendor *models.Vendor) error
	Delete(id uint) error
	GetSummary() (*models.VendorSummary, error)

	// Vendor Category CRUD
	GetAllCategories() ([]models.VendorCategory, error)
	GetCategoryByID(id uint) (*models.VendorCategory, error)
	CreateCategory(category *models.VendorCategory) error
	UpdateCategory(category *models.VendorCategory) error
	DeleteCategory(id uint) error

	// Vendor Material relationship
	GetVendorMaterials(vendorID uint) ([]models.VendorMaterial, error)
	GetMaterialVendors(materialID uint) ([]models.VendorMaterial, error)
	AddVendorMaterial(vm *models.VendorMaterial) error
	UpdateVendorMaterial(vm *models.VendorMaterial) error
	RemoveVendorMaterial(vendorID, materialID uint) error
	GetPreferredVendor(materialID uint) (*models.VendorMaterial, error)
}

type vendorRepository struct {
	db *gorm.DB
}

func NewVendorRepository(db *gorm.DB) VendorRepository {
	return &vendorRepository{db: db}
}

// Vendor CRUD
func (r *vendorRepository) GetAll(filter map[string]interface{}) ([]models.Vendor, error) {
	var vendors []models.Vendor
	query := r.db.Where("deleted_at IS NULL")

	if categoryID, ok := filter["category_id"]; ok {
		query = query.Where("category_id = ?", categoryID)
	}
	if isActive, ok := filter["is_active"]; ok {
		query = query.Where("is_active = ?", isActive)
	}
	if search, ok := filter["search"]; ok && search != "" {
		searchStr := "%" + search.(string) + "%"
		query = query.Where("code ILIKE ? OR name ILIKE ? OR contact_person ILIKE ?", searchStr, searchStr, searchStr)
	}
	if city, ok := filter["city"]; ok && city != "" {
		query = query.Where("city ILIKE ?", "%"+city.(string)+"%")
	}

	err := query.
		Preload("Category").
		Order("name ASC").
		Find(&vendors).Error
	return vendors, err
}

func (r *vendorRepository) GetByID(id uint) (*models.Vendor, error) {
	var vendor models.Vendor
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).
		Preload("Category").
		First(&vendor).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("vendor not found")
		}
		return nil, err
	}
	return &vendor, nil
}

func (r *vendorRepository) GetByCode(code string) (*models.Vendor, error) {
	var vendor models.Vendor
	err := r.db.Where("code = ? AND deleted_at IS NULL", code).
		Preload("Category").
		First(&vendor).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("vendor not found")
		}
		return nil, err
	}
	return &vendor, nil
}

func (r *vendorRepository) Create(vendor *models.Vendor) error {
	var count int64
	r.db.Model(&models.Vendor{}).Where("code = ? AND deleted_at IS NULL", vendor.Code).Count(&count)
	if count > 0 {
		return errors.New("vendor code already exists")
	}
	return r.db.Create(vendor).Error
}

func (r *vendorRepository) Update(vendor *models.Vendor) error {
	var count int64
	r.db.Model(&models.Vendor{}).
		Where("code = ? AND id != ? AND deleted_at IS NULL", vendor.Code, vendor.ID).
		Count(&count)
	if count > 0 {
		return errors.New("vendor code already exists")
	}
	return r.db.Save(vendor).Error
}

func (r *vendorRepository) Delete(id uint) error {
	// Check if used in purchase requests
	var prCount int64
	r.db.Model(&models.PurchaseRequest{}).Where("vendor_id = ? AND deleted_at IS NULL", id).Count(&prCount)
	if prCount > 0 {
		return errors.New("cannot delete vendor used in purchase requests")
	}

	return r.db.Model(&models.Vendor{}).Where("id = ?", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (r *vendorRepository) GetSummary() (*models.VendorSummary, error) {
	summary := &models.VendorSummary{}

	r.db.Model(&models.Vendor{}).Where("deleted_at IS NULL").Count(&summary.TotalVendors)
	r.db.Model(&models.Vendor{}).Where("deleted_at IS NULL AND is_active = true").Count(&summary.ActiveVendors)
	r.db.Model(&models.VendorCategory{}).Where("deleted_at IS NULL").Count(&summary.TotalCategories)

	r.db.Model(&models.Vendor{}).
		Where("deleted_at IS NULL AND rating > 0").
		Select("COALESCE(AVG(rating), 0)").
		Scan(&summary.AverageRating)

	return summary, nil
}

// Vendor Category CRUD
func (r *vendorRepository) GetAllCategories() ([]models.VendorCategory, error) {
	var categories []models.VendorCategory
	err := r.db.Where("deleted_at IS NULL").Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *vendorRepository) GetCategoryByID(id uint) (*models.VendorCategory, error) {
	var category models.VendorCategory
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("vendor category not found")
		}
		return nil, err
	}
	return &category, nil
}

func (r *vendorRepository) CreateCategory(category *models.VendorCategory) error {
	var count int64
	r.db.Model(&models.VendorCategory{}).Where("code = ? AND deleted_at IS NULL", category.Code).Count(&count)
	if count > 0 {
		return errors.New("category code already exists")
	}
	return r.db.Create(category).Error
}

func (r *vendorRepository) UpdateCategory(category *models.VendorCategory) error {
	var count int64
	r.db.Model(&models.VendorCategory{}).
		Where("code = ? AND id != ? AND deleted_at IS NULL", category.Code, category.ID).
		Count(&count)
	if count > 0 {
		return errors.New("category code already exists")
	}
	return r.db.Save(category).Error
}

func (r *vendorRepository) DeleteCategory(id uint) error {
	// Check if used by vendors
	var vendorCount int64
	r.db.Model(&models.Vendor{}).Where("category_id = ? AND deleted_at IS NULL", id).Count(&vendorCount)
	if vendorCount > 0 {
		return errors.New("cannot delete category used by vendors")
	}

	return r.db.Model(&models.VendorCategory{}).Where("id = ?", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

// Vendor Material relationship
func (r *vendorRepository) GetVendorMaterials(vendorID uint) ([]models.VendorMaterial, error) {
	var vms []models.VendorMaterial
	err := r.db.Where("vendor_id = ? AND deleted_at IS NULL", vendorID).
		Preload("Material").
		Preload("Material.Category").
		Find(&vms).Error
	return vms, err
}

func (r *vendorRepository) GetMaterialVendors(materialID uint) ([]models.VendorMaterial, error) {
	var vms []models.VendorMaterial
	err := r.db.Where("material_id = ? AND deleted_at IS NULL", materialID).
		Preload("Vendor").
		Preload("Vendor.Category").
		Order("is_preferred DESC, unit_price ASC").
		Find(&vms).Error
	return vms, err
}

func (r *vendorRepository) AddVendorMaterial(vm *models.VendorMaterial) error {
	// Check if already exists
	var count int64
	r.db.Model(&models.VendorMaterial{}).
		Where("vendor_id = ? AND material_id = ? AND deleted_at IS NULL", vm.VendorID, vm.MaterialID).
		Count(&count)
	if count > 0 {
		return errors.New("vendor-material relationship already exists")
	}
	return r.db.Create(vm).Error
}

func (r *vendorRepository) UpdateVendorMaterial(vm *models.VendorMaterial) error {
	return r.db.Save(vm).Error
}

func (r *vendorRepository) RemoveVendorMaterial(vendorID, materialID uint) error {
	return r.db.Model(&models.VendorMaterial{}).
		Where("vendor_id = ? AND material_id = ?", vendorID, materialID).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (r *vendorRepository) GetPreferredVendor(materialID uint) (*models.VendorMaterial, error) {
	var vm models.VendorMaterial
	err := r.db.Where("material_id = ? AND is_preferred = true AND deleted_at IS NULL", materialID).
		Preload("Vendor").
		First(&vm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return cheapest vendor if no preferred
			err = r.db.Where("material_id = ? AND deleted_at IS NULL", materialID).
				Preload("Vendor").
				Order("unit_price ASC").
				First(&vm).Error
			if err != nil {
				return nil, errors.New("no vendor found for this material")
			}
		} else {
			return nil, err
		}
	}
	return &vm, nil
}
