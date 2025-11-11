package models

import "time"

// BudgetVsActualReport - Report Budget vs Actual by COA Group
type BudgetVsActualReport struct {
	ReportDate    time.Time                `json:"report_date"`
	ProjectID     *uint                    `json:"project_id,omitempty"`
	ProjectName   string                   `json:"project_name,omitempty"`
	StartDate     time.Time                `json:"start_date"`
	EndDate       time.Time                `json:"end_date"`
	COAGroups     []BudgetVsActualCOAGroup `json:"coa_groups"`
	TotalBudget   float64                  `json:"total_budget"`
	TotalActual   float64                  `json:"total_actual"`
	TotalVariance float64                  `json:"total_variance"`
	VarianceRate  float64                  `json:"variance_rate"` // percentage
}

type BudgetVsActualCOAGroup struct {
	COACode       string  `json:"coa_code"`
	COAName       string  `json:"coa_name"`
	COAType       string  `json:"coa_type"`
	Budget        float64 `json:"budget"`
	Actual        float64 `json:"actual"`
	Variance      float64 `json:"variance"`      // Actual - Budget
	VarianceRate  float64 `json:"variance_rate"` // percentage
	Status        string  `json:"status"`        // OVER_BUDGET, UNDER_BUDGET, ON_TARGET
}

// ProfitabilityReport - Report Profitability per Project
type ProfitabilityReport struct {
	ReportDate        time.Time             `json:"report_date"`
	StartDate         time.Time             `json:"start_date"`
	EndDate           time.Time             `json:"end_date"`
	Projects          []ProjectProfitability `json:"projects"`
	TotalRevenue      float64               `json:"total_revenue"`
	TotalDirectCost   float64               `json:"total_direct_cost"`
	TotalOperational  float64               `json:"total_operational"`
	TotalProfit       float64               `json:"total_profit"`
	OverallMargin     float64               `json:"overall_margin"` // percentage
}

type ProjectProfitability struct {
	ProjectID          uint    `json:"project_id"`
	ProjectCode        string  `json:"project_code"`
	ProjectName        string  `json:"project_name"`
	ProjectStatus      string  `json:"project_status"`
	Revenue            float64 `json:"revenue"`
	DirectCost         float64 `json:"direct_cost"`   // Material, Labour, Equipment
	OperationalCost    float64 `json:"operational_cost"` // Overhead, Admin
	TotalCost          float64 `json:"total_cost"`
	GrossProfit        float64 `json:"gross_profit"`        // Revenue - DirectCost
	NetProfit          float64 `json:"net_profit"`          // Revenue - TotalCost
	GrossProfitMargin  float64 `json:"gross_profit_margin"`  // percentage
	NetProfitMargin    float64 `json:"net_profit_margin"`    // percentage
}

// CashFlowReport - Cash Flow per Project
type CashFlowReport struct {
	ReportDate       time.Time            `json:"report_date"`
	StartDate        time.Time            `json:"start_date"`
	EndDate          time.Time            `json:"end_date"`
	Projects         []ProjectCashFlow    `json:"projects"`
	TotalCashIn      float64              `json:"total_cash_in"`
	TotalCashOut     float64              `json:"total_cash_out"`
	NetCashFlow      float64              `json:"net_cash_flow"`
	BeginningBalance float64              `json:"beginning_balance"`
	EndingBalance    float64              `json:"ending_balance"`
}

type ProjectCashFlow struct {
	ProjectID        uint                `json:"project_id"`
	ProjectCode      string              `json:"project_code"`
	ProjectName      string              `json:"project_name"`
	CashIn           float64             `json:"cash_in"`
	CashOut          float64             `json:"cash_out"`
	NetCashFlow      float64             `json:"net_cash_flow"`
	CashInDetails    []CashFlowDetail    `json:"cash_in_details"`
	CashOutDetails   []CashFlowDetail    `json:"cash_out_details"`
}

type CashFlowDetail struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	COACode     string    `json:"coa_code"`
	COAName     string    `json:"coa_name"`
	Amount      float64   `json:"amount"`
	Source      string    `json:"source"` // SALE, PAYMENT, EXPENSE, etc
}

// CostSummaryReport - Cost Summary by Category
type CostSummaryReport struct {
	ReportDate      time.Time         `json:"report_date"`
	ProjectID       *uint             `json:"project_id,omitempty"`
	ProjectName     string            `json:"project_name,omitempty"`
	StartDate       time.Time         `json:"start_date"`
	EndDate         time.Time         `json:"end_date"`
	Categories      []CostCategory    `json:"categories"`
	TotalCost       float64           `json:"total_cost"`
	LargestCategory string            `json:"largest_category"`
	LargestAmount   float64           `json:"largest_amount"`
}

type CostCategory struct {
	CategoryCode  string         `json:"category_code"`
	CategoryName  string         `json:"category_name"`
	TotalAmount   float64        `json:"total_amount"`
	Percentage    float64        `json:"percentage"` // of total cost
	ItemCount     int            `json:"item_count"`
	Items         []CostItem     `json:"items"`
}

type CostItem struct {
	Date        time.Time `json:"date"`
	COACode     string    `json:"coa_code"`
	COAName     string    `json:"coa_name"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	ProjectCode string    `json:"project_code,omitempty"`
	ProjectName string    `json:"project_name,omitempty"`
}

// Export Parameters
type ProjectReportParams struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	ProjectID *uint     `json:"project_id,omitempty"`
	Format    string    `json:"format"` // json, pdf, csv
}
