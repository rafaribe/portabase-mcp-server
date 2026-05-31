package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "usage: setup-apikey <base-url> <email> <password>\n")
		os.Exit(1)
	}
	baseURL, email, password := os.Args[1], os.Args[2], os.Args[3]

	// Sign in
	body := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
	resp, err := http.Post(baseURL+"/api/auth/sign-in/email", "application/json", strings.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "sign-in failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "sign-in returned %d\n", resp.StatusCode)
		os.Exit(1)
	}

	// Extract session cookie
	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if strings.Contains(c.Name, "session") {
			sessionCookie = c
			break
		}
	}

	// Also try response body for token
	var signInBody map[string]any
	json.NewDecoder(resp.Body).Decode(&signInBody)

	if sessionCookie == nil {
		// Try bearer token from body
		if token, ok := signInBody["token"].(string); ok && token != "" {
			sessionCookie = &http.Cookie{Name: "better-auth.session_token", Value: token}
		}
	}

	if sessionCookie == nil {
		fmt.Fprintf(os.Stderr, "no session in sign-in response, cookies: %v, body: %v\n", resp.Cookies(), signInBody)
		os.Exit(1)
	}

	// Create API key
	req, _ := http.NewRequest("POST", baseURL+"/api/auth/api-key/create",
		strings.NewReader(`{"name":"e2e-test-key","configId":"standard"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)

	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create api-key failed: %v\n", err)
		os.Exit(1)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "create api-key returned %d\n", resp2.StatusCode)
		os.Exit(1)
	}

	var keyResp map[string]any
	json.NewDecoder(resp2.Body).Decode(&keyResp)

	if key, ok := keyResp["key"].(string); ok && key != "" {
		fmt.Print(key)
		return
	}

	fmt.Fprintf(os.Stderr, "no key in response: %v\n", keyResp)
	os.Exit(1)
}
