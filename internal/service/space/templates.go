// internal/service/space/templates.go

package space

import (
	"time"

	"essg/internal/domain/space"
	"essg/internal/domain/trend"
)

// BaseTemplate provides common functionality for all templates
type BaseTemplate struct {
	templateType space.TemplateType
	features     []space.Feature
	isGeoAware   bool
}

// GetType returns the template type
func (t *BaseTemplate) GetType() space.TemplateType {
	return t.templateType
}

// GetFeatures returns the features enabled for this template
func (t *BaseTemplate) GetFeatures() []space.Feature {
	return t.features
}

// IsGeoAware returns true if this template supports location features
func (t *BaseTemplate) IsGeoAware() bool {
	return t.isGeoAware
}

// GeneralTemplate is a general purpose space template
type GeneralTemplate struct {
	BaseTemplate
}

// NewGeneralTemplate creates a new general template
func NewGeneralTemplate() *GeneralTemplate {
	return &GeneralTemplate{
		BaseTemplate: BaseTemplate{
			templateType: space.TemplateGeneral,
			features: []space.Feature{
				{
					ID:          "messaging",
					Name:        "Messaging",
					Description: "Basic text messaging",
					IsEnabled:   true,
				},
				{
					ID:          "reactions",
					Name:        "Reactions",
					Description: "Message reactions",
					IsEnabled:   true,
				},
				{
					ID:          "media",
					Name:        "Media Sharing",
					Description: "Share images and links",
					IsEnabled:   true,
				},
			},
			isGeoAware: false,
		},
	}
}

// Instantiate creates a new space instance from this template
func (t *GeneralTemplate) Instantiate(trend trend.Trend) *space.Space {
	return &space.Space{
		Title:          trend.Topic,
		Description:    trend.Description,
		TrendID:        trend.ID,
		TemplateType:   t.templateType,
		Features:       t.features,
		LifecycleStage: space.StageCreating,
		CreatedAt:      time.Now(),
		LastActive:     time.Now(),
		TopicTags:      trend.Keywords,
		IsGeoLocal:     false,
	}
}

// BreakingNewsTemplate is for breaking news discussions
type BreakingNewsTemplate struct {
	BaseTemplate
}

// NewBreakingNewsTemplate creates a new breaking news template
func NewBreakingNewsTemplate() *BreakingNewsTemplate {
	return &BreakingNewsTemplate{
		BaseTemplate: BaseTemplate{
			templateType: space.TemplateBreakingNews,
			features: []space.Feature{
				{
					ID:          "messaging",
					Name:        "Messaging",
					Description: "Basic text messaging",
					IsEnabled:   true,
				},
				{
					ID:          "reactions",
					Name:        "Reactions",
					Description: "Message reactions",
					IsEnabled:   true,
				},
				{
					ID:          "media",
					Name:        "Media Sharing",
					Description: "Share images and links",
					IsEnabled:   true,
				},
				{
					ID:          "source_validation",
					Name:        "Source Validation",
					Description: "Validate news sources",
					IsEnabled:   true,
					Config: map[string]interface{}{
						"trusted_domains": []string{
							"reuters.com",
							"apnews.com",
							"bbc.com",
							"nytimes.com",
							"washingtonpost.com",
						},
					},
				},
				{
					ID:          "timeline",
					Name:        "Timeline",
					Description: "Event timeline",
					IsEnabled:   true,
				},
			},
			isGeoAware: true,
		},
	}
}

// Instantiate creates a new space instance from this template
func (t *BreakingNewsTemplate) Instantiate(trend trend.Trend) *space.Space {
	s := &space.Space{
		Title:          trend.Topic,
		Description:    trend.Description,
		TrendID:        trend.ID,
		TemplateType:   t.templateType,
		Features:       t.features,
		LifecycleStage: space.StageCreating,
		CreatedAt:      time.Now(),
		LastActive:     time.Now(),
		TopicTags:      trend.Keywords,
		IsGeoLocal:     trend.IsGeoLocal,
	}

	// Add location data if available
	if trend.Location != nil {
		s.Location = trend.Location
		s.LocationRadius = trend.LocationRadius
	}

	return s
}

// EventTemplate is for event-based discussions
type EventTemplate struct {
	BaseTemplate
}

// NewEventTemplate creates a new event template
func NewEventTemplate() *EventTemplate {
	return &EventTemplate{
		BaseTemplate: BaseTemplate{
			templateType: space.TemplateEvent,
			features: []space.Feature{
				{
					ID:          "messaging",
					Name:        "Messaging",
					Description: "Basic text messaging",
					IsEnabled:   true,
				},
				{
					ID:          "reactions",
					Name:        "Reactions",
					Description: "Message reactions",
					IsEnabled:   true,
				},
				{
					ID:          "media",
					Name:        "Media Sharing",
					Description: "Share images and links",
					IsEnabled:   true,
				},
				{
					ID:          "attendees",
					Name:        "Attendees",
					Description: "Track event attendees",
					IsEnabled:   true,
				},
				{
					ID:          "event_details",
					Name:        "Event Details",
					Description: "Show event details",
					IsEnabled:   true,
				},
			},
			isGeoAware: true,
		},
	}
}

// Instantiate creates a new space instance from this template
func (t *EventTemplate) Instantiate(trend trend.Trend) *space.Space {
	s := &space.Space{
		Title:          trend.Topic,
		Description:    trend.Description,
		TrendID:        trend.ID,
		TemplateType:   t.templateType,
		Features:       t.features,
		LifecycleStage: space.StageCreating,
		CreatedAt:      time.Now(),
		LastActive:     time.Now(),
		TopicTags:      trend.Keywords,
		IsGeoLocal:     trend.IsGeoLocal,
	}

	// Add location data if available
	if trend.Location != nil {
		s.Location = trend.Location
		s.LocationRadius = trend.LocationRadius
	}

	return s
}

// DiscussionTemplate is for in-depth discussions
type DiscussionTemplate struct {
	BaseTemplate
}

// NewDiscussionTemplate creates a new discussion template
func NewDiscussionTemplate() *DiscussionTemplate {
	return &DiscussionTemplate{
		BaseTemplate: BaseTemplate{
			templateType: space.TemplateDiscussion,
			features: []space.Feature{
				{
					ID:          "messaging",
					Name:        "Messaging",
					Description: "Basic text messaging",
					IsEnabled:   true,
				},
				{
					ID:          "reactions",
					Name:        "Reactions",
					Description: "Message reactions",
					IsEnabled:   true,
				},
				{
					ID:          "media",
					Name:        "Media Sharing",
					Description: "Share images and links",
					IsEnabled:   true,
				},
				{
					ID:          "threading",
					Name:        "Threaded Replies",
					Description: "Threaded conversation replies",
					IsEnabled:   true,
				},
				{
					ID:          "pinned_messages",
					Name:        "Pinned Messages",
					Description: "Pin important messages",
					IsEnabled:   true,
				},
				{
					ID:          "polls",
					Name:        "Polls",
					Description: "Create and vote in polls",
					IsEnabled:   true,
				},
			},
			isGeoAware: false,
		},
	}
}

// Instantiate creates a new space instance from this template
func (t *DiscussionTemplate) Instantiate(trend trend.Trend) *space.Space {
	return &space.Space{
		Title:          trend.Topic,
		Description:    trend.Description,
		TrendID:        trend.ID,
		TemplateType:   t.templateType,
		Features:       t.features,
		LifecycleStage: space.StageCreating,
		CreatedAt:      time.Now(),
		LastActive:     time.Now(),
		TopicTags:      trend.Keywords,
		IsGeoLocal:     false,
	}
}

// LocalTemplate is for location-based discussions
type LocalTemplate struct {
	BaseTemplate
}

// NewLocalTemplate creates a new local template
func NewLocalTemplate() *LocalTemplate {
	return &LocalTemplate{
		BaseTemplate: BaseTemplate{
			templateType: space.TemplateLocal,
			features: []space.Feature{
				{
					ID:          "messaging",
					Name:        "Messaging",
					Description: "Basic text messaging",
					IsEnabled:   true,
				},
				{
					ID:          "reactions",
					Name:        "Reactions",
					Description: "Message reactions",
					IsEnabled:   true,
				},
				{
					ID:          "media",
					Name:        "Media Sharing",
					Description: "Share images and links",
					IsEnabled:   true,
				},
				{
					ID:          "geo_context",
					Name:        "Geographic Context",
					Description: "Adds location context to messages",
					IsEnabled:   true,
				},
				{
					ID:          "proximity",
					Name:        "Proximity Indicators",
					Description: "Show proximity of users",
					IsEnabled:   true,
				},
				{
					ID:          "local_tags",
					Name:        "Local Tags",
					Description: "Location-specific tags",
					IsEnabled:   true,
				},
			},
			isGeoAware: true,
		},
	}
}

// Instantiate creates a new space instance from this template
func (t *LocalTemplate) Instantiate(trend trend.Trend) *space.Space {
	s := &space.Space{
		Title:          trend.Topic,
		Description:    trend.Description,
		TrendID:        trend.ID,
		TemplateType:   t.templateType,
		Features:       t.features,
		LifecycleStage: space.StageCreating,
		CreatedAt:      time.Now(),
		LastActive:     time.Now(),
		TopicTags:      trend.Keywords,
		IsGeoLocal:     true,
	}

	// Add location data if available
	if trend.Location != nil {
		s.Location = trend.Location
		s.LocationRadius = trend.LocationRadius
	}

	return s
}
