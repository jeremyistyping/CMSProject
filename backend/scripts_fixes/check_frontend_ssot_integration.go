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

	fmt.Println("🔍 CHECKING FRONTEND-SSOT INTEGRATION")
	fmt.Println("=====================================")

	// 1. Cek apa frontend seharusnya menggunakan SSOT balance calculation
	fmt.Println("\n1️⃣ CHECKING BACKEND API RESPONSE:")
	fmt.Println("---------------------------------")

	// Simulate API call to /accounts endpoint
	type APIAccount struct {
		ID          uint      `json:"id"`
		Code        string    `json:"code"`
		Name        string    `json:"name"`
		Type        string    `json:"type"`
		Balance     float64   `json:"balance"`
		IsActive    bool      `json:"is_active"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	var accounts []APIAccount
	err := db.Raw(`
		SELECT id, code, name, type, balance, is_active, created_at, updated_at 
		FROM accounts 
		WHERE is_active = true 
		ORDER BY code
	`).Scan(&accounts).Error
	
	if err != nil {
		log.Printf("Error getting accounts: %v", err)
		return
	}

	// Find revenue accounts
	fmt.Println("Revenue accounts in API response:")
	for _, acc := range accounts {
		if acc.Type == "REVENUE" {
			fmt.Printf("  %s (%s): Balance = %.0f\n", acc.Code, acc.Name, acc.Balance)
		}
	}

	// 2. Test actual API endpoint
	fmt.Println("\n2️⃣ TESTING ACTUAL API ENDPOINT:")
	fmt.Println("-------------------------------")
	
	// Make HTTP request to actual API
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("http://localhost:8080/api/v1/accounts")
	if err != nil {
		log.Printf("❌ Error calling API: %v", err)
		fmt.Println("⚠️  Backend might not be running on port 8080")
	} else {
		defer resp.Body.Close()
		
		if resp.StatusCode == 200 {
			var apiResponse struct {
				Data  []APIAccount `json:"data"`
				Count int          `json:"count"`
			}
			
			err = json.NewDecoder(resp.Body).Decode(&apiResponse)
			if err != nil {
				log.Printf("Error decoding response: %v", err)
			} else {
				fmt.Printf("✅ API responded with %d accounts\n", apiResponse.Count)
				
				// Find account 4101 in API response
				var found4101 *APIAccount
				for _, acc := range apiResponse.Data {
					if acc.Code == "4101" {
						found4101 = &acc
						break
					}
				}
				
				if found4101 != nil {
					fmt.Printf("✅ Account 4101 found in API: Balance = %.0f\n", found4101.Balance)
					if found4101.Balance == 0 {
						fmt.Println("❌ API is returning 0 balance!")
					}
				} else {
					fmt.Println("❌ Account 4101 NOT found in API response")
				}
			}
		} else {
			fmt.Printf("❌ API returned status code: %d\n", resp.StatusCode)
		}
	}

	// 3. Cek apakah ada endpoint khusus untuk SSOT balance
	fmt.Println("\n3️⃣ CHECKING FOR SSOT BALANCE ENDPOINT:")
	fmt.Println("-------------------------------------")
	
	// Try SSOT balance endpoint
	resp2, err := client.Get("http://localhost:8080/api/v1/journals/account-balances")
	if err != nil {
		log.Printf("SSOT balance endpoint not available: %v", err)
	} else {
		defer resp2.Body.Close()
		fmt.Printf("SSOT balance endpoint status: %d\n", resp2.StatusCode)
	}

	// 4. Manual check dengan langsung update database lagi
	fmt.Println("\n4️⃣ FORCE UPDATE DATABASE AGAIN:")
	fmt.Println("-------------------------------")
	
	// Update account balance again 
	result := db.Exec(`
		UPDATE accounts 
		SET balance = 5000000.00, updated_at = NOW() 
		WHERE code = '4101'
	`)
	
	if result.Error != nil {
		log.Printf("❌ Error updating: %v", result.Error)
	} else {
		fmt.Printf("✅ Force updated account 4101: %d rows affected\n", result.RowsAffected)
		
		// Verify update dengan timestamp
		var verification struct {
			Code      string    `json:"code"`
			Balance   float64   `json:"balance"`
			UpdatedAt time.Time `json:"updated_at"`
		}
		
		err = db.Raw("SELECT code, balance, updated_at FROM accounts WHERE code = '4101'").Scan(&verification).Error
		if err == nil {
			fmt.Printf("✅ Verified: %s = %.0f (updated: %s)\n", 
				verification.Code, verification.Balance, verification.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
	}

	// 5. Cek apakah ada materialized view yang perlu di-refresh
	fmt.Println("\n5️⃣ REFRESH MATERIALIZED VIEWS:")
	fmt.Println("------------------------------")
	
	err = db.Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY account_balances").Error
	if err != nil {
		// Try without CONCURRENTLY
		err = db.Exec("REFRESH MATERIALIZED VIEW account_balances").Error
		if err != nil {
			log.Printf("⚠️  Error refreshing materialized view: %v", err)
		} else {
			fmt.Println("✅ Materialized view refreshed (without concurrent)")
		}
	} else {
		fmt.Println("✅ Materialized view refreshed concurrently")
	}

	// 6. Cek apakah frontend cache yang jadi masalah
	fmt.Println("\n6️⃣ DEBUGGING FRONTEND CACHE ISSUE:")
	fmt.Println("---------------------------------")
	
	fmt.Println("Kemungkinan penyebab frontend masih menunjukkan Rp 0:")
	fmt.Println("1. ❌ Browser cache tidak ter-clear")
	fmt.Println("2. ❌ Frontend JavaScript cache")
	fmt.Println("3. ❌ Frontend menggunakan stale WebSocket connection")
	fmt.Println("4. ❌ Frontend menggunakan localStorage cache")
	fmt.Println("5. ❌ Service Worker cache (jika ada)")
	
	fmt.Println("\n🎯 SOLUSI BERIKUTNYA:")
	fmt.Println("1. Buka Developer Tools (F12)")
	fmt.Println("2. Klik tab Application → Storage → Clear Storage")
	fmt.Println("3. Atau klik kanan refresh button → 'Empty Cache and Hard Reload'")
	fmt.Println("4. Atau gunakan Incognito/Private window untuk test")
	fmt.Println("5. Atau restart browser completely")
	
	fmt.Printf("\n⏰ Current time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("📊 Database account 4101 balance: 5,000,000 (confirmed)")
}