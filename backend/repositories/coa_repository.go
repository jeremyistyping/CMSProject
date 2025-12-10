package repositories

import (
	"app-sistem-akuntansi/models"
	"errors"

	"gorm.io/gorm"
)

type COARepository interface {
	GetAll(filter map[string]interface{}) ([]models.COAAccount, error)
	GetByID(id uint) (*models.COAAccount, error)
	GetByCode(code string) (*models.COAAccount, error)
	GetTree() ([]*models.COATreeNode, error)
	GetByType(coaType string) ([]models.COAAccount, error)
	GetByCategory(category string) ([]models.COAAccount, error)
	GetByBudgetCategory(budgetCategory string) ([]models.COAAccount, error)
	Create(coa *models.COAAccount) error
	Update(coa *models.COAAccount) error
	Delete(id uint) error
	GetChildren(parentID uint) ([]models.COAAccount, error)
}

type coaRepository struct {
	db *gorm.DB
}

func NewCOARepository(db *gorm.DB) COARepository {
	return &coaRepository{db: db}
}

func (r *coaRepository) GetAll(filter map[string]interface{}) ([]models.COAAccount, error) {
	var accounts []models.COAAccount
	query := r.db.Where("deleted_at IS NULL")

	if coaType, ok := filter["type"]; ok {
		query = query.Where("type = ?", coaType)
	}
	if category, ok := filter["category"]; ok {
		query = query.Where("category = ?", category)
	}
	if budgetCategory, ok := filter["budget_category"]; ok {
		query = query.Where("budget_category = ?", budgetCategory)
	}
	if workPackage, ok := filter["work_package"]; ok {
		query = query.Where("work_package = ?", workPackage)
	}
	if isActive, ok := filter["is_active"]; ok {
		query = query.Where("is_active = ?", isActive)
	}
	if isHeader, ok := filter["is_header"]; ok {
		query = query.Where("is_header = ?", isHeader)
	}
	if search, ok := filter["search"]; ok && search != "" {
		searchStr := "%" + search.(string) + "%"
		query = query.Where("code ILIKE ? OR name ILIKE ?", searchStr, searchStr)
	}

	err := query.Order("code ASC").Find(&accounts).Error
	return accounts, err
}

func (r *coaRepository) GetByID(id uint) (*models.COAAccount, error) {
	var account models.COAAccount
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).
		Preload("Parent").
		First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("COA account not found")
		}
		return nil, err
	}
	return &account, nil
}

func (r *coaRepository) GetByCode(code string) (*models.COAAccount, error) {
	var account models.COAAccount
	err := r.db.Where("code = ? AND deleted_at IS NULL", code).First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("COA account not found")
		}
		return nil, err
	}
	return &account, nil
}

func (r *coaRepository) GetTree() ([]*models.COATreeNode, error) {
	var accounts []models.COAAccount
	err := r.db.Where("deleted_at IS NULL").Order("code ASC").Find(&accounts).Error
	if err != nil {
		return nil, err
	}

	// Build tree structure
	nodeMap := make(map[uint]*models.COATreeNode)
	for _, acc := range accounts {
		nodeMap[acc.ID] = &models.COATreeNode{
			ID:             acc.ID,
			Code:           acc.Code,
			Name:           acc.Name,
			Type:           acc.Type,
			Category:       acc.Category,
			BudgetCategory: acc.BudgetCategory,
			WorkPackage:    acc.WorkPackage,
			Level:          acc.Level,
			IsActive:       acc.IsActive,
			IsHeader:       acc.IsHeader,
			Balance:        acc.Balance,
			Children:       []*models.COATreeNode{},
		}
	}

	var rootNodes []*models.COATreeNode
	for _, acc := range accounts {
		node := nodeMap[acc.ID]
		if acc.ParentID == nil {
			rootNodes = append(rootNodes, node)
		} else {
			if parent, ok := nodeMap[*acc.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return rootNodes, nil
}

func (r *coaRepository) GetByType(coaType string) ([]models.COAAccount, error) {
	var accounts []models.COAAccount
	err := r.db.Where("type = ? AND deleted_at IS NULL", coaType).
		Order("code ASC").Find(&accounts).Error
	return accounts, err
}

func (r *coaRepository) GetByCategory(category string) ([]models.COAAccount, error) {
	var accounts []models.COAAccount
	err := r.db.Where("category = ? AND deleted_at IS NULL AND is_active = true", category).
		Order("code ASC").Find(&accounts).Error
	return accounts, err
}

func (r *coaRepository) GetByBudgetCategory(budgetCategory string) ([]models.COAAccount, error) {
	var accounts []models.COAAccount
	err := r.db.Where("budget_category = ? AND deleted_at IS NULL AND is_active = true", budgetCategory).
		Order("code ASC").Find(&accounts).Error
	return accounts, err
}

func (r *coaRepository) Create(coa *models.COAAccount) error {
	// Check for duplicate code
	var count int64
	r.db.Model(&models.COAAccount{}).Where("code = ? AND deleted_at IS NULL", coa.Code).Count(&count)
	if count > 0 {
		return errors.New("COA code already exists")
	}

	// Calculate level based on parent
	if coa.ParentID != nil {
		var parent models.COAAccount
		if err := r.db.First(&parent, *coa.ParentID).Error; err != nil {
			return errors.New("parent COA not found")
		}
		coa.Level = parent.Level + 1
	} else {
		coa.Level = 1
	}

	return r.db.Create(coa).Error
}

func (r *coaRepository) Update(coa *models.COAAccount) error {
	// Check for duplicate code (excluding current)
	var count int64
	r.db.Model(&models.COAAccount{}).
		Where("code = ? AND id != ? AND deleted_at IS NULL", coa.Code, coa.ID).
		Count(&count)
	if count > 0 {
		return errors.New("COA code already exists")
	}

	return r.db.Save(coa).Error
}

func (r *coaRepository) Delete(id uint) error {
	// Check if has children
	var childCount int64
	r.db.Model(&models.COAAccount{}).Where("parent_id = ? AND deleted_at IS NULL", id).Count(&childCount)
	if childCount > 0 {
		return errors.New("cannot delete COA with children")
	}

	// Check if used in materials
	var materialCount int64
	r.db.Model(&models.Material{}).Where("coa_account_id = ? AND deleted_at IS NULL", id).Count(&materialCount)
	if materialCount > 0 {
		return errors.New("cannot delete COA used by materials")
	}

	// Check if used in CBS nodes
	var cbsCount int64
	r.db.Model(&models.CBSNode{}).Where("coa_account_id = ? AND deleted_at IS NULL", id).Count(&cbsCount)
	if cbsCount > 0 {
		return errors.New("cannot delete COA used by CBS nodes")
	}

	return r.db.Model(&models.COAAccount{}).Where("id = ?", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (r *coaRepository) GetChildren(parentID uint) ([]models.COAAccount, error) {
	var accounts []models.COAAccount
	err := r.db.Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Order("code ASC").Find(&accounts).Error
	return accounts, err
}
