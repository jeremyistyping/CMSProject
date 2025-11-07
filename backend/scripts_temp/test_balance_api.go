package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
)

type BalanceResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data    BalanceSummaryData `json:"data"`
}

type BalanceSummaryData struct {
	TotalCash    float64            `json:"total_cash"`
	TotalBank    float64            `json:"total_bank"`
	TotalBalance float64            `json:"total_balance"`
	ByAccount    []AccountBalance   `json:"by_account"`
	ByCurrency   map[string]float64 `json:"by_currency"`
}

type AccountBalance struct {
	AccountID   uint    `json:"account_id"`
	AccountName string  `json:"account_name"`
	AccountType string  `json:"account_type"`
	Balance     float64 `json:"balance"`
	Currency    string  `json:"currency"`
}

func main() {
	// Test the API endpoint
	resp, err := http.Get("http://localhost:8080/api/cash-bank/reports/balance-summary")
	if err != nil {
		log.Fatal("Error calling API:", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response:", err)
	}

	fmt.Printf("HTTP Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n\n", string(body))

	// Parse JSON response
	var balanceResp BalanceResponse
	err = json.Unmarshal(body, &balanceResp)
	if err != nil {
		log.Fatal("Error parsing JSON:", err)
	}

	// Display formatted results
	fmt.Println("=== BALANCE SUMMARY RESULTS ===")
	fmt.Printf("Total Cash: IDR %.2f\n", balanceResp.Data.TotalCash)
	fmt.Printf("Total Bank: IDR %.2f\n", balanceResp.Data.TotalBank)
	fmt.Printf("Total Balance: IDR %.2f\n", balanceResp.Data.TotalBalance)
	
	fmt.Println("\n=== BY ACCOUNT ===")
	for _, account := range balanceResp.Data.ByAccount {
		fmt.Printf("ID: %d, Name: %s, Type: %s, Balance: %.2f %s\n", 
			account.AccountID, account.AccountName, account.AccountType, 
			account.Balance, account.Currency)
	}
	
	fmt.Println("\n=== BY CURRENCY ===")
	for currency, total := range balanceResp.Data.ByCurrency {
		fmt.Printf("%s: %.2f\n", currency, total)
	}
}