package listening

import (
	"context"

	"essg/internal/domain/trend"
)

// GeoTagger adds location information to content
type GeoTagger struct {
}

// NewGeoTagger creates a new geo tagger
func NewGeoTagger() *GeoTagger {
	return &GeoTagger{}
}

// TagContent adds location information to content
func (g *GeoTagger) TagContent(ctx context.Context, content map[string]interface{}) (*trend.Location, error) {
	// Implementation will come later
	return nil, nil
}

// GetSignificantLocations returns locations with significant activity
func (g *GeoTagger) GetSignificantLocations(ctx context.Context) ([]trend.Location, error) {
	// Implementation will come later
	return []trend.Location{}, nil
}

// IsLocalTrend determines if a trend is primarily local
func (g *GeoTagger) IsLocalTrend(ctx context.Context, t *trend.Trend) (bool, error) {
	// Simple implementation for now
	return t.IsGeoLocal, nil
}

// GetLocationRadius calculates an appropriate radius for a location-based trend
func (g *GeoTagger) GetLocationRadius(ctx context.Context, t *trend.Trend) (float64, error) {
	// Simple implementation for now
	if t.LocationRadius > 0 {
		return t.LocationRadius, nil
	}
	return 5.0, nil // Default 5km radius
}
