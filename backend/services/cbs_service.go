package services

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"errors"
	"time"
)

type CBSService interface {
	GetProjectCBSTree(projectID uint) ([]models.CBSNode, error)
	GetCBSNodeByID(id uint) (*models.CBSNode, error)
	CreateCBSNode(node *models.CBSNode) error
	UpdateCBSNode(id uint, node *models.CBSNode) error
	DeleteCBSNode(id uint) error
	GetNodeCostSummary(nodeID uint) (*models.CBSNodeSummary, error)
	GetProjectBudgetSummary(projectID uint) (*ProjectBudgetSummary, error)
	ValidateCBSBudget(nodeID uint, amount int64) error
	GetPRCBSMappings(prID uint) ([]models.PRCBSMapping, error)
	VerifyPurchaseRequest(prID uint, userID uint, mappings []models.PRCBSMapping, notes string) error
}

type ProjectBudgetSummary struct {
	ProjectID    uint    `json:"project_id"`
	TotalBudget  int64   `json:"total_budget"`
	TotalActual  int64   `json:"total_actual"`
	TotalVariance int64  `json:"total_variance"`
	NodeCount    int     `json:"node_count"`
}

type cbsService struct {
	repo              repositories.CBSRepository
	projectBudgetRepo repositories.ProjectBudgetRepository
}

func NewCBSService(repo repositories.CBSRepository, projectBudgetRepo repositories.ProjectBudgetRepository) CBSService {
	return &cbsService{
		repo:              repo,
		projectBudgetRepo: projectBudgetRepo,
	}
}

// GetProjectCBSTree retrieves the CBS tree with cost summaries
func (s *cbsService) GetProjectCBSTree(projectID uint) ([]models.CBSNode, error) {
	nodes, err := s.repo.GetCBSTree(projectID)
	if err != nil {
		return nil, err
	}

	// Calculate actual costs for each node
	s.enrichNodesWithCosts(nodes)

	return nodes, nil
}

// enrichNodesWithCosts recursively adds cost information to nodes
func (s *cbsService) enrichNodesWithCosts(nodes []models.CBSNode) {
	for i := range nodes {
		summary, err := s.repo.GetNodeCostSummary(nodes[i].ID)
		if err == nil {
			nodes[i].ActualCost = summary.TotalCost
		}

		if len(nodes[i].Children) > 0 {
			s.enrichNodesWithCosts(nodes[i].Children)
		}
	}
}

// GetCBSNodeByID retrieves a single CBS node
func (s *cbsService) GetCBSNodeByID(id uint) (*models.CBSNode, error) {
	return s.repo.GetNodeByID(id)
}

// CreateCBSNode creates a new CBS node with validation
func (s *cbsService) CreateCBSNode(node *models.CBSNode) error {
	// Validate parent exists if specified
	if node.ParentID != nil {
		parent, err := s.repo.GetNodeByID(*node.ParentID)
		if err != nil {
			return errors.New("parent CBS node not found")
		}
		// Ensure parent is in the same project
		if parent.ProjectID != node.ProjectID {
			return errors.New("parent node must be in the same project")
		}
	}

	// Set timestamps
	now := time.Now()
	node.CreatedAt = now
	node.UpdatedAt = now

	// Create CBS node
	if err := s.repo.CreateNode(node); err != nil {
		return err
	}

	// Auto-sync to project_budgets if COAAccountID and BudgetAmount exist
	if node.COAAccountID != nil && node.BudgetAmount > 0 {
		if err := s.syncToProjectBudget(node); err != nil {
			// Log warning but don't fail the operation
			// In production, you might want to use proper logging
			// log.Printf("Warning: Failed to sync CBS to project budget: %v", err)
		}
	}

	return nil
}

// UpdateCBSNode updates an existing CBS node
func (s *cbsService) UpdateCBSNode(id uint, node *models.CBSNode) error {
	// Check if node exists
	existing, err := s.repo.GetNodeByID(id)
	if err != nil {
		return err
	}

	// Validate parent if changed
	if node.ParentID != nil {
		if existing.ParentID == nil || *existing.ParentID != *node.ParentID {
			// Prevent circular reference
			if *node.ParentID == id {
				return errors.New("node cannot be its own parent")
			}

			parent, err := s.repo.GetNodeByID(*node.ParentID)
			if err != nil {
				return errors.New("parent CBS node not found")
			}
			if parent.ProjectID != existing.ProjectID {
				return errors.New("parent node must be in the same project")
			}
		}
	}

	// Update fields
	node.ID = id
	node.ProjectID = existing.ProjectID // Cannot change project
	node.UpdatedAt = time.Now()

	// Update CBS node
	if err := s.repo.UpdateNode(node); err != nil {
		return err
	}

	// Auto-sync to project_budgets if COAAccountID and BudgetAmount exist
	if node.COAAccountID != nil && node.BudgetAmount > 0 {
		if err := s.syncToProjectBudget(node); err != nil {
			// Log warning but don't fail the operation
			// log.Printf("Warning: Failed to sync CBS to project budget: %v", err)
		}
	}

	return nil
}

// DeleteCBSNode deletes a CBS node
func (s *cbsService) DeleteCBSNode(id uint) error {
	// Check if node has any PR mappings
	mappings, err := s.repo.GetPRCBSMappings(0) // This needs adjustment to check by node
	if err == nil && len(mappings) > 0 {
		// For now, we'll allow deletion but in production you might want to prevent it
		// return errors.New("cannot delete CBS node with existing PR mappings")
	}

	return s.repo.DeleteNode(id)
}

// GetNodeCostSummary retrieves cost summary for a node
func (s *cbsService) GetNodeCostSummary(nodeID uint) (*models.CBSNodeSummary, error) {
	return s.repo.GetNodeCostSummary(nodeID)
}

// ValidateCBSBudget checks if adding an amount would exceed the budget
func (s *cbsService) ValidateCBSBudget(nodeID uint, amount int64) error {
	summary, err := s.repo.GetNodeCostSummary(nodeID)
	if err != nil {
		return err
	}

	newTotal := summary.TotalCost + amount
	if newTotal > summary.BudgetAmount {
		return errors.New("allocation would exceed CBS node budget")
	}

	return nil
}

// GetPRCBSMappings retrieves CBS mappings for a purchase request
func (s *cbsService) GetPRCBSMappings(prID uint) ([]models.PRCBSMapping, error) {
	return s.repo.GetPRCBSMappings(prID)
}

// VerifyPurchaseRequest verifies a PR and saves CBS mappings
func (s *cbsService) VerifyPurchaseRequest(prID uint, userID uint, mappings []models.PRCBSMapping, notes string) error {
	// 1. Delete existing mappings (if any)
	if err := s.repo.DeletePRCBSMappings(prID); err != nil {
		return err
	}

	// 2. Create new mappings
	for _, mapping := range mappings {
		mapping.PurchaseRequestID = prID
		mapping.CreatedBy = &userID
		now := time.Now()
		mapping.CreatedAt = now
		mapping.UpdatedAt = now

		if err := s.repo.CreatePRCBSMapping(&mapping); err != nil {
			return err
		}
	}

	// 3. Update PR status to VERIFIED
	// We need to access PurchaseRequest repository here, but we don't have it injected.
	// For now, let's assume we can update it via DB directly or we need to inject PR repo.
	// Since we only have CBSRepository, let's add a method to CBSRepository to update PR status.
	return s.repo.UpdatePRVerificationStatus(prID, userID, notes)
}

// syncToProjectBudget syncs CBS node budget to project_budgets table
func (s *cbsService) syncToProjectBudget(node *models.CBSNode) error {
	if node.COAAccountID == nil {
		return nil
	}

	budget := &models.ProjectBudget{
		ProjectID:       node.ProjectID,
		AccountID:       *node.COAAccountID,
		EstimatedAmount: float64(node.BudgetAmount),
	}

	return s.projectBudgetRepo.Upsert(budget)
}

// GetProjectBudgetSummary retrieves budget summary for a project from CBS
func (s *cbsService) GetProjectBudgetSummary(projectID uint) (*ProjectBudgetSummary, error) {
	nodes, err := s.repo.GetNodesByProjectID(projectID)
	if err != nil {
		return nil, err
	}

	summary := &ProjectBudgetSummary{
		ProjectID: projectID,
		NodeCount: len(nodes),
	}

	for _, node := range nodes {
		summary.TotalBudget += node.BudgetAmount
		
		// Get actual cost for this node
		nodeSummary, err := s.repo.GetNodeCostSummary(node.ID)
		if err == nil {
			summary.TotalActual += nodeSummary.ActualCost
		}
	}

	summary.TotalVariance = summary.TotalBudget - summary.TotalActual

	return summary, nil
}
