// internal/domain/geo/service.go

package geo

import (
	"context"

	"essg/internal/domain/space"
	"essg/internal/domain/trend"
)

// LocationContext provides additional information about a geographic location
type LocationContext struct {
	PlaceID       string
	Name          string
	Neighborhood  string
	Locality      string // City
	AdminArea     string // State/Province
	Country       string
	PostalCode    string
	FormattedAddr string
	Population    int64
	Timezone      string
	Types         []string // e.g., "locality", "point_of_interest"
}

// PopulationDensity represents population density information for an area
type PopulationDensity struct {
	Location      trend.Location
	RadiusKm      float64
	Population    int64
	DensityPerKm2 float64
}

// LocalTrend represents a trend specific to a geographic area
type LocalTrend struct {
	TrendID         string
	Score           float64
	LocationScore   float64 // How strongly tied to the location
	Location        trend.Location
	LocationContext LocationContext
	LocalRelevance  map[string]float64 // Different dimensions of local relevance
}

// LocalSource defines sources of local information
type LocalSource interface {
	// Name returns the name of the source
	Name() string

	// GetTrendsNear returns trends near a location
	GetTrendsNear(ctx context.Context, location trend.Location, radiusKm float64) ([]LocalTrend, error)
}

// Service defines the interface for geospatial services
type Service interface {
	// FindNearbySpaces returns spaces near a location
	FindNearbySpaces(ctx context.Context, location trend.Location, radiusKm float64) ([]space.Space, error)

	// GetLocalTrends returns trends specific to a location
	GetLocalTrends(ctx context.Context, location trend.Location, radiusKm float64) ([]LocalTrend, error)

	// GetLocationContext returns context information for a location
	GetLocationContext(ctx context.Context, location trend.Location) (*LocationContext, error)

	// FuzzLocation reduces the precision of a location for privacy
	FuzzLocation(location trend.Location, precisionLevel string) trend.Location

	// CalculateDistance calculates the distance between two locations
	CalculateDistance(a, b trend.Location) float64

	// GetPopulationDensity returns population density for an area
	GetPopulationDensity(ctx context.Context, location trend.Location, radiusKm float64) (*PopulationDensity, error)

	// ClusterLocations groups nearby locations together
	ClusterLocations(locations []trend.Location, maxDistanceKm float64) [][]trend.Location

	// GetOptimalRadius determines the best radius for a location based on population density
	GetOptimalRadius(ctx context.Context, location trend.Location) (float64, error)

	// IsWithinBounds checks if a location is within a specified geographic boundary
	IsWithinBounds(location trend.Location, centerPoint trend.Location, radiusKm float64) bool

	// AddLocalSource adds a source of local information
	AddLocalSource(source LocalSource) error
}

// GeoPrivacyManager manages location privacy
type GeoPrivacyManager interface {
	// ApplyPrivacySettings applies privacy settings to a location
	ApplyPrivacySettings(location trend.Location, privacySetting string, contextRadius float64) trend.Location

	// GetPrivacyLevels returns available privacy levels
	GetPrivacyLevels() []string

	// ValidatePrivacySetting checks if a privacy setting is valid
	ValidatePrivacySetting(setting string) bool
}
