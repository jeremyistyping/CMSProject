package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
	User    struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"user"`
}

type AssetCategoryRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func main() {
	baseURL := "http://localhost:8080/api"
	
	// Step 1: Login to get token
	fmt.Println("=== LOGIN TEST ===")
	token, err := login(baseURL)
	if err != nil {
		log.Fatal("Login failed:", err)
	}
	fmt.Printf("✅ Login successful, token obtained\n\n")

	// Step 2: Get existing asset categories
	fmt.Println("=== GET ASSET CATEGORIES (BEFORE) ===")
	categories, err := getAssetCategories(baseURL, token)
	if err != nil {
		log.Printf("❌ Failed to get asset categories: %v", err)
	} else {
		fmt.Printf("✅ Found %d asset categories\n", len(categories))
	}

	// Step 3: Create new asset category (similar to what user did)
	fmt.Println("\n=== CREATE ASSET CATEGORY ===")
	testCategory := AssetCategoryRequest{
		Code:        "TEST01",
		Name:        "test01", // Same as user's test
		Description: "Testing category persistence",
		IsActive:    true,
	}

	createdID, err := createAssetCategory(baseURL, token, testCategory)
	if err != nil {
		log.Printf("❌ Failed to create asset category: %v", err)
	} else {
		fmt.Printf("✅ Asset category created with ID: %d\n", createdID)
	}

	// Step 4: Verify creation
	fmt.Println("\n=== GET ASSET CATEGORIES (AFTER CREATE) ===")
	categories, err = getAssetCategories(baseURL, token)
	if err != nil {
		log.Printf("❌ Failed to get asset categories: %v", err)
	} else {
		fmt.Printf("✅ Found %d asset categories\n", len(categories))
		
		// Check if our test category exists
		found := false
		for _, cat := range categories {
			if name, ok := cat["name"].(string); ok && name == "test01" {
				fmt.Printf("✅ Test category 'test01' found in database\n")
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("❌ Test category 'test01' NOT found in database\n")
		}
	}

	// Wait a moment
	fmt.Println("\n=== WAITING (Simulating browser refresh) ===")
	time.Sleep(2 * time.Second)

	// Step 5: Get asset categories again (simulating browser refresh)
	fmt.Println("\n=== GET ASSET CATEGORIES (AFTER REFRESH) ===")
	categories, err = getAssetCategories(baseURL, token)
	if err != nil {
		log.Printf("❌ Failed to get asset categories: %v", err)
	} else {
		fmt.Printf("✅ Found %d asset categories\n", len(categories))
		
		// Check if our test category still exists
		found := false
		for _, cat := range categories {
			if name, ok := cat["name"].(string); ok && name == "test01" {
				fmt.Printf("✅ Test category 'test01' STILL found after refresh - PERSISTENT!\n")
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("❌ Test category 'test01' DISAPPEARED after refresh - THIS IS THE BUG!\n")
		}
	}

	fmt.Println("\n=== TEST COMPLETE ===")
}

func login(baseURL string) (string, error) {
	loginReq := LoginRequest{
		Email:    "admin@example.com",
		Password: "admin123",
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	err = json.Unmarshal(body, &loginResp)
	if err != nil {
		return "", err
	}

	return loginResp.Token, nil
}

func getAssetCategories(baseURL, token string) ([]map[string]interface{}, error) {
	req, err := http.NewRequest("GET", baseURL+"/v1/assets/categories", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get categories failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	categories := make([]map[string]interface{}, len(data))
	for i, item := range data {
		if cat, ok := item.(map[string]interface{}); ok {
			categories[i] = cat
		}
	}

	return categories, nil
}

func createAssetCategory(baseURL, token string, category AssetCategoryRequest) (int, error) {
	jsonData, err := json.Marshal(category)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", baseURL+"/v1/assets/categories", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Create response status: %d\n", resp.StatusCode)
	fmt.Printf("Create response body: %s\n", string(body))

	if resp.StatusCode != 201 {
		return 0, fmt.Errorf("create category failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("unexpected response format")
	}

	id, ok := data["id"].(float64)
	if !ok {
		return 0, fmt.Errorf("could not extract ID from response")
	}

	return int(id), nil
}