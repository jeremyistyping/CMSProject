package main

import (
	"fmt"
	"app-sistem-akuntansi/database"
)

type AccountBalance struct {
	Code    string
	Name    string
	Balance float64
}

func main() {
	db := database.ConnectDB()
	
	type AccountBalanceExt struct {
		Code      string
		Name      string
		Balance   float64
		DeletedAt *string
	}
	
	var accounts []AccountBalanceExt
	db.Raw(`
		SELECT code, name, balance, deleted_at::text as deleted_at
		FROM accounts 
		WHERE code IN ('3101', '3201', '4101', '5101', '1102-001', '1301', '2101')
		ORDER BY code, id
	`).Scan(&accounts)
	
	fmt.Println("\n=== ACCOUNT BALANCES (from accounts table) ===")
	for _, acc := range accounts {
		deletedStatus := ""
		if acc.DeletedAt != nil {
			deletedStatus = " [DELETED]"
		}
		fmt.Printf("%s - %s: %.2f%s\n", acc.Code, acc.Name, acc.Balance, deletedStatus)
	}
}
