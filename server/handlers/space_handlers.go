package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// CreateSpaceFromTrendRequest represents the request to create a space from a trend
type CreateSpaceFromTrendRequest struct {
	TrendID string `json:"trendId"`
	Source  string `json:"source"`
}

// CreateSpaceFromTrend handles requests to create a new space based on a social media trend
func (h *SpaceHandler) CreateSpaceFromTrend(w http.ResponseWriter, r *http.Request) {
	// Get anonymous user ID from header
	userID := r.Header.Get("x-anonymous-user-id")
	if userID == "" {
		http.Error(w, "Anonymous user ID required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req CreateSpaceFromTrendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.TrendID == "" {
		http.Error(w, "Trend ID is required", http.StatusBadRequest)
		return
	}

	// Get trend details from social service
	// This is a simplified example - in a real implementation,
	// you would fetch the actual trend details from your social service
	var trendName string
	var trendDescription string

	if req.Source == "twitter" {
		// Fetch trend details from Twitter
		// This is a placeholder - you would implement this based on your Twitter client
		trendName = "Twitter Trend #" + req.TrendID
		trendDescription = "A space for discussing the trending topic on Twitter"
	} else {
		http.Error(w, "Unsupported social media source", http.StatusBadRequest)
		return
	}

	// Create a new space based on the trend
	space := &models.Space{
		ID:             uuid.New().String(),
		Title:          trendName,
		Description:    trendDescription,
		TemplateType:   "social_trend",
		LifecycleStage: "growing",
		CreatedAt:      time.Now(),
		LastActive:     time.Now(),
		UserCount:      1, // Start with the creator
		MessageCount:   0,
		IsGeoLocal:     false, // Social trends are typically not geo-local
		TopicTags:      []string{req.Source, "trend"},
		CreatedBy:      userID,
	}

	// Add expiration (e.g., 24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour)
	space.ExpiresAt = &expiresAt

	// Save the space
	if err := h.spaceService.CreateSpace(space); err != nil {
		http.Error(w, "Failed to create space: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Auto-join the creator to the space
	if err := h.spaceService.JoinSpace(space.ID, userID); err != nil {
		// Log the error but don't fail the request
		log.Printf("Failed to auto-join creator to space: %v", err)
	}

	// Return the created space
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(space)
}
