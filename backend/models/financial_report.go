package models

import (
	"time"
)

// Financial Report Types
const (
	ReportTypeProfitLoss    = "PROFIT_LOSS"
	ReportTypeBalanceSheet  = "BALANCE_SHEET"
	ReportTypeCashFlow      = "CASH_FLOW"
	ReportTypeTrialBalance  = "TRIAL_BALANCE"
	ReportTypeGeneralLedger = "GENERAL_LEDGER"
)

// Report Periods
const (
	PeriodDaily     = "DAILY"
	PeriodWeekly    = "WEEKLY"
	PeriodMonthly   = "MONTHLY"
	PeriodQuarterly = "QUARTERLY"
	PeriodYearly    = "YEARLY"
	PeriodCustom    = "CUSTOM"
)

// Financial Report Request
type FinancialReportRequest struct {
	ReportType  string    `json:"report_type" binding:"required"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
	Period      string    `json:"period"`
	Comparative bool      `json:"comparative"`          // Compare with previous period
	ShowZero    bool      `json:"show_zero"`            // Show accounts with zero balance
	Format      string    `json:"format"`               // JSON, PDF, EXCEL
}

// Profit & Loss Statement
type ProfitLossStatement struct {
	ReportHeader  ReportHeader           `json:"report_header"`
	Revenue       []AccountLineItem      `json:"revenue"`
	TotalRevenue  float64                `json:"total_revenue"`
	COGS          []AccountLineItem      `json:"cost_of_goods_sold"`
	TotalCOGS     float64                `json:"total_cogs"`
	GrossProfit   float64                `json:"gross_profit"`
	Expenses      []AccountLineItem      `json:"expenses"`
	TotalExpenses float64                `json:"total_expenses"`
	NetIncome     float64                `json:"net_income"`
	Comparative   *ProfitLossComparative `json:"comparative,omitempty"`
}

type ProfitLossComparative struct {
	PreviousPeriod ProfitLossStatement `json:"previous_period"`
	Variance       ProfitLossVariance  `json:"variance"`
}

type ProfitLossVariance struct {
	RevenueVariance  float64 `json:"revenue_variance"`
	COGSVariance     float64 `json:"cogs_variance"`
	GrossProfitVariance float64 `json:"gross_profit_variance"`
	ExpenseVariance  float64 `json:"expense_variance"`
	NetIncomeVariance float64 `json:"net_income_variance"`
}

// Balance Sheet
type BalanceSheet struct {
	ReportHeader        ReportHeader             `json:"report_header"`
	Assets              BalanceSheetSection      `json:"assets"`
	Liabilities         BalanceSheetSection      `json:"liabilities"`
	Equity              BalanceSheetSection      `json:"equity"`
	TotalAssets         float64                  `json:"total_assets"`
	TotalLiabilities    float64                  `json:"total_liabilities"`
	TotalEquity         float64                  `json:"total_equity"`
	IsBalanced          bool                     `json:"is_balanced"`
	Comparative         *BalanceSheetComparative `json:"comparative,omitempty"`
}

type BalanceSheetSection struct {
	Categories []BalanceSheetCategory `json:"categories"`
	Total      float64                `json:"total"`
}

type BalanceSheetCategory struct {
	Name     string            `json:"name"`
	Accounts []AccountLineItem `json:"accounts"`
	Total    float64           `json:"total"`
}

type BalanceSheetComparative struct {
	PreviousPeriod BalanceSheet        `json:"previous_period"`
	Variance       BalanceSheetVariance `json:"variance"`
}

type BalanceSheetVariance struct {
	AssetsVariance      float64 `json:"assets_variance"`
	LiabilitiesVariance float64 `json:"liabilities_variance"`
	EquityVariance      float64 `json:"equity_variance"`
}

// Cash Flow Statement
type CashFlowStatement struct {
	ReportHeader        ReportHeader               `json:"report_header"`
	OperatingActivities CashFlowSection            `json:"operating_activities"`
	InvestingActivities CashFlowSection            `json:"investing_activities"`
	FinancingActivities CashFlowSection            `json:"financing_activities"`
	NetCashFlow         float64                    `json:"net_cash_flow"`
	BeginningCash       float64                    `json:"beginning_cash"`
	EndingCash          float64                    `json:"ending_cash"`
	Comparative         *CashFlowComparative       `json:"comparative,omitempty"`
}

type CashFlowSection struct {
	Items []CashFlowItem `json:"items"`
	Total float64        `json:"total"`
}

type CashFlowItem struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name"`
}

type CashFlowComparative struct {
	PreviousPeriod CashFlowStatement `json:"previous_period"`
	Variance       CashFlowVariance  `json:"variance"`
}

type CashFlowVariance struct {
	OperatingVariance float64 `json:"operating_variance"`
	InvestingVariance float64 `json:"investing_variance"`
	FinancingVariance float64 `json:"financing_variance"`
	NetCashVariance   float64 `json:"net_cash_variance"`
}

// Trial Balance
type TrialBalance struct {
	ReportHeader ReportHeader        `json:"report_header"`
	Accounts     []TrialBalanceItem  `json:"accounts"`
	TotalDebits  float64             `json:"total_debits"`
	TotalCredits float64             `json:"total_credits"`
	IsBalanced   bool                `json:"is_balanced"`
}

type TrialBalanceItem struct {
	AccountID     uint    `json:"account_id"`
	AccountCode   string  `json:"account_code"`
	AccountName   string  `json:"account_name"`
	AccountType   string  `json:"account_type"`
	DebitBalance  float64 `json:"debit_balance"`
	CreditBalance float64 `json:"credit_balance"`
}

// General Ledger
type GeneralLedger struct {
	ReportHeader ReportHeader         `json:"report_header"`
	Account      Account              `json:"account"`
	Transactions []GeneralLedgerEntry `json:"transactions"`
	BeginningBalance float64          `json:"beginning_balance"`
	EndingBalance    float64          `json:"ending_balance"`
	TotalDebits      float64          `json:"total_debits"`
	TotalCredits     float64          `json:"total_credits"`
}

type GeneralLedgerEntry struct {
	Date          time.Time `json:"date"`
	JournalCode   string    `json:"journal_code"`
	Description   string    `json:"description"`
	Reference     string    `json:"reference"`
	DebitAmount   float64   `json:"debit_amount"`
	CreditAmount  float64   `json:"credit_amount"`
	Balance       float64   `json:"balance"`
}

// Common structures
type ReportHeader struct {
	ReportType    string    `json:"report_type"`
	CompanyName   string    `json:"company_name"`
	ReportTitle   string    `json:"report_title"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	GeneratedAt   time.Time `json:"generated_at"`
	GeneratedBy   string    `json:"generated_by"`
	Currency      string    `json:"currency"`
	IsComparative bool      `json:"is_comparative"`
}

type AccountLineItem struct {
	AccountID     uint    `json:"account_id"`
	AccountCode   string  `json:"account_code"`
	AccountName   string  `json:"account_name"`
	AccountType   string  `json:"account_type"`
	Category      string  `json:"category"`
	Balance       float64 `json:"balance"`
	PreviousBalance float64 `json:"previous_balance,omitempty"`
	Variance      float64 `json:"variance,omitempty"`
	VariancePercent float64 `json:"variance_percent,omitempty"`
}

// Financial Ratios
// Report Metadata for API documentation
type ReportMetadata struct {
	ReportType          string   `json:"report_type"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	SupportsComparative bool     `json:"supports_comparative"`
	RequiredParams      []string `json:"required_params"`
	OptionalParams      []string `json:"optional_params"`
}

// Export Request
type ReportExportRequest struct {
	ReportType string      `json:"report_type"`
	ReportData interface{} `json:"report_data"`
	Format     string      `json:"format"` // PDF, EXCEL, CSV
	Filename   string      `json:"filename,omitempty"`
}

// Report Generation Summary
type ReportGenerationSummary struct {
	TotalReportsGenerated int                    `json:"total_reports_generated"`
	RecentReports         []ReportGenerationLog `json:"recent_reports"`
	PopularReports        []ReportPopularity    `json:"popular_reports"`
	LastGeneratedAt       time.Time             `json:"last_generated_at"`
}

type ReportGenerationLog struct {
	ReportType  string    `json:"report_type"`
	GeneratedAt time.Time `json:"generated_at"`
	GeneratedBy string    `json:"generated_by"`
	Format      string    `json:"format"`
	Status      string    `json:"status"`
}

type ReportPopularity struct {
	ReportType      string `json:"report_type"`
	GenerationCount int    `json:"generation_count"`
}

// Quick Financial Stats
type QuickFinancialStats struct {
	CashBalance         float64   `json:"cash_balance"`
	TodayRevenue        float64   `json:"today_revenue"`
	TodayExpenses       float64   `json:"today_expenses"`
	MonthToDateRevenue  float64   `json:"month_to_date_revenue"`
	MonthToDateExpenses float64   `json:"month_to_date_expenses"`
	YearToDateRevenue   float64   `json:"year_to_date_revenue"`
	YearToDateExpenses  float64   `json:"year_to_date_expenses"`
	PendingReceivables  float64   `json:"pending_receivables"`
	PendingPayables     float64   `json:"pending_payables"`
	InventoryValue      float64   `json:"inventory_value"`
	LastUpdated         time.Time `json:"last_updated"`
}

// Financial Ratios
type FinancialRatios struct {
	// Liquidity Ratios
	CurrentRatio      float64 `json:"current_ratio"`
	QuickRatio        float64 `json:"quick_ratio"`
	CashRatio         float64 `json:"cash_ratio"`
	
	// Profitability Ratios
	GrossProfitMargin float64 `json:"gross_profit_margin"`
	NetProfitMargin   float64 `json:"net_profit_margin"`
	ROA               float64 `json:"return_on_assets"`
	ROE               float64 `json:"return_on_equity"`
	
	// Leverage Ratios
	DebtRatio         float64 `json:"debt_ratio"`
	DebtToEquityRatio float64 `json:"debt_to_equity_ratio"`
	
	// Efficiency Ratios
	AssetTurnover     float64 `json:"asset_turnover"`
	InventoryTurnover float64 `json:"inventory_turnover"`
	
	// Report Info
	CalculatedAt      time.Time `json:"calculated_at"`
	PeriodStart       time.Time `json:"period_start"`
	PeriodEnd         time.Time `json:"period_end"`
}

// Financial Dashboard Summary
type FinancialDashboard struct {
	ReportDate      time.Time                `json:"report_date"`
	KeyMetrics      FinancialKeyMetrics      `json:"key_metrics"`
	Ratios          FinancialRatios          `json:"ratios"`
	CashPosition    CashPositionSummary      `json:"cash_position"`
	AccountBalances []AccountBalanceSummary  `json:"account_balances"`
	RecentActivity  []RecentActivityItem     `json:"recent_activity"`
	Alerts          []FinancialAlert         `json:"alerts"`
}

type FinancialKeyMetrics struct {
	TotalRevenue      float64 `json:"total_revenue"`
	TotalExpenses     float64 `json:"total_expenses"`
	NetIncome         float64 `json:"net_income"`
	TotalAssets       float64 `json:"total_assets"`
	TotalLiabilities  float64 `json:"total_liabilities"`
	TotalEquity       float64 `json:"total_equity"`
	CashBalance       float64 `json:"cash_balance"`
	AccountsReceivable float64 `json:"accounts_receivable"`
	AccountsPayable   float64 `json:"accounts_payable"`
	Inventory         float64 `json:"inventory"`
}

type CashPositionSummary struct {
	TotalCash     float64          `json:"total_cash"`
	CashAccounts  []CashAccount    `json:"cash_accounts"`
	BankAccounts  []BankAccount    `json:"bank_accounts"`
	CashFlow30Day float64          `json:"cash_flow_30_day"`
}

type CashAccount struct {
	AccountID   uint    `json:"account_id"`
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name"`
	Balance     float64 `json:"balance"`
}

type BankAccount struct {
	AccountID     uint    `json:"account_id"`
	AccountCode   string  `json:"account_code"`
	AccountName   string  `json:"account_name"`
	BankName      string  `json:"bank_name"`
	AccountNumber string  `json:"account_number"`
	Balance       float64 `json:"balance"`
}

type AccountBalanceSummary struct {
	AccountType string  `json:"account_type"`
	Balance     float64 `json:"balance"`
	Count       int64   `json:"count"`
	Percentage  float64 `json:"percentage"`
}

type RecentActivityItem struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"`
	Reference   string    `json:"reference"`
}

type FinancialAlert struct {
	Type        string    `json:"type"`        // LOW_CASH, HIGH_DEBT, NEGATIVE_EQUITY, etc.
	Severity    string    `json:"severity"`    // LOW, MEDIUM, HIGH, CRITICAL
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	CreatedAt   time.Time `json:"created_at"`
}

// Report Generation Status
type ReportGenerationJob struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	JobID       string    `json:"job_id" gorm:"unique;not null;size:50"`
	ReportType  string    `json:"report_type" gorm:"not null;size:50"`
	Parameters  string    `json:"parameters" gorm:"type:text"`        // JSON parameters
	Status      string    `json:"status" gorm:"size:20;default:'PENDING'"` // PENDING, PROCESSING, COMPLETED, FAILED
	Progress    int       `json:"progress" gorm:"default:0"`          // 0-100
	Result      string    `json:"result" gorm:"type:text"`            // JSON result or error message
	FileURL     string    `json:"file_url" gorm:"size:500"`           // URL to generated file
	UserID      uint      `json:"user_id" gorm:"not null;index"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// Report Generation Status Constants
const (
	ReportJobStatusPending    = "PENDING"
	ReportJobStatusProcessing = "PROCESSING"
	ReportJobStatusCompleted  = "COMPLETED"
	ReportJobStatusFailed     = "FAILED"
)

// Advanced Report Filters
type AdvancedReportFilter struct {
	ReportType      string    `json:"report_type"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	AccountTypes    []string  `json:"account_types"`
	AccountIDs      []uint    `json:"account_ids"`
	ExcludeInactive bool      `json:"exclude_inactive"`
	Consolidated    bool      `json:"consolidated"`    // Group by categories
	ShowDetails     bool      `json:"show_details"`    // Show individual transactions
	Currency        string    `json:"currency"`
	ExchangeRate    float64   `json:"exchange_rate"`
}

// Account Movement Analysis
type AccountMovementAnalysis struct {
	Account        Account                    `json:"account"`
	Period         string                     `json:"period"`
	StartDate      time.Time                  `json:"start_date"`
	EndDate        time.Time                  `json:"end_date"`
	OpeningBalance float64                    `json:"opening_balance"`
	ClosingBalance float64                    `json:"closing_balance"`
	TotalDebits    float64                    `json:"total_debits"`
	TotalCredits   float64                    `json:"total_credits"`
	NetMovement    float64                    `json:"net_movement"`
	Movements      []AccountMovementDetail    `json:"movements"`
	Statistics     AccountMovementStatistics  `json:"statistics"`
}

type AccountMovementDetail struct {
	Date            time.Time `json:"date"`
	JournalCode     string    `json:"journal_code"`
	Description     string    `json:"description"`
	Reference       string    `json:"reference"`
	ReferenceType   string    `json:"reference_type"`
	DebitAmount     float64   `json:"debit_amount"`
	CreditAmount    float64   `json:"credit_amount"`
	RunningBalance  float64   `json:"running_balance"`
}

type AccountMovementStatistics struct {
	TransactionCount    int64   `json:"transaction_count"`
	AverageTransaction  float64 `json:"average_transaction"`
	MaxTransaction      float64 `json:"max_transaction"`
	MinTransaction      float64 `json:"min_transaction"`
	DebitTransactions   int64   `json:"debit_transactions"`
	CreditTransactions  int64   `json:"credit_transactions"`
}

// Financial Trend Analysis
type FinancialTrendAnalysis struct {
	ReportType    string                      `json:"report_type"`
	Period        string                      `json:"period"`
	StartDate     time.Time                   `json:"start_date"`
	EndDate       time.Time                   `json:"end_date"`
	TrendData     []FinancialTrendDataPoint   `json:"trend_data"`
	Summary       FinancialTrendSummary       `json:"summary"`
	Forecasting   *FinancialForecast          `json:"forecasting,omitempty"`
}

type FinancialTrendDataPoint struct {
	Period        string  `json:"period"`
	Date          time.Time `json:"date"`
	Revenue       float64 `json:"revenue"`
	Expenses      float64 `json:"expenses"`
	NetIncome     float64 `json:"net_income"`
	Assets        float64 `json:"assets"`
	Liabilities   float64 `json:"liabilities"`
	Equity        float64 `json:"equity"`
	CashFlow      float64 `json:"cash_flow"`
}

type FinancialTrendSummary struct {
	RevenueGrowth    float64 `json:"revenue_growth"`
	ExpenseGrowth    float64 `json:"expense_growth"`
	NetIncomeGrowth  float64 `json:"net_income_growth"`
	AssetGrowth      float64 `json:"asset_growth"`
	EquityGrowth     float64 `json:"equity_growth"`
	Volatility       float64 `json:"volatility"`
}

type FinancialForecast struct {
	NextPeriod      FinancialTrendDataPoint `json:"next_period"`
	Confidence      float64                 `json:"confidence"`
	ForecastMethod  string                  `json:"forecast_method"`
	Assumptions     []string                `json:"assumptions"`
}

// Real-time Financial Metrics (for dashboards)
type RealTimeFinancialMetrics struct {
	AsOfDate         time.Time `json:"as_of_date"`
	CashPosition     float64   `json:"cash_position"`
	DailyRevenue     float64   `json:"daily_revenue"`
	DailyExpenses    float64   `json:"daily_expenses"`
	DailyNetIncome   float64   `json:"daily_net_income"`
	MonthlyRevenue   float64   `json:"monthly_revenue"`
	MonthlyExpenses  float64   `json:"monthly_expenses"`
	MonthlyNetIncome float64   `json:"monthly_net_income"`
	YearlyRevenue    float64   `json:"yearly_revenue"`
	YearlyExpenses   float64   `json:"yearly_expenses"`
	YearlyNetIncome  float64   `json:"yearly_net_income"`
	PendingReceivables float64 `json:"pending_receivables"`
	PendingPayables    float64 `json:"pending_payables"`
	InventoryValue     float64 `json:"inventory_value"`
	LastUpdated        time.Time `json:"last_updated"`
}

// Budget vs Actual Analysis
type BudgetAnalysis struct {
	ReportHeader ReportHeader         `json:"report_header"`
	Categories   []BudgetCategory     `json:"categories"`
	TotalBudget  float64              `json:"total_budget"`
	TotalActual  float64              `json:"total_actual"`
	TotalVariance float64             `json:"total_variance"`
	VariancePercent float64           `json:"variance_percent"`
}

type BudgetCategory struct {
	CategoryName     string              `json:"category_name"`
	CategoryType     string              `json:"category_type"`
	BudgetedAmount   float64             `json:"budgeted_amount"`
	ActualAmount     float64             `json:"actual_amount"`
	Variance         float64             `json:"variance"`
	VariancePercent  float64             `json:"variance_percent"`
	Accounts         []BudgetAccountItem `json:"accounts"`
}

type BudgetAccountItem struct {
	AccountLineItem
	BudgetedAmount  float64 `json:"budgeted_amount"`
	Variance        float64 `json:"variance"`
	VariancePercent float64 `json:"variance_percent"`
}

// Financial Health Score
type FinancialHealthScore struct {
	OverallScore    float64                    `json:"overall_score"`     // 0-100
	ScoreGrade      string                     `json:"score_grade"`       // A+, A, B+, B, C+, C, D, F
	Components      FinancialHealthComponents  `json:"components"`
	Recommendations []HealthRecommendation     `json:"recommendations"`
	CalculatedAt    time.Time                  `json:"calculated_at"`
}

type FinancialHealthComponents struct {
	LiquidityScore      float64 `json:"liquidity_score"`
	ProfitabilityScore  float64 `json:"profitability_score"`
	LeverageScore       float64 `json:"leverage_score"`
	EfficiencyScore     float64 `json:"efficiency_score"`
	GrowthScore         float64 `json:"growth_score"`
}

type HealthRecommendation struct {
	Category    string `json:"category"`
	Priority    string `json:"priority"`     // HIGH, MEDIUM, LOW
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action"`
}
