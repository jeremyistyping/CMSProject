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

type AccountsResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    []CashBank  `json:"data"`
}

type CashBank struct {
	ID          uint    `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Balance     float64 `json:"balance"`
	Currency    string  `json:"currency"`
	IsActive    bool    `json:"is_active"`
	BankName    string  `json:"bank_name"`
	AccountNo   string  `json:"account_no"`
	Description string  `json:"description"`
}

func login() string {
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
	
	var loginResult LoginResponse
	err = json.Unmarshal(loginBody, &loginResult)
	if err != nil {
		log.Fatal("Error parsing login JSON:", err)
	}
	
	if !loginResult.Success {
		log.Fatal("Login failed:", loginResult.Message)
	}
	
	return loginResult.Data.Token
}

func testEndpoint(token, url string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error calling API:", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response:", err)
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n\n", string(body))

	if resp.StatusCode == 200 {
		var accountsResp AccountsResponse
		err = json.Unmarshal(body, &accountsResp)
		if err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			return
		}

		fmt.Printf("Total accounts returned: %d\n", len(accountsResp.Data))
		fmt.Println("\n=== INDIVIDUAL ACCOUNTS ===")
		
		cashCount := 0
		bankCount := 0
		totalCash := 0.0
		totalBank := 0.0
		
		for _, account := range accountsResp.Data {
			fmt.Printf("ID: %d | Code: %s | Name: %s | Type: %s | Balance: %.2f | Currency: %s | Active: %t\n", 
				account.ID, account.Code, account.Name, account.Type, account.Balance, account.Currency, account.IsActive)
			
			if account.Type == "CASH" {
				cashCount++
				totalCash += account.Balance
			} else if account.Type == "BANK" {
				bankCount++
				totalBank += account.Balance
			}
		}
		
		fmt.Printf("\n=== SUMMARY FROM INDIVIDUAL ACCOUNTS ===\n")
		fmt.Printf("Cash Accounts: %d | Total Cash: %.2f\n", cashCount, totalCash)
		fmt.Printf("Bank Accounts: %d | Total Bank: %.2f\n", bankCount, totalBank)
		fmt.Printf("Grand Total: %.2f\n", totalCash + totalBank)
	}
}

func main() {
	token := login()
	fmt.Printf("Login successful! Token: %s\n\n", token[:20]+"...")
	
	fmt.Println("=== TESTING CASH-BANK ACCOUNTS ENDPOINT ===")
	testEndpoint(token, "http://localhost:8080/api/v1/cash-bank/accounts")
}