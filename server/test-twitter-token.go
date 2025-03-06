package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	// Load environment variables from .env file
	loadEnvFromFile()

	// Get bearer token
	bearerToken := os.Getenv("TWITTER_BEARER_TOKEN")
	if bearerToken == "" {
		fmt.Println("Error: TWITTER_BEARER_TOKEN environment variable is not set")
		return
	}

	// Print token length and preview
	tokenPreview := bearerToken
	if len(bearerToken) > 10 {
		tokenPreview = bearerToken[:10] + "..."
	}
	fmt.Printf("Testing Twitter API with token: %s (length: %d)\n", tokenPreview, len(bearerToken))

	// Try a simple, public endpoint
	// This is the GET /2/tweets/search/recent endpoint with minimal query
	url := "https://api.twitter.com/2/tweets/search/recent?query=twitter&max_results=10&tweet.fields=public_metrics"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+bearerToken)
	req.Header.Add("User-Agent", "PostmanRuntime/7.32.3") // Adding a user agent to avoid some blocks

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error connecting to Twitter API: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("Twitter API response status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))
}

// loadEnvFromFile loads environment variables from .env file
func loadEnvFromFile() {
	// Try to find and load the .env file
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		// If not in current directory, try server directory
		envFile = "server/.env"
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			// If not in server directory, try parent directory
			envFile = "../.env"
			if _, err := os.Stat(envFile); os.IsNotExist(err) {
				// Give up if we can't find it
				fmt.Println("Warning: Could not find .env file")
				return
			}
		}
	}

	fmt.Printf("Loading environment variables from %s\n", envFile)

	// Read and parse .env file
	content, err := ioutil.ReadFile(envFile)
	if err != nil {
		fmt.Printf("Error reading .env file: %v\n", err)
		return
	}

	// Process each line
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip comments and empty lines
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only set if not already set in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
			if strings.Contains(key, "TOKEN") || strings.Contains(key, "KEY") || strings.Contains(key, "SECRET") {
				fmt.Printf("Loaded credential: %s (length: %d)\n", key, len(value))
			} else {
				fmt.Printf("Loaded environment variable: %s=%s\n", key, value)
			}
		}
	}
}
