package social

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// TwitterClient handles interactions with the Twitter API
type TwitterClient struct {
	BearerToken string
	BaseURL     string
	HTTPClient  *http.Client
}

// TwitterTrend represents a trending topic from Twitter
type TwitterTrend struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Query       string `json:"query"`
	TweetVolume int    `json:"tweet_volume"`
}

// TwitterTrendsResponse represents the response from Twitter's trends/place endpoint
type TwitterTrendsResponse []struct {
	Trends    []TwitterTrend `json:"trends"`
	AsOf      time.Time      `json:"as_of"`
	CreatedAt time.Time      `json:"created_at"`
	Locations []struct {
		Name  string `json:"name"`
		WoeID int    `json:"woeid"`
	} `json:"locations"`
}

// NewTwitterClient creates a new Twitter API client
func NewTwitterClient() *TwitterClient {
	return &TwitterClient{
		BearerToken: os.Getenv("TWITTER_BEARER_TOKEN"),
		BaseURL:     "https://api.twitter.com/1.1",
		HTTPClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// GetTrends fetches trending topics from Twitter for a specific location
func (c *TwitterClient) GetTrends(woeid string) ([]TwitterTrend, error) {
	if c.BearerToken == "" {
		return nil, fmt.Errorf("Twitter bearer token not configured")
	}

	// Default to worldwide (woeid: 1) if not specified
	if woeid == "" {
		woeid = "1"
	}

	url := fmt.Sprintf("%s/trends/place.json?id=%s", c.BaseURL, woeid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.BearerToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Twitter API returned status code %d", resp.StatusCode)
	}

	var trendsResp TwitterTrendsResponse
	if err := json.NewDecoder(resp.Body).Decode(&trendsResp); err != nil {
		return nil, err
	}

	if len(trendsResp) == 0 {
		return []TwitterTrend{}, nil
	}

	return trendsResp[0].Trends, nil
}

// GetAvailableLocations fetches the locations that Twitter has trending topic information for
func (c *TwitterClient) GetAvailableLocations() ([]map[string]interface{}, error) {
	if c.BearerToken == "" {
		return nil, fmt.Errorf("Twitter bearer token not configured")
	}

	url := fmt.Sprintf("%s/trends/available.json", c.BaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.BearerToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Twitter API returned status code %d", resp.StatusCode)
	}

	var locations []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&locations); err != nil {
		return nil, err
	}

	return locations, nil
}

// GetClosestLocation finds the closest location to the given coordinates
func (c *TwitterClient) GetClosestLocation(lat, lng float64) (string, error) {
	if c.BearerToken == "" {
		return "", fmt.Errorf("Twitter bearer token not configured")
	}

	url := fmt.Sprintf("%s/trends/closest.json?lat=%f&long=%f", c.BaseURL, lat, lng)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+c.BearerToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Twitter API returned status code %d", resp.StatusCode)
	}

	var locations []struct {
		WoeID int `json:"woeid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&locations); err != nil {
		return "", err
	}

	if len(locations) == 0 {
		return "1", nil // Default to worldwide
	}

	return fmt.Sprintf("%d", locations[0].WoeID), nil
}
