package models

// StockValidationItem represents the check result for a single line item in the sales form
type StockValidationItem struct {
	ProductID     uint   `json:"product_id"`
	ProductCode   string `json:"product_code"`
	ProductName   string `json:"product_name"`
	RequestedQty  int    `json:"requested_qty"`
	AvailableQty  int    `json:"available_qty"`
	MinStock      int    `json:"min_stock"`
	ReorderLevel  int    `json:"reorder_level"`
	IsService     bool   `json:"is_service"`
	IsSufficient  bool   `json:"is_sufficient"`
	LowStock      bool   `json:"low_stock"`
	AtOrBelowMin  bool   `json:"at_or_below_min"`
	AtOrBelowReorder bool `json:"at_or_below_reorder"`
	IsZeroStock   bool   `json:"is_zero_stock"`
	Warning       string `json:"warning"`
}

// StockValidationRequest is the payload from sales create form to validate stock
// It reuses the SaleCreateRequest items shape but only requires items
type StockValidationRequest struct {
	Items []SaleItemRequest `json:"items"`
}

// StockValidationResponse aggregates per-item results and overall flags
type StockValidationResponse struct {
	Items              []StockValidationItem `json:"items"`
	HasInsufficient    bool                  `json:"has_insufficient"`
	HasLowStock        bool                  `json:"has_low_stock"`
	HasMinStockAlerts  bool                  `json:"has_min_stock_alerts"`
	HasReorderAlerts   bool                  `json:"has_reorder_alerts"`
	HasZeroStock       bool                  `json:"has_zero_stock"`
}
