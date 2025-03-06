package social

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RedditClient handles interactions with the Reddit API
type RedditClient struct {
	HTTPClient *http.Client
	BaseURL    string
}

// RedditPost represents a post from Reddit
type RedditPost struct {
	Title         string  `json:"title"`
	URL           string  `json:"url"`
	Permalink     string  `json:"permalink"`
	Score         int     `json:"score"`
	NumComments   int     `json:"num_comments"`
	Subreddit     string  `json:"subreddit"`
	Created       float64 `json:"created_utc"`
	IsVideo       bool    `json:"is_video"`
	Thumbnail     string  `json:"thumbnail"`
	SelfText      string  `json:"selftext"`
	Author        string  `json:"author"`
	PostHint      string  `json:"post_hint,omitempty"`
	Distinguished string  `json:"distinguished,omitempty"`
}

// RedditResponse represents the structure of the Reddit API response
type RedditResponse struct {
	Kind string `json:"kind"`
	Data struct {
		After    string `json:"after"`
		Children []struct {
			Kind string     `json:"kind"`
			Data RedditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// NewRedditClient creates a new Reddit API client
func NewRedditClient() *RedditClient {
	return &RedditClient{
		HTTPClient: &http.Client{
			Timeout: time.Second * 10,
		},
		BaseURL: "https://www.reddit.com",
	}
}

// GetTrending fetches trending posts from Reddit
// timeRange can be: hour, day, week, month, year, all
func (c *RedditClient) GetTrending(subreddit string, limit int, timeRange string) ([]RedditPost, error) {
	if subreddit == "" {
		subreddit = "popular" // Default to r/popular if no subreddit is specified
	}

	if limit <= 0 {
		limit = 25 // Default limit
	}

	if timeRange == "" {
		timeRange = "day" // Default time range
	}

	// Construct the URL for the Reddit API
	url := fmt.Sprintf("%s/r/%s/top.json?limit=%d&t=%s", c.BaseURL, subreddit, limit, timeRange)
	fmt.Printf("Making request to Reddit API: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set a User-Agent header to avoid rate limiting
	req.Header.Set("User-Agent", "essg-app/1.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Reddit API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Reddit API returned status code %d", resp.StatusCode)
	}

	var redditResp RedditResponse
	if err := json.NewDecoder(resp.Body).Decode(&redditResp); err != nil {
		return nil, fmt.Errorf("failed to decode Reddit API response: %w", err)
	}

	// Extract the posts from the response
	var posts []RedditPost
	for _, child := range redditResp.Data.Children {
		posts = append(posts, child.Data)
	}

	fmt.Printf("Received %d posts from Reddit API\n", len(posts))
	return posts, nil
}

// GetPopularSubreddits fetches popular subreddits
func (c *RedditClient) GetPopularSubreddits(limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 25 // Default limit
	}

	// Construct the URL for the Reddit API
	url := fmt.Sprintf("%s/subreddits/popular.json?limit=%d", c.BaseURL, limit)
	fmt.Printf("Making request to Reddit API: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set a User-Agent header to avoid rate limiting
	req.Header.Set("User-Agent", "essg-app/1.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Reddit API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Reddit API returned status code %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Reddit API response: %w", err)
	}

	// Extract the subreddits from the response
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format from Reddit API")
	}

	children, ok := data["children"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format from Reddit API")
	}

	var subreddits []map[string]interface{}
	for _, child := range children {
		childMap, ok := child.(map[string]interface{})
		if !ok {
			continue
		}

		subredditData, ok := childMap["data"].(map[string]interface{})
		if !ok {
			continue
		}

		subreddits = append(subreddits, subredditData)
	}

	fmt.Printf("Received %d subreddits from Reddit API\n", len(subreddits))
	return subreddits, nil
}

// SearchSubreddits searches for subreddits matching the query
func (c *RedditClient) SearchSubreddits(query string, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 25 // Default limit
	}

	// Construct the URL for the Reddit API
	url := fmt.Sprintf("%s/subreddits/search.json?q=%s&limit=%d", c.BaseURL, query, limit)
	fmt.Printf("Making request to Reddit API: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set a User-Agent header to avoid rate limiting
	req.Header.Set("User-Agent", "essg-app/1.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Reddit API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Reddit API returned status code %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Reddit API response: %w", err)
	}

	// Extract the subreddits from the response
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format from Reddit API")
	}

	children, ok := data["children"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format from Reddit API")
	}

	var subreddits []map[string]interface{}
	for _, child := range children {
		childMap, ok := child.(map[string]interface{})
		if !ok {
			continue
		}

		subredditData, ok := childMap["data"].(map[string]interface{})
		if !ok {
			continue
		}

		subreddits = append(subreddits, subredditData)
	}

	fmt.Printf("Received %d subreddits from Reddit API\n", len(subreddits))
	return subreddits, nil
}
