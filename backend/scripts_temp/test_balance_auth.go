package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    LoginData   `json:"data"`
}

type LoginData struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type BalanceResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
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
	// Step 1: Login to get token
	loginReq := LoginRequest{
		Email:    "admin@company.com",
		Password: "admin123",
	}
	
	loginJSON, err := json.Marshal(loginReq)
	if err != nil {
		log.Fatal("Error marshaling login request:", err)
	}
	
	loginResp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		log.Fatal("Error calling login API:", err)
	}
	defer loginResp.Body.Close()
	
	loginBody, err := ioutil.ReadAll(loginResp.Body)
	if err != nil {
		log.Fatal("Error reading login response:", err)
	}
	
	fmt.Printf("Login Status: %d\n", loginResp.StatusCode)
	
	var loginResult LoginResponse
	err = json.Unmarshal(loginBody, &loginResult)
	if err != nil {
		log.Fatal("Error parsing login JSON:", err)
	}
	
	if !loginResult.Success {
		log.Fatal("Login failed:", loginResult.Message)
	}
	
	token := loginResult.Data.Token
	fmt.Printf("Login successful! Token: %s\n\n", token[:20]+"...")
	
	// Step 2: Call balance summary with authorization
	req, err := http.NewRequest("GET", "http://localhost:8080/api/cash-bank/reports/balance-summary", nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error calling balance API:", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading balance response:", err)
	}

	fmt.Printf("Balance API Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n\n", string(body))

	// Parse JSON response
	if resp.StatusCode == 200 {
		var balanceResp BalanceResponse
		err = json.Unmarshal(body, &balanceResp)
		if err != nil {
			log.Fatal("Error parsing balance JSON:", err)
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
}