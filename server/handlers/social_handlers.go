package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"essg/server/services/social"

	"github.com/gorilla/mux"
)

// SocialTrendResponse is the standardized response format for social trends
type SocialTrendResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Query       string `json:"query"`
	TweetVolume int    `json:"tweet_volume"`
	Source      string `json:"source"`
	URL         string `json:"url,omitempty"`
}

// SocialHandler handles requests related to social media trends
type SocialHandler struct {
	TwitterClient *social.TwitterClient
	// Add more social media clients as needed
}

// NewSocialHandler creates a new social handler
func NewSocialHandler() *SocialHandler {
	return &SocialHandler{
		TwitterClient: social.NewTwitterClient(),
	}
}

// RegisterRoutes registers the social API routes
func (h *SocialHandler) RegisterRoutes(r *mux.Router) {
	fmt.Println("Registering social API routes...")
	r.HandleFunc("/api/social/trends", h.GetSocialTrends).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/social/locations", h.GetAvailableLocations).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/social/debug", h.DebugTwitterAPI).Methods("GET", "OPTIONS")
	fmt.Println("Social API routes registered.")
}

// GetSocialTrends handles requests to get trending topics from social media
func (h *SocialHandler) GetSocialTrends(w http.ResponseWriter, r *http.Request) {
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	source := r.URL.Query().Get("source")
	location := r.URL.Query().Get("location")

	// Log the request for debugging
	fmt.Printf("GetSocialTrends request: source=%s, location=%s\n", source, location)

	// Check if bearer token is set
	bearerToken := os.Getenv("TWITTER_BEARER_TOKEN")
	if bearerToken == "" {
		fmt.Println("Error: Twitter bearer token not configured")
		http.Error(w, "Twitter bearer token not configured", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Twitter Bearer Token: %s...\n", bearerToken[:min(10, len(bearerToken))]+"...")

	// Default to Twitter if no source specified
	if source == "" || source == "twitter" {
		trends, err := h.getTwitterTrends(location)
		if err != nil {
			fmt.Printf("Error fetching Twitter trends: %v\n", err)
			http.Error(w, "Failed to fetch Twitter trends: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("Successfully fetched %d Twitter trends\n", len(trends))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trends)
		return
	}

	// Handle other sources as they are implemented
	http.Error(w, "Unsupported social media source", http.StatusBadRequest)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getTwitterTrends fetches trending topics from Twitter
func (h *SocialHandler) getTwitterTrends(location string) ([]SocialTrendResponse, error) {
	fmt.Printf("Fetching Twitter trends for location: %s\n", location)
	twitterTrends, err := h.TwitterClient.GetTrends(location)
	if err != nil {
		fmt.Printf("Error in GetTrends: %v\n", err)
		return nil, err
	}

	fmt.Printf("Received %d raw Twitter trends\n", len(twitterTrends))

	var trends []SocialTrendResponse
	for i, trend := range twitterTrends {
		// Skip trends with no tweet volume or very low volume
		if trend.TweetVolume < 1000 && trend.TweetVolume != 0 {
			continue
		}

		trends = append(trends, SocialTrendResponse{
			ID:          strconv.Itoa(i),
			Name:        trend.Name,
			Query:       trend.Query,
			TweetVolume: trend.TweetVolume,
			Source:      "twitter",
			URL:         trend.URL,
		})

		// Limit to top 20 trends
		if len(trends) >= 20 {
			break
		}
	}

	// If no trends with volume, include some without volume
	if len(trends) == 0 {
		fmt.Println("No trends with sufficient volume, including some without volume")
		for i, trend := range twitterTrends {
			if i >= 10 {
				break
			}

			trends = append(trends, SocialTrendResponse{
				ID:          strconv.Itoa(i),
				Name:        trend.Name,
				Query:       trend.Query,
				TweetVolume: trend.TweetVolume,
				Source:      "twitter",
				URL:         trend.URL,
			})
		}
	}

	fmt.Printf("Returning %d filtered Twitter trends\n", len(trends))
	return trends, nil
}

// GetAvailableLocations handles requests to get available locations for trends
func (h *SocialHandler) GetAvailableLocations(w http.ResponseWriter, r *http.Request) {
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	source := r.URL.Query().Get("source")

	// Default to Twitter if no source specified
	if source == "" || source == "twitter" {
		locations, err := h.TwitterClient.GetAvailableLocations()
		if err != nil {
			http.Error(w, "Failed to fetch Twitter locations: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(locations)
		return
	}

	// Handle other sources as they are implemented
	http.Error(w, "Unsupported social media source", http.StatusBadRequest)
}

// DebugTwitterAPI provides a debug endpoint to test Twitter API connectivity
func (h *SocialHandler) DebugTwitterAPI(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Debug Twitter API endpoint called")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check if bearer token is set
	bearerToken := os.Getenv("TWITTER_BEARER_TOKEN")
	if bearerToken == "" {
		fmt.Println("Error: Twitter bearer token not configured")
		http.Error(w, "Twitter bearer token not configured", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Twitter Bearer Token: %s...\n", bearerToken[:min(10, len(bearerToken))]+"...")

	// Make a simple request to the Twitter API to test connectivity
	url := "https://api.twitter.com/1.1/trends/place.json?id=1" // Worldwide trends
	fmt.Printf("Making request to Twitter API: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Add("Authorization", "Bearer "+bearerToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error connecting to Twitter API: %v\n", err)
		http.Error(w, "Failed to connect to Twitter API: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		http.Error(w, "Failed to read response body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Twitter API response status: %d\n", resp.StatusCode)

	// Return debug information
	debugInfo := map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
		"body":        string(body),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(debugInfo)
}
