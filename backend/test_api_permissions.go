package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	baseURL := "http://localhost:8080/api/v1"

	// 1. Login
	loginPayload := map[string]string{
		"email":    "purchasing@company.com",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(loginPayload)

	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Printf("Login failed with status %d: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var loginResp map[string]interface{}
	json.Unmarshal(body, &loginResp)
	token := loginResp["token"].(string)
	fmt.Println("Login successful, token received.")

	// 2. Get Permissions
	req, _ := http.NewRequest("GET", baseURL+"/permissions/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("Permissions request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("Permissions Response Status: %d\n", resp.StatusCode)

	// Pretty print JSON
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, body, "", "  ")
	fmt.Println(prettyJSON.String())
}
