package models

import (
	"time"
)

// Space represents a conversation space
type Space struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	TemplateType   string     `json:"templateType"`
	LifecycleStage string     `json:"lifecycleStage"`
	CreatedAt      time.Time  `json:"createdAt"`
	LastActive     time.Time  `json:"lastActive"`
	UserCount      int        `json:"userCount"`
	MessageCount   int        `json:"messageCount"`
	IsGeoLocal     bool       `json:"isGeoLocal"`
	TopicTags      []string   `json:"topicTags"`
	TrendName      string     `json:"trendName,omitempty"`
	CreatedBy      string     `json:"createdBy,omitempty"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	Location       *Location  `json:"location,omitempty"`
	LocationRadius *float64   `json:"locationRadius,omitempty"`
	RelatedSpaces  []string   `json:"relatedSpaces,omitempty"`
}

// Location represents geographic coordinates
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
