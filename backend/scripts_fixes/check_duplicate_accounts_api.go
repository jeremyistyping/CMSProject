package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"app-sistem-akuntansi/database"
)

func main() {
	// Initialize database connection
	db := database.ConnectDB()

	fmt.Println("🔍 CHECKING FOR DUPLICATE ACCOUNTS & API RESPONSE")
	fmt.Println("=================================================")

	// 1. Check for all accounts with code 4101 (including inactive)
	fmt.Println("\n1️⃣ ALL ACCOUNTS WITH CODE '4101' (INCLUDING INACTIVE):")
	fmt.Println("----------------------------------------------------")
	
	type FullAccount struct {
		ID        uint      `json:"id"`
		Code      string    `json:"code"`
		Name      string    `json:"name"`
		Type      string    `json:"type"`
		Balance   float64   `json:"balance"`
		IsActive  bool      `json:"is_active"`
		UpdatedAt time.Time `json:"updated_at"`
		DeletedAt *time.Time `json:"deleted_at"`
	}
	
	var allAccounts []FullAccount
	err := db.Raw(`
		SELECT id, code, name, type, balance, is_active, updated_at, deleted_at 
		FROM accounts 
		WHERE code = '4101'
		ORDER BY id
	`).Scan(&allAccounts).Error
	
	if err != nil {
		log.Printf("Error getting all 4101 accounts: %v", err)
		return
	}
	
	fmt.Printf("Found %d accounts with code '4101':\n", len(allAccounts))
	for i, acc := range allAccounts {
		deletedStatus := "ACTIVE"
		if acc.DeletedAt != nil {
			deletedStatus = "DELETED"
		}
		fmt.Printf("  %d. ID:%d Balance:%.0f IsActive:%v %s Updated:%s\n", 
			i+1, acc.ID, acc.Balance, acc.IsActive, deletedStatus, acc.UpdatedAt.Format("2006-01-02 15:04:05"))
	}

	// 2. Check what the actual API endpoint returns (simulate)
	fmt.Println("\n2️⃣ SIMULATING EXACT API ENDPOINT RESPONSE:")
	fmt.Println("-----------------------------------------")
	
	// Simulate the exact query that AccountHandler.ListAccounts uses
	var apiAccounts []FullAccount
	err = db.Raw(`
		SELECT id, code, name, type, balance, is_active, updated_at, deleted_at 
		FROM accounts 
		WHERE is_active = true AND deleted_at IS NULL
		ORDER BY code
	`).Scan(&apiAccounts).Error
	
	if err != nil {
		log.Printf("Error simulating API: %v", err)
		return
	}
	
	// Find account 4101 in API results
	var found4101 *FullAccount
	for _, acc := range apiAccounts {
		if acc.Code == "4101" {
			found4101 = &acc
			break
		}
	}
	
	if found4101 != nil {
		fmt.Printf("✅ Account 4101 found in API simulation:\n")
		fmt.Printf("   ID: %d\n", found4101.ID)
		fmt.Printf("   Balance: %.0f\n", found4101.Balance)
		fmt.Printf("   IsActive: %v\n", found4101.IsActive)
		fmt.Printf("   UpdatedAt: %s\n", found4101.UpdatedAt.Format("2006-01-02 15:04:05"))
		
		if found4101.Balance == 0 {
			fmt.Println("❌ API simulation shows 0 balance!")
		}
	} else {
		fmt.Println("❌ Account 4101 NOT found in API simulation")
		fmt.Println("   This could be why frontend shows nothing")
	}

	// 3. Test actual live API call
	fmt.Println("\n3️⃣ TESTING LIVE API ENDPOINT:")
	fmt.Println("-----------------------------")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("http://localhost:8080/api/v1/accounts")
	if err != nil {
		fmt.Printf("❌ Error calling live API: %v\n", err)
		fmt.Println("   Backend might be down or on different port")
	} else {
		defer resp.Body.Close()
		
		fmt.Printf("API Status Code: %d\n", resp.StatusCode)
		
		if resp.StatusCode == 200 {
			var apiResponse struct {
				Data  []map[string]interface{} `json:"data"`
				Count int                      `json:"count"`
			}
			
			err = json.NewDecoder(resp.Body).Decode(&apiResponse)
			if err != nil {
				fmt.Printf("❌ Error decoding response: %v\n", err)
			} else {
				fmt.Printf("✅ API returned %d accounts\n", apiResponse.Count)
				
				// Find account 4101 in API response
				var foundInAPI map[string]interface{}
				for _, acc := range apiResponse.Data {
					if code, ok := acc["code"].(string); ok && code == "4101" {
						foundInAPI = acc
						break
					}
				}
				
				if foundInAPI != nil {
					fmt.Printf("✅ Account 4101 found in live API:\n")
					if balance, ok := foundInAPI["balance"].(float64); ok {
						fmt.Printf("   Balance: %.0f\n", balance)
						if balance == 0 {
							fmt.Println("❌ LIVE API RETURNS 0 BALANCE!")
							fmt.Println("   This confirms the issue is in backend data")
						} else {
							fmt.Println("✅ Live API returns correct balance")
							fmt.Println("   Frontend must have different issue")
						}
					}
					fmt.Printf("   Full data: %+v\n", foundInAPI)
				} else {
					fmt.Println("❌ Account 4101 NOT found in live API response")
				}
			}
		} else if resp.StatusCode == 401 {
			fmt.Println("❌ API requires authentication")
			fmt.Println("   Need to test with valid JWT token")
		}
	}

	// 4. Check if there are any GORM model issues
	fmt.Println("\n4️⃣ CHECKING GORM MODEL MAPPING:")
	fmt.Println("-------------------------------")
	
	// Check if the Account model has proper field mapping
	type TestAccount struct {
		ID       uint    `json:"id" gorm:"primaryKey"`
		Code     string  `json:"code" gorm:"uniqueIndex"`
		Name     string  `json:"name"`
		Type     string  `json:"type" gorm:"column:type"`
		Balance  float64 `json:"balance" gorm:"column:balance"`
		IsActive bool    `json:"is_active" gorm:"column:is_active"`
	}
	
	var testAccount TestAccount
	err = db.Table("accounts").Where("code = ?", "4101").First(&testAccount).Error
	if err != nil {
		fmt.Printf("❌ GORM model test failed: %v\n", err)
	} else {
		fmt.Printf("✅ GORM model test:\n")
		fmt.Printf("   Balance: %.0f\n", testAccount.Balance)
		
		if testAccount.Balance == 0 {
			fmt.Println("❌ GORM returns 0 balance!")
			fmt.Println("   Issue might be in GORM field mapping or column name")
		}
	}

	// 5. Force refresh all balance calculations
	fmt.Println("\n5️⃣ FORCE REFRESH ALL CALCULATIONS:")
	fmt.Println("---------------------------------")
	
	// Update with a slightly different value to trigger change
	result := db.Exec(`
		UPDATE accounts 
		SET balance = 5000001.00, 
		    updated_at = NOW() 
		WHERE code = '4101'
	`)
	
	if result.Error != nil {
		fmt.Printf("❌ Error updating: %v\n", result.Error)
	} else {
		fmt.Printf("✅ Updated account 4101 balance to 5,000,001\n")
		
		// Wait a moment for any async processes
		time.Sleep(1 * time.Second)
		
		// Verify
		var newBalance float64
		err = db.Raw("SELECT balance FROM accounts WHERE code = '4101'").Scan(&newBalance).Error
		if err == nil {
			fmt.Printf("✅ Verified new balance: %.0f\n", newBalance)
		}
	}

	fmt.Println("\n🎯 ACTION ITEMS:")
	fmt.Println("1. Check if live API returns correct balance")
	fmt.Println("2. If API correct but frontend wrong → frontend JavaScript issue")
	fmt.Println("3. If API wrong → backend GORM/model mapping issue") 
	fmt.Println("4. Check network tab in browser DevTools for actual API response")
	fmt.Println("5. Check if frontend uses different endpoint")
}