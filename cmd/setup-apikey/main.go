package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "usage: setup-apikey <base-url> <email> <password>\n")
		os.Exit(1)
	}
	baseURL, email, password := os.Args[1], os.Args[2], os.Args[3]

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 10 * time.Second}

	// Retry sign-in up to 5 times (Portabase may still be seeding)
	var signInResp *http.Response
	var err error
	for attempt := 1; attempt <= 5; attempt++ {
		body := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
		req, _ := http.NewRequest("POST", baseURL+"/api/auth/sign-in/email", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", baseURL)
		signInResp, err = client.Do(req)
		if err == nil && signInResp.StatusCode < 500 {
			break
		}
		if signInResp != nil {
			signInResp.Body.Close()
		}
		fmt.Fprintf(os.Stderr, "sign-in attempt %d failed (err=%v, status=%d), retrying...\n", attempt, err, func() int {
			if signInResp != nil {
				return signInResp.StatusCode
			}
			return 0
		}())
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "sign-in failed after retries: %v\n", err)
		os.Exit(1)
	}
	defer signInResp.Body.Close()

	if signInResp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(signInResp.Body)
		fmt.Fprintf(os.Stderr, "sign-in returned %d: %s\n", signInResp.StatusCode, string(respBody))
		os.Exit(1)
	}

	// The cookie jar automatically captures the session cookie.
	// Also check response body for token (some better-auth configs return it inline).
	var signInBody map[string]any
	json.NewDecoder(signInResp.Body).Decode(&signInBody)

	// Create API key using the session (cookie jar handles auth)
	req, _ := http.NewRequest("POST", baseURL+"/api/auth/api-key/create",
		strings.NewReader(`{"name":"e2e-test-key","configId":"standard"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", baseURL)

	// If we got a token in the body, also set it as bearer
	if token, ok := signInBody["token"].(string); ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create api-key request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "create api-key returned %d: %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var keyResp map[string]any
	json.NewDecoder(resp.Body).Decode(&keyResp)

	if key, ok := keyResp["key"].(string); ok && key != "" {
		fmt.Print(key)
		return
	}

	fmt.Fprintf(os.Stderr, "no key in response: %v\n", keyResp)
	os.Exit(1)
}
