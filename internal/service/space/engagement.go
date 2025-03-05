// internal/service/space/engagement.go

package space

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nats-io/nats.go"

	"essg/internal/domain/geo"
	spaceDomain "essg/internal/domain/space"
	"essg/internal/domain/trend"
)

// EngagementAnalyzerConfig contains configuration for the engagement analyzer
type EngagementAnalyzerConfig struct {
	MonitoringInterval time.Duration
}

// EngagementAnalyzer implements the space.EngagementAnalyzer interface
type EngagementAnalyzer struct {
	db         *pgxpool.Pool
	eventBus   *nats.Conn
	geoService geo.Service
	config     EngagementAnalyzerConfig
	monitoring sync.Map
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewEngagementAnalyzer creates a new engagement analyzer
func NewEngagementAnalyzer(
	db *pgxpool.Pool,
	eventBus *nats.Conn,
	geoService geo.Service,
	config EngagementAnalyzerConfig,
) *EngagementAnalyzer {
	ctx, cancel := context.WithCancel(context.Background())

	return &EngagementAnalyzer{
		db:         db,
		eventBus:   eventBus,
		geoService: geoService,
		config:     config,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// AnalyzeEngagement calculates engagement metrics for a space
func (e *EngagementAnalyzer) AnalyzeEngagement(ctx context.Context, spaceID string) (map[string]float64, error) {
	// Fetch space data
	s, err := e.getSpace(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("error fetching space: %w", err)
	}

	// Fetch recent messages
	recentMessages, err := e.getRecentMessages(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("error fetching messages: %w", err)
	}

	// Fetch active users
	activeUsers, err := e.getActiveUsers(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("error fetching active users: %w", err)
	}

	// Calculate activity metrics
	messageVelocity := e.calculateMessageVelocity(recentMessages)
	userRetention := e.calculateUserRetention(activeUsers, s.UserCount)
	messageDepth := e.calculateMessageDepth(recentMessages)

	// Create base metrics
	metrics := map[string]float64{
		"message_velocity": messageVelocity,
		"user_retention":   userRetention,
		"message_depth":    messageDepth,
	}

	// Calculate overall engagement score
	metrics["engagement_score"] = e.calculateEngagementScore(metrics)

	// Apply geo factors for local spaces
	if s.IsGeoLocal && s.Location != nil {
		geoFactors, err := e.calculateGeoFactors(ctx, s)
		if err != nil {
			// Log the error but continue with base metrics
			fmt.Printf("Error calculating geo factors: %v\n", err)
		} else {
			// Add geo-specific metrics
			for k, v := range geoFactors {
				metrics[k] = v
			}

			// Adjust engagement score for geo-local spaces
			metrics["engagement_score"] = metrics["engagement_score"] * metrics["geo_multiplier"]
		}
	}

	return metrics, nil
}

// DetermineLifecycleStage determines the appropriate lifecycle stage
func (e *EngagementAnalyzer) DetermineLifecycleStage(ctx context.Context, s *spaceDomain.Space) (spaceDomain.LifecycleStage, error) {
	// If space is already in terminal states, don't change
	if s.LifecycleStage == spaceDomain.StageDissolved ||
		s.LifecycleStage == spaceDomain.StageDevolving {
		return s.LifecycleStage, nil
	}

	// Get engagement metrics
	metrics, err := e.AnalyzeEngagement(ctx, s.ID)
	if err != nil {
		return s.LifecycleStage, fmt.Errorf("error analyzing engagement: %w", err)
	}

	// Get engagement score
	score := metrics["engagement_score"]

	// Determine stage based on score and current stage
	switch s.LifecycleStage {
	case spaceDomain.StageCreating:
		// Always move from Creating to Growing
		return spaceDomain.StageGrowing, nil

	case spaceDomain.StageGrowing:
		if score > 70 {
			return spaceDomain.StagePeak, nil
		}
		return spaceDomain.StageGrowing, nil

	case spaceDomain.StagePeak:
		if score < 40 {
			return spaceDomain.StageWaning, nil
		}
		return spaceDomain.StagePeak, nil

	case spaceDomain.StageWaning:
		if score > 60 {
			return spaceDomain.StagePeak, nil
		}
		if score < 20 {
			return spaceDomain.StageDevolving, nil
		}
		return spaceDomain.StageWaning, nil

	default:
		return spaceDomain.StageGrowing, nil
	}
}

// ShouldDissolve determines if a space should begin dissolution
func (e *EngagementAnalyzer) ShouldDissolve(ctx context.Context, s *spaceDomain.Space) (bool, error) {
	// If space is already dissolving or dissolved, no need to check
	if s.LifecycleStage == spaceDomain.StageDevolving ||
		s.LifecycleStage == spaceDomain.StageDissolved {
		return false, nil
	}

	// Get engagement metrics
	metrics, err := e.AnalyzeEngagement(ctx, s.ID)
	if err != nil {
		return false, fmt.Errorf("error analyzing engagement: %w", err)
	}

	// Get engagement score
	score := metrics["engagement_score"]

	// Check time since last activity
	timeSinceLastActive := time.Since(s.LastActive)

	// Check message velocity
	messageVelocity := metrics["message_velocity"]

	// Check user retention
	userRetention := metrics["user_retention"]

	// Dissolution criteria

	// Very low engagement score for a significant period
	if score < 10 && timeSinceLastActive > 2*time.Hour {
		return true, nil
	}

	// No messages for a long time
	if messageVelocity < 0.01 && timeSinceLastActive > 6*time.Hour {
		return true, nil
	}

	// Very low user retention and low message velocity
	if userRetention < 0.1 && messageVelocity < 0.5 && timeSinceLastActive > 4*time.Hour {
		return true, nil
	}

	// For geo-local spaces, use different criteria
	if s.IsGeoLocal {
		geoMultiplier, ok := metrics["geo_multiplier"]
		if ok && geoMultiplier < 0.2 && timeSinceLastActive > 4*time.Hour {
			return true, nil
		}
	}

	return false, nil
}

// StartMonitoring begins monitoring a space's engagement
func (e *EngagementAnalyzer) StartMonitoring(ctx context.Context, spaceID string) error {
	// Check if already monitoring
	if _, exists := e.monitoring.Load(spaceID); exists {
		return nil
	}

	// Start monitoring goroutine
	monitorCtx, cancel := context.WithCancel(e.ctx)
	e.monitoring.Store(spaceID, cancel)

	go e.monitorSpace(monitorCtx, spaceID)

	return nil
}

// StopMonitoring stops monitoring a space's engagement
func (e *EngagementAnalyzer) StopMonitoring(ctx context.Context, spaceID string) error {
	// Get cancel function
	cancelI, exists := e.monitoring.Load(spaceID)
	if !exists {
		return nil
	}

	// Call cancel function
	if cancel, ok := cancelI.(context.CancelFunc); ok {
		cancel()
	}

	// Remove from monitoring map
	e.monitoring.Delete(spaceID)

	return nil
}

// monitorSpace continuously monitors a space's engagement
func (e *EngagementAnalyzer) monitorSpace(ctx context.Context, spaceID string) {
	ticker := time.NewTicker(e.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get space
			s, err := e.getSpace(ctx, spaceID)
			if err != nil {
				fmt.Printf("Error getting space %s: %v\n", spaceID, err)
				continue
			}

			// Skip spaces in terminal states
			if s.LifecycleStage == spaceDomain.StageDissolved {
				e.StopMonitoring(context.Background(), spaceID)
				return
			}

			// Analyze engagement
			metrics, err := e.AnalyzeEngagement(ctx, spaceID)
			if err != nil {
				fmt.Printf("Error analyzing engagement for space %s: %v\n", spaceID, err)
				continue
			}

			// Store analytics
			if err := e.storeAnalytics(ctx, spaceID, s.LifecycleStage, metrics); err != nil {
				fmt.Printf("Error storing analytics for space %s: %v\n", spaceID, err)
			}

			// Publish metrics
			if err := e.publishMetrics(spaceID, metrics); err != nil {
				fmt.Printf("Error publishing metrics for space %s: %v\n", spaceID, err)
			}
		}
	}
}

// getSpace fetches a space by ID
func (e *EngagementAnalyzer) getSpace(ctx context.Context, spaceID string) (*spaceDomain.Space, error) {
	query := `
		SELECT id, title, description, trend_id, template_type::text, lifecycle_stage::text,
			created_at, last_active, user_count, message_count, 
			ST_X(location::geometry) as lat, ST_Y(location::geometry) as lng,
			location_radius, is_geo_local, topic_tags
		FROM spaces
		WHERE id = $1
	`

	var s spaceDomain.Space
	var lat, lng *float64
	var lifecycleStage, templateType string
	var topicTags []string

	err := e.db.QueryRow(ctx, query, spaceID).Scan(
		&s.ID, &s.Title, &s.Description, &s.TrendID, &templateType, &lifecycleStage,
		&s.CreatedAt, &s.LastActive, &s.UserCount, &s.MessageCount,
		&lat, &lng, &s.LocationRadius, &s.IsGeoLocal, &topicTags,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying space: %w", err)
	}

	// Set location if coordinates are present
	if lat != nil && lng != nil {
		s.Location = &trend.Location{
			Latitude:  *lat,
			Longitude: *lng,
			Timestamp: s.CreatedAt,
		}
	}

	// Set topic tags
	s.TopicTags = topicTags

	// Set enum values
	s.LifecycleStage = spaceDomain.LifecycleStage(lifecycleStage)
	s.TemplateType = spaceDomain.TemplateType(templateType)

	return &s, nil
}

// getRecentMessages fetches recent messages for a space
func (e *EngagementAnalyzer) getRecentMessages(ctx context.Context, spaceID string) ([]Message, error) {
	query := `
		SELECT id, user_id, created_at, reply_to_id
		FROM messages
		WHERE space_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := e.db.Query(ctx, query, spaceID)
	if err != nil {
		return nil, fmt.Errorf("error querying messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var replyToID *string

		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.CreatedAt, &replyToID); err != nil {
			return nil, fmt.Errorf("error scanning message: %w", err)
		}

		if replyToID != nil {
			msg.ReplyToID = *replyToID
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// Message represents a simplified message for engagement analysis
type Message struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	ReplyToID string
}

// getActiveUsers fetches active users for a space
func (e *EngagementAnalyzer) getActiveUsers(ctx context.Context, spaceID string) ([]string, error) {
	query := `
		SELECT DISTINCT user_id
		FROM messages
		WHERE space_id = $1
		AND created_at > NOW() - INTERVAL '1 hour'
	`

	rows, err := e.db.Query(ctx, query, spaceID)
	if err != nil {
		return nil, fmt.Errorf("error querying active users: %w", err)
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("error scanning user ID: %w", err)
		}
		users = append(users, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// calculateMessageVelocity calculates message velocity (messages per minute)
func (e *EngagementAnalyzer) calculateMessageVelocity(messages []Message) float64 {
	if len(messages) == 0 {
		return 0
	}

	// If only one message, use time since it was created
	if len(messages) == 1 {
		duration := time.Since(messages[0].CreatedAt)
		if duration.Minutes() < 1 {
			return 1.0 // At least one message per minute
		}
		return 1.0 / duration.Minutes()
	}

	// Get newest and oldest message times
	newest := messages[0].CreatedAt
	oldest := messages[len(messages)-1].CreatedAt

	// Calculate duration
	duration := newest.Sub(oldest)

	// Calculate messages per minute
	if duration.Minutes() < 1 {
		return float64(len(messages))
	}

	return float64(len(messages)) / duration.Minutes()
}

// calculateUserRetention calculates user retention
func (e *EngagementAnalyzer) calculateUserRetention(activeUsers []string, totalUsers int) float64 {
	if totalUsers == 0 {
		return 0
	}

	return float64(len(activeUsers)) / float64(totalUsers)
}

// calculateMessageDepth calculates message depth (replies / total messages)
func (e *EngagementAnalyzer) calculateMessageDepth(messages []Message) float64 {
	if len(messages) == 0 {
		return 0
	}

	// Count replies
	replies := 0
	for _, msg := range messages {
		if msg.ReplyToID != "" {
			replies++
		}
	}

	return float64(replies) / float64(len(messages))
}

// calculateEngagementScore calculates overall engagement score
func (e *EngagementAnalyzer) calculateEngagementScore(metrics map[string]float64) float64 {
	// Weights for different metrics
	weights := map[string]float64{
		"message_velocity": 0.5,
		"user_retention":   0.3,
		"message_depth":    0.2,
	}

	// Calculate weighted score
	var score float64
	for metric, weight := range weights {
		if value, ok := metrics[metric]; ok {
			score += value * weight
		}
	}

	// Normalize to 0-100 scale
	normalizedScore := math.Min(100, score*100)

	return normalizedScore
}

// calculateGeoFactors calculates geo-specific engagement factors
func (e *EngagementAnalyzer) calculateGeoFactors(ctx context.Context, s *spaceDomain.Space) (map[string]float64, error) {
	if s.Location == nil {
		return nil, fmt.Errorf("space has no location")
	}

	// Get population density
	density, err := e.geoService.GetPopulationDensity(ctx, *s.Location, s.LocationRadius)
	if err != nil {
		return nil, fmt.Errorf("error getting population density: %w", err)
	}

	// Count local users
	localUsers, err := e.countLocalUsers(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("error counting local users: %w", err)
	}

	// Calculate local user ratio
	localUserRatio := 0.0
	if s.UserCount > 0 {
		localUserRatio = float64(localUsers) / float64(s.UserCount)
	}

	// Calculate geo multiplier
	// Higher for spaces with more local users relative to population density
	geoMultiplier := 1.0
	if density.DensityPerKm2 > 0 {
		// Adjust for population density
		// More credit for engagement in less dense areas
		densityFactor := 1.0 / math.Log10(math.Max(10, density.DensityPerKm2))

		// Combine with local user ratio
		geoMultiplier = 1.0 + (localUserRatio * densityFactor * 2.0)
	}

	return map[string]float64{
		"geo_multiplier":     geoMultiplier,
		"local_user_ratio":   localUserRatio,
		"population_density": density.DensityPerKm2,
	}, nil
}

// countLocalUsers counts users within the space's location radius
func (e *EngagementAnalyzer) countLocalUsers(ctx context.Context, s *spaceDomain.Space) (int, error) {
	if s.Location == nil {
		return 0, fmt.Errorf("space has no location")
	}

	query := `
		SELECT COUNT(DISTINCT ei.user_id)
		FROM ephemeral_identities ei
		WHERE ei.space_id = $1
		AND ei.location IS NOT NULL
		AND ST_DWithin(
			geography(ei.location),
			geography(ST_MakePoint($2, $3)),
			$4 * 1000
		)
	`

	var count int
	err := e.db.QueryRow(
		ctx,
		query,
		s.ID,
		s.Location.Longitude,
		s.Location.Latitude,
		s.LocationRadius,
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("error querying local users: %w", err)
	}

	return count, nil
}

// storeAnalytics stores space analytics
func (e *EngagementAnalyzer) storeAnalytics(
	ctx context.Context,
	spaceID string,
	lifecycleStage spaceDomain.LifecycleStage,
	metrics map[string]float64,
) error {
	query := `
		INSERT INTO space_analytics (
			space_id, timestamp, user_count, message_count,
			active_users, engagement_score, lifecycle_stage, metrics
		) VALUES ($1, $2, $3, $4, $5, $6, $7::lifecycle_stage, $8)
	`

	// Get space data
	s, err := e.getSpace(ctx, spaceID)
	if err != nil {
		return fmt.Errorf("error getting space: %w", err)
	}

	// Get active users
	activeUsers, err := e.getActiveUsers(ctx, spaceID)
	if err != nil {
		return fmt.Errorf("error getting active users: %w", err)
	}

	// Convert metrics to JSONB
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("error marshaling metrics: %w", err)
	}

	// Execute query
	_, err = e.db.Exec(
		ctx,
		query,
		spaceID,
		time.Now(),
		s.UserCount,
		s.MessageCount,
		len(activeUsers),
		metrics["engagement_score"],
		string(lifecycleStage),
		metricsJSON,
	)

	if err != nil {
		return fmt.Errorf("error inserting analytics: %w", err)
	}

	return nil
}

// publishMetrics publishes engagement metrics to NATS
func (e *EngagementAnalyzer) publishMetrics(spaceID string, metrics map[string]float64) error {
	// Convert metrics to JSON
	metricsJSON, err := json.Marshal(map[string]interface{}{
		"space_id": spaceID,
		"time":     time.Now(),
		"metrics":  metrics,
	})

	if err != nil {
		return fmt.Errorf("error marshaling metrics: %w", err)
	}

	// Publish to NATS
	topic := fmt.Sprintf("space.%s.metrics", spaceID)
	if err := e.eventBus.Publish(topic, metricsJSON); err != nil {
		return fmt.Errorf("error publishing metrics: %w", err)
	}

	return nil
}
