// internal/server/handlers/space.go

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"essg/internal/domain/messaging"
	"essg/internal/domain/space"
	"essg/internal/domain/trend"
)

// SpaceHandler handles space-related HTTP requests
type SpaceHandler struct {
	manager space.Manager
}

// NewSpaceHandler creates a new space handler
func NewSpaceHandler(manager space.Manager) *SpaceHandler {
	return &SpaceHandler{
		manager: manager,
	}
}

// ListSpaces returns a list of spaces
func (h *SpaceHandler) ListSpaces(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	var filter space.SpaceFilter

	// Parse lifecycle stages
	if stagesStr := r.URL.Query().Get("stages"); stagesStr != "" {
		// Parse comma-separated list
		// In a real implementation, we'd properly parse and validate this
		filter.LifecycleStages = []space.LifecycleStage{space.LifecycleStage(stagesStr)}
	} else {
		// Default to active stages
		filter.LifecycleStages = []space.LifecycleStage{
			space.StageGrowing,
			space.StagePeak,
		}
	}

	// Parse template types
	if typesStr := r.URL.Query().Get("types"); typesStr != "" {
		// Parse comma-separated list
		// In a real implementation, we'd properly parse and validate this
		filter.TemplateTypes = []space.TemplateType{space.TemplateType(typesStr)}
	}

	// Parse geo flag
	if geoStr := r.URL.Query().Get("geo"); geoStr != "" {
		geoOnly, err := strconv.ParseBool(geoStr)
		if err == nil {
			filter.IsGeoLocal = &geoOnly
		}
	}

	// Parse pagination
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Set defaults if not specified
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	// Get spaces
	spaces, err := h.manager.ListSpaces(r.Context(), filter)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list spaces", err)
		return
	}

	respondWithJSON(w, http.StatusOK, spaces)
}

// CreateSpace creates a new space
func (h *SpaceHandler) CreateSpace(w http.ResponseWriter, r *http.Request) {
	// Define request body struct
	type createSpaceRequest struct {
		Title       string          `json:"title"`
		Description string          `json:"description"`
		TopicTags   []string        `json:"topic_tags"`
		Location    *trend.Location `json:"location"`
		IsGeoLocal  bool            `json:"is_geo_local"`
	}

	// Parse request body
	var req createSpaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Create a simple trend from the request
	t := trend.Trend{
		Topic:       req.Title,
		Description: req.Description,
		Keywords:    req.TopicTags,
		Location:    req.Location,
		IsGeoLocal:  req.IsGeoLocal,
	}

	// Create space
	s, err := h.manager.CreateSpace(r.Context(), t)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create space", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, s)
}

// GetSpace returns a specific space by ID
func (h *SpaceHandler) GetSpace(w http.ResponseWriter, r *http.Request) {
	// Get space ID from URL
	id := chi.URLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Missing space ID", nil)
		return
	}

	// Get space
	s, err := h.manager.GetSpace(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Space not found", nil)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to get space", err)
		}
		return
	}

	respondWithJSON(w, http.StatusOK, s)
}

// GetNearbySpaces returns spaces near a specific location
func (h *SpaceHandler) GetNearbySpaces(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	radiusStr := r.URL.Query().Get("radius")

	if latStr == "" || lngStr == "" {
		respondWithError(w, http.StatusBadRequest, "Missing location parameters", nil)
		return
	}

	// Parse coordinates
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid latitude", err)
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid longitude", err)
		return
	}

	// Parse radius (default to 5km)
	radius := 5.0
	if radiusStr != "" {
		radius, err = strconv.ParseFloat(radiusStr, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid radius", err)
			return
		}
	}

	// Create location
	location := trend.Location{
		Latitude:  lat,
		Longitude: lng,
		Timestamp: time.Now(),
	}

	// Get nearby spaces
	spaces, err := h.manager.GetNearbySpaces(r.Context(), location, radius)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get nearby spaces", err)
		return
	}

	respondWithJSON(w, http.StatusOK, spaces)
}

// SendMessage sends a message to a space
func (h *SpaceHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	// Get space ID from URL
	spaceID := chi.URLParam(r, "id")
	if spaceID == "" {
		respondWithError(w, http.StatusBadRequest, "Missing space ID", nil)
		return
	}

	// Check if space exists
	_, err := h.manager.GetSpace(r.Context(), spaceID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Space not found", nil)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to get space", err)
		}
		return
	}

	// Define request body struct
	type sendMessageRequest struct {
		Content     string          `json:"content"`
		UserID      string          `json:"user_id"`
		Location    *trend.Location `json:"location"`
		IsAnonymous bool            `json:"is_anonymous"`
		MediaURLs   []string        `json:"media_urls"`
	}

	// Parse request body
	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Create message - in a real implementation, we would use the message service
	// This is a simplified example that returns a mock response
	message := messaging.Message{
		SpaceID:     spaceID,
		UserID:      req.UserID,
		Type:        messaging.TypeText,
		Content:     req.Content,
		MediaURLs:   req.MediaURLs,
		Location:    req.Location,
		IsAnonymous: req.IsAnonymous,
		CreatedAt:   time.Now(),
		Status:      messaging.StatusDelivered,
	}

	// In a real implementation, we would call the message service
	// response, err := h.messageService.SendMessage(r.Context(), message)
	// Here we're just mocking a successful response

	// Add an ID to the message for the response
	message.ID = "msg_" + strconv.FormatInt(time.Now().UnixNano(), 10)

	respondWithJSON(w, http.StatusCreated, message)
}

// GetMessages returns messages for a space
func (h *SpaceHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	// Get space ID from URL
	spaceID := chi.URLParam(r, "id")
	if spaceID == "" {
		respondWithError(w, http.StatusBadRequest, "Missing space ID", nil)
		return
	}

	// Check if space exists
	_, err := h.manager.GetSpace(r.Context(), spaceID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Space not found", nil)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to get space", err)
		}
		return
	}

	// Parse query parameters for filtering
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Default limit
	// limit := 50
	// offset := 0

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			// limit = parsedLimit
		}
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			// offset = parsedOffset
		}
	}

	// In a real implementation, we would call the message service
	// Here we're just returning mock data
	messages := []messaging.Message{
		{
			ID:        "msg_1",
			SpaceID:   spaceID,
			UserID:    "user_1",
			Type:      messaging.TypeText,
			Content:   "Hello from the space!",
			CreatedAt: time.Now().Add(-30 * time.Minute),
			Status:    messaging.StatusDelivered,
			Reactions: map[string]int{"👍": 3, "❤️": 2},
		},
		{
			ID:        "msg_2",
			SpaceID:   spaceID,
			UserID:    "user_2",
			Type:      messaging.TypeText,
			Content:   "This is an interesting discussion.",
			CreatedAt: time.Now().Add(-25 * time.Minute),
			Status:    messaging.StatusDelivered,
			Reactions: map[string]int{"👍": 1},
		},
	}

	respondWithJSON(w, http.StatusOK, messages)
}
