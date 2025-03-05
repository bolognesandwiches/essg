// internal/domain/space/manager.go

package space

import (
	"context"
	"time"

	"essg/internal/domain/trend"
)

// Template defines the structure and features for a space type
type Template interface {
	// GetType returns the template type
	GetType() TemplateType

	// Instantiate creates a new space instance from this template
	Instantiate(trend trend.Trend) *Space

	// GetFeatures returns the features enabled for this template
	GetFeatures() []Feature

	// IsGeoAware returns true if this template supports location features
	IsGeoAware() bool
}

// Manager defines the interface for space management
type Manager interface {
	// CreateSpace creates a new ephemeral space from a detected trend
	CreateSpace(ctx context.Context, trend trend.Trend) (*Space, error)

	// GetSpace returns a space by ID
	GetSpace(ctx context.Context, id string) (*Space, error)

	// ListSpaces returns spaces matching the given filter
	ListSpaces(ctx context.Context, filter SpaceFilter) ([]Space, error)

	// UpdateLifecycle updates a space's lifecycle stage
	UpdateLifecycle(ctx context.Context, spaceID string, stage LifecycleStage) error

	// InitiateDissolution begins the dissolution process for a space
	InitiateDissolution(ctx context.Context, spaceID string, gracePeriod time.Duration) error

	// GetNearbySpaces returns spaces near a specific location
	GetNearbySpaces(ctx context.Context, location trend.Location, radiusKm float64) ([]Space, error)

	// RegisterLifecycleHandler registers a callback for lifecycle changes
	RegisterLifecycleHandler(handler func(Space, LifecycleStage) error) error
}

// SpaceFilter defines criteria for filtering spaces
type SpaceFilter struct {
	LifecycleStages []LifecycleStage
	TemplateTypes   []TemplateType
	IsGeoLocal      *bool
	Location        *trend.Location
	WithinKm        float64
	MinUserCount    int
	MaxUserCount    int
	CreatedAfter    time.Time
	CreatedBefore   time.Time
	SearchTerms     string
	TopicTags       []string
	Limit           int
	Offset          int
}

// EngagementAnalyzer defines the interface for analyzing space engagement
type EngagementAnalyzer interface {
	// AnalyzeEngagement calculates engagement metrics for a space
	AnalyzeEngagement(ctx context.Context, spaceID string) (map[string]float64, error)

	// DetermineLifecycleStage determines the appropriate lifecycle stage
	DetermineLifecycleStage(ctx context.Context, space *Space) (LifecycleStage, error)

	// ShouldDissolve determines if a space should begin dissolution
	ShouldDissolve(ctx context.Context, space *Space) (bool, error)

	// StartMonitoring begins monitoring a space's engagement
	StartMonitoring(ctx context.Context, spaceID string) error

	// StopMonitoring stops monitoring a space's engagement
	StopMonitoring(ctx context.Context, spaceID string) error
}
