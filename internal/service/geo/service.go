// internal/service/geo/service.go

package geo

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"

	"essg/internal/domain/geo"
	"essg/internal/domain/space"
	"essg/internal/domain/trend"
)

// NewGeocoderService creates a new instance of a geocoder service.
func NewGeocoderService() GeocoderService {
	return &defaultGeocoderService{}
}

// defaultGeocoderService is a basic implementation of the GeocoderService interface.
type defaultGeocoderService struct{}

// ReverseGeocode provides a dummy implementation for reverse geocoding.
func (g *defaultGeocoderService) ReverseGeocode(ctx context.Context, lat, lng float64) (*geo.LocationContext, error) {
	// This should call an actual geocoding API (e.g., Google Maps API, OpenStreetMap, etc.)
	return &geo.LocationContext{
		FormattedAddr: fmt.Sprintf("Lat: %f, Lng: %f", lat, lng), // Use FormattedAddr instead
	}, nil
}

// Geocode provides a dummy implementation for geocoding.
func (g *defaultGeocoderService) Geocode(ctx context.Context, address string) (*trend.Location, error) {
	// This should call an actual geocoding API.
	return &trend.Location{
		Latitude:  37.7749, // Example: San Francisco
		Longitude: -122.4194,
	}, nil
}

// LocalSourceRegistry manages local information sources
type LocalSourceRegistry struct {
	sources []geo.LocalSource
	mu      sync.RWMutex
}

// NewLocalSourceRegistry creates a new registry for local sources
func NewLocalSourceRegistry() *LocalSourceRegistry {
	return &LocalSourceRegistry{
		sources: []geo.LocalSource{},
	}
}

// AddSource adds a source to the registry
func (r *LocalSourceRegistry) AddSource(source geo.LocalSource) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sources = append(r.sources, source)
}

// GetSources returns all registered sources
func (r *LocalSourceRegistry) GetSources() []geo.LocalSource {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a copy to avoid race conditions
	sources := make([]geo.LocalSource, len(r.sources))
	copy(sources, r.sources)

	return sources
}

// GeoSpatialConfig contains configuration for the geospatial service
type GeoSpatialConfig struct {
	DefaultRadius               float64
	MinRadius                   float64
	MaxRadius                   float64
	ClusterThreshold            float64
	PopulationDensityThresholds map[string]float64
}

// GeoSpatialService implements the geo.Service interface
type GeoSpatialService struct {
	db           *pgxpool.Pool
	localSources *LocalSourceRegistry
	geocoder     GeocoderService
	config       GeoSpatialConfig
}

// GeocoderService provides geocoding functionality
type GeocoderService interface {
	// ReverseGeocode gets location context from coordinates
	ReverseGeocode(ctx context.Context, lat, lng float64) (*geo.LocationContext, error)

	// Geocode gets coordinates from an address
	Geocode(ctx context.Context, address string) (*trend.Location, error)
}

// NewGeoSpatialService creates a new geospatial service
func NewGeoSpatialService(
	db *pgxpool.Pool,
	localSources *LocalSourceRegistry,
	geocoder GeocoderService,
	config GeoSpatialConfig,
) *GeoSpatialService {
	return &GeoSpatialService{
		db:           db,
		localSources: localSources,
		geocoder:     geocoder,
		config:       config,
	}
}

// FindNearbySpaces returns spaces near a location
func (s *GeoSpatialService) FindNearbySpaces(
	ctx context.Context,
	location trend.Location,
	radiusKm float64,
) ([]space.Space, error) {
	// Use PostGIS ST_DWithin for efficient spatial query
	query := `
		SELECT id, title, description, created_at, user_count, topic_tags, 
		       trend_id, template_type, lifecycle_stage, is_geo_local,
		       ST_X(location::geometry) as lat, ST_Y(location::geometry) as lng, 
		       location_radius
		FROM spaces 
		WHERE ST_DWithin(
			geography(location),
			geography(ST_MakePoint($1, $2)),
			$3 * 1000
		)
		AND lifecycle_stage IN ('growing', 'active', 'peak')
		ORDER BY 
			user_count DESC,
			ST_Distance(geography(location), geography(ST_MakePoint($1, $2)))
		LIMIT 20
	`

	rows, err := s.db.Query(ctx, query, location.Longitude, location.Latitude, radiusKm)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var spaces []space.Space
	for rows.Next() {
		var s space.Space
		var lat, lng *float64

		// Scan row into space struct and location variables
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.CreatedAt, &s.UserCount, &s.TopicTags,
			&s.TrendID, &s.TemplateType, &s.LifecycleStage, &s.IsGeoLocal,
			&lat, &lng, &s.LocationRadius,
		); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// Set location if coordinates are present
		if lat != nil && lng != nil {
			s.Location = &trend.Location{
				Latitude:  *lat,
				Longitude: *lng,
				Timestamp: s.CreatedAt,
			}
		}

		spaces = append(spaces, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return spaces, nil
}

// GetLocalTrends returns trends specific to a location
func (s *GeoSpatialService) GetLocalTrends(
	ctx context.Context,
	location trend.Location,
	radiusKm float64,
) ([]geo.LocalTrend, error) {
	// Aggregate trends from all local sources
	var allTrends []geo.LocalTrend

	// Get sources
	sources := s.localSources.GetSources()

	// Use a wait group to handle concurrent requests
	var wg sync.WaitGroup
	trendChan := make(chan []geo.LocalTrend, len(sources))
	errChan := make(chan error, len(sources))

	// Query each source concurrently
	for _, source := range sources {
		wg.Add(1)
		go func(src geo.LocalSource) {
			defer wg.Done()

			sourceTrends, err := src.GetTrendsNear(ctx, location, radiusKm)
			if err != nil {
				errChan <- fmt.Errorf("error from source %s: %w", src.Name(), err)
				return
			}

			trendChan <- sourceTrends
		}(source)
	}

	// Wait for all queries to complete
	wg.Wait()
	close(trendChan)
	close(errChan)

	// Check for errors
	select {
	case err := <-errChan:
		return nil, err
	default:
		// Continue if no errors
	}

	// Collect all trends
	for trends := range trendChan {
		allTrends = append(allTrends, trends...)
	}

	// Cluster similar trends
	clusters := s.clusterSimilarTrends(allTrends)

	// Create aggregated trends from clusters
	var aggregatedTrends []geo.LocalTrend
	for _, cluster := range clusters {
		aggregatedTrends = append(aggregatedTrends, s.createAggregatedTrend(cluster))
	}

	// Sort by score
	sort.Slice(aggregatedTrends, func(i, j int) bool {
		return aggregatedTrends[i].Score > aggregatedTrends[j].Score
	})

	return aggregatedTrends, nil
}

// clusterSimilarTrends groups similar trends together
func (s *GeoSpatialService) clusterSimilarTrends(trends []geo.LocalTrend) [][]geo.LocalTrend {
	if len(trends) == 0 {
		return nil
	}

	// Simple clustering for now - in a real implementation, we'd use more sophisticated
	// NLP techniques for similarity detection
	var clusters [][]geo.LocalTrend
	assigned := make(map[int]bool)

	for i, trend := range trends {
		if assigned[i] {
			continue
		}

		// Create a new cluster with this trend
		cluster := []geo.LocalTrend{trend}
		assigned[i] = true

		// Find similar trends to add to this cluster
		for j, otherTrend := range trends {
			if i == j || assigned[j] {
				continue
			}

			// Simplified similarity check - real implementation would use semantic similarity
			if s.areTrendsSimilar(trend, otherTrend) {
				cluster = append(cluster, otherTrend)
				assigned[j] = true
			}
		}

		clusters = append(clusters, cluster)
	}

	return clusters
}

// areTrendsSimilar determines if two trends are similar
func (s *GeoSpatialService) areTrendsSimilar(a, b geo.LocalTrend) bool {
	// In a real implementation, this would use NLP to compare trend topics
	// This is a placeholder that assumes trends with the same ID are similar
	return a.TrendID == b.TrendID
}

// createAggregatedTrend creates an aggregated trend from a cluster
func (s *GeoSpatialService) createAggregatedTrend(cluster []geo.LocalTrend) geo.LocalTrend {
	if len(cluster) == 0 {
		return geo.LocalTrend{}
	}

	// If only one trend, return it directly
	if len(cluster) == 1 {
		return cluster[0]
	}

	// Use the most complete trend as the base
	baseTrend := cluster[0]

	// Calculate average score and location score
	var totalScore, totalLocationScore float64
	for _, t := range cluster {
		totalScore += t.Score
		totalLocationScore += t.LocationScore
	}

	avgScore := totalScore / float64(len(cluster))
	avgLocationScore := totalLocationScore / float64(len(cluster))

	// Create aggregated trend with averaged scores
	aggregated := geo.LocalTrend{
		TrendID:         baseTrend.TrendID,
		Score:           avgScore,
		LocationScore:   avgLocationScore,
		Location:        baseTrend.Location,
		LocationContext: baseTrend.LocationContext,
		LocalRelevance:  make(map[string]float64),
	}

	// Combine local relevance factors
	for _, t := range cluster {
		for factor, score := range t.LocalRelevance {
			if existingScore, exists := aggregated.LocalRelevance[factor]; exists {
				// Average the scores if the factor already exists
				aggregated.LocalRelevance[factor] = (existingScore + score) / 2
			} else {
				aggregated.LocalRelevance[factor] = score
			}
		}
	}

	return aggregated
}

// GetLocationContext returns context information for a location
func (s *GeoSpatialService) GetLocationContext(
	ctx context.Context,
	location trend.Location,
) (*geo.LocationContext, error) {
	return s.geocoder.ReverseGeocode(ctx, location.Latitude, location.Longitude)
}

// FuzzLocation reduces the precision of a location for privacy
func (s *GeoSpatialService) FuzzLocation(location trend.Location, precisionLevel string) trend.Location {
	fuzzed := location

	switch precisionLevel {
	case "precise":
		// No fuzzing
		return location

	case "neighborhood":
		// Fuzz within ~500m
		fuzzed.Latitude += (rand.Float64() - 0.5) * 0.01
		fuzzed.Longitude += (rand.Float64() - 0.5) * 0.01

	case "approximate":
		// Fuzz within ~2km
		fuzzed.Latitude += (rand.Float64() - 0.5) * 0.04
		fuzzed.Longitude += (rand.Float64() - 0.5) * 0.04

	case "disabled":
		// Return a nil location or an error in a real implementation
		// Here we'll just reset coordinates to zero
		fuzzed.Latitude = 0
		fuzzed.Longitude = 0

	default:
		// Default to neighborhood level
		fuzzed.Latitude += (rand.Float64() - 0.5) * 0.01
		fuzzed.Longitude += (rand.Float64() - 0.5) * 0.01
	}

	return fuzzed
}

// CalculateDistance calculates the distance between two locations in kilometers
func (s *GeoSpatialService) CalculateDistance(a, b trend.Location) float64 {
	// Implementation of the Haversine formula for distance on a sphere
	const earthRadiusKm = 6371.0

	// Convert latitude and longitude from degrees to radians
	lat1 := a.Latitude * math.Pi / 180.0
	lon1 := a.Longitude * math.Pi / 180.0
	lat2 := b.Latitude * math.Pi / 180.0
	lon2 := b.Longitude * math.Pi / 180.0

	// Haversine formula
	dLat := lat2 - lat1
	dLon := lon2 - lon1

	hSin := math.Sin(dLat / 2)
	hSin *= hSin

	vSin := math.Sin(dLon / 2)
	vSin *= vSin

	h := hSin + math.Cos(lat1)*math.Cos(lat2)*vSin

	return 2 * earthRadiusKm * math.Asin(math.Sqrt(h))
}

// GetPopulationDensity returns population density for an area
func (s *GeoSpatialService) GetPopulationDensity(
	ctx context.Context,
	location trend.Location,
	radiusKm float64,
) (*geo.PopulationDensity, error) {
	// In a production implementation, this would query a real population database
	// This is a simplified version that uses the location context

	// First get location context
	locationContext, err := s.GetLocationContext(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("error getting location context: %w", err)
	}

	// Get population from location context
	population := locationContext.Population

	// If population data is missing, estimate based on location type
	if population <= 0 {
		// Rough estimates based on location types
		if contains(locationContext.Types, "locality") {
			// City - rough estimate
			population = 100000
		} else if contains(locationContext.Types, "neighborhood") {
			// Neighborhood
			population = 10000
		} else if contains(locationContext.Types, "administrative_area_level_1") {
			// State/Province
			population = 1000000
		} else {
			// Default value
			population = 5000
		}
	}

	// Calculate area in square kilometers
	areaKm2 := math.Pi * radiusKm * radiusKm

	// Calculate density
	densityPerKm2 := float64(population) / areaKm2

	return &geo.PopulationDensity{
		Location:      location,
		RadiusKm:      radiusKm,
		Population:    population,
		DensityPerKm2: densityPerKm2,
	}, nil
}

// ClusterLocations groups nearby locations together
func (s *GeoSpatialService) ClusterLocations(
	locations []trend.Location,
	maxDistanceKm float64,
) [][]trend.Location {
	if len(locations) == 0 {
		return nil
	}

	var clusters [][]trend.Location
	visited := make(map[int]bool)

	for i, loc := range locations {
		if visited[i] {
			continue
		}

		// Start a new cluster with this location
		cluster := []trend.Location{loc}
		visited[i] = true

		// Find nearby locations to add to this cluster
		for j, otherLoc := range locations {
			if i == j || visited[j] {
				continue
			}

			if s.CalculateDistance(loc, otherLoc) <= maxDistanceKm {
				cluster = append(cluster, otherLoc)
				visited[j] = true
			}
		}

		clusters = append(clusters, cluster)
	}

	return clusters
}

// GetOptimalRadius determines the best radius for a location based on population density
func (s *GeoSpatialService) GetOptimalRadius(
	ctx context.Context,
	location trend.Location,
) (float64, error) {
	// Start with default radius
	radius := s.config.DefaultRadius

	// Get population density for the area
	density, err := s.GetPopulationDensity(ctx, location, radius)
	if err != nil {
		// Return default radius if we can't get density
		return radius, nil
	}

	// Adjust radius based on population density
	if density.DensityPerKm2 > s.config.PopulationDensityThresholds["urban"] {
		// Urban area - smaller radius
		radius = math.Max(s.config.MinRadius, radius*0.5)
	} else if density.DensityPerKm2 < s.config.PopulationDensityThresholds["rural"] {
		// Rural area - larger radius
		radius = math.Min(s.config.MaxRadius, radius*2.0)
	}

	return radius, nil
}

// IsWithinBounds checks if a location is within a specified geographic boundary
func (s *GeoSpatialService) IsWithinBounds(
	location trend.Location,
	centerPoint trend.Location,
	radiusKm float64,
) bool {
	distance := s.CalculateDistance(location, centerPoint)
	return distance <= radiusKm
}

// AddLocalSource adds a source of local information
func (s *GeoSpatialService) AddLocalSource(source geo.LocalSource) error {
	s.localSources.AddSource(source)
	return nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GeoPrivacyManagerImpl implements the GeoPrivacyManager interface
type GeoPrivacyManagerImpl struct {
	privacyLevels map[string]float64
}

// NewGeoPrivacyManager creates a new geo privacy manager
func NewGeoPrivacyManager() *GeoPrivacyManagerImpl {
	return &GeoPrivacyManagerImpl{
		privacyLevels: map[string]float64{
			"disabled":     0.0,    // No location sharing
			"approximate":  0.05,   // ~5km precision
			"neighborhood": 0.01,   // ~1km precision
			"precise":      0.0001, // ~10m precision
		},
	}
}

// ApplyPrivacySettings applies privacy settings to a location
func (g *GeoPrivacyManagerImpl) ApplyPrivacySettings(
	location trend.Location,
	privacySetting string,
	contextRadius float64,
) trend.Location {
	// If disabled, return a nil location or an error in a real implementation
	// Here we'll just return a location with zeroed coordinates
	if privacySetting == "disabled" {
		return trend.Location{
			Latitude:  0,
			Longitude: 0,
			Timestamp: location.Timestamp,
		}
	}

	// Get precision level for this setting
	precision, ok := g.privacyLevels[privacySetting]
	if !ok {
		// Default to neighborhood if setting not found
		precision = g.privacyLevels["neighborhood"]
	}

	// Apply fuzzing based on precision
	fuzzedLat := location.Latitude
	fuzzedLng := location.Longitude

	// Add random noise based on precision level
	if precision > 0 {
		fuzzedLat = math.Round(fuzzedLat/precision) * precision
		fuzzedLng = math.Round(fuzzedLng/precision) * precision

		// Add small random offset for additional privacy
		fuzzedLat += (rand.Float64() - 0.5) * precision * 0.5
		fuzzedLng += (rand.Float64() - 0.5) * precision * 0.5
	}

	return trend.Location{
		Latitude:  fuzzedLat,
		Longitude: fuzzedLng,
		Timestamp: location.Timestamp,
		Accuracy:  precision * 111000, // Convert degrees to meters (roughly)
	}
}

// GetPrivacyLevels returns available privacy levels
func (g *GeoPrivacyManagerImpl) GetPrivacyLevels() []string {
	levels := make([]string, 0, len(g.privacyLevels))
	for level := range g.privacyLevels {
		levels = append(levels, level)
	}
	return levels
}

// ValidatePrivacySetting checks if a privacy setting is valid
func (g *GeoPrivacyManagerImpl) ValidatePrivacySetting(setting string) bool {
	_, valid := g.privacyLevels[setting]
	return valid
}
