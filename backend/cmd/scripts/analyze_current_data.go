package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔍 Analyzing Current Data Structure for PSAK Compliance...")
	
	// Load configuration
	cfg := config.LoadConfig()
	
	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	log.Println("✅ Database connected successfully")
	
	// 1. Analyze Account Structure
	analyzeAccountStructure(db)
	
	// 2. Analyze CashBank Structure
	analyzeCashBankStructure(db)
	
	// 3. Check Account-CashBank Links
	checkAccountCashBankLinks(db)
	
	// 4. Identify PSAK Compliance Issues
	identifyPSAKIssues(db)
	
	// 5. Check for Orphaned Data
	checkOrphanedData(db)
	
	log.Println("🎉 Analysis completed!")
}

func analyzeAccountStructure(db *gorm.DB) {
	log.Println("\n📊 ACCOUNT STRUCTURE ANALYSIS:")
	log.Println("=" + strings.Repeat("=", 50))
	
	var accounts []models.Account
	db.Find(&accounts)
	
	// Group by type and level
	typeGroups := make(map[string][]models.Account)
	levelGroups := make(map[int][]models.Account)
	
	for _, acc := range accounts {
		typeGroups[acc.Type] = append(typeGroups[acc.Type], acc)
		levelGroups[acc.Level] = append(levelGroups[acc.Level], acc)
	}
	
	// Show by type
	log.Println("\n📋 Accounts by Type:")
	for accountType, accs := range typeGroups {
		log.Printf("  %s: %d accounts", accountType, len(accs))
		
		// Show asset accounts in detail (cash/bank related)
		if accountType == "ASSET" {
			cashBankAccounts := []models.Account{}
			for _, acc := range accs {
				if strings.Contains(strings.ToUpper(acc.Name), "KAS") || 
				   strings.Contains(strings.ToUpper(acc.Name), "BANK") ||
				   strings.Contains(strings.ToUpper(acc.Code), "110") {
					cashBankAccounts = append(cashBankAccounts, acc)
				}
			}
			
			if len(cashBankAccounts) > 0 {
				log.Printf("    💰 Cash/Bank related accounts (%d):", len(cashBankAccounts))
				sort.Slice(cashBankAccounts, func(i, j int) bool {
					return cashBankAccounts[i].Code < cashBankAccounts[j].Code
				})
				
				for _, acc := range cashBankAccounts {
					parentInfo := "ROOT"
					if acc.ParentID != nil {
						parentInfo = fmt.Sprintf("Parent ID: %d", *acc.ParentID)
					}
					headerMark := ""
					if acc.IsHeader {
						headerMark = " [HEADER]"
					}
					
					log.Printf("      %s - %s (Level: %d, %s)%s Balance: %.2f", 
						acc.Code, acc.Name, acc.Level, parentInfo, headerMark, acc.Balance)
				}
			}
		}
	}
	
	// Show hierarchy issues
	log.Println("\n🔍 Hierarchy Analysis:")
	for level, accs := range levelGroups {
		log.Printf("  Level %d: %d accounts", level, len(accs))
	}
}

func analyzeCashBankStructure(db *gorm.DB) {
	log.Println("\n🏦 CASH/BANK ACCOUNTS ANALYSIS:")
	log.Println("=" + strings.Repeat("=", 50))
	
	var cashBanks []models.CashBank
	db.Preload("Account").Find(&cashBanks)
	
	log.Printf("Total CashBank records: %d", len(cashBanks))
	
	typeGroups := make(map[string][]models.CashBank)
	linkedCount := 0
	unlinkedCount := 0
	
	for _, cb := range cashBanks {
		typeGroups[cb.Type] = append(typeGroups[cb.Type], cb)
		if cb.AccountID > 0 {
			linkedCount++
		} else {
			unlinkedCount++
		}
	}
	
	log.Printf("📈 Linked to GL Account: %d", linkedCount)
	log.Printf("📉 Not linked to GL Account: %d", unlinkedCount)
	
	log.Println("\n💰 By Type:")
	for cbType, cbs := range typeGroups {
		log.Printf("  %s: %d accounts", cbType, len(cbs))
		
		sort.Slice(cbs, func(i, j int) bool {
			return cbs[i].Code < cbs[j].Code
		})
		
		for _, cb := range cbs {
			glInfo := "NO GL LINK"
			if cb.AccountID > 0 && cb.Account.Code != "" {
				glInfo = fmt.Sprintf("GL: %s (%s)", cb.Account.Code, cb.Account.Name)
			}
			
			log.Printf("    %s - %s | %s | Balance: %.2f", 
				cb.Code, cb.Name, glInfo, cb.Balance)
		}
	}
}

func checkAccountCashBankLinks(db *gorm.DB) {
	log.Println("\n🔗 ACCOUNT-CASHBANK LINK ANALYSIS:")
	log.Println("=" + strings.Repeat("=", 50))
	
	// Find accounts that should be linked to cash/bank but aren't
	var accounts []models.Account
	db.Where("type = ? AND (name ILIKE ? OR name ILIKE ? OR code LIKE ?)", 
		"ASSET", "%kas%", "%bank%", "110%").Find(&accounts)
	
	var cashBanks []models.CashBank
	db.Find(&cashBanks)
	
	// Create map of linked account IDs
	linkedAccountIDs := make(map[uint]bool)
	for _, cb := range cashBanks {
		if cb.AccountID > 0 {
			linkedAccountIDs[cb.AccountID] = true
		}
	}
	
	log.Println("🎯 Asset accounts that might be cash/bank related:")
	orphanedAssets := []models.Account{}
	
	for _, acc := range accounts {
		isLinked := linkedAccountIDs[acc.ID]
		linkStatus := "❌ NOT LINKED"
		if isLinked {
			linkStatus = "✅ LINKED"
		} else {
			orphanedAssets = append(orphanedAssets, acc)
		}
		
		log.Printf("  %s - %s | %s", acc.Code, acc.Name, linkStatus)
	}
	
	if len(orphanedAssets) > 0 {
		log.Printf("\n⚠️  Found %d asset accounts that might need CashBank links:", len(orphanedAssets))
		for _, acc := range orphanedAssets {
			log.Printf("    %s - %s (Balance: %.2f)", acc.Code, acc.Name, acc.Balance)
		}
	}
}

func identifyPSAKIssues(db *gorm.DB) {
	log.Println("\n🚨 PSAK COMPLIANCE ISSUES:")
	log.Println("=" + strings.Repeat("=", 50))
	
	issues := []string{}
	
	// 1. Check for wrong code format
	var accounts []models.Account
	db.Where("type = ?", "ASSET").Find(&accounts)
	
	for _, acc := range accounts {
		// Check asset accounts that don't start with 1
		if !strings.HasPrefix(acc.Code, "1") {
			issues = append(issues, fmt.Sprintf("❌ Asset account '%s' doesn't start with 1", acc.Code))
		}
		
		// Check for inconsistent child format
		if strings.Contains(acc.Code, "-") {
			parts := strings.Split(acc.Code, "-")
			if len(parts) != 2 {
				issues = append(issues, fmt.Sprintf("❌ Invalid child format: '%s'", acc.Code))
			} else if len(parts[0]) != 4 || len(parts[1]) != 3 {
				issues = append(issues, fmt.Sprintf("❌ Wrong child format: '%s' should be XXXX-XXX", acc.Code))
			}
		}
		
		// Check for header accounts without children
		if acc.IsHeader {
			var childCount int64
			db.Model(&models.Account{}).Where("parent_id = ?", acc.ID).Count(&childCount)
			if childCount == 0 {
				issues = append(issues, fmt.Sprintf("❌ Header account '%s' has no children", acc.Code))
			}
		}
	}
	
	// 2. Check CashBank account code issues
	var cashBanks []models.CashBank
	db.Find(&cashBanks)
	
	for _, cb := range cashBanks {
		// Check for wrong CashBank code format
		if !strings.Contains(cb.Code, "-") && len(cb.Code) > 4 {
			issues = append(issues, fmt.Sprintf("❌ CashBank code format issue: '%s'", cb.Code))
		}
	}
	
	// 3. Check for missing parent accounts
	requiredParents := map[string]string{
		"1100": "KAS DAN BANK",
		"1101": "KAS",
		"1102": "BANK BCA", 
		"1103": "BANK MANDIRI",
		"1104": "BANK UOB",
	}
	
	for code, expectedName := range requiredParents {
		var count int64
		db.Model(&models.Account{}).Where("code = ?", code).Count(&count)
		if count == 0 {
			issues = append(issues, fmt.Sprintf("❌ Missing required parent account: %s (%s)", code, expectedName))
		}
	}
	
	// Display all issues
	if len(issues) > 0 {
		log.Printf("Found %d PSAK compliance issues:", len(issues))
		for i, issue := range issues {
			log.Printf("  %d. %s", i+1, issue)
		}
	} else {
		log.Println("✅ No major PSAK compliance issues found!")
	}
}

func checkOrphanedData(db *gorm.DB) {
	log.Println("\n👻 ORPHANED DATA ANALYSIS:")
	log.Println("=" + strings.Repeat("=", 50))
	
	// Check for CashBank records with invalid account_id
	var orphanedCB []models.CashBank
	db.Where("account_id > 0 AND account_id NOT IN (SELECT id FROM accounts WHERE deleted_at IS NULL)").Find(&orphanedCB)
	
	if len(orphanedCB) > 0 {
		log.Printf("❌ Found %d CashBank records with invalid account_id:", len(orphanedCB))
		for _, cb := range orphanedCB {
			log.Printf("  CashBank ID: %d, Code: %s, Invalid Account ID: %d", cb.ID, cb.Code, cb.AccountID)
		}
	} else {
		log.Println("✅ No orphaned CashBank records found")
	}
	
	// Check for accounts that might be duplicated
	duplicateMap := make(map[string][]models.Account)
	var allAccounts []models.Account
	db.Find(&allAccounts)
	
	for _, acc := range allAccounts {
		key := strings.ToLower(strings.TrimSpace(acc.Name))
		duplicateMap[key] = append(duplicateMap[key], acc)
	}
	
	duplicates := []string{}
	for name, accs := range duplicateMap {
		if len(accs) > 1 {
			codes := []string{}
			for _, acc := range accs {
				codes = append(codes, acc.Code)
			}
			duplicates = append(duplicates, fmt.Sprintf("Name: '%s' -> Codes: %v", name, codes))
		}
	}
	
	if len(duplicates) > 0 {
		log.Printf("⚠️  Found %d potential duplicate account names:", len(duplicates))
		for _, dup := range duplicates {
			log.Printf("  %s", dup)
		}
	} else {
		log.Println("✅ No duplicate account names found")
	}
}