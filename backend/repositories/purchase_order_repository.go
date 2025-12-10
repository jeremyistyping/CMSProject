package repositories

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type PurchaseOrderRepository struct {
	db *gorm.DB
}

func NewPurchaseOrderRepository(db *gorm.DB) *PurchaseOrderRepository {
	return &PurchaseOrderRepository{db: db}
}

// GeneratePOCode generates a unique PO code
func (r *PurchaseOrderRepository) GeneratePOCode() (string, error) {
	var count int64
	today := time.Now().Format("20060102")
	
	r.db.Model(&models.PurchaseOrder{}).
		Where("code LIKE ?", fmt.Sprintf("PO-%s%%", today)).
		Count(&count)
	
	code := fmt.Sprintf("PO-%s%04d", today, count+1)
	return code, nil
}

// Create creates a new purchase order
func (r *PurchaseOrderRepository) Create(po *models.PurchaseOrder) error {
	return r.db.Create(po).Error
}

// GetByID retrieves a purchase order by ID with relations
func (r *PurchaseOrderRepository) GetByID(id uint) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	err := r.db.Preload("PurchaseRequest").
		Preload("Project").
		Preload("Vendor").
		Preload("Creator").
		Preload("Items").
		Preload("Items.Material").
		Preload("GoodsReceipts").
		Preload("GoodsReceipts.Items").
		First(&po, id).Error
	if err != nil {
		return nil, err
	}
	return &po, nil
}

// GetByPRID retrieves purchase orders by purchase request ID
func (r *PurchaseOrderRepository) GetByPRID(prID uint) ([]models.PurchaseOrder, error) {
	var pos []models.PurchaseOrder
	err := r.db.Where("purchase_request_id = ?", prID).
		Preload("Items").
		Preload("Vendor").
		Find(&pos).Error
	return pos, err
}

// GetAll retrieves all purchase orders with optional filters
func (r *PurchaseOrderRepository) GetAll(filter map[string]interface{}) ([]models.PurchaseOrder, error) {
	var pos []models.PurchaseOrder
	query := r.db.Preload("PurchaseRequest").
		Preload("Project").
		Preload("Vendor").
		Preload("Creator").
		Preload("Items")

	if projectID, ok := filter["project_id"]; ok {
		query = query.Where("project_id = ?", projectID)
	}
	if status, ok := filter["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if vendorID, ok := filter["vendor_id"]; ok {
		query = query.Where("vendor_id = ?", vendorID)
	}

	err := query.Order("created_at DESC").Find(&pos).Error
	return pos, err
}

// Update updates a purchase order
func (r *PurchaseOrderRepository) Update(po *models.PurchaseOrder) error {
	return r.db.Save(po).Error
}

// UpdateStatus updates the status of a purchase order
func (r *PurchaseOrderRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.PurchaseOrder{}).Where("id = ?", id).Update("status", status).Error
}

// Delete soft deletes a purchase order
func (r *PurchaseOrderRepository) Delete(id uint) error {
	return r.db.Delete(&models.PurchaseOrder{}, id).Error
}

// CreateItem creates a purchase order item
func (r *PurchaseOrderRepository) CreateItem(item *models.PurchaseOrderItem) error {
	return r.db.Create(item).Error
}

// UpdateItemReceivedQty updates the received quantity of a PO item
func (r *PurchaseOrderRepository) UpdateItemReceivedQty(itemID uint, qty float64) error {
	return r.db.Model(&models.PurchaseOrderItem{}).
		Where("id = ?", itemID).
		Update("received_quantity", gorm.Expr("received_quantity + ?", qty)).Error
}

// CheckAllItemsReceived checks if all items in a PO have been fully received
func (r *PurchaseOrderRepository) CheckAllItemsReceived(poID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.PurchaseOrderItem{}).
		Where("purchase_order_id = ? AND received_quantity < quantity", poID).
		Count(&count).Error
	return count == 0, err
}

// Goods Receipt methods

// GenerateGRCode generates a unique GR code
func (r *PurchaseOrderRepository) GenerateGRCode() (string, error) {
	var count int64
	today := time.Now().Format("20060102")
	
	r.db.Model(&models.GoodsReceipt{}).
		Where("code LIKE ?", fmt.Sprintf("GR-%s%%", today)).
		Count(&count)
	
	code := fmt.Sprintf("GR-%s%04d", today, count+1)
	return code, nil
}

// CreateGoodsReceipt creates a new goods receipt
func (r *PurchaseOrderRepository) CreateGoodsReceipt(gr *models.GoodsReceipt) error {
	return r.db.Create(gr).Error
}

// GetGoodsReceiptByID retrieves a goods receipt by ID
func (r *PurchaseOrderRepository) GetGoodsReceiptByID(id uint) (*models.GoodsReceipt, error) {
	var gr models.GoodsReceipt
	err := r.db.Preload("PurchaseOrder").
		Preload("Project").
		Preload("Receiver").
		Preload("Items").
		Preload("Items.POItem").
		First(&gr, id).Error
	if err != nil {
		return nil, err
	}
	return &gr, nil
}

// GetGoodsReceiptsByPOID retrieves all goods receipts for a PO
func (r *PurchaseOrderRepository) GetGoodsReceiptsByPOID(poID uint) ([]models.GoodsReceipt, error) {
	var grs []models.GoodsReceipt
	err := r.db.Where("purchase_order_id = ?", poID).
		Preload("Items").
		Preload("Receiver").
		Find(&grs).Error
	return grs, err
}

// CreateGoodsReceiptItem creates a goods receipt item
func (r *PurchaseOrderRepository) CreateGoodsReceiptItem(item *models.GoodsReceiptItem) error {
	return r.db.Create(item).Error
}
