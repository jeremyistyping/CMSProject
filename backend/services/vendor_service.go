package services

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"time"
)

type VendorService interface {
	// Vendor operations
	GetAll(filter map[string]interface{}) ([]models.Vendor, error)
	GetByID(id uint) (*models.Vendor, error)
	GetByCode(code string) (*models.Vendor, error)
	Create(dto *models.CreateVendorDTO, userID uint) (*models.Vendor, error)
	Update(id uint, dto *models.UpdateVendorDTO) (*models.Vendor, error)
	Delete(id uint) error
	GetSummary() (*models.VendorSummary, error)

	// Category operations
	GetAllCategories() ([]models.VendorCategory, error)
	GetCategoryByID(id uint) (*models.VendorCategory, error)
	CreateCategory(dto *models.CreateVendorCategoryDTO) (*models.VendorCategory, error)
	UpdateCategory(id uint, dto *models.CreateVendorCategoryDTO) (*models.VendorCategory, error)
	DeleteCategory(id uint) error

	// Vendor-Material operations
	GetVendorMaterials(vendorID uint) ([]models.VendorMaterial, error)
	GetMaterialVendors(materialID uint) ([]models.VendorMaterial, error)
	AddVendorMaterial(vendorID, materialID uint, unitPrice float64, leadTime int, isPreferred bool) error
	RemoveVendorMaterial(vendorID, materialID uint) error
	GetPreferredVendor(materialID uint) (*models.VendorMaterial, error)

	SeedDefaultCategories() error
}

type vendorService struct {
	repo repositories.VendorRepository
}

func NewVendorService(repo repositories.VendorRepository) VendorService {
	return &vendorService{repo: repo}
}


// Vendor operations
func (s *vendorService) GetAll(filter map[string]interface{}) ([]models.Vendor, error) {
	return s.repo.GetAll(filter)
}

func (s *vendorService) GetByID(id uint) (*models.Vendor, error) {
	return s.repo.GetByID(id)
}

func (s *vendorService) GetByCode(code string) (*models.Vendor, error) {
	return s.repo.GetByCode(code)
}

func (s *vendorService) Create(dto *models.CreateVendorDTO, userID uint) (*models.Vendor, error) {
	vendor := &models.Vendor{
		Code:          dto.Code,
		Name:          dto.Name,
		ContactPerson: dto.ContactPerson,
		Email:         dto.Email,
		Phone:         dto.Phone,
		Address:       dto.Address,
		City:          dto.City,
		Province:      dto.Province,
		PostalCode:    dto.PostalCode,
		NPWP:          dto.NPWP,
		BankName:      dto.BankName,
		BankAccount:   dto.BankAccount,
		BankBranch:    dto.BankBranch,
		PaymentTerms:  dto.PaymentTerms,
		CategoryID:    dto.CategoryID,
		Notes:         dto.Notes,
		IsActive:      true,
		CreatedBy:     userID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if vendor.PaymentTerms == 0 {
		vendor.PaymentTerms = 30
	}

	if err := s.repo.Create(vendor); err != nil {
		return nil, err
	}

	return vendor, nil
}

func (s *vendorService) Update(id uint, dto *models.UpdateVendorDTO) (*models.Vendor, error) {
	vendor, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if dto.Code != "" {
		vendor.Code = dto.Code
	}
	if dto.Name != "" {
		vendor.Name = dto.Name
	}
	vendor.ContactPerson = dto.ContactPerson
	vendor.Email = dto.Email
	vendor.Phone = dto.Phone
	vendor.Address = dto.Address
	vendor.City = dto.City
	vendor.Province = dto.Province
	vendor.PostalCode = dto.PostalCode
	vendor.NPWP = dto.NPWP
	vendor.BankName = dto.BankName
	vendor.BankAccount = dto.BankAccount
	vendor.BankBranch = dto.BankBranch
	if dto.PaymentTerms > 0 {
		vendor.PaymentTerms = dto.PaymentTerms
	}
	if dto.CategoryID != nil {
		vendor.CategoryID = dto.CategoryID
	}
	if dto.Rating > 0 {
		vendor.Rating = dto.Rating
	}
	vendor.Notes = dto.Notes
	if dto.IsActive != nil {
		vendor.IsActive = *dto.IsActive
	}
	vendor.UpdatedAt = time.Now()

	if err := s.repo.Update(vendor); err != nil {
		return nil, err
	}

	return vendor, nil
}

func (s *vendorService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *vendorService) GetSummary() (*models.VendorSummary, error) {
	return s.repo.GetSummary()
}


// Category operations
func (s *vendorService) GetAllCategories() ([]models.VendorCategory, error) {
	return s.repo.GetAllCategories()
}

func (s *vendorService) GetCategoryByID(id uint) (*models.VendorCategory, error) {
	return s.repo.GetCategoryByID(id)
}

func (s *vendorService) CreateCategory(dto *models.CreateVendorCategoryDTO) (*models.VendorCategory, error) {
	category := &models.VendorCategory{
		Code:        dto.Code,
		Name:        dto.Name,
		Description: dto.Description,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateCategory(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *vendorService) UpdateCategory(id uint, dto *models.CreateVendorCategoryDTO) (*models.VendorCategory, error) {
	category, err := s.repo.GetCategoryByID(id)
	if err != nil {
		return nil, err
	}

	category.Code = dto.Code
	category.Name = dto.Name
	category.Description = dto.Description
	category.UpdatedAt = time.Now()

	if err := s.repo.UpdateCategory(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *vendorService) DeleteCategory(id uint) error {
	return s.repo.DeleteCategory(id)
}

// Vendor-Material operations
func (s *vendorService) GetVendorMaterials(vendorID uint) ([]models.VendorMaterial, error) {
	return s.repo.GetVendorMaterials(vendorID)
}

func (s *vendorService) GetMaterialVendors(materialID uint) ([]models.VendorMaterial, error) {
	return s.repo.GetMaterialVendors(materialID)
}

func (s *vendorService) AddVendorMaterial(vendorID, materialID uint, unitPrice float64, leadTime int, isPreferred bool) error {
	vm := &models.VendorMaterial{
		VendorID:     vendorID,
		MaterialID:   materialID,
		UnitPrice:    unitPrice,
		LeadTimeDays: leadTime,
		IsPreferred:  isPreferred,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	return s.repo.AddVendorMaterial(vm)
}

func (s *vendorService) RemoveVendorMaterial(vendorID, materialID uint) error {
	return s.repo.RemoveVendorMaterial(vendorID, materialID)
}

func (s *vendorService) GetPreferredVendor(materialID uint) (*models.VendorMaterial, error) {
	return s.repo.GetPreferredVendor(materialID)
}

func (s *vendorService) SeedDefaultCategories() error {
	defaultCategories := []models.VendorCategory{
		{Code: "MAT", Name: "Supplier Material", Description: "Supplier bahan bangunan"},
		{Code: "EQP", Name: "Rental Alat", Description: "Penyewaan alat berat dan scaffolding"},
		{Code: "SUB", Name: "Subkontraktor", Description: "Subkontraktor pekerjaan"},
		{Code: "SVC", Name: "Jasa", Description: "Penyedia jasa lainnya"},
		{Code: "OTH", Name: "Lainnya", Description: "Vendor lainnya"},
	}

	for _, cat := range defaultCategories {
		existing, _ := s.repo.GetAllCategories()
		found := false
		for _, e := range existing {
			if e.Code == cat.Code {
				found = true
				break
			}
		}
		if !found {
			cat.IsActive = true
			cat.CreatedAt = time.Now()
			cat.UpdatedAt = time.Now()
			s.repo.CreateCategory(&cat)
		}
	}

	return nil
}
