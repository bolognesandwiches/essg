// internal/adapter/storage/space_store.go

package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"

	"essg/internal/domain/space"
	"essg/internal/domain/trend"
)

// SpaceStore implements storage for spaces
type SpaceStore struct {
	db *pgxpool.Pool
}

// NewSpaceStore creates a new space store
func NewSpaceStore(db *pgxpool.Pool) *SpaceStore {
	return &SpaceStore{
		db: db,
	}
}

// SaveSpace saves a space to storage
func (s *SpaceStore) SaveSpace(ctx context.Context, sp space.Space) error {
	query := `
		INSERT INTO spaces (
			id, title, description, trend_id, template_type, lifecycle_stage,
			created_at, last_active, expires_at, user_count, message_count,
			location, location_radius, is_geo_local,
			topic_tags, related_spaces, engagement_metrics, features
		) VALUES (
			$1, $2, $3, $4, $5::template_type, $6::lifecycle_stage,
			$7, $8, $9, $10, $11,
			ST_MakePoint($12, $13)::geography, $14, $15,
			$16, $17, $18, $19
		)
		ON CONFLICT (id) DO UPDATE
		SET
			title = $2,
			description = $3,
			trend_id = $4,
			template_type = $5::template_type,
			lifecycle_stage = $6::lifecycle_stage,
			last_active = $8,
			expires_at = $9,
			user_count = $10,
			message_count = $11,
			location = CASE WHEN $12 IS NOT NULL AND $13 IS NOT NULL THEN ST_MakePoint($12, $13)::geography ELSE spaces.location END,
			location_radius = $14,
			is_geo_local = $15,
			topic_tags = $16,
			related_spaces = $17,
			engagement_metrics = $18,
			features = $19
	`

	// Prepare location data
	var lng, lat *float64
	if sp.Location != nil {
		lng = &sp.Location.Longitude
		lat = &sp.Location.Latitude
	}

	// Convert JSON fields
	metricsJSON, err := json.Marshal(sp.EngagementMetrics)
	if err != nil {
		return fmt.Errorf("error marshaling engagement metrics: %w", err)
	}

	featuresJSON, err := json.Marshal(sp.Features)
	if err != nil {
		return fmt.Errorf("error marshaling features: %w", err)
	}

	// Execute query
	_, err = s.db.Exec(
		ctx,
		query,
		sp.ID,
		sp.Title,
		sp.Description,
		sp.TrendID,
		string(sp.TemplateType),
		string(sp.LifecycleStage),
		sp.CreatedAt,
		sp.LastActive,
		sp.ExpiresAt,
		sp.UserCount,
		sp.MessageCount,
		lng,
		lat,
		sp.LocationRadius,
		sp.IsGeoLocal,
		sp.TopicTags,
		sp.RelatedSpaces,
		metricsJSON,
		featuresJSON,
	)

	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil
}

// GetSpace retrieves a space by ID
func (s *SpaceStore) GetSpace(ctx context.Context, id string) (*space.Space, error) {
	query := `
		SELECT
			id, title, description, trend_id, template_type::text, lifecycle_stage::text,
			created_at, last_active, expires_at, user_count, message_count,
			ST_X(location::geometry) as lng, ST_Y(location::geometry) as lat,
			location_radius, is_geo_local,
			topic_tags, related_spaces, engagement_metrics, features
		FROM spaces
		WHERE id = $1
	`

	var sp space.Space
	var templateType, lifecycleStage string
	var lng, lat *float64
	var metricsJSON, featuresJSON []byte

	err := s.db.QueryRow(ctx, query, id).Scan(
		&sp.ID,
		&sp.Title,
		&sp.Description,
		&sp.TrendID,
		&templateType,
		&lifecycleStage,
		&sp.CreatedAt,
		&sp.LastActive,
		&sp.ExpiresAt,
		&sp.UserCount,
		&sp.MessageCount,
		&lng,
		&lat,
		&sp.LocationRadius,
		&sp.IsGeoLocal,
		&sp.TopicTags,
		&sp.RelatedSpaces,
		&metricsJSON,
		&featuresJSON,
	)

	if err != nil {
		return nil, fmt.Errorf("error querying space: %w", err)
	}

	// Set location if coordinates are present
	if lng != nil && lat != nil {
		sp.Location = &trend.Location{
			Longitude: *lng,
			Latitude:  *lat,
			Timestamp: sp.CreatedAt,
		}
	}

	// Parse JSON fields
	if err := json.Unmarshal(metricsJSON, &sp.EngagementMetrics); err != nil {
		return nil, fmt.Errorf("error unmarshaling engagement metrics: %w", err)
	}

	if err := json.Unmarshal(featuresJSON, &sp.Features); err != nil {
		return nil, fmt.Errorf("error unmarshaling features: %w", err)
	}

	// Set enum types
	sp.TemplateType = space.TemplateType(templateType)
	sp.LifecycleStage = space.LifecycleStage(lifecycleStage)

	return &sp, nil
}

// FindSpaces finds spaces matching the filter
func (s *SpaceStore) FindSpaces(ctx context.Context, filter space.SpaceFilter) ([]space.Space, error) {
	// Build dynamic query
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT
			id, title, description, trend_id, template_type::text, lifecycle_stage::text,
			created_at, last_active, user_count, message_count,
			ST_X(location::geometry) as lng, ST_Y(location::geometry) as lat,
			location_radius, is_geo_local,
			topic_tags
		FROM spaces
		WHERE 1=1
	`)

	args := []interface{}{}
	argIndex := 1

	// Add lifecycle stage filter
	if len(filter.LifecycleStages) > 0 {
		queryBuilder.WriteString(" AND lifecycle_stage IN (")

		for i, stage := range filter.LifecycleStages {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(fmt.Sprintf("$%d::lifecycle_stage", argIndex))
			args = append(args, string(stage))
			argIndex++
		}

		queryBuilder.WriteString(")")
	}

	// Add template type filter
	if len(filter.TemplateTypes) > 0 {
		queryBuilder.WriteString(" AND template_type IN (")

		for i, templateType := range filter.TemplateTypes {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(fmt.Sprintf("$%d::template_type", argIndex))
			args = append(args, string(templateType))
			argIndex++
		}

		queryBuilder.WriteString(")")
	}

	// Add geo filter
	if filter.IsGeoLocal != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND is_geo_local = $%d", argIndex))
		args = append(args, *filter.IsGeoLocal)
		argIndex++
	}

	// Add location filter
	if filter.Location != nil && filter.WithinKm > 0 {
		queryBuilder.WriteString(fmt.Sprintf(
			" AND ST_DWithin(geography(location), geography(ST_MakePoint($%d, $%d)), $%d * 1000)",
			argIndex, argIndex+1, argIndex+2,
		))
		args = append(args, filter.Location.Longitude, filter.Location.Latitude, filter.WithinKm)
		argIndex += 3
	}

	// Add user count filters
	if filter.MinUserCount > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" AND user_count >= $%d", argIndex))
		args = append(args, filter.MinUserCount)
		argIndex++
	}

	if filter.MaxUserCount > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" AND user_count <= $%d", argIndex))
		args = append(args, filter.MaxUserCount)
		argIndex++
	}

	// Add time filters
	if !filter.CreatedAfter.IsZero() {
		queryBuilder.WriteString(fmt.Sprintf(" AND created_at >= $%d", argIndex))
		args = append(args, filter.CreatedAfter)
		argIndex++
	}

	if !filter.CreatedBefore.IsZero() {
		queryBuilder.WriteString(fmt.Sprintf(" AND created_at <= $%d", argIndex))
		args = append(args, filter.CreatedBefore)
		argIndex++
	}

	// Add text search
	if filter.SearchTerms != "" {
		queryBuilder.WriteString(fmt.Sprintf(
			" AND (title ILIKE $%d OR description ILIKE $%d)",
			argIndex, argIndex,
		))
		args = append(args, "%"+filter.SearchTerms+"%")
		argIndex++
	}

	// Add topic tags filter
	if len(filter.TopicTags) > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" AND topic_tags && $%d", argIndex))
		args = append(args, filter.TopicTags)
		argIndex++
	}

	// Add ordering, limit and offset
	queryBuilder.WriteString(" ORDER BY created_at DESC")

	if filter.Limit > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d", argIndex))
		args = append(args, filter.Limit)
		argIndex++
	} else {
		queryBuilder.WriteString(" LIMIT 20") // Default limit
	}

	if filter.Offset > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" OFFSET $%d", argIndex))
		args = append(args, filter.Offset)
	}

	// Execute query
	rows, err := s.db.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	// Parse results
	var spaces []space.Space
	for rows.Next() {
		var sp space.Space
		var templateType, lifecycleStage string
		var lng, lat *float64

		err := rows.Scan(
			&sp.ID,
			&sp.Title,
			&sp.Description,
			&sp.TrendID,
			&templateType,
			&lifecycleStage,
			&sp.CreatedAt,
			&sp.LastActive,
			&sp.UserCount,
			&sp.MessageCount,
			&lng,
			&lat,
			&sp.LocationRadius,
			&sp.IsGeoLocal,
			&sp.TopicTags,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning space: %w", err)
		}

		// Set location if coordinates are present
		if lng != nil && lat != nil {
			sp.Location = &trend.Location{
				Longitude: *lng,
				Latitude:  *lat,
				Timestamp: sp.CreatedAt,
			}
		}

		// Set enum types
		sp.TemplateType = space.TemplateType(templateType)
		sp.LifecycleStage = space.LifecycleStage(lifecycleStage)

		spaces = append(spaces, sp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating spaces: %w", err)
	}

	return spaces, nil
}

// FindNearbySpaces finds spaces near a location
func (s *SpaceStore) FindNearbySpaces(
	ctx context.Context,
	location trend.Location,
	radiusKm float64,
) ([]space.Space, error) {
	query := `
		SELECT
			id, title, description, trend_id, template_type::text, lifecycle_stage::text,
			created_at, last_active, user_count, message_count,
			ST_X(location::geometry) as lng, ST_Y(location::geometry) as lat,
			location_radius, is_geo_local,
			topic_tags,
			ST_Distance(geography(location), geography(ST_MakePoint($1, $2))) / 1000 as distance
		FROM spaces
		WHERE location IS NOT NULL
		AND ST_DWithin(geography(location), geography(ST_MakePoint($1, $2)), $3 * 1000)
		AND lifecycle_stage IN ('growing', 'peak', 'waning')
		ORDER BY 
			CASE 
				WHEN is_geo_local THEN 0 
				ELSE 1 
			END, -- Prioritize geo-local spaces
			user_count DESC,
			distance ASC
		LIMIT 20
	`

	// Execute query
	rows, err := s.db.Query(ctx, query, location.Longitude, location.Latitude, radiusKm)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	// Parse results
	var spaces []space.Space
	for rows.Next() {
		var sp space.Space
		var templateType, lifecycleStage string
		var lng, lat *float64
		var distance float64

		err := rows.Scan(
			&sp.ID,
			&sp.Title,
			&sp.Description,
			&sp.TrendID,
			&templateType,
			&lifecycleStage,
			&sp.CreatedAt,
			&sp.LastActive,
			&sp.UserCount,
			&sp.MessageCount,
			&lng,
			&lat,
			&sp.LocationRadius,
			&sp.IsGeoLocal,
			&sp.TopicTags,
			&distance,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning space: %w", err)
		}

		// Set location if coordinates are present
		if lng != nil && lat != nil {
			sp.Location = &trend.Location{
				Longitude: *lng,
				Latitude:  *lat,
				Timestamp: sp.CreatedAt,
			}
		}

		// Set enum types
		sp.TemplateType = space.TemplateType(templateType)
		sp.LifecycleStage = space.LifecycleStage(lifecycleStage)

		spaces = append(spaces, sp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating spaces: %w", err)
	}

	return spaces, nil
}
