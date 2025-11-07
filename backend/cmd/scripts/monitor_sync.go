package main

import (
	"fmt"
	"time"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔍 Real-time CashBank-COA Sync Monitor...")
	
	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	
	fmt.Println("📊 Current CashBank-COA Status:")
	fmt.Println("=====================================")
	
	// Monitor loop
	for i := 0; i < 10; i++ {
		var linkedAccounts []struct {
			CashBankID      uint    `db:"cashbank_id"`
			CashBankCode    string  `db:"cashbank_code"`
			CashBankName    string  `db:"cashbank_name"`
			CashBankBalance float64 `db:"cashbank_balance"`
			AccountCode     string  `db:"account_code"`
			AccountName     string  `db:"account_name"`
			COABalance      float64 `db:"coa_balance"`
			IsSynced        string  `db:"is_synced"`
		}
		
		db.Raw(`
			SELECT 
				cb.id as cashbank_id,
				cb.code as cashbank_code,
				COALESCE(cb.name, 'UNNAMED') as cashbank_name, 
				cb.balance as cashbank_balance,
				a.code as account_code,
				a.name as account_name,
				a.balance as coa_balance,
				CASE 
					WHEN cb.balance = a.balance THEN '✅ SYNCED'
					ELSE '❌ OUT OF SYNC'
				END as is_synced
			FROM cash_banks cb 
			JOIN accounts a ON cb.account_id = a.id 
			WHERE cb.deleted_at IS NULL 
			  AND cb.is_active = true
			  AND cb.account_id IS NOT NULL
			  AND cb.account_id > 0
			ORDER BY cb.code, a.code
		`).Scan(&linkedAccounts)
		
		fmt.Printf("\n⏰ Check #%d - %s\n", i+1, time.Now().Format("15:04:05"))
		fmt.Printf("Found %d linked accounts:\n", len(linkedAccounts))
		
		for _, acc := range linkedAccounts {
			fmt.Printf("  %s %s: CB=%.0f ↔ COA=%s %.0f %s\n", 
				acc.CashBankCode, acc.CashBankName,
				acc.CashBankBalance,
				acc.AccountCode, acc.COABalance,
				acc.IsSynced)
		}
		
		// Check recent audit logs
		var recentAudits []struct {
			Action    string    `db:"action"`
			TableName string    `db:"table_name"`
			CreatedAt time.Time `db:"created_at"`
		}
		
		db.Raw(`
			SELECT action, table_name, created_at
			FROM audit_logs 
			WHERE table_name IN ('coa_to_cashbank_sync', 'cashbank_coa_sync')
			AND created_at > NOW() - INTERVAL '1 minute'
			ORDER BY created_at DESC
			LIMIT 5
		`).Scan(&recentAudits)
		
		if len(recentAudits) > 0 {
			fmt.Printf("\n📋 Recent sync activity:\n")
			for _, audit := range recentAudits {
				fmt.Printf("  %s - %s (%s)\n", 
					audit.CreatedAt.Format("15:04:05"), 
					audit.Action, audit.TableName)
			}
		}
		
		if i < 9 {
			fmt.Println("\n⏳ Waiting 5 seconds... (modify COA balance in UI to test)")
			time.Sleep(5 * time.Second)
		}
	}
	
	fmt.Println("\n✅ Monitoring completed!")
	fmt.Println("\n📖 Instructions:")
	fmt.Println("1. Open Chart of Accounts in your browser")
	fmt.Println("2. Edit any Bank account balance (1102, 1103, 1105)")
	fmt.Println("3. Save changes")
	fmt.Println("4. Check if CashBank balance updates automatically")
	fmt.Println("5. If working, you should see sync activity in logs above")
}
