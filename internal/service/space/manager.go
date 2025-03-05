// internal/service/space/manager.go

package space

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"essg/internal/domain/space"
	"essg/internal/domain/trend"
)

// SpaceStore defines the storage interface for spaces
type SpaceStore interface {
	// SaveSpace saves a space to storage
	SaveSpace(ctx context.Context, s space.Space) error

	// GetSpace retrieves a space by ID
	GetSpace(ctx context.Context, id string) (*space.Space, error)

	// FindSpaces finds spaces matching the filter
	FindSpaces(ctx context.Context, filter space.SpaceFilter) ([]space.Space, error)

	// FindNearbySpaces finds spaces near a location
	FindNearbySpaces(ctx context.Context, location trend.Location, radiusKm float64) ([]space.Space, error)
}

// SpaceManagerConfig contains configuration for the space manager
type SpaceManagerConfig struct {
	EventsTopic         string
	DefaultGracePeriod  time.Duration
	MonitoringInterval  time.Duration
	MaxConcurrentSpaces int
}

// SpaceManager implements the space.Manager interface
type SpaceManager struct {
	spaceStore         SpaceStore
	spaceTemplates     map[space.TemplateType]space.Template
	engagementAnalyzer space.EngagementAnalyzer
	eventBus           *nats.Conn
	config             SpaceManagerConfig
	lifecycleHandlers  []func(space.Space, space.LifecycleStage) error
	activeSpaces       sync.Map
	ctx                context.Context
	cancel             context.CancelFunc
	mu                 sync.RWMutex
	wg                 sync.WaitGroup
}

// NewSpaceManager creates a new space manager
func NewSpaceManager(
	spaceStore SpaceStore,
	engagementAnalyzer space.EngagementAnalyzer,
	eventBus *nats.Conn,
	config SpaceManagerConfig,
) *SpaceManager {
	ctx, cancel := context.WithCancel(context.Background())

	sm := &SpaceManager{
		spaceStore:         spaceStore,
		spaceTemplates:     make(map[space.TemplateType]space.Template),
		engagementAnalyzer: engagementAnalyzer,
		eventBus:           eventBus,
		config:             config,
		lifecycleHandlers:  []func(space.Space, space.LifecycleStage) error{},
		ctx:                ctx,
		cancel:             cancel,
	}

	// Start background monitoring of active spaces
	go sm.monitorActiveSpaces()

	return sm
}

// RegisterTemplate registers a space template
func (sm *SpaceManager) RegisterTemplate(template space.Template) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.spaceTemplates[template.GetType()] = template
}

// CreateSpace creates a new ephemeral space from a detected trend
func (sm *SpaceManager) CreateSpace(ctx context.Context, trend trend.Trend) (*space.Space, error) {
	// Select the best template for this trend
	template := sm.selectBestTemplate(trend)
	if template == nil {
		return nil, fmt.Errorf("no suitable template found for trend")
	}

	// Create new space from template
	s := template.Instantiate(trend)

	// Generate unique ID if not present
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	// Set created time
	s.CreatedAt = time.Now()
	s.LastActive = time.Now()

	// Set initial lifecycle stage
	s.LifecycleStage = space.StageCreating

	// Set location data if trend has location
	if trend.Location != nil {
		s.Location = trend.Location
		s.LocationRadius = trend.LocationRadius
		s.IsGeoLocal = trend.IsGeoLocal
	}

	// Save to storage
	if err := sm.spaceStore.SaveSpace(ctx, *s); err != nil {
		return nil, fmt.Errorf("error saving space: %w", err)
	}

	// Track in active spaces
	sm.activeSpaces.Store(s.ID, s)

	// Start engagement monitoring
	if err := sm.engagementAnalyzer.StartMonitoring(ctx, s.ID); err != nil {
		return nil, fmt.Errorf("error starting engagement monitoring: %w", err)
	}

	// Publish space created event
	if err := sm.publishSpaceEvent(*s, "created"); err != nil {
		// Log error but continue
		fmt.Printf("Error publishing space created event: %v\n", err)
	}

	// Update lifecycle to growing
	if err := sm.UpdateLifecycle(ctx, s.ID, space.StageGrowing); err != nil {
		// Log error but continue
		fmt.Printf("Error updating lifecycle to growing: %v\n", err)
	}

	return s, nil
}

// GetSpace returns a space by ID
func (sm *SpaceManager) GetSpace(ctx context.Context, id string) (*space.Space, error) {
	return sm.spaceStore.GetSpace(ctx, id)
}

// ListSpaces returns spaces matching the given filter
func (sm *SpaceManager) ListSpaces(ctx context.Context, filter space.SpaceFilter) ([]space.Space, error) {
	return sm.spaceStore.FindSpaces(ctx, filter)
}

// UpdateLifecycle updates a space's lifecycle stage
func (sm *SpaceManager) UpdateLifecycle(ctx context.Context, spaceID string, stage space.LifecycleStage) error {
	// Get current space
	s, err := sm.spaceStore.GetSpace(ctx, spaceID)
	if err != nil {
		return fmt.Errorf("error getting space: %w", err)
	}

	// If stage is the same, do nothing
	if s.LifecycleStage == stage {
		return nil
	}

	// Special handling for dissolution
	if stage == space.StageDevolving {
		return sm.InitiateDissolution(ctx, spaceID, sm.config.DefaultGracePeriod)
	}

	// Update stage
	prevStage := s.LifecycleStage
	s.LifecycleStage = stage

	// Save updated space
	if err := sm.spaceStore.SaveSpace(ctx, *s); err != nil {
		return fmt.Errorf("error saving space with updated lifecycle: %w", err)
	}

	// Publish lifecycle changed event
	if err := sm.publishLifecycleEvent(*s, prevStage, stage); err != nil {
		// Log error but continue
		fmt.Printf("Error publishing lifecycle event: %v\n", err)
	}

	// Call lifecycle handlers
	sm.callLifecycleHandlers(*s, stage)

	return nil
}

// InitiateDissolution begins the dissolution process for a space
func (sm *SpaceManager) InitiateDissolution(ctx context.Context, spaceID string, gracePeriod time.Duration) error {
	// Get current space
	s, err := sm.spaceStore.GetSpace(ctx, spaceID)
	if err != nil {
		return fmt.Errorf("error getting space: %w", err)
	}

	// Set dissolution time
	expiresAt := time.Now().Add(gracePeriod)
	s.ExpiresAt = &expiresAt

	// Update lifecycle stage
	prevStage := s.LifecycleStage
	s.LifecycleStage = space.StageDevolving

	// Save updated space
	if err := sm.spaceStore.SaveSpace(ctx, *s); err != nil {
		return fmt.Errorf("error saving space with dissolution info: %w", err)
	}

	// Schedule final dissolution
	go func() {
		select {
		case <-sm.ctx.Done():
			return
		case <-time.After(gracePeriod):
			// Final dissolution logic
			dissolutionCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Update to dissolved state
			finalSpace, err := sm.spaceStore.GetSpace(dissolutionCtx, spaceID)
			if err != nil {
				fmt.Printf("Error getting space during final dissolution: %v\n", err)
				return
			}

			finalSpace.LifecycleStage = space.StageDissolved

			// Save final state
			if err := sm.spaceStore.SaveSpace(dissolutionCtx, *finalSpace); err != nil {
				fmt.Printf("Error saving final dissolution state: %v\n", err)
				return
			}

			// Stop monitoring
			if err := sm.engagementAnalyzer.StopMonitoring(dissolutionCtx, spaceID); err != nil {
				fmt.Printf("Error stopping engagement monitoring: %v\n", err)
			}

			// Remove from active spaces
			sm.activeSpaces.Delete(spaceID)

			// Publish dissolved event
			if err := sm.publishLifecycleEvent(*finalSpace, space.StageDevolving, space.StageDissolved); err != nil {
				fmt.Printf("Error publishing final dissolution event: %v\n", err)
			}

			// Call lifecycle handlers
			sm.callLifecycleHandlers(*finalSpace, space.StageDissolved)
		}
	}()

	// Publish dissolution initiated event
	if err := sm.publishLifecycleEvent(*s, prevStage, space.StageDevolving); err != nil {
		// Log error but continue
		fmt.Printf("Error publishing dissolution event: %v\n", err)
	}

	// Call lifecycle handlers
	sm.callLifecycleHandlers(*s, space.StageDevolving)

	return nil
}

// GetNearbySpaces returns spaces near a specific location
func (sm *SpaceManager) GetNearbySpaces(ctx context.Context, location trend.Location, radiusKm float64) ([]space.Space, error) {
	return sm.spaceStore.FindNearbySpaces(ctx, location, radiusKm)
}

// RegisterLifecycleHandler registers a callback for lifecycle changes
func (sm *SpaceManager) RegisterLifecycleHandler(handler func(space.Space, space.LifecycleStage) error) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.lifecycleHandlers = append(sm.lifecycleHandlers, handler)
	return nil
}

// Stop gracefully stops the space manager
func (sm *SpaceManager) Stop(ctx context.Context) error {
	// Signal all goroutines to stop
	sm.cancel()

	// Create channel for wait group completion
	c := make(chan struct{})
	go func() {
		sm.wg.Wait()
		close(c)
	}()

	// Wait for all goroutines with timeout
	select {
	case <-c:
		// All goroutines finished
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// selectBestTemplate selects the most appropriate template for a trend
func (sm *SpaceManager) selectBestTemplate(t trend.Trend) space.Template {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// If this is a geo-local trend, prioritize local template
	if t.IsGeoLocal && sm.spaceTemplates[space.TemplateLocal] != nil {
		return sm.spaceTemplates[space.TemplateLocal]
	}

	// Check for breaking news
	if isBreakingNews(t) && sm.spaceTemplates[space.TemplateBreakingNews] != nil {
		return sm.spaceTemplates[space.TemplateBreakingNews]
	}

	// Check for event-based trends
	if isEventBased(t) && sm.spaceTemplates[space.TemplateEvent] != nil {
		return sm.spaceTemplates[space.TemplateEvent]
	}

	// Check for discussion-based trends
	if isDiscussionBased(t) && sm.spaceTemplates[space.TemplateDiscussion] != nil {
		return sm.spaceTemplates[space.TemplateDiscussion]
	}

	// Fall back to general template
	return sm.spaceTemplates[space.TemplateGeneral]
}

// Helper functions for trend categorization
func isBreakingNews(t trend.Trend) bool {
	// Simplified logic - in a real implementation, this would be more sophisticated
	if entityTypes, ok := t.EntityTypes["news"]; ok && entityTypes > 0.7 {
		return true
	}
	return t.Velocity > 5.0 // High velocity often indicates breaking news
}

func isEventBased(t trend.Trend) bool {
	// Check for event entity types
	if entityTypes, ok := t.EntityTypes["event"]; ok && entityTypes > 0.6 {
		return true
	}
	return false
}

func isDiscussionBased(t trend.Trend) bool {
	// Discussion trends often have more varied sources and lower velocity
	return len(t.Sources) > 2 && t.Velocity < 3.0
}

// monitorActiveSpaces regularly checks all active spaces for lifecycle updates
func (sm *SpaceManager) monitorActiveSpaces() {
	ticker := time.NewTicker(sm.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.checkActiveSpaces()
		}
	}
}

// checkActiveSpaces checks all active spaces for potential lifecycle updates
func (sm *SpaceManager) checkActiveSpaces() {
	// Create context with timeout for this check
	ctx, cancel := context.WithTimeout(sm.ctx, 30*time.Second)
	defer cancel()

	// Process each active space
	sm.activeSpaces.Range(func(key, value interface{}) bool {
		spaceID, ok := key.(string)
		if !ok {
			return true // Continue to next item
		}

		// Get latest space data
		s, err := sm.spaceStore.GetSpace(ctx, spaceID)
		if err != nil {
			fmt.Printf("Error getting space during monitoring: %v\n", err)
			return true
		}

		// Skip spaces in terminal states
		if s.LifecycleStage == space.StageDissolved {
			sm.activeSpaces.Delete(spaceID)
			return true
		}

		// Skip spaces already dissolving
		if s.LifecycleStage == space.StageDevolving {
			return true
		}

		// Analyze engagement to determine lifecycle stage
		stage, err := sm.engagementAnalyzer.DetermineLifecycleStage(ctx, s)
		if err != nil {
			fmt.Printf("Error determining lifecycle stage: %v\n", err)
			return true
		}

		// Check if stage is different and update if needed
		if stage != s.LifecycleStage {
			if err := sm.UpdateLifecycle(ctx, spaceID, stage); err != nil {
				fmt.Printf("Error updating lifecycle: %v\n", err)
			}
		}

		// Check if space should be dissolved
		if s.LifecycleStage != space.StageDevolving {
			shouldDissolve, err := sm.engagementAnalyzer.ShouldDissolve(ctx, s)
			if err != nil {
				fmt.Printf("Error checking dissolution: %v\n", err)
			} else if shouldDissolve {
				if err := sm.InitiateDissolution(ctx, spaceID, sm.config.DefaultGracePeriod); err != nil {
					fmt.Printf("Error initiating dissolution: %v\n", err)
				}
			}
		}

		return true
	})
}

// publishSpaceEvent publishes a space event to the event bus
func (sm *SpaceManager) publishSpaceEvent(s space.Space, eventType string) error {
	// In a real implementation, we'd properly serialize the space to JSON
	data := []byte(fmt.Sprintf(`{"id":"%s","title":"%s","event":"%s"}`, s.ID, s.Title, eventType))

	topic := fmt.Sprintf("%s.%s", sm.config.EventsTopic, eventType)
	return sm.eventBus.Publish(topic, data)
}

// publishLifecycleEvent publishes a lifecycle change event
func (sm *SpaceManager) publishLifecycleEvent(s space.Space, prevStage, newStage space.LifecycleStage) error {
	// In a real implementation, we'd properly serialize the event to JSON
	data := []byte(fmt.Sprintf(
		`{"id":"%s","title":"%s","prevStage":"%s","newStage":"%s"}`,
		s.ID, s.Title, prevStage, newStage,
	))

	topic := fmt.Sprintf("%s.lifecycle.changed", sm.config.EventsTopic)
	return sm.eventBus.Publish(topic, data)
}

// callLifecycleHandlers calls all registered lifecycle handlers
func (sm *SpaceManager) callLifecycleHandlers(s space.Space, stage space.LifecycleStage) {
	sm.mu.RLock()
	handlers := make([]func(space.Space, space.LifecycleStage) error, len(sm.lifecycleHandlers))
	copy(handlers, sm.lifecycleHandlers)
	sm.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(s, stage); err != nil {
			fmt.Printf("Error in lifecycle handler: %v\n", err)
		}
	}
}
