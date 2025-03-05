package space

import (
	"time"

	"essg/internal/domain/trend"
)

// LifecycleStage represents the current stage in a space's lifecycle
type LifecycleStage string

const (
	StageCreating  LifecycleStage = "creating"
	StageGrowing   LifecycleStage = "growing"
	StagePeak      LifecycleStage = "peak"
	StageWaning    LifecycleStage = "waning"
	StageDevolving LifecycleStage = "dissolving"
	StageDissolved LifecycleStage = "dissolved"
)

// TemplateType identifies the type of space template
type TemplateType string

const (
	TemplateGeneral      TemplateType = "general"
	TemplateBreakingNews TemplateType = "breaking_news"
	TemplateEvent        TemplateType = "event"
	TemplateDiscussion   TemplateType = "discussion"
	TemplateLocal        TemplateType = "local"
)

// Feature represents a feature enabled for a space
type Feature struct {
	ID          string
	Name        string
	Description string
	Config      map[string]interface{}
	IsEnabled   bool
}

// Space represents an ephemeral conversation space
type Space struct {
	ID                string
	Title             string
	Description       string
	TrendID           string
	TemplateType      TemplateType
	Features          []Feature
	LifecycleStage    LifecycleStage
	CreatedAt         time.Time
	LastActive        time.Time
	ExpiresAt         *time.Time
	UserCount         int
	MessageCount      int
	Location          *trend.Location
	LocationRadius    float64
	IsGeoLocal        bool
	TopicTags         []string
	RelatedSpaces     []string
	EngagementMetrics map[string]float64
}
