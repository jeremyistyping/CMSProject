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
	ValidateCBSBudget(nodeID uint, amount int64) error
}

type cbsService struct {
	repo repositories.CBSRepository
}

func NewCBSService(repo repositories.CBSRepository) CBSService {
	return &cbsService{repo: repo}
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

	return s.repo.CreateNode(node)
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

	return s.repo.UpdateNode(node)
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
