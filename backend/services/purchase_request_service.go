package services

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"fmt"
	"time"
)

type PurchaseRequestService interface {
	CreatePR(pr *models.PurchaseRequest) error
	UpdatePR(id uint, pr *models.PurchaseRequest) error
	DeletePR(id uint) error
	GetPRByID(id uint) (*models.PurchaseRequest, error)
	GetAllPRs(filter map[string]interface{}) ([]models.PurchaseRequest, error)
	UpdateStatus(id uint, status string, approverID uint, reason string) error
}

type purchaseRequestService struct {
	repo repositories.PurchaseRequestRepository
}

func NewPurchaseRequestService(repo repositories.PurchaseRequestRepository) PurchaseRequestService {
	return &purchaseRequestService{repo}
}

func (s *purchaseRequestService) CreatePR(pr *models.PurchaseRequest) error {
	// Generate Code
	pr.Code = fmt.Sprintf("PR-%d-%s", pr.ProjectID, time.Now().Format("20060102150405"))
	pr.Status = models.PRStatusPending

	// Calculate Total Amount
	var total float64
	for _, item := range pr.Items {
		total += item.TotalPrice
	}
	pr.TotalAmount = total

	return s.repo.Create(pr)
}

func (s *purchaseRequestService) UpdatePR(id uint, pr *models.PurchaseRequest) error {
	existingPR, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if existingPR.Status != models.PRStatusPending && existingPR.Status != models.PRStatusRevision {
		return fmt.Errorf("cannot update PR with status %s", existingPR.Status)
	}

	// Update fields
	existingPR.ProjectID = pr.ProjectID
	existingPR.RequestDate = pr.RequestDate
	existingPR.RequiredDate = pr.RequiredDate
	existingPR.VendorID = pr.VendorID
	existingPR.Notes = pr.Notes
	existingPR.Items = pr.Items

	// Recalculate Total
	var total float64
	for _, item := range pr.Items {
		total += item.TotalPrice
	}
	existingPR.TotalAmount = total

	return s.repo.Update(existingPR)
}

func (s *purchaseRequestService) DeletePR(id uint) error {
	existingPR, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if existingPR.Status != models.PRStatusPending {
		return fmt.Errorf("cannot delete PR with status %s", existingPR.Status)
	}

	return s.repo.Delete(id)
}

func (s *purchaseRequestService) GetPRByID(id uint) (*models.PurchaseRequest, error) {
	return s.repo.FindByID(id)
}

func (s *purchaseRequestService) GetAllPRs(filter map[string]interface{}) ([]models.PurchaseRequest, error) {
	return s.repo.FindAll(filter)
}

func (s *purchaseRequestService) UpdateStatus(id uint, status string, approverID uint, reason string) error {
	pr, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	pr.Status = status
	if status == models.PRStatusApproved {
		pr.ApprovedBy = &approverID
		now := time.Now()
		pr.ApprovedAt = &now
	} else if status == models.PRStatusRejected || status == models.PRStatusRevision {
		pr.RejectionReason = reason
	}

	return s.repo.Update(pr)
}
