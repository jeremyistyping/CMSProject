package repositories

import (
	"app-sistem-akuntansi/models"
	"errors"

	"gorm.io/gorm"
)

type CBSRepository interface {
	GetCBSTree(projectID uint) ([]models.CBSNode, error)
	GetNodeByID(id uint) (*models.CBSNode, error)
	GetNodesByProjectID(projectID uint) ([]models.CBSNode, error)
	CreateNode(node *models.CBSNode) error
	UpdateNode(node *models.CBSNode) error
	DeleteNode(id uint) error
	GetNodeCostSummary(nodeID uint) (*models.CBSNodeSummary, error)
	GetPRCBSMappings(prID uint) ([]models.PRCBSMapping, error)
	CreatePRCBSMapping(mapping *models.PRCBSMapping) error
	DeletePRCBSMappings(prID uint) error
}

type cbsRepository struct {
	db *gorm.DB
}

func NewCBSRepository(db *gorm.DB) CBSRepository {
	return &cbsRepository{db: db}
}

// GetCBSTree retrieves the CBS tree structure for a project
func (r *cbsRepository) GetCBSTree(projectID uint) ([]models.CBSNode, error) {
	var nodes []models.CBSNode

	// Get all nodes for the project, ordered by code
	err := r.db.Where("project_id = ? AND deleted_at IS NULL", projectID).
		Order("code ASC").
		Preload("COAAccount").
		Find(&nodes).Error

	if err != nil {
		return nil, err
	}

	// Build tree structure
	nodeMap := make(map[uint]*models.CBSNode)
	for i := range nodes {
		nodeMap[nodes[i].ID] = &nodes[i]
		nodes[i].Children = []models.CBSNode{}
	}

	var rootNodes []models.CBSNode
	for i := range nodes {
		if nodes[i].ParentID == nil {
			rootNodes = append(rootNodes, nodes[i])
		} else {
			if parent, ok := nodeMap[*nodes[i].ParentID]; ok {
				parent.Children = append(parent.Children, nodes[i])
			}
		}
	}

	return rootNodes, nil
}

// GetNodeByID retrieves a single CBS node by ID
func (r *cbsRepository) GetNodeByID(id uint) (*models.CBSNode, error) {
	var node models.CBSNode
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).
		Preload("Parent").
		Preload("COAAccount").
		First(&node).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("CBS node not found")
		}
		return nil, err
	}
	return &node, nil
}

// GetNodesByProjectID retrieves all nodes for a project (flat list)
func (r *cbsRepository) GetNodesByProjectID(projectID uint) ([]models.CBSNode, error) {
	var nodes []models.CBSNode
	err := r.db.Where("project_id = ? AND deleted_at IS NULL", projectID).
		Order("code ASC").
		Preload("COAAccount").
		Find(&nodes).Error
	return nodes, err
}

// CreateNode creates a new CBS node
func (r *cbsRepository) CreateNode(node *models.CBSNode) error {
	// Check for duplicate code within project
	var count int64
	r.db.Model(&models.CBSNode{}).
		Where("project_id = ? AND code = ? AND deleted_at IS NULL", node.ProjectID, node.Code).
		Count(&count)

	if count > 0 {
		return errors.New("CBS code already exists in this project")
	}

	return r.db.Create(node).Error
}

// UpdateNode updates an existing CBS node
func (r *cbsRepository) UpdateNode(node *models.CBSNode) error {
	// Check for duplicate code (excluding current node)
	var count int64
	r.db.Model(&models.CBSNode{}).
		Where("project_id = ? AND code = ? AND id != ? AND deleted_at IS NULL",
			node.ProjectID, node.Code, node.ID).
		Count(&count)

	if count > 0 {
		return errors.New("CBS code already exists in this project")
	}

	return r.db.Save(node).Error
}

// DeleteNode soft deletes a CBS node
func (r *cbsRepository) DeleteNode(id uint) error {
	// Check if node has children
	var childCount int64
	r.db.Model(&models.CBSNode{}).
		Where("parent_id = ? AND deleted_at IS NULL", id).
		Count(&childCount)

	if childCount > 0 {
		return errors.New("cannot delete CBS node with children")
	}

	// Soft delete
	return r.db.Model(&models.CBSNode{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

// GetNodeCostSummary calculates cost summary for a node
func (r *cbsRepository) GetNodeCostSummary(nodeID uint) (*models.CBSNodeSummary, error) {
	var node models.CBSNode
	if err := r.db.First(&node, nodeID).Error; err != nil {
		return nil, err
	}

	summary := &models.CBSNodeSummary{
		NodeID:       nodeID,
		BudgetAmount: node.BudgetAmount,
	}

	// Calculate actual cost from PR CBS mappings
	r.db.Model(&models.PRCBSMapping{}).
		Select("COALESCE(SUM(allocated_amount), 0)").
		Where("cbs_node_id = ?", nodeID).
		Scan(&summary.ActualCost)

	// Calculate children cost (recursive)
	var childIDs []uint
	r.db.Model(&models.CBSNode{}).
		Where("parent_id = ? AND deleted_at IS NULL", nodeID).
		Pluck("id", &childIDs)

	for _, childID := range childIDs {
		childSummary, err := r.GetNodeCostSummary(childID)
		if err == nil {
			summary.ChildrenCost += childSummary.TotalCost
		}
	}

	summary.TotalCost = summary.ActualCost + summary.ChildrenCost
	summary.Variance = summary.BudgetAmount - summary.TotalCost

	return summary, nil
}

// GetPRCBSMappings retrieves all CBS mappings for a PR
func (r *cbsRepository) GetPRCBSMappings(prID uint) ([]models.PRCBSMapping, error) {
	var mappings []models.PRCBSMapping
	err := r.db.Where("purchase_request_id = ?", prID).
		Preload("CBSNode").
		Preload("CBSNode.COAAccount").
		Preload("PRItem").
		Find(&mappings).Error
	return mappings, err
}

// CreatePRCBSMapping creates a new PR CBS mapping
func (r *cbsRepository) CreatePRCBSMapping(mapping *models.PRCBSMapping) error {
	return r.db.Create(mapping).Error
}

// DeletePRCBSMappings deletes all CBS mappings for a PR
func (r *cbsRepository) DeletePRCBSMappings(prID uint) error {
	return r.db.Where("purchase_request_id = ?", prID).
		Delete(&models.PRCBSMapping{}).Error
}
