// internal/service/listening/detector.go

package listening

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"essg/internal/domain/trend"
)

// SocialPlatform defines an interface for social platform data sources
type SocialPlatform interface {
	// Name returns the platform name
	Name() string

	// Start begins monitoring the platform
	Start(ctx context.Context) error

	// Stop stops monitoring the platform
	Stop() error

	// GetTrends returns current trends from this platform
	GetTrends(ctx context.Context) ([]trend.Trend, error)
}

// TrendDetectorConfig contains configuration for the trend detector
type TrendDetectorConfig struct {
	TrendThreshold         float64
	ScanInterval           time.Duration
	GeoScanInterval        time.Duration
	CorrelationThreshold   float64
	MaxConcurrentPlatforms int
	EventsTopic            string
}

// TrendDetector implements the trend.Detector interface
type TrendDetector struct {
	platforms     map[string]SocialPlatform
	analyzer      trend.Analyzer
	geoTagger     trend.GeoTagger
	config        TrendDetectorConfig
	eventBus      *nats.Conn
	trendHandlers []func(trend.Trend) error
	trendStore    TrendStore
	mu            sync.RWMutex
	platformsLock sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// TrendStore defines storage for trends
type TrendStore interface {
	SaveTrend(ctx context.Context, t trend.Trend) error
	GetTrend(ctx context.Context, id string) (*trend.Trend, error)
	FindTrends(ctx context.Context, filter trend.Filter) ([]trend.Trend, error)
	FindTrendsForLocation(ctx context.Context, location trend.Location, radiusKm float64) ([]trend.Trend, error)
}

// NewTrendDetector creates a new trend detector
func NewTrendDetector(
	analyzer trend.Analyzer,
	geoTagger trend.GeoTagger,
	trendStore TrendStore,
	eventBus *nats.Conn,
	config TrendDetectorConfig,
) *TrendDetector {
	ctx, cancel := context.WithCancel(context.Background())

	return &TrendDetector{
		platforms:     make(map[string]SocialPlatform),
		analyzer:      analyzer,
		geoTagger:     geoTagger,
		config:        config,
		eventBus:      eventBus,
		trendHandlers: []func(trend.Trend) error{},
		trendStore:    trendStore,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins the trend detection process
func (td *TrendDetector) Start(ctx context.Context) error {
	// Start platform monitoring goroutines
	td.platformsLock.RLock()
	for _, platform := range td.platforms {
		td.wg.Add(1)
		go func(p SocialPlatform) {
			defer td.wg.Done()
			if err := p.Start(ctx); err != nil {
				// Log the error but continue with other platforms
				fmt.Printf("Error starting platform %s: %v\n", p.Name(), err)
			}
		}(platform)
	}
	td.platformsLock.RUnlock()

	// Start cross-platform analysis
	td.wg.Add(1)
	go td.analyzeCrossPlatformTrends(ctx)

	// Start geo-specific analysis
	td.wg.Add(1)
	go td.analyzeGeoTrends(ctx)

	return nil
}

// GetTrends returns currently detected trends filtered by the provided criteria
func (td *TrendDetector) GetTrends(ctx context.Context, filter trend.Filter) ([]trend.Trend, error) {
	return td.trendStore.FindTrends(ctx, filter)
}

// GetTrendByID returns a specific trend by ID
func (td *TrendDetector) GetTrendByID(ctx context.Context, id string) (*trend.Trend, error) {
	return td.trendStore.GetTrend(ctx, id)
}

// GetTrendsForLocation returns trends relevant to a specific location
func (td *TrendDetector) GetTrendsForLocation(ctx context.Context, location trend.Location, radiusKm float64) ([]trend.Trend, error) {
	return td.trendStore.FindTrendsForLocation(ctx, location, radiusKm)
}

// AddPlatform adds a platform to monitor
func (td *TrendDetector) AddPlatform(ctx context.Context, platformConfig map[string]interface{}) error {
	// This is a simplified placeholder. In a real implementation, we'd create the
	// appropriate platform instance based on the config.

	// Example with validation
	if platformConfig["type"] == nil {
		return fmt.Errorf("platform type is required")
	}

	platformType, ok := platformConfig["type"].(string)
	if !ok {
		return fmt.Errorf("platform type must be a string")
	}

	// Platform factory implementation would go here
	var platform SocialPlatform
	switch platformType {
	case "twitter":
		// Create Twitter platform
		// platform = twitter.NewPlatform(platformConfig)
		return fmt.Errorf("twitter platform not implemented")
	case "reddit":
		// Create Reddit platform
		// platform = reddit.NewPlatform(platformConfig)
		return fmt.Errorf("reddit platform not implemented")
	// Add other platforms
	default:
		return fmt.Errorf("unsupported platform type: %s", platformType)
	}

	// Add platform to the map
	td.platformsLock.Lock()
	td.platforms[platform.Name()] = platform
	td.platformsLock.Unlock()

	// If detector is already running, start the platform
	if td.ctx.Err() == nil {
		return platform.Start(td.ctx)
	}

	return nil
}

// RemovePlatform removes a platform from monitoring
func (td *TrendDetector) RemovePlatform(ctx context.Context, platformID string) error {
	td.platformsLock.RLock()
	platform, exists := td.platforms[platformID]
	td.platformsLock.RUnlock()

	if !exists {
		return fmt.Errorf("platform not found: %s", platformID)
	}

	// Stop the platform
	if err := platform.Stop(); err != nil {
		return fmt.Errorf("error stopping platform: %v", err)
	}

	// Remove from map
	td.platformsLock.Lock()
	delete(td.platforms, platformID)
	td.platformsLock.Unlock()

	return nil
}

// RegisterTrendHandler registers a callback function for when new trends are detected
func (td *TrendDetector) RegisterTrendHandler(handler func(trend.Trend) error) error {
	td.mu.Lock()
	defer td.mu.Unlock()

	td.trendHandlers = append(td.trendHandlers, handler)
	return nil
}

// analyzeCrossPlatformTrends analyzes trends across platforms
func (td *TrendDetector) analyzeCrossPlatformTrends(ctx context.Context) {
	defer td.wg.Done()

	ticker := time.NewTicker(td.config.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			td.processCrossPlatformTrends(ctx)
		}
	}
}

// processCrossPlatformTrends processes trends from all platforms
func (td *TrendDetector) processCrossPlatformTrends(ctx context.Context) {
	// Collect trends from all platforms
	platformTrends := make(map[string][]trend.Trend)

	td.platformsLock.RLock()
	for name, platform := range td.platforms {
		trends, err := platform.GetTrends(ctx)
		if err != nil {
			fmt.Printf("Error getting trends from %s: %v\n", name, err)
			continue
		}
		platformTrends[name] = trends
	}
	td.platformsLock.RUnlock()

	// Correlate trends across platforms
	correlatedTrends, err := td.analyzer.CorrelateAcrossPlatforms(ctx, platformTrends)
	if err != nil {
		fmt.Printf("Error correlating trends: %v\n", err)
		return
	}

	// Process each correlated trend
	for i := range correlatedTrends {
		trend := correlatedTrends[i]

		// Calculate final trend score
		score, err := td.analyzer.CalculateTrendScore(ctx, &trend)
		if err != nil {
			fmt.Printf("Error calculating trend score: %v\n", err)
			continue
		}

		trend.Score = score

		// Check if trend exceeds threshold
		if trend.Score < td.config.TrendThreshold {
			continue
		}

		// Generate ID if not exists
		if trend.ID == "" {
			trend.ID = uuid.New().String()
		}

		// Save trend
		if err := td.trendStore.SaveTrend(ctx, trend); err != nil {
			fmt.Printf("Error saving trend: %v\n", err)
			continue
		}

		// Publish trend detected event
		if err := td.publishTrendEvent(trend); err != nil {
			fmt.Printf("Error publishing trend event: %v\n", err)
		}

		// Call registered handlers
		td.callTrendHandlers(trend)
	}
}

// analyzeGeoTrends analyzes location-specific trends
func (td *TrendDetector) analyzeGeoTrends(ctx context.Context) {
	defer td.wg.Done()

	ticker := time.NewTicker(td.config.GeoScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			td.processGeoTrends(ctx)
		}
	}
}

// processGeoTrends processes location-specific trends
func (td *TrendDetector) processGeoTrends(ctx context.Context) {
	// Get significant locations
	locations, err := td.geoTagger.GetSignificantLocations(ctx)
	if err != nil {
		fmt.Printf("Error getting significant locations: %v\n", err)
		return
	}

	for _, location := range locations {
		// Get trends for this location
		filter := trend.Filter{
			Location: &location,
			WithinKm: 50, // Adjustable radius
		}

		trends, err := td.GetTrends(ctx, filter)
		if err != nil {
			fmt.Printf("Error getting trends for location: %v\n", err)
			continue
		}

		for i := range trends {
			t := trends[i]

			// Check if this is primarily a local trend
			isLocal, err := td.geoTagger.IsLocalTrend(ctx, &t)
			if err != nil {
				fmt.Printf("Error determining if trend is local: %v\n", err)
				continue
			}

			// Only process local trends
			if !isLocal {
				continue
			}

			// Update trend with geo information
			t.IsGeoLocal = true

			// Calculate appropriate radius
			radius, err := td.geoTagger.GetLocationRadius(ctx, &t)
			if err != nil {
				fmt.Printf("Error getting location radius: %v\n", err)
			} else {
				t.LocationRadius = radius
			}

			// Save updated trend
			if err := td.trendStore.SaveTrend(ctx, t); err != nil {
				fmt.Printf("Error saving geo trend: %v\n", err)
				continue
			}

			// Publish geo trend event
			if err := td.publishGeoTrendEvent(t); err != nil {
				fmt.Printf("Error publishing geo trend event: %v\n", err)
			}

			// Call registered handlers
			td.callTrendHandlers(t)
		}
	}
}

// publishTrendEvent publishes a trend detected event
func (td *TrendDetector) publishTrendEvent(t trend.Trend) error {
	// Serialize the trend to JSON
	// In a real implementation, we'd use proper serialization
	data := []byte(fmt.Sprintf(`{"id":"%s","topic":"%s","score":%f}`, t.ID, t.Topic, t.Score))

	// Publish to event bus
	topic := fmt.Sprintf("%s.detected", td.config.EventsTopic)
	return td.eventBus.Publish(topic, data)
}

// publishGeoTrendEvent publishes a geo trend detected event
func (td *TrendDetector) publishGeoTrendEvent(t trend.Trend) error {
	// Serialize the trend to JSON
	// In a real implementation, we'd use proper serialization
	data := []byte(fmt.Sprintf(`{"id":"%s","topic":"%s","score":%f,"isGeoLocal":true}`, t.ID, t.Topic, t.Score))

	// Publish to event bus
	topic := fmt.Sprintf("%s.geo.detected", td.config.EventsTopic)
	return td.eventBus.Publish(topic, data)
}

// callTrendHandlers calls all registered trend handlers
func (td *TrendDetector) callTrendHandlers(t trend.Trend) {
	td.mu.RLock()
	handlers := make([]func(trend.Trend) error, len(td.trendHandlers))
	copy(handlers, td.trendHandlers)
	td.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(t); err != nil {
			fmt.Printf("Error in trend handler: %v\n", err)
		}
	}
}

// Stop gracefully stops the trend detection process
func (td *TrendDetector) Stop(ctx context.Context) error {
	// Signal all goroutines to stop
	td.cancel()

	// Stop all platforms
	td.platformsLock.RLock()
	for _, platform := range td.platforms {
		if err := platform.Stop(); err != nil {
			// Log error but continue stopping other platforms
			fmt.Printf("Error stopping platform %s: %v\n", platform.Name(), err)
		}
	}
	td.platformsLock.RUnlock()

	// Wait for all goroutines to finish with a timeout
	c := make(chan struct{})
	go func() {
		td.wg.Wait()
		close(c)
	}()

	select {
	case <-c:
		// All goroutines finished
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
