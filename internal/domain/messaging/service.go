// internal/domain/messaging/service.go

package messaging

import (
	"context"
	"time"

	"essg/internal/domain/geo"
	"essg/internal/domain/identity"
	"essg/internal/domain/trend"
)

// MessageType defines the type of message
type MessageType string

const (
	TypeText     MessageType = "text"
	TypeMedia    MessageType = "media"
	TypeSystem   MessageType = "system"
	TypeEvent    MessageType = "event"
	TypeLocation MessageType = "location"
)

// MessageStatus defines the current status of a message
type MessageStatus string

const (
	StatusSending   MessageStatus = "sending"
	StatusDelivered MessageStatus = "delivered"
	StatusRead      MessageStatus = "read"
	StatusFailed    MessageStatus = "failed"
	StatusRemoved   MessageStatus = "removed"
)

// Message represents a single message in a space
type Message struct {
	ID                 string
	SpaceID            string
	UserID             string
	EphemeralIdentity  *identity.EphemeralIdentity
	Type               MessageType
	Content            string
	MediaURLs          []string
	Metadata           map[string]interface{}
	ReplyToID          string
	Status             MessageStatus
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Reactions          map[string]int
	Location           *trend.Location
	LocationContext    *geo.LocationContext
	DistanceFromCenter float64 // For geo-local spaces
	IsAnonymous        bool
	VisibleToRoles     []string // Empty means visible to everyone
}

// Service defines the interface for messaging services
type Service interface {
	// SendMessage sends a message to a space
	SendMessage(ctx context.Context, message Message) (*Message, error)

	// GetMessage retrieves a message by ID
	GetMessage(ctx context.Context, id string) (*Message, error)

	// GetMessages retrieves messages for a space
	GetMessages(ctx context.Context, spaceID string, filter MessageFilter) ([]Message, error)

	// UpdateMessage updates a message
	UpdateMessage(ctx context.Context, message Message) error

	// DeleteMessage marks a message as removed
	DeleteMessage(ctx context.Context, id string) error

	// AddReaction adds a reaction to a message
	AddReaction(ctx context.Context, messageID, userID, reaction string) error

	// RemoveReaction removes a reaction from a message
	RemoveReaction(ctx context.Context, messageID, userID, reaction string) error

	// SubscribeToSpace subscribes to real-time messages for a space
	SubscribeToSpace(ctx context.Context, spaceID string, callback func(Message)) (string, error)

	// UnsubscribeFromSpace unsubscribes from a space
	UnsubscribeFromSpace(ctx context.Context, subscriptionID string) error

	// EnrichWithGeoContext adds location context to a message
	EnrichWithGeoContext(ctx context.Context, message Message) (Message, error)
}

// MessageFilter defines criteria for filtering messages
type MessageFilter struct {
	Types         []MessageType
	FromUserID    string
	CreatedAfter  time.Time
	CreatedBefore time.Time
	ReplyToID     string
	WithLocation  bool
	Limit         int
	Offset        int
}

// RateLimiter defines rate limiting for messaging
type RateLimiter interface {
	// CheckLimit checks if an action exceeds rate limits
	CheckLimit(userID, actionType, resourceID string) (bool, error)

	// RecordAction records an action for rate limiting
	RecordAction(userID, actionType, resourceID string) error

	// GetRemainingLimit gets remaining actions allowed
	GetRemainingLimit(userID, actionType, resourceID string) (int, time.Time, error)
}

// MessageProcessor handles content processing for messages
type MessageProcessor interface {
	// Process processes a message before sending
	Process(ctx context.Context, message Message) (Message, error)

	// Filter applies content filtering to a message
	Filter(ctx context.Context, message Message) (Message, bool, error)

	// Enrich adds additional information to a message
	Enrich(ctx context.Context, message Message) (Message, error)
}
