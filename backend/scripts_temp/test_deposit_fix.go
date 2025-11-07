package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type CashBankAccount struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

type DepositRequest struct {
	AccountID uint   `json:"account_id"`
	Date      string `json:"date"`
	Amount    float64 `json:"amount"`
	Notes     string `json:"notes"`
}

func main() {
	fmt.Println("🧪 Testing Deposit Bug Fix")
	
	// Step 1: Login to get token
	fmt.Println("1. Logging in...")
	token, err := login()
	if err != nil {
		fmt.Printf("❌ Login failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Login successful, token: %s...\n", token[:20])

	// Step 2: Get cash bank accounts
	fmt.Println("2. Getting cash bank accounts...")
	accounts, err := getCashBankAccounts(token)
	if err != nil {
		fmt.Printf("❌ Failed to get accounts: %v\n", err)
		return
	}

	if len(accounts) == 0 {
		fmt.Println("❌ No cash bank accounts found")
		return
	}

	// Use the first account for testing
	testAccount := accounts[0]
	fmt.Printf("✅ Found %d accounts. Testing with: %s (ID: %d, Balance: %.2f)\n", 
		len(accounts), testAccount.Name, testAccount.ID, testAccount.Balance)

	// Step 3: Record original balance
	originalBalance := testAccount.Balance
	fmt.Printf("📊 Original balance: %.2f\n", originalBalance)

	// Step 4: Make a deposit
	depositAmount := 100.0
	fmt.Printf("3. Making deposit of %.2f...\n", depositAmount)
	
	err = makeDeposit(token, testAccount.ID, depositAmount)
	if err != nil {
		fmt.Printf("❌ Deposit failed: %v\n", err)
		return
	}
	fmt.Println("✅ Deposit request completed successfully")

	// Step 5: Wait a moment and check new balance
	fmt.Println("4. Waiting 2 seconds then checking balance...")
	time.Sleep(2 * time.Second)

	// Get updated account balance
	updatedAccounts, err := getCashBankAccounts(token)
	if err != nil {
		fmt.Printf("❌ Failed to get updated accounts: %v\n", err)
		return
	}

	var updatedAccount *CashBankAccount
	for _, acc := range updatedAccounts {
		if acc.ID == testAccount.ID {
			updatedAccount = &acc
			break
		}
	}

	if updatedAccount == nil {
		fmt.Println("❌ Could not find test account in updated list")
		return
	}

	// Step 6: Verify the fix
	expectedBalance := originalBalance + depositAmount
	actualBalance := updatedAccount.Balance
	
	fmt.Printf("📊 Balance check:\n")
	fmt.Printf("   Original: %.2f\n", originalBalance)
	fmt.Printf("   Deposit:  %.2f\n", depositAmount)
	fmt.Printf("   Expected: %.2f\n", expectedBalance)
	fmt.Printf("   Actual:   %.2f\n", actualBalance)

	if actualBalance == expectedBalance {
		fmt.Printf("✅ SUCCESS! Deposit bug is FIXED! Balance updated correctly from %.2f to %.2f\n", 
			originalBalance, actualBalance)
	} else {
		fmt.Printf("❌ FAILED! Balance bug still exists. Expected %.2f but got %.2f\n", 
			expectedBalance, actualBalance)
	}
}

func login() (string, error) {
	loginReq := LoginRequest{
		Username: "admin@company.com", // Correct admin user
		Password: "password123", // Correct password
	}

	jsonData, _ := json.Marshal(loginReq)
	resp, err := http.Post("http://localhost:8080/api/v1/auth/login", 
		"application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", err
	}

	return loginResp.Token, nil
}

func getCashBankAccounts(token string) ([]CashBankAccount, error) {
	req, err := http.NewRequest("GET", "http://localhost:8080/api/v1/cashbank/accounts", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get accounts failed with status %d: %s", resp.StatusCode, string(body))
	}

	var accounts []CashBankAccount
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, err
	}

	return accounts, nil
}

func makeDeposit(token string, accountID int, amount float64) error {
	depositReq := DepositRequest{
		AccountID: uint(accountID),
		Date:      time.Now().Format("2006-01-02"),
		Amount:    amount,
		Notes:     "Test deposit for bug fix verification",
	}

	jsonData, _ := json.Marshal(depositReq)
	
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/cashbank/deposit", 
		bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("deposit failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}