// internal/server/handlers/geo.go

package handlers

import (
	"net/http"
	"strconv"
	"time"

	"essg/internal/domain/geo"
	"essg/internal/domain/trend"
)

// GeoHandler handles geospatial-related HTTP requests
type GeoHandler struct {
	service geo.Service
}

// NewGeoHandler creates a new geo handler
func NewGeoHandler(service geo.Service) *GeoHandler {
	return &GeoHandler{
		service: service,
	}
}

// GetLocationContext returns context information for a location
func (h *GeoHandler) GetLocationContext(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")

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

	// Create location
	location := trend.Location{
		Latitude:  lat,
		Longitude: lng,
		Timestamp: time.Now(),
	}

	// Get location context
	context, err := h.service.GetLocationContext(r.Context(), location)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get location context", err)
		return
	}

	respondWithJSON(w, http.StatusOK, context)
}

// GetLocalTrends returns trends specific to a location
func (h *GeoHandler) GetLocalTrends(w http.ResponseWriter, r *http.Request) {
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

	// Get local trends
	trends, err := h.service.GetLocalTrends(r.Context(), location, radius)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get local trends", err)
		return
	}

	respondWithJSON(w, http.StatusOK, trends)
}
