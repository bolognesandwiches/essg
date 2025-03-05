package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bolognesandwiches/essg/server/services/social"
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
	r.HandleFunc("/api/social/trends", h.GetSocialTrends).Methods("GET")
	r.HandleFunc("/api/social/locations", h.GetAvailableLocations).Methods("GET")
}

// GetSocialTrends handles requests to get trending topics from social media
func (h *SocialHandler) GetSocialTrends(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("source")
	location := r.URL.Query().Get("location")

	// Default to Twitter if no source specified
	if source == "" || source == "twitter" {
		trends, err := h.getTwitterTrends(location)
		if err != nil {
			http.Error(w, "Failed to fetch Twitter trends: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trends)
		return
	}

	// Handle other sources as they are implemented
	http.Error(w, "Unsupported social media source", http.StatusBadRequest)
}

// getTwitterTrends fetches trending topics from Twitter
func (h *SocialHandler) getTwitterTrends(location string) ([]SocialTrendResponse, error) {
	twitterTrends, err := h.TwitterClient.GetTrends(location)
	if err != nil {
		return nil, err
	}

	var trends []SocialTrendResponse
	for i, trend := range twitterTrends {
		// Skip trends with no tweet volume or very low volume
		if trend.TweetVolume < 1000 {
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

	return trends, nil
}

// GetAvailableLocations handles requests to get available locations for trends
func (h *SocialHandler) GetAvailableLocations(w http.ResponseWriter, r *http.Request) {
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
