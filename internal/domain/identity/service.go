// internal/domain/identity/service.go

package identity

import (
	"context"
	"time"

	"essg/internal/domain/trend"
)

// LocationSharingLevel defines how precisely a user's location is shared
type LocationSharingLevel string

const (
	LocationSharingDisabled     LocationSharingLevel = "disabled"
	LocationSharingApproximate  LocationSharingLevel = "approximate" // City level
	LocationSharingNeighborhood LocationSharingLevel = "neighborhood"
	LocationSharingPrecise      LocationSharingLevel = "precise"
)

// User represents a user of the system
type User struct {
	ID                        string
	ExternalIDs               map[string]string // Platform to ID mapping
	CreatedAt                 time.Time
	LastSeen                  time.Time
	Location                  *trend.Location
	LocationSharingPreference LocationSharingLevel
	DefaultAnonymity          bool
	NotificationPreferences   map[string]bool
}

// EphemeralIdentity represents a temporary identity in a specific space
type EphemeralIdentity struct {
	ID                 string
	UserID             string
	SpaceID            string
	Nickname           string
	Avatar             string
	IsAnonymous        bool
	Location           *trend.Location
	LocationShareLevel LocationSharingLevel
	CreatedAt          time.Time
	LastActive         time.Time
	Reputation         map[string]float64 // Different dimensions of reputation
}

// Service defines the interface for identity services
type Service interface {
	// GetOrCreateUser gets an existing user or creates a new one
	GetOrCreateUser(ctx context.Context, platformID, platform string) (*User, error)

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id string) (*User, error)

	// UpdateUserLocation updates a user's location
	UpdateUserLocation(ctx context.Context, userID string, location trend.Location) error

	// UpdateLocationSharing updates a user's location sharing preferences
	UpdateLocationSharing(ctx context.Context, userID string, level LocationSharingLevel) error

	// GetOrCreateEphemeralIdentity gets or creates an ephemeral identity for a space
	GetOrCreateEphemeralIdentity(ctx context.Context, userID, spaceID string) (*EphemeralIdentity, error)

	// UpdateEphemeralIdentity updates an ephemeral identity
	UpdateEphemeralIdentity(ctx context.Context, identity EphemeralIdentity) error

	// GetIdentitiesInSpace retrieves all identities in a space
	GetIdentitiesInSpace(ctx context.Context, spaceID string) ([]EphemeralIdentity, error)
}

// TokenManager handles authentication tokens
type TokenManager interface {
	// GenerateToken generates a token for a user
	GenerateToken(userID string, ttl time.Duration) (string, error)

	// ValidateToken validates a token and returns the user ID
	ValidateToken(token string) (string, error)

	// RevokeToken revokes a token
	RevokeToken(token string) error

	// RevokeAllForUser revokes all tokens for a user
	RevokeAllForUser(userID string) error
}

// PrivacyConfig defines privacy settings
type PrivacyConfig interface {
	// GetLocationSharingOptions returns available location sharing options
	GetLocationSharingOptions() []LocationSharingLevel

	// GetDefaultLocationSharing returns the default location sharing level
	GetDefaultLocationSharing() LocationSharingLevel

	// GetAnonymityOptions returns anonymity options
	GetAnonymityOptions() map[string]interface{}

	// GetDataRetentionPolicy returns the data retention policy
	GetDataRetentionPolicy() map[string]interface{}
}

// LocationPrivacyManager manages location privacy
type LocationPrivacyManager interface {
	// ApplyPrivacySettings applies privacy settings to a location
	ApplyPrivacySettings(location trend.Location, preference LocationSharingLevel, contextRadius float64) *trend.Location

	// GetLocationPrecisionLevel returns the precision level for a sharing preference
	GetLocationPrecisionLevel(preference LocationSharingLevel) float64
}
