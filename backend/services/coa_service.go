package services

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"errors"
	"time"
)

type COAService interface {
	GetAll(filter map[string]interface{}) ([]models.COAAccount, error)
	GetByID(id uint) (*models.COAAccount, error)
	GetByCode(code string) (*models.COAAccount, error)
	GetTree() ([]*models.COATreeNode, error)
	GetByType(coaType string) ([]models.COAAccount, error)
	GetByCategory(category string) ([]models.COAAccount, error)
	GetByBudgetCategory(budgetCategory string) ([]models.COAAccount, error)
	Create(dto *models.CreateCOADTO) (*models.COAAccount, error)
	Update(id uint, dto *models.UpdateCOADTO) (*models.COAAccount, error)
	Delete(id uint) error
	SeedDefaultCOA() error
}

type coaService struct {
	repo repositories.COARepository
}

func NewCOAService(repo repositories.COARepository) COAService {
	return &coaService{repo: repo}
}

func (s *coaService) GetAll(filter map[string]interface{}) ([]models.COAAccount, error) {
	return s.repo.GetAll(filter)
}

func (s *coaService) GetByID(id uint) (*models.COAAccount, error) {
	return s.repo.GetByID(id)
}

func (s *coaService) GetByCode(code string) (*models.COAAccount, error) {
	return s.repo.GetByCode(code)
}

func (s *coaService) GetTree() ([]*models.COATreeNode, error) {
	return s.repo.GetTree()
}

func (s *coaService) GetByType(coaType string) ([]models.COAAccount, error) {
	return s.repo.GetByType(coaType)
}

func (s *coaService) GetByCategory(category string) ([]models.COAAccount, error) {
	return s.repo.GetByCategory(category)
}

func (s *coaService) GetByBudgetCategory(budgetCategory string) ([]models.COAAccount, error) {
	return s.repo.GetByBudgetCategory(budgetCategory)
}

func (s *coaService) Create(dto *models.CreateCOADTO) (*models.COAAccount, error) {
	// Validate type
	validTypes := map[string]bool{
		models.COATypeAsset:     true,
		models.COATypeLiability: true,
		models.COATypeEquity:    true,
		models.COATypeRevenue:   true,
		models.COATypeExpense:   true,
	}
	if !validTypes[dto.Type] {
		return nil, errors.New("invalid COA type")
	}

	coa := &models.COAAccount{
		Code:           dto.Code,
		Name:           dto.Name,
		Description:    dto.Description,
		Type:           dto.Type,
		Category:       dto.Category,
		BudgetCategory: dto.BudgetCategory,
		WorkPackage:    dto.WorkPackage,
		ParentID:       dto.ParentID,
		IsHeader:       dto.IsHeader,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.Create(coa); err != nil {
		return nil, err
	}

	return coa, nil
}

func (s *coaService) Update(id uint, dto *models.UpdateCOADTO) (*models.COAAccount, error) {
	coa, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if dto.Code != "" {
		coa.Code = dto.Code
	}
	if dto.Name != "" {
		coa.Name = dto.Name
	}
	if dto.Description != "" {
		coa.Description = dto.Description
	}
	if dto.Type != "" {
		coa.Type = dto.Type
	}
	if dto.Category != "" {
		coa.Category = dto.Category
	}
	if dto.BudgetCategory != "" {
		coa.BudgetCategory = dto.BudgetCategory
	}
	if dto.WorkPackage != "" {
		coa.WorkPackage = dto.WorkPackage
	}
	if dto.ParentID != nil {
		coa.ParentID = dto.ParentID
	}
	if dto.IsActive != nil {
		coa.IsActive = *dto.IsActive
	}
	if dto.IsHeader != nil {
		coa.IsHeader = *dto.IsHeader
	}
	coa.UpdatedAt = time.Now()

	if err := s.repo.Update(coa); err != nil {
		return nil, err
	}

	return coa, nil
}

func (s *coaService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// SeedDefaultCOA creates default COA for construction projects
func (s *coaService) SeedDefaultCOA() error {
	defaultCOAs := []models.COAAccount{
		// EXPENSE - Header
		{Code: "5000", Name: "Biaya Proyek", Type: models.COATypeExpense, IsHeader: true, Level: 1},

		// Material Costs
		{Code: "5100", Name: "Biaya Material", Type: models.COATypeExpense, Category: models.COACategoryMaterial, IsHeader: true, Level: 2},
		{Code: "5101", Name: "Material Struktural", Type: models.COATypeExpense, Category: models.COACategoryMaterial, Level: 3},
		{Code: "5102", Name: "Material Finishing", Type: models.COATypeExpense, Category: models.COACategoryMaterial, Level: 3},
		{Code: "5103", Name: "Material MEP", Type: models.COATypeExpense, Category: models.COACategoryMaterial, Level: 3},
		{Code: "5104", Name: "Material Lainnya", Type: models.COATypeExpense, Category: models.COACategoryMaterial, Level: 3},

		// Labor Costs
		{Code: "5200", Name: "Biaya Tenaga Kerja", Type: models.COATypeExpense, Category: models.COACategoryLabor, IsHeader: true, Level: 2},
		{Code: "5201", Name: "Upah Pekerja Harian", Type: models.COATypeExpense, Category: models.COACategoryLabor, Level: 3},
		{Code: "5202", Name: "Upah Pekerja Borongan", Type: models.COATypeExpense, Category: models.COACategoryLabor, Level: 3},
		{Code: "5203", Name: "Upah Mandor", Type: models.COATypeExpense, Category: models.COACategoryLabor, Level: 3},
		{Code: "5204", Name: "Upah Tenaga Ahli", Type: models.COATypeExpense, Category: models.COACategoryLabor, Level: 3},

		// Equipment Costs
		{Code: "5300", Name: "Biaya Peralatan", Type: models.COATypeExpense, Category: models.COACategoryEquipment, IsHeader: true, Level: 2},
		{Code: "5301", Name: "Sewa Alat Berat", Type: models.COATypeExpense, Category: models.COACategoryEquipment, Level: 3},
		{Code: "5302", Name: "Sewa Scaffolding", Type: models.COATypeExpense, Category: models.COACategoryEquipment, Level: 3},
		{Code: "5303", Name: "Sewa Alat Bantu", Type: models.COATypeExpense, Category: models.COACategoryEquipment, Level: 3},
		{Code: "5304", Name: "BBM & Pelumas", Type: models.COATypeExpense, Category: models.COACategoryEquipment, Level: 3},

		// Subcontractor Costs
		{Code: "5400", Name: "Biaya Subkontraktor", Type: models.COATypeExpense, Category: models.COACategorySubcontractor, IsHeader: true, Level: 2},
		{Code: "5401", Name: "Subkon Struktural", Type: models.COATypeExpense, Category: models.COACategorySubcontractor, Level: 3},
		{Code: "5402", Name: "Subkon MEP", Type: models.COATypeExpense, Category: models.COACategorySubcontractor, Level: 3},
		{Code: "5403", Name: "Subkon Finishing", Type: models.COATypeExpense, Category: models.COACategorySubcontractor, Level: 3},

		// Overhead Costs
		{Code: "5500", Name: "Biaya Overhead", Type: models.COATypeExpense, Category: models.COACategoryOverhead, IsHeader: true, Level: 2},
		{Code: "5501", Name: "Biaya Kantor Proyek", Type: models.COATypeExpense, Category: models.COACategoryOverhead, Level: 3},
		{Code: "5502", Name: "Biaya Keamanan", Type: models.COATypeExpense, Category: models.COACategoryOverhead, Level: 3},
		{Code: "5503", Name: "Biaya Utilitas", Type: models.COATypeExpense, Category: models.COACategoryOverhead, Level: 3},
		{Code: "5504", Name: "Biaya Transportasi", Type: models.COATypeExpense, Category: models.COACategoryOverhead, Level: 3},
		{Code: "5505", Name: "Biaya Asuransi", Type: models.COATypeExpense, Category: models.COACategoryOverhead, Level: 3},
		{Code: "5506", Name: "Biaya Lain-lain", Type: models.COATypeExpense, Category: models.COACategoryOverhead, Level: 3},
	}

	// Set parent IDs
	parentMap := map[string]uint{}
	for i, coa := range defaultCOAs {
		// Check if already exists
		existing, _ := s.repo.GetByCode(coa.Code)
		if existing != nil {
			parentMap[coa.Code] = existing.ID
			continue
		}

		// Set parent based on code prefix
		if len(coa.Code) == 4 && coa.Code[0] == '5' && coa.Code[1] != '0' {
			if parentID, ok := parentMap["5000"]; ok {
				defaultCOAs[i].ParentID = &parentID
			}
		} else if len(coa.Code) == 4 && coa.Code[3] != '0' {
			parentCode := coa.Code[:3] + "0"
			if parentID, ok := parentMap[parentCode]; ok {
				defaultCOAs[i].ParentID = &parentID
			}
		}

		defaultCOAs[i].IsActive = true
		defaultCOAs[i].CreatedAt = time.Now()
		defaultCOAs[i].UpdatedAt = time.Now()

		if err := s.repo.Create(&defaultCOAs[i]); err != nil {
			continue // Skip if error (might already exist)
		}
		parentMap[coa.Code] = defaultCOAs[i].ID
	}

	return nil
}
