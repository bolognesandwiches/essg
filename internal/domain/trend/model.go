package trend

import (
	"time"
)

// Location represents a geographic point with optional metadata
type Location struct {
	Latitude  float64
	Longitude float64
	Accuracy  float64
	Timestamp time.Time
}

// Source identifies where a trend originated
type Source struct {
	Platform    string
	ExternalID  string
	URL         string
	AccessLevel string
}

// Trend represents a detected trending topic across platforms
type Trend struct {
	ID             string
	Topic          string
	Description    string
	Keywords       []string
	Score          float64
	Velocity       float64
	Sources        []Source
	Location       *Location
	LocationRadius float64
	IsGeoLocal     bool
	FirstDetected  time.Time
	LastUpdated    time.Time
	RelatedTrends  []string
	EntityTypes    map[string]float64
	RawData        map[string]interface{}
}

// Filter defines criteria for filtering trends
type Filter struct {
	MinScore          float64
	IncludePlatforms  []string
	ExcludePlatforms  []string
	GeoOnly           bool
	WithinKm          float64
	Location          *Location
	IncludeEntityType []string
}
