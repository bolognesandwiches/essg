// internal/server/handlers/trend.go

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"essg/internal/domain/trend"
)

// TrendHandler handles trend-related HTTP requests
type TrendHandler struct {
	detector trend.Detector
}

// NewTrendHandler creates a new trend handler
func NewTrendHandler(detector trend.Detector) *TrendHandler {
	return &TrendHandler{
		detector: detector,
	}
}

// GetTrends returns a list of trends
func (h *TrendHandler) GetTrends(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	minScore, _ := strconv.ParseFloat(r.URL.Query().Get("min_score"), 64)
	if minScore <= 0 {
		minScore = 10.0 // Default min score
	}

	// Create filter
	filter := trend.Filter{
		MinScore: minScore,
	}

	// Get platforms filter
	if platforms := r.URL.Query().Get("platforms"); platforms != "" {
		filter.IncludePlatforms = []string{platforms}
	}

	// Get entity type filter
	if entityType := r.URL.Query().Get("entity_type"); entityType != "" {
		filter.IncludeEntityType = []string{entityType}
	}

	// Get trends
	trends, err := h.detector.GetTrends(r.Context(), filter)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get trends", err)
		return
	}

	respondWithJSON(w, http.StatusOK, trends)
}

// GetTrend returns a specific trend by ID
func (h *TrendHandler) GetTrend(w http.ResponseWriter, r *http.Request) {
	// Get trend ID from URL
	id := chi.URLParam(r, "id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Missing trend ID", nil)
		return
	}

	// Get trend
	t, err := h.detector.GetTrendByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Trend not found", nil)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to get trend", err)
		}
		return
	}

	respondWithJSON(w, http.StatusOK, t)
}

// GetGeoTrends returns trends near a specific location
func (h *TrendHandler) GetGeoTrends(w http.ResponseWriter, r *http.Request) {
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
	}

	// Get trends for location
	trends, err := h.detector.GetTrendsForLocation(r.Context(), location, radius)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get geo trends", err)
		return
	}

	respondWithJSON(w, http.StatusOK, trends)
}

// Helper for JSON responses
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to marshal response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Helper for error responses
func respondWithError(w http.ResponseWriter, code int, message string, err error) {
	response := map[string]string{"error": message}

	if err != nil && code >= 500 {
		// Log server errors
		// logger.Error("HTTP error", "code", code, "message", message, "error", err)
	}

	jsonResponse, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(jsonResponse)
}

// Common errors
var (
	ErrNotFound = errors.New("not found")
)
