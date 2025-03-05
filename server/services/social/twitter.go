package social

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	twitter "github.com/g8rswimmer/go-twitter/v2"
)

// TwitterClient handles interactions with the Twitter API
type TwitterClient struct {
	client *twitter.Client
}

// TwitterTrend represents a trending topic from Twitter
type TwitterTrend struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Query       string `json:"query"`
	TweetVolume int    `json:"tweet_volume"`
}

// twitterAuth implements the twitter.Authorizer interface
type twitterAuth struct {
	Token string
}

func (a twitterAuth) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

// NewTwitterClient creates a new Twitter API client
func NewTwitterClient() *TwitterClient {
	bearerToken := os.Getenv("TWITTER_BEARER_TOKEN")
	if bearerToken == "" {
		fmt.Println("Warning: TWITTER_BEARER_TOKEN environment variable is not set")
		return &TwitterClient{}
	}

	// Print a truncated version of the token for debugging
	tokenPreview := bearerToken
	if len(bearerToken) > 10 {
		tokenPreview = bearerToken[:10] + "..."
	}
	fmt.Printf("Initializing Twitter client with token: %s\n", tokenPreview)

	client := &twitter.Client{
		Authorizer: twitterAuth{
			Token: bearerToken,
		},
		Client: &http.Client{
			Timeout: time.Second * 10,
		},
		Host: "https://api.twitter.com",
	}

	return &TwitterClient{
		client: client,
	}
}

// GetTrends fetches trending topics from Twitter for a specific location
// Note: Twitter v2 API doesn't have a direct trends endpoint, so we're still using v1.1
// This is a custom implementation that doesn't use the go-twitter v2 package
func (c *TwitterClient) GetTrends(woeid string) ([]TwitterTrend, error) {
	bearerToken := os.Getenv("TWITTER_BEARER_TOKEN")
	if bearerToken == "" {
		return nil, fmt.Errorf("Twitter bearer token not configured")
	}

	// Default to worldwide (woeid: 1) if not specified
	if woeid == "" {
		woeid = "1"
	}

	url := fmt.Sprintf("https://api.twitter.com/1.1/trends/place.json?id=%s", woeid)
	fmt.Printf("Making request to Twitter API: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+bearerToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Twitter API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Twitter API returned status code %d", resp.StatusCode)
	}

	// Parse the response
	var trendsResp []struct {
		Trends []TwitterTrend `json:"trends"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&trendsResp); err != nil {
		return nil, fmt.Errorf("failed to decode Twitter API response: %w", err)
	}

	if len(trendsResp) == 0 {
		return []TwitterTrend{}, nil
	}

	fmt.Printf("Received %d trends from Twitter API\n", len(trendsResp[0].Trends))
	return trendsResp[0].Trends, nil
}

// GetAvailableLocations fetches the locations that Twitter has trending topic information for
// Note: Twitter v2 API doesn't have a direct locations endpoint, so we're still using v1.1
// This is a custom implementation that doesn't use the go-twitter v2 package
func (c *TwitterClient) GetAvailableLocations() ([]map[string]interface{}, error) {
	bearerToken := os.Getenv("TWITTER_BEARER_TOKEN")
	if bearerToken == "" {
		return nil, fmt.Errorf("Twitter bearer token not configured")
	}

	url := "https://api.twitter.com/1.1/trends/available.json"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+bearerToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Twitter API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Twitter API returned status code %d", resp.StatusCode)
	}

	var locations []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&locations); err != nil {
		return nil, fmt.Errorf("failed to decode Twitter API response: %w", err)
	}

	return locations, nil
}

// GetClosestLocation finds the closest location to the given coordinates
// Note: Twitter v2 API doesn't have a direct closest location endpoint, so we're still using v1.1
// This is a custom implementation that doesn't use the go-twitter v2 package
func (c *TwitterClient) GetClosestLocation(lat, lng float64) (string, error) {
	bearerToken := os.Getenv("TWITTER_BEARER_TOKEN")
	if bearerToken == "" {
		return "", fmt.Errorf("Twitter bearer token not configured")
	}

	url := fmt.Sprintf("https://api.twitter.com/1.1/trends/closest.json?lat=%f&long=%f", lat, lng)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+bearerToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Twitter API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Twitter API returned status code %d", resp.StatusCode)
	}

	var locations []struct {
		WoeID int `json:"woeid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&locations); err != nil {
		return "", fmt.Errorf("failed to decode Twitter API response: %w", err)
	}

	if len(locations) == 0 {
		return "1", nil // Default to worldwide
	}

	return fmt.Sprintf("%d", locations[0].WoeID), nil
}

// SearchTweets searches for tweets using the Twitter v2 API
func (c *TwitterClient) SearchTweets(query string, maxResults int) (interface{}, error) {
	if c.client == nil {
		return nil, fmt.Errorf("Twitter client not initialized (bearer token missing)")
	}

	opts := twitter.TweetSearchOpts{
		MaxResults: maxResults,
		TweetFields: []twitter.TweetField{
			twitter.TweetFieldCreatedAt,
			twitter.TweetFieldPublicMetrics,
		},
	}

	tweetResponse, err := c.client.TweetSearch(context.Background(), query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search tweets: %w", err)
	}

	return tweetResponse, nil
}

// GetUserByUsername gets a Twitter user by username using the Twitter v2 API
func (c *TwitterClient) GetUserByUsername(username string) (interface{}, error) {
	if c.client == nil {
		return nil, fmt.Errorf("Twitter client not initialized (bearer token missing)")
	}

	opts := twitter.UserLookupOpts{
		UserFields: []twitter.UserField{
			twitter.UserFieldDescription,
			twitter.UserFieldProfileImageURL,
			twitter.UserFieldPublicMetrics,
		},
	}

	userResponse, err := c.client.UserNameLookup(context.Background(), []string{username}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup user: %w", err)
	}

	return userResponse, nil
}
