package services

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"time"
)

type MaterialService interface {
	// Material operations
	GetAll(filter map[string]interface{}) ([]models.Material, error)
	GetByID(id uint) (*models.Material, error)
	GetByCode(code string) (*models.Material, error)
	Create(dto *models.CreateMaterialDTO, userID uint) (*models.Material, error)
	Update(id uint, dto *models.UpdateMaterialDTO) (*models.Material, error)
	Delete(id uint) error
	GetSummary() (*models.MaterialSummary, error)

	// Category operations
	GetAllCategories() ([]models.MaterialCategory, error)
	GetCategoryByID(id uint) (*models.MaterialCategory, error)
	CreateCategory(dto *models.CreateMaterialCategoryDTO) (*models.MaterialCategory, error)
	UpdateCategory(id uint, dto *models.CreateMaterialCategoryDTO) (*models.MaterialCategory, error)
	DeleteCategory(id uint) error
	GetCategoryTree() ([]models.MaterialCategory, error)

	// UoM operations
	GetAllUoM() ([]models.UnitOfMeasure, error)
	SeedDefaultUoM() error
	SeedDefaultCategories() error
}

type materialService struct {
	repo repositories.MaterialRepository
}

func NewMaterialService(repo repositories.MaterialRepository) MaterialService {
	return &materialService{repo: repo}
}

// Material operations
func (s *materialService) GetAll(filter map[string]interface{}) ([]models.Material, error) {
	return s.repo.GetAll(filter)
}

func (s *materialService) GetByID(id uint) (*models.Material, error) {
	return s.repo.GetByID(id)
}

func (s *materialService) GetByCode(code string) (*models.Material, error) {
	return s.repo.GetByCode(code)
}

func (s *materialService) Create(dto *models.CreateMaterialDTO, userID uint) (*models.Material, error) {
	material := &models.Material{
		Code:         dto.Code,
		Name:         dto.Name,
		Description:  dto.Description,
		CategoryID:   dto.CategoryID,
		Unit:         dto.Unit,
		UnitPrice:    dto.UnitPrice,
		MinStock:     dto.MinStock,
		MaxStock:     dto.MaxStock,
		COAAccountID: dto.COAAccountID,
		IsActive:     true,
		CreatedBy:    userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(material); err != nil {
		return nil, err
	}

	return material, nil
}

func (s *materialService) Update(id uint, dto *models.UpdateMaterialDTO) (*models.Material, error) {
	material, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if dto.Code != "" {
		material.Code = dto.Code
	}
	if dto.Name != "" {
		material.Name = dto.Name
	}
	if dto.Description != "" {
		material.Description = dto.Description
	}
	if dto.CategoryID != nil {
		material.CategoryID = dto.CategoryID
	}
	if dto.Unit != "" {
		material.Unit = dto.Unit
	}
	if dto.UnitPrice > 0 {
		material.UnitPrice = dto.UnitPrice
	}
	if dto.MinStock >= 0 {
		material.MinStock = dto.MinStock
	}
	if dto.MaxStock >= 0 {
		material.MaxStock = dto.MaxStock
	}
	if dto.COAAccountID != nil {
		material.COAAccountID = dto.COAAccountID
	}
	if dto.IsActive != nil {
		material.IsActive = *dto.IsActive
	}
	material.UpdatedAt = time.Now()

	if err := s.repo.Update(material); err != nil {
		return nil, err
	}

	return material, nil
}

func (s *materialService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *materialService) GetSummary() (*models.MaterialSummary, error) {
	return s.repo.GetSummary()
}

// Category operations
func (s *materialService) GetAllCategories() ([]models.MaterialCategory, error) {
	return s.repo.GetAllCategories()
}

func (s *materialService) GetCategoryByID(id uint) (*models.MaterialCategory, error) {
	return s.repo.GetCategoryByID(id)
}

func (s *materialService) CreateCategory(dto *models.CreateMaterialCategoryDTO) (*models.MaterialCategory, error) {
	category := &models.MaterialCategory{
		Code:        dto.Code,
		Name:        dto.Name,
		Description: dto.Description,
		ParentID:    dto.ParentID,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateCategory(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *materialService) UpdateCategory(id uint, dto *models.CreateMaterialCategoryDTO) (*models.MaterialCategory, error) {
	category, err := s.repo.GetCategoryByID(id)
	if err != nil {
		return nil, err
	}

	category.Code = dto.Code
	category.Name = dto.Name
	category.Description = dto.Description
	category.ParentID = dto.ParentID
	category.UpdatedAt = time.Now()

	if err := s.repo.UpdateCategory(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *materialService) DeleteCategory(id uint) error {
	return s.repo.DeleteCategory(id)
}

func (s *materialService) GetCategoryTree() ([]models.MaterialCategory, error) {
	return s.repo.GetCategoryTree()
}

// UoM operations
func (s *materialService) GetAllUoM() ([]models.UnitOfMeasure, error) {
	return s.repo.GetAllUoM()
}

func (s *materialService) SeedDefaultUoM() error {
	defaultUoMs := []models.UnitOfMeasure{
		// Length
		{Code: "m", Name: "Meter", Symbol: "m", Category: models.UoMCategoryLength},
		{Code: "cm", Name: "Centimeter", Symbol: "cm", Category: models.UoMCategoryLength},
		{Code: "mm", Name: "Milimeter", Symbol: "mm", Category: models.UoMCategoryLength},

		// Area
		{Code: "m2", Name: "Meter Persegi", Symbol: "m²", Category: models.UoMCategoryArea},

		// Volume
		{Code: "m3", Name: "Meter Kubik", Symbol: "m³", Category: models.UoMCategoryVolume},
		{Code: "ltr", Name: "Liter", Symbol: "L", Category: models.UoMCategoryVolume},

		// Weight
		{Code: "kg", Name: "Kilogram", Symbol: "kg", Category: models.UoMCategoryWeight},
		{Code: "ton", Name: "Ton", Symbol: "ton", Category: models.UoMCategoryWeight},
		{Code: "zak", Name: "Zak", Symbol: "zak", Category: models.UoMCategoryWeight},

		// Quantity
		{Code: "pcs", Name: "Pieces", Symbol: "pcs", Category: models.UoMCategoryQuantity},
		{Code: "unit", Name: "Unit", Symbol: "unit", Category: models.UoMCategoryQuantity},
		{Code: "set", Name: "Set", Symbol: "set", Category: models.UoMCategoryQuantity},
		{Code: "btg", Name: "Batang", Symbol: "btg", Category: models.UoMCategoryQuantity},
		{Code: "lbr", Name: "Lembar", Symbol: "lbr", Category: models.UoMCategoryQuantity},
		{Code: "roll", Name: "Roll", Symbol: "roll", Category: models.UoMCategoryQuantity},
		{Code: "dus", Name: "Dus", Symbol: "dus", Category: models.UoMCategoryQuantity},
	}

	for _, uom := range defaultUoMs {
		existing, _ := s.repo.GetAllUoM()
		found := false
		for _, e := range existing {
			if e.Code == uom.Code {
				found = true
				break
			}
		}
		if !found {
			uom.IsActive = true
			uom.CreatedAt = time.Now()
			uom.UpdatedAt = time.Now()
			s.repo.CreateUoM(&uom)
		}
	}

	return nil
}

func (s *materialService) SeedDefaultCategories() error {
	defaultCategories := []models.MaterialCategory{
		{Code: "STR", Name: "Material Struktural", Description: "Besi, beton, semen, dll"},
		{Code: "FIN", Name: "Material Finishing", Description: "Cat, keramik, plafon, dll"},
		{Code: "MEP", Name: "Material MEP", Description: "Pipa, kabel, fitting, dll"},
		{Code: "WOD", Name: "Material Kayu", Description: "Kayu, triplek, dll"},
		{Code: "SAN", Name: "Sanitary", Description: "Closet, wastafel, kran, dll"},
		{Code: "ELC", Name: "Elektrikal", Description: "Kabel, saklar, lampu, dll"},
		{Code: "OTH", Name: "Lain-lain", Description: "Material lainnya"},
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
