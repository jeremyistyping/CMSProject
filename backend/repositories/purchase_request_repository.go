package repositories

import (
	"app-sistem-akuntansi/models"

	"gorm.io/gorm"
)

type PurchaseRequestRepository interface {
	Create(pr *models.PurchaseRequest) error
	Update(pr *models.PurchaseRequest) error
	Delete(id uint) error
	FindByID(id uint) (*models.PurchaseRequest, error)
	FindAll(filter map[string]interface{}) ([]models.PurchaseRequest, error)
	GetByID(id uint) (*models.PurchaseRequest, error)
	UpdateStatus(id uint, status string) error
}

type purchaseRequestRepository struct {
	db *gorm.DB
}

func NewPurchaseRequestRepository(db *gorm.DB) PurchaseRequestRepository {
	return &purchaseRequestRepository{db}
}

func (r *purchaseRequestRepository) Create(pr *models.PurchaseRequest) error {
	return r.db.Create(pr).Error
}

func (r *purchaseRequestRepository) Update(pr *models.PurchaseRequest) error {
	return r.db.Save(pr).Error
}

func (r *purchaseRequestRepository) Delete(id uint) error {
	return r.db.Delete(&models.PurchaseRequest{}, id).Error
}

func (r *purchaseRequestRepository) FindByID(id uint) (*models.PurchaseRequest, error) {
	var pr models.PurchaseRequest
	err := r.db.Preload("Items").Preload("Project").Preload("Vendor").Preload("Requester").Preload("Approver").First(&pr, id).Error
	return &pr, err
}

func (r *purchaseRequestRepository) FindAll(filter map[string]interface{}) ([]models.PurchaseRequest, error) {
	var prs []models.PurchaseRequest
	query := r.db.Preload("Items").Preload("Project").Preload("Vendor").Preload("Requester").Preload("Approver")

	if projectID, ok := filter["project_id"]; ok {
		query = query.Where("project_id = ?", projectID)
	}
	if status, ok := filter["status"]; ok {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at desc").Find(&prs).Error
	return prs, err
}

// GetByID is an alias for FindByID for compatibility
func (r *purchaseRequestRepository) GetByID(id uint) (*models.PurchaseRequest, error) {
	return r.FindByID(id)
}

// UpdateStatus updates the status of a purchase request
func (r *purchaseRequestRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.PurchaseRequest{}).Where("id = ?", id).Update("status", status).Error
}
