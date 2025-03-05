// internal/domain/trend/detector.go

package trend

import (
	"context"
)

// Detector defines the interface for trend detection
type Detector interface {
	// Start begins the trend detection process
	Start(ctx context.Context) error

	// Stop gracefully stops the trend detection process
	Stop(ctx context.Context) error

	// GetTrends returns currently detected trends filtered by the provided criteria
	GetTrends(ctx context.Context, filter Filter) ([]Trend, error)

	// GetTrendByID returns a specific trend by ID
	GetTrendByID(ctx context.Context, id string) (*Trend, error)

	// GetTrendsForLocation returns trends relevant to a specific location
	GetTrendsForLocation(ctx context.Context, location Location, radiusKm float64) ([]Trend, error)

	// AddPlatform adds a platform to monitor
	AddPlatform(ctx context.Context, platformConfig map[string]interface{}) error

	// RemovePlatform removes a platform from monitoring
	RemovePlatform(ctx context.Context, platformID string) error

	// RegisterTrendHandler registers a callback function for when new trends are detected
	RegisterTrendHandler(handler func(Trend) error) error
}

// Analyzer defines the interface for analyzing content to identify trends
type Analyzer interface {
	// AnalyzeContent processes content to extract trend information
	AnalyzeContent(ctx context.Context, content map[string]interface{}, source Source) ([]Trend, error)

	// CorrelateAcrossPlatforms identifies the same trends across different platforms
	CorrelateAcrossPlatforms(ctx context.Context, platformTrends map[string][]Trend) ([]Trend, error)

	// CalculateTrendScore computes a normalized score for a trend
	CalculateTrendScore(ctx context.Context, trend *Trend) (float64, error)
}

// GeoTagger defines the interface for adding location context to trends
type GeoTagger interface {
	// TagContent adds location information to content
	TagContent(ctx context.Context, content map[string]interface{}) (*Location, error)

	// GetSignificantLocations returns locations with significant activity
	GetSignificantLocations(ctx context.Context) ([]Location, error)

	// IsLocalTrend determines if a trend is primarily local
	IsLocalTrend(ctx context.Context, trend *Trend) (bool, error)

	// GetLocationRadius calculates an appropriate radius for a location-based trend
	GetLocationRadius(ctx context.Context, trend *Trend) (float64, error)
}
