package services

import (
	"fmt"
	"strings"
	"time"
	"gorm.io/gorm"
)

// SSOTCashFlowService generates Cash Flow reports from SSOT Journal System
type SSOTCashFlowService struct {
	db *gorm.DB
}

// NewSSOTCashFlowService creates a new SSOT Cash Flow service
func NewSSOTCashFlowService(db *gorm.DB) *SSOTCashFlowService {
	return &SSOTCashFlowService{
		db: db,
	}
}

// SSOTCashFlowData represents the comprehensive Cash Flow statement structure for SSOT
type SSOTCashFlowData struct {
	Company     CompanyInfo `json:"company"`
	StartDate   time.Time   `json:"start_date"`
	EndDate     time.Time   `json:"end_date"`
	Currency    string      `json:"currency"`

	// Operating Activities
	OperatingActivities struct {
		NetIncome                   float64                 `json:"net_income"`
		Adjustments                 struct {
			Depreciation            float64                 `json:"depreciation"`
			Amortization           float64                 `json:"amortization"`
			BadDebtExpense         float64                 `json:"bad_debt_expense"`
			GainLossOnAssetDisposal float64               `json:"gain_loss_on_asset_disposal"`
			OtherNonCashItems      float64                 `json:"other_non_cash_items"`
			TotalAdjustments       float64                 `json:"total_adjustments"`
			Items                  []CFSectionItem         `json:"items"`
		} `json:"adjustments"`
		
		WorkingCapitalChanges       struct {
			AccountsReceivableChange float64                `json:"accounts_receivable_change"`
			InventoryChange         float64                 `json:"inventory_change"`
			PrepaidExpensesChange   float64                 `json:"prepaid_expenses_change"`
			AccountsPayableChange   float64                 `json:"accounts_payable_change"`
			AccruedLiabilitiesChange float64               `json:"accrued_liabilities_change"`
			OtherWorkingCapitalChange float64              `json:"other_working_capital_change"`
			TotalWorkingCapitalChanges float64             `json:"total_working_capital_changes"`
			Items                   []CFSectionItem         `json:"items"`
		} `json:"working_capital_changes"`
		
		TotalOperatingCashFlow      float64                 `json:"total_operating_cash_flow"`
	} `json:"operating_activities"`

	// Investing Activities
	InvestingActivities struct {
		PurchaseOfFixedAssets      float64                  `json:"purchase_of_fixed_assets"`
		SaleOfFixedAssets         float64                   `json:"sale_of_fixed_assets"`
		PurchaseOfInvestments     float64                   `json:"purchase_of_investments"`
		SaleOfInvestments         float64                   `json:"sale_of_investments"`
		IntangibleAssetPurchases  float64                   `json:"intangible_asset_purchases"`
		OtherInvestingActivities  float64                   `json:"other_investing_activities"`
		TotalInvestingCashFlow    float64                   `json:"total_investing_cash_flow"`
		Items                     []CFSectionItem           `json:"items"`
	} `json:"investing_activities"`

	// Financing Activities
	FinancingActivities struct {
		ShareCapitalIncrease      float64                   `json:"share_capital_increase"`
		ShareCapitalDecrease      float64                   `json:"share_capital_decrease"`
		LongTermDebtIncrease      float64                   `json:"long_term_debt_increase"`
		LongTermDebtDecrease      float64                   `json:"long_term_debt_decrease"`
		ShortTermDebtIncrease     float64                   `json:"short_term_debt_increase"`
		ShortTermDebtDecrease     float64                   `json:"short_term_debt_decrease"`
		DividendsPaid             float64                   `json:"dividends_paid"`
		OtherFinancingActivities  float64                   `json:"other_financing_activities"`
		TotalFinancingCashFlow    float64                   `json:"total_financing_cash_flow"`
		Items                     []CFSectionItem           `json:"items"`
	} `json:"financing_activities"`

	// Summary
	NetCashFlow               float64                     `json:"net_cash_flow"`
	CashAtBeginning          float64                     `json:"cash_at_beginning"`
	CashAtEnd                float64                     `json:"cash_at_end"`
	
	// Analysis
	CashFlowRatios           struct {
		OperatingCashFlowRatio   float64                 `json:"operating_cash_flow_ratio"`
		CashFlowToDebtRatio     float64                  `json:"cash_flow_to_debt_ratio"`
		FreeCashFlow            float64                  `json:"free_cash_flow"`
		CashFlowPerShare        float64                  `json:"cash_flow_per_share,omitempty"`
	} `json:"cash_flow_ratios"`

	GeneratedAt              time.Time                    `json:"generated_at"`
	Enhanced                 bool                         `json:"enhanced"`

	// Account Details for Drilldown
	AccountDetails           []SSOTAccountBalance         `json:"account_details,omitempty"`
}

// CFSectionItem represents an item within a Cash Flow section
type CFSectionItem struct {
	AccountCode   string  `json:"account_code"`
	AccountName   string  `json:"account_name"`
	Amount        float64 `json:"amount"`
	AccountID     uint    `json:"account_id,omitempty"`
	Type          string  `json:"type"` // increase, decrease, inflow, outflow
}

// GenerateSSOTCashFlow generates Cash Flow statement from SSOT journal system
func (s *SSOTCashFlowService) GenerateSSOTCashFlow(startDate, endDate string) (*SSOTCashFlowData, error) {
	// Default to current fiscal year when parameters are empty
	if strings.TrimSpace(startDate) == "" || strings.TrimSpace(endDate) == "" {
		settingsSvc := NewSettingsService(s.db)
		fyStart, fyEnd, err := settingsSvc.GetCurrentFiscalYearRange()
		if err == nil {
			if strings.TrimSpace(startDate) == "" { startDate = fyStart }
			if strings.TrimSpace(endDate) == "" { endDate = fyEnd }
		}
	}
	// Parse dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %v", err)
	}
	
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %v", err)
	}
	
	// Make end date inclusive by using exclusive upper bound (end + 1 day)
	endExclusive := end.AddDate(0, 0, 1).Format("2006-01-02")
	
	// Get cash flow data from SSOT journal entries (inclusive date range)
	cashFlowTransactions, err := s.getCashFlowDataFromSSOT(startDate, endExclusive)
	if err != nil {
		return nil, fmt.Errorf("failed to get cash flow data: %v", err)
	}
	
	// Get net income for the period (from P&L)
	netIncome, err := s.getNetIncomeFromSSOT(startDate, endExclusive)
	if err != nil {
		return nil, fmt.Errorf("failed to get net income: %v", err)
	}
	
	// Get beginning and ending cash balances
	beginningCash, endingCash, err := s.getCashBalances(startDate, endExclusive)
	if err != nil {
		return nil, fmt.Errorf("failed to get cash balances: %v", err)
	}
	
	// Generate Cash Flow data structure
	cfData := s.generateCashFlowFromTransactions(cashFlowTransactions, netIncome, beginningCash, endingCash, start, end)
	
	return cfData, nil
}

// getCashFlowDataFromSSOT retrieves cash flow related transactions from SSOT journal system
func (s *SSOTCashFlowService) getCashFlowDataFromSSOT(startDate, endDate string) ([]SSOTAccountBalance, error) {
	var transactions []SSOTAccountBalance
	
	query := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			COALESCE(SUM(ujl.debit_amount), 0) as debit_total,
			COALESCE(SUM(ujl.credit_amount), 0) as credit_total,
			CASE 
				WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
					COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
				ELSE 
					COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
			END as net_balance
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE uje.status = 'POSTED' 
			AND uje.entry_date >= ? 
			AND uje.entry_date < ?
		GROUP BY a.id, a.code, a.name, a.type
		HAVING COALESCE(SUM(ujl.debit_amount), 0) > 0 OR COALESCE(SUM(ujl.credit_amount), 0) > 0
		ORDER BY a.code
	`
	
	if err := s.db.Raw(query, startDate, endDate).Scan(&transactions).Error; err != nil {
		return nil, fmt.Errorf("error executing cash flow transactions query: %v", err)
	}
	
	return transactions, nil
}

// getNetIncomeFromSSOT calculates net income from revenue and expense accounts
func (s *SSOTCashFlowService) getNetIncomeFromSSOT(startDate, endDate string) (float64, error) {
	var result struct {
		NetIncome float64 `json:"net_income"`
	}
	
	query := `
		SELECT 
			COALESCE(
				SUM(CASE 
					WHEN a.code LIKE '4%' THEN 
						COALESCE(ujl.credit_amount, 0) - COALESCE(ujl.debit_amount, 0)
					WHEN a.code LIKE '5%' OR a.code LIKE '6%' OR a.code LIKE '7%' THEN 
						COALESCE(ujl.debit_amount, 0) - COALESCE(ujl.credit_amount, 0)
					ELSE 0
				END), 0
			) * -1 as net_income
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE uje.status = 'POSTED' 
			AND uje.entry_date >= ? 
			AND uje.entry_date < ?
			AND (a.code LIKE '4%' OR a.code LIKE '5%' OR a.code LIKE '6%' OR a.code LIKE '7%')
	`
	
	if err := s.db.Raw(query, startDate, endDate).Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("error calculating net income: %v", err)
	}
	
	return result.NetIncome, nil
}

// getCashBalances gets beginning and ending cash balances
func (s *SSOTCashFlowService) getCashBalances(startDate, endDate string) (float64, float64, error) {
	var beginningCash, endingCash float64
	
	// Get beginning cash balance (before start date)
	// Include all Cash (1101%) and Bank (1102%) accounts per Indonesian COA standard
	beginningQuery := `
		SELECT 
			COALESCE(SUM(
				CASE 
					WHEN a.type = 'ASSET' THEN 
						COALESCE(ujl.debit_amount, 0) - COALESCE(ujl.credit_amount, 0)
					ELSE 0
				END
			), 0) as cash_balance
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE uje.status = 'POSTED' 
			AND uje.entry_date < ?
			AND (a.code LIKE '1101%' OR a.code LIKE '1102%')
	`
	
	// Get ending cash balance (up to end date)
	// Include all Cash (1101%) and Bank (1102%) accounts per Indonesian COA standard
	endingQuery := `
		SELECT 
			COALESCE(SUM(
				CASE 
					WHEN a.type = 'ASSET' THEN 
						COALESCE(ujl.debit_amount, 0) - COALESCE(ujl.credit_amount, 0)
					ELSE 0
				END
			), 0) as cash_balance
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE uje.status = 'POSTED' 
			AND uje.entry_date < ?
			AND (a.code LIKE '1101%' OR a.code LIKE '1102%')
	`
	
	var beginResult struct {
		CashBalance float64 `json:"cash_balance"`
	}
	var endResult struct {
		CashBalance float64 `json:"cash_balance"`
	}
	
	if err := s.db.Raw(beginningQuery, startDate).Scan(&beginResult).Error; err != nil {
		return 0, 0, fmt.Errorf("error getting beginning cash balance: %v", err)
	}
	
	if err := s.db.Raw(endingQuery, endDate).Scan(&endResult).Error; err != nil {
		return 0, 0, fmt.Errorf("error getting ending cash balance: %v", err)
	}
	
	beginningCash = beginResult.CashBalance
	endingCash = endResult.CashBalance
	
	return beginningCash, endingCash, nil
}

// generateCashFlowFromTransactions creates the Cash Flow structure from transactions
func (s *SSOTCashFlowService) generateCashFlowFromTransactions(transactions []SSOTAccountBalance, netIncome, beginningCash, endingCash float64, start, end time.Time) *SSOTCashFlowData {
	// Get company information from settings
	settingsSvc := NewSettingsService(s.db)
	settings, err := settingsSvc.GetSettings()
	
	var companyInfo CompanyInfo
	if err == nil && settings != nil {
		// Use actual settings from database
		companyInfo = CompanyInfo{
			Name:    settings.CompanyName,
			Address: settings.CompanyAddress,
			Phone:   settings.CompanyPhone,
			Email:   settings.CompanyEmail,
		}
	} else {
		// Fallback to default values if settings cannot be retrieved
		companyInfo = CompanyInfo{
			Name:    "PT. Sistem Akuntansi Indonesia",
			Address: "Jl. Sudirman Kav. 45-46, Jakarta Pusat 10210, Indonesia",
			Phone:   "+62-21-5551234",
			Email:   "info@sistemakuntansi.co.id",
		}
	}
	
	cfData := &SSOTCashFlowData{
		Company:         companyInfo,
		StartDate:       start,
		EndDate:         end,
		Currency:        "IDR",
		Enhanced:        true,
		GeneratedAt:     time.Now(),
		CashAtBeginning: beginningCash,
		CashAtEnd:       endingCash,
		AccountDetails:  transactions,
	}
	
	// Initialize sections
	cfData.OperatingActivities.NetIncome = netIncome
	cfData.OperatingActivities.Adjustments.Items = []CFSectionItem{}
	cfData.OperatingActivities.WorkingCapitalChanges.Items = []CFSectionItem{}
	cfData.InvestingActivities.Items = []CFSectionItem{}
	cfData.FinancingActivities.Items = []CFSectionItem{}
	
	// Process each transaction
	for _, transaction := range transactions {
		code := transaction.AccountCode
		amount := transaction.NetBalance
		
		// Skip if amount is zero
		if amount == 0 {
			continue
		}
		
		item := CFSectionItem{
			AccountCode: transaction.AccountCode,
			AccountName: transaction.AccountName,
			Amount:      amount,
			AccountID:   transaction.AccountID,
		}
		
		// Categorize based on account codes
		s.categorizeCashFlowTransaction(cfData, item, code, amount)
	}
	
	// Calculate totals and ratios
	s.calculateCashFlowTotals(cfData)
	
	return cfData
}

// categorizeCashFlowTransaction categorizes transactions into cash flow activities
func (s *SSOTCashFlowService) categorizeCashFlowTransaction(cfData *SSOTCashFlowData, item CFSectionItem, code string, amount float64) {
	switch {
	// Operating Activities - Non-cash adjustments
	case strings.Contains(strings.ToLower(item.AccountName), "depreciation") || 
		 strings.Contains(strings.ToLower(item.AccountName), "depresiasi"):
		cfData.OperatingActivities.Adjustments.Depreciation += amount
		item.Type = "increase"
		cfData.OperatingActivities.Adjustments.Items = append(cfData.OperatingActivities.Adjustments.Items, item)
		
	case strings.Contains(strings.ToLower(item.AccountName), "amortization") || 
		 strings.Contains(strings.ToLower(item.AccountName), "amortisasi"):
		cfData.OperatingActivities.Adjustments.Amortization += amount
		item.Type = "increase"
		cfData.OperatingActivities.Adjustments.Items = append(cfData.OperatingActivities.Adjustments.Items, item)
		
	case strings.Contains(strings.ToLower(item.AccountName), "bad debt") || 
		 strings.Contains(strings.ToLower(item.AccountName), "piutang"):
		cfData.OperatingActivities.Adjustments.BadDebtExpense += amount
		item.Type = "increase"
		cfData.OperatingActivities.Adjustments.Items = append(cfData.OperatingActivities.Adjustments.Items, item)
		
	// Operating Activities - Working Capital Changes
	case strings.HasPrefix(code, "12"): // Accounts Receivable (12xx)
		// Per COA: 12xx = Piutang. Kenaikan piutang = arus kas keluar (indirect method)
		cfData.OperatingActivities.WorkingCapitalChanges.AccountsReceivableChange += amount * -1
		item.Amount = amount * -1  // Update item amount to match the calculation
		if amount > 0 {
			item.Type = "decrease"
		} else {
			item.Type = "increase"
		}
		cfData.OperatingActivities.WorkingCapitalChanges.Items = append(cfData.OperatingActivities.WorkingCapitalChanges.Items, item)
		
	case strings.HasPrefix(code, "13"): // Inventory (13xx)
		// Per COA: 13xx = Persediaan. Kenaikan persediaan = arus kas keluar
		cfData.OperatingActivities.WorkingCapitalChanges.InventoryChange += amount * -1
		item.Amount = amount * -1  // Update item amount to match the calculation
		if amount > 0 {
			item.Type = "decrease"
		} else {
			item.Type = "increase"
		}
		cfData.OperatingActivities.WorkingCapitalChanges.Items = append(cfData.OperatingActivities.WorkingCapitalChanges.Items, item)
		
	case strings.HasPrefix(code, "114") || strings.HasPrefix(code, "115"): // Prepaid Expenses
		cfData.OperatingActivities.WorkingCapitalChanges.PrepaidExpensesChange += amount * -1
		item.Amount = amount * -1  // Update item amount to match the calculation
		if amount > 0 {
			item.Type = "decrease"
		} else {
			item.Type = "increase"
		}
		cfData.OperatingActivities.WorkingCapitalChanges.Items = append(cfData.OperatingActivities.WorkingCapitalChanges.Items, item)
		
	case strings.HasPrefix(code, "124"): // PPN Masukan (Tax Prepaid/Input VAT - 1240)
		// PPN Masukan adalah asset. Penurunan = kas keluar untuk kompensasi pajak
		// Diperlakukan seperti prepaid: kenaikan = kas keluar, penurunan = kas masuk
		cfData.OperatingActivities.WorkingCapitalChanges.OtherWorkingCapitalChange += amount * -1
		item.Amount = amount * -1
		item.AccountName = item.AccountName + " (PPN Masukan)"
		if amount > 0 {
			item.Type = "decrease" // Asset increase = cash outflow
		} else {
			item.Type = "increase" // Asset decrease = cash inflow (kompensasi)
		}
		cfData.OperatingActivities.WorkingCapitalChanges.Items = append(cfData.OperatingActivities.WorkingCapitalChanges.Items, item)
		
	case strings.HasPrefix(code, "210") && !strings.HasPrefix(code, "2103"): // Accounts Payable (exclude PPN Keluaran)
		cfData.OperatingActivities.WorkingCapitalChanges.AccountsPayableChange += amount // Increase in A/P is cash inflow
		if amount > 0 {
			item.Type = "increase"
		} else {
			item.Type = "decrease"
		}
		cfData.OperatingActivities.WorkingCapitalChanges.Items = append(cfData.OperatingActivities.WorkingCapitalChanges.Items, item)
		
	case strings.HasPrefix(code, "2103"): // PPN Keluaran (Output VAT Payable - 2103)
		// PPN Keluaran adalah liability. Penurunan = kas keluar (bayar pajak ke negara)
		// Increase in liability = cash inflow (belum bayar), decrease = cash outflow (sudah bayar)
		cfData.OperatingActivities.WorkingCapitalChanges.OtherWorkingCapitalChange += amount
		item.AccountName = item.AccountName + " (PPN Keluaran)"
		if amount > 0 {
			item.Type = "increase" // Liability increase = cash inflow (defer payment)
		} else {
			item.Type = "decrease" // Liability decrease = cash outflow (paid tax)
		}
		cfData.OperatingActivities.WorkingCapitalChanges.Items = append(cfData.OperatingActivities.WorkingCapitalChanges.Items, item)
		
	case strings.HasPrefix(code, "212") || strings.HasPrefix(code, "213"): // Accrued Liabilities
		cfData.OperatingActivities.WorkingCapitalChanges.AccruedLiabilitiesChange += amount
		if amount > 0 {
			item.Type = "increase"
		} else {
			item.Type = "decrease"
		}
		cfData.OperatingActivities.WorkingCapitalChanges.Items = append(cfData.OperatingActivities.WorkingCapitalChanges.Items, item)
		
	// Investing Activities
	case strings.HasPrefix(code, "16") || strings.HasPrefix(code, "17"): // Fixed Assets (Activa Tetap)
		if amount > 0 {
			cfData.InvestingActivities.PurchaseOfFixedAssets += amount * -1 // Purchase is cash outflow
			item.Type = "outflow"
		} else {
			cfData.InvestingActivities.SaleOfFixedAssets += amount * -1 // Sale is cash inflow
			item.Type = "inflow"
		}
		cfData.InvestingActivities.Items = append(cfData.InvestingActivities.Items, item)
		
	case strings.HasPrefix(code, "15"): // Investments
		if amount > 0 {
			cfData.InvestingActivities.PurchaseOfInvestments += amount * -1
			item.Type = "outflow"
		} else {
			cfData.InvestingActivities.SaleOfInvestments += amount * -1
			item.Type = "inflow"
		}
		cfData.InvestingActivities.Items = append(cfData.InvestingActivities.Items, item)
		
	case strings.HasPrefix(code, "14"): // Intangible Assets
		cfData.InvestingActivities.IntangibleAssetPurchases += amount * -1
		item.Type = "outflow"
		cfData.InvestingActivities.Items = append(cfData.InvestingActivities.Items, item)
		
	// Financing Activities
	case strings.HasPrefix(code, "31"): // Share Capital
		if amount > 0 {
			cfData.FinancingActivities.ShareCapitalIncrease += amount
			item.Type = "inflow"
		} else {
			cfData.FinancingActivities.ShareCapitalDecrease += amount * -1
			item.Type = "outflow"
		}
		cfData.FinancingActivities.Items = append(cfData.FinancingActivities.Items, item)
		
	case strings.HasPrefix(code, "22"): // Long-term Debt
		if amount > 0 {
			cfData.FinancingActivities.LongTermDebtIncrease += amount
			item.Type = "inflow"
		} else {
			cfData.FinancingActivities.LongTermDebtDecrease += amount * -1
			item.Type = "outflow"
		}
		cfData.FinancingActivities.Items = append(cfData.FinancingActivities.Items, item)
		
	case strings.HasPrefix(code, "211"): // Short-term Debt
		if amount > 0 {
			cfData.FinancingActivities.ShortTermDebtIncrease += amount
			item.Type = "inflow"
		} else {
			cfData.FinancingActivities.ShortTermDebtDecrease += amount * -1
			item.Type = "outflow"
		}
		cfData.FinancingActivities.Items = append(cfData.FinancingActivities.Items, item)
		
	case strings.Contains(strings.ToLower(item.AccountName), "dividend") || 
		 strings.Contains(strings.ToLower(item.AccountName), "dividen"):
		cfData.FinancingActivities.DividendsPaid += amount * -1
		item.Type = "outflow"
		cfData.FinancingActivities.Items = append(cfData.FinancingActivities.Items, item)
	}
}

// calculateCashFlowTotals calculates all totals and ratios
func (s *SSOTCashFlowService) calculateCashFlowTotals(cfData *SSOTCashFlowData) {
	// Calculate operating activities adjustments total
	cfData.OperatingActivities.Adjustments.TotalAdjustments = 
		cfData.OperatingActivities.Adjustments.Depreciation +
		cfData.OperatingActivities.Adjustments.Amortization +
		cfData.OperatingActivities.Adjustments.BadDebtExpense +
		cfData.OperatingActivities.Adjustments.GainLossOnAssetDisposal +
		cfData.OperatingActivities.Adjustments.OtherNonCashItems
	
	// Calculate working capital changes total
	cfData.OperatingActivities.WorkingCapitalChanges.TotalWorkingCapitalChanges = 
		cfData.OperatingActivities.WorkingCapitalChanges.AccountsReceivableChange +
		cfData.OperatingActivities.WorkingCapitalChanges.InventoryChange +
		cfData.OperatingActivities.WorkingCapitalChanges.PrepaidExpensesChange +
		cfData.OperatingActivities.WorkingCapitalChanges.AccountsPayableChange +
		cfData.OperatingActivities.WorkingCapitalChanges.AccruedLiabilitiesChange +
		cfData.OperatingActivities.WorkingCapitalChanges.OtherWorkingCapitalChange
	
	// Calculate total operating cash flow
	cfData.OperatingActivities.TotalOperatingCashFlow = 
		cfData.OperatingActivities.NetIncome +
		cfData.OperatingActivities.Adjustments.TotalAdjustments +
		cfData.OperatingActivities.WorkingCapitalChanges.TotalWorkingCapitalChanges
	
	// Calculate investing activities total
	cfData.InvestingActivities.TotalInvestingCashFlow = 
		cfData.InvestingActivities.PurchaseOfFixedAssets +
		cfData.InvestingActivities.SaleOfFixedAssets +
		cfData.InvestingActivities.PurchaseOfInvestments +
		cfData.InvestingActivities.SaleOfInvestments +
		cfData.InvestingActivities.IntangibleAssetPurchases +
		cfData.InvestingActivities.OtherInvestingActivities
	
	// Calculate financing activities total
	cfData.FinancingActivities.TotalFinancingCashFlow = 
		cfData.FinancingActivities.ShareCapitalIncrease +
		cfData.FinancingActivities.ShareCapitalDecrease +
		cfData.FinancingActivities.LongTermDebtIncrease +
		cfData.FinancingActivities.LongTermDebtDecrease +
		cfData.FinancingActivities.ShortTermDebtIncrease +
		cfData.FinancingActivities.ShortTermDebtDecrease +
		cfData.FinancingActivities.DividendsPaid +
		cfData.FinancingActivities.OtherFinancingActivities
	
	// Calculate net cash flow
	cfData.NetCashFlow = 
		cfData.OperatingActivities.TotalOperatingCashFlow +
		cfData.InvestingActivities.TotalInvestingCashFlow +
		cfData.FinancingActivities.TotalFinancingCashFlow
	
	// Calculate ratios
	if cfData.OperatingActivities.TotalOperatingCashFlow > 0 {
		// Calculate free cash flow (Operating CF - Capital Expenditures)
		capEx := cfData.InvestingActivities.PurchaseOfFixedAssets * -1
		cfData.CashFlowRatios.FreeCashFlow = cfData.OperatingActivities.TotalOperatingCashFlow - capEx
		
		// Operating cash flow ratio (simplified)
		cfData.CashFlowRatios.OperatingCashFlowRatio = cfData.OperatingActivities.TotalOperatingCashFlow / cfData.OperatingActivities.NetIncome
		if cfData.OperatingActivities.NetIncome == 0 {
			cfData.CashFlowRatios.OperatingCashFlowRatio = 0
		}
	}
	
	// Validate cash balance reconciliation
	expectedEndingCash := cfData.CashAtBeginning + cfData.NetCashFlow
	if expectedEndingCash != cfData.CashAtEnd {
		// There might be untracked cash transactions, adjust if needed
		cfData.OperatingActivities.Adjustments.OtherNonCashItems += cfData.CashAtEnd - expectedEndingCash
		cfData.OperatingActivities.Adjustments.TotalAdjustments += cfData.CashAtEnd - expectedEndingCash
		cfData.OperatingActivities.TotalOperatingCashFlow += cfData.CashAtEnd - expectedEndingCash
		cfData.NetCashFlow += cfData.CashAtEnd - expectedEndingCash
	}
}

