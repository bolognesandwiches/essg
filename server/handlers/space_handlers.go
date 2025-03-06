package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"essg/server/models"
	"essg/server/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// SpaceHandler handles requests related to spaces
type SpaceHandler struct {
	spaceService *services.SpaceService
}

// NewSpaceHandler creates a new space handler
func NewSpaceHandler(spaceService *services.SpaceService) *SpaceHandler {
	return &SpaceHandler{
		spaceService: spaceService,
	}
}

// RegisterRoutes registers the space API routes
func (h *SpaceHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/spaces/trending", h.GetTrendingSpaces).Methods("GET")
	r.HandleFunc("/api/spaces/nearby", h.GetNearbySpaces).Methods("GET")
	r.HandleFunc("/api/spaces/joined", h.GetJoinedSpaces).Methods("GET")
	r.HandleFunc("/api/spaces/check-exists", h.CheckSpaceExists).Methods("GET")
	r.HandleFunc("/api/spaces/{id}", h.GetSpaceById).Methods("GET")
	r.HandleFunc("/api/spaces/{id}/join", h.JoinSpace).Methods("POST")
	r.HandleFunc("/api/spaces/{id}/leave", h.LeaveSpace).Methods("POST")
	r.HandleFunc("/api/spaces", h.CreateSpace).Methods("POST")
	// Add more routes as needed
}

// GetTrendingSpaces handles requests to get trending spaces
func (h *SpaceHandler) GetTrendingSpaces(w http.ResponseWriter, r *http.Request) {
	// Implementation will go here
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]models.Space{})
}

// GetNearbySpaces handles requests to get nearby spaces
func (h *SpaceHandler) GetNearbySpaces(w http.ResponseWriter, r *http.Request) {
	// Implementation will go here
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]models.Space{})
}

// GetSpaceById handles requests to get a space by ID
func (h *SpaceHandler) GetSpaceById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spaceID := vars["id"]

	space, err := h.spaceService.GetSpaceByID(spaceID)
	if err != nil {
		http.Error(w, "Failed to get space: "+err.Error(), http.StatusNotFound)
		return
	}

	fmt.Printf("Retrieved space with ID %s, created at %s, last active %s\n",
		space.ID,
		space.CreatedAt.Format(time.RFC3339),
		space.LastActive.Format(time.RFC3339))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(space)
}

// GetJoinedSpaces handles requests to get spaces joined by a user
func (h *SpaceHandler) GetJoinedSpaces(w http.ResponseWriter, r *http.Request) {
	// Get user ID from headers
	userID := r.Header.Get("x-anonymous-user-id")
	if userID == "" {
		http.Error(w, "Anonymous user ID required", http.StatusBadRequest)
		return
	}

	fmt.Printf("Getting joined spaces for user: %s\n", userID)

	// Get spaces joined by the user
	spaces, err := h.spaceService.GetJoinedSpaces(userID)
	if err != nil {
		http.Error(w, "Failed to get joined spaces: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Found %d joined spaces for user %s\n", len(spaces), userID)

	// Return the spaces
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spaces)
}

// CreateSpaceRequest represents the request body for creating a space
type CreateSpaceRequest struct {
	TrendID string `json:"trendId"`
	Source  string `json:"source"`
}

// JoinSpace handles requests to join a space
func (h *SpaceHandler) JoinSpace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spaceID := vars["id"]

	// Get user ID from headers
	userID := r.Header.Get("x-anonymous-user-id")
	if userID == "" {
		http.Error(w, "Anonymous user ID required", http.StatusBadRequest)
		return
	}

	// Join the space
	err := h.spaceService.JoinSpace(spaceID, userID)
	if err != nil {
		http.Error(w, "Failed to join space: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "joined"})
}

// LeaveSpace handles requests to leave a space
func (h *SpaceHandler) LeaveSpace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spaceID := vars["id"]

	// Get user ID from headers
	userID := r.Header.Get("x-anonymous-user-id")
	if userID == "" {
		http.Error(w, "Anonymous user ID required", http.StatusBadRequest)
		return
	}

	// Leave the space
	err := h.spaceService.LeaveSpace(spaceID, userID)
	if err != nil {
		http.Error(w, "Failed to leave space: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "left"})
}

// CreateSpace creates a new space based on a trend
func (h *SpaceHandler) CreateSpace(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req CreateSpaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("Creating space from trend: %s (source: %s)\n", req.TrendID, req.Source)

	// Get user ID from headers (anonymous user)
	userID := r.Header.Get("x-anonymous-user-id")
	if userID == "" {
		// Generate a random ID if user is not identified
		userID = uuid.New().String()
	}

	// First, check if a space for this trend already exists
	existingSpace, err := h.spaceService.GetSpaceByTrend(req.Source, req.TrendID)
	if err == nil && existingSpace != nil {
		// Space already exists for this trend, join it instead of creating a new one
		fmt.Printf("Space already exists for trend '%s' (ID: %s), joining instead\n", req.TrendID, existingSpace.ID)

		// Join the user to the existing space
		err := h.spaceService.JoinSpace(existingSpace.ID, userID)
		if err != nil {
			http.Error(w, "Failed to join existing space: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Return the existing space
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // Use 200 instead of 201 since we're not creating a new space
		json.NewEncoder(w).Encode(existingSpace)
		return
	}

	// Create a space based on the trend
	space := &models.Space{
		ID:             uuid.New().String(),
		Title:          req.TrendID, // Use the trend name directly
		Description:    fmt.Sprintf("A space for discussing the trending topic '%s' from %s", req.TrendID, req.Source),
		TemplateType:   "social_trend",
		LifecycleStage: "active",
		CreatedAt:      time.Now(),
		LastActive:     time.Now(),
		UserCount:      1,
		MessageCount:   0,
		IsGeoLocal:     false,
		TopicTags:      []string{req.Source, "trend", "social_media"},
		TrendName:      req.TrendID, // Store the original trend name
		CreatedBy:      userID,
	}

	// Set an expiration time (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour)
	space.ExpiresAt = &expiresAt

	fmt.Printf("Created space with ID %s and tags: %v\n", space.ID, space.TopicTags)

	// Save the space
	err = h.spaceService.CreateSpace(space)
	if err != nil {
		http.Error(w, "Failed to create space: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Join the user to the space
	err = h.spaceService.JoinSpace(space.ID, userID)
	if err != nil {
		http.Error(w, "Failed to join space: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the created space
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(space)
}

// CheckSpaceExists checks if a space exists for a specific trend
func (h *SpaceHandler) CheckSpaceExists(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	trendName := r.URL.Query().Get("trendName")
	source := r.URL.Query().Get("source")

	if trendName == "" || source == "" {
		http.Error(w, "trendName and source are required", http.StatusBadRequest)
		return
	}

	// Check if space exists
	existingSpace, err := h.spaceService.GetSpaceByTrend(source, trendName)

	// Prepare response
	response := struct {
		Exists  bool   `json:"exists"`
		SpaceId string `json:"spaceId,omitempty"`
	}{
		Exists: err == nil && existingSpace != nil,
	}

	// Add the space ID if found
	if existingSpace != nil {
		response.SpaceId = existingSpace.ID
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
