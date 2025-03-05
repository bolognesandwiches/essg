// internal/adapter/storage/trend_store.go

package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"essg/internal/domain/trend"
)

// TrendStore implements storage for trends
type TrendStore struct {
	db *pgxpool.Pool
}

// NewTrendStore creates a new trend store
func NewTrendStore(db *pgxpool.Pool) *TrendStore {
	return &TrendStore{
		db: db,
	}
}

// SaveTrend saves a trend to storage
func (s *TrendStore) SaveTrend(ctx context.Context, t trend.Trend) error {
	query := `
		INSERT INTO trends (
			id, topic, description, keywords, score, velocity,
			location, location_radius, is_geo_local,
			first_detected, last_updated, related_trends,
			entity_types, sources, raw_data
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			ST_MakePoint($7, $8)::geography, $9, $10,
			$11, $12, $13,
			$14, $15, $16
		)
		ON CONFLICT (id) DO UPDATE
		SET
			topic = $2,
			description = $3,
			keywords = $4,
			score = $5,
			velocity = $6,
			location = CASE WHEN $7 IS NOT NULL AND $8 IS NOT NULL THEN ST_MakePoint($7, $8)::geography ELSE trends.location END,
			location_radius = $9,
			is_geo_local = $10,
			last_updated = $12,
			related_trends = $13,
			entity_types = $14,
			sources = $15,
			raw_data = $16
	`

	// Set timestamps if not provided
	if t.FirstDetected.IsZero() {
		t.FirstDetected = time.Now()
	}
	if t.LastUpdated.IsZero() {
		t.LastUpdated = time.Now()
	}

	// Prepare location data
	var lng, lat *float64
	if t.Location != nil {
		lng = &t.Location.Longitude
		lat = &t.Location.Latitude
	}

	// Convert JSON fields
	entityTypesJSON, err := json.Marshal(t.EntityTypes)
	if err != nil {
		return fmt.Errorf("error marshaling entity types: %w", err)
	}

	sourcesJSON, err := json.Marshal(t.Sources)
	if err != nil {
		return fmt.Errorf("error marshaling sources: %w", err)
	}

	rawDataJSON, err := json.Marshal(t.RawData)
	if err != nil {
		return fmt.Errorf("error marshaling raw data: %w", err)
	}

	// Execute query
	_, err = s.db.Exec(
		ctx,
		query,
		t.ID,
		t.Topic,
		t.Description,
		t.Keywords,
		t.Score,
		t.Velocity,
		lng,
		lat,
		t.LocationRadius,
		t.IsGeoLocal,
		t.FirstDetected,
		t.LastUpdated,
		t.RelatedTrends,
		entityTypesJSON,
		sourcesJSON,
		rawDataJSON,
	)

	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil
}

// GetTrend retrieves a trend by ID
func (s *TrendStore) GetTrend(ctx context.Context, id string) (*trend.Trend, error) {
	query := `
		SELECT
			id, topic, description, keywords, score, velocity,
			ST_X(location::geometry) as lng, ST_Y(location::geometry) as lat,
			location_radius, is_geo_local,
			first_detected, last_updated, related_trends,
			entity_types, sources, raw_data
		FROM trends
		WHERE id = $1
	`

	var t trend.Trend
	var lng, lat *float64
	var entityTypesJSON, sourcesJSON, rawDataJSON []byte

	err := s.db.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.Topic,
		&t.Description,
		&t.Keywords,
		&t.Score,
		&t.Velocity,
		&lng,
		&lat,
		&t.LocationRadius,
		&t.IsGeoLocal,
		&t.FirstDetected,
		&t.LastUpdated,
		&t.RelatedTrends,
		&entityTypesJSON,
		&sourcesJSON,
		&rawDataJSON,
	)

	if err != nil {
		return nil, fmt.Errorf("error querying trend: %w", err)
	}

	// Set location if coordinates are present
	if lng != nil && lat != nil {
		t.Location = &trend.Location{
			Longitude: *lng,
			Latitude:  *lat,
			Timestamp: t.LastUpdated,
		}
	}

	// Parse JSON fields
	if err := json.Unmarshal(entityTypesJSON, &t.EntityTypes); err != nil {
		return nil, fmt.Errorf("error unmarshaling entity types: %w", err)
	}

	if err := json.Unmarshal(sourcesJSON, &t.Sources); err != nil {
		return nil, fmt.Errorf("error unmarshaling sources: %w", err)
	}

	if err := json.Unmarshal(rawDataJSON, &t.RawData); err != nil {
		return nil, fmt.Errorf("error unmarshaling raw data: %w", err)
	}

	return &t, nil
}

// FindTrends finds trends matching the filter
func (s *TrendStore) FindTrends(ctx context.Context, filter trend.Filter) ([]trend.Trend, error) {
	// Build query with filters
	query := `
		SELECT
			id, topic, description, keywords, score, velocity,
			ST_X(location::geometry) as lng, ST_Y(location::geometry) as lat,
			location_radius, is_geo_local,
			first_detected, last_updated, related_trends,
			entity_types, sources
		FROM trends
		WHERE score >= $1
	`

	args := []interface{}{filter.MinScore}
	argIndex := 2

	// Add platform filters
	if len(filter.IncludePlatforms) > 0 {
		platformsJSON, err := json.Marshal(filter.IncludePlatforms)
		if err != nil {
			return nil, fmt.Errorf("error marshaling include platforms: %w", err)
		}

		query += fmt.Sprintf(" AND sources @> $%d", argIndex)
		args = append(args, platformsJSON)
		argIndex++
	}

	// Add exclude platform filters
	if len(filter.ExcludePlatforms) > 0 {
		for _, platform := range filter.ExcludePlatforms {
			query += fmt.Sprintf(" AND NOT sources ? $%d", argIndex)
			args = append(args, platform)
			argIndex++
		}
	}

	// Add geo filters
	if filter.GeoOnly {
		query += " AND is_geo_local = true"
	}

	if filter.Location != nil && filter.WithinKm > 0 {
		query += fmt.Sprintf(
			" AND ST_DWithin(geography(location), geography(ST_MakePoint($%d, $%d)), $%d * 1000)",
			argIndex, argIndex+1, argIndex+2,
		)
		args = append(args, filter.Location.Longitude, filter.Location.Latitude, filter.WithinKm)
		argIndex += 3
	}

	// Add entity type filters
	if len(filter.IncludeEntityType) > 0 {
		for _, entityType := range filter.IncludeEntityType {
			query += fmt.Sprintf(" AND entity_types ? $%d", argIndex)
			args = append(args, entityType)
			argIndex++
		}
	}

	// Add ordering and limit
	query += " ORDER BY score DESC LIMIT 100"

	// Execute query
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	// Parse results
	var trends []trend.Trend
	for rows.Next() {
		var t trend.Trend
		var lng, lat *float64
		var entityTypesJSON, sourcesJSON []byte

		err := rows.Scan(
			&t.ID,
			&t.Topic,
			&t.Description,
			&t.Keywords,
			&t.Score,
			&t.Velocity,
			&lng,
			&lat,
			&t.LocationRadius,
			&t.IsGeoLocal,
			&t.FirstDetected,
			&t.LastUpdated,
			&t.RelatedTrends,
			&entityTypesJSON,
			&sourcesJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning trend: %w", err)
		}

		// Set location if coordinates are present
		if lng != nil && lat != nil {
			t.Location = &trend.Location{
				Longitude: *lng,
				Latitude:  *lat,
				Timestamp: t.LastUpdated,
			}
		}

		// Parse JSON fields
		if err := json.Unmarshal(entityTypesJSON, &t.EntityTypes); err != nil {
			return nil, fmt.Errorf("error unmarshaling entity types: %w", err)
		}

		if err := json.Unmarshal(sourcesJSON, &t.Sources); err != nil {
			return nil, fmt.Errorf("error unmarshaling sources: %w", err)
		}

		trends = append(trends, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trends: %w", err)
	}

	return trends, nil
}

// FindTrendsForLocation finds trends near a location
func (s *TrendStore) FindTrendsForLocation(
	ctx context.Context,
	location trend.Location,
	radiusKm float64,
) ([]trend.Trend, error) {
	query := `
		SELECT
			id, topic, description, keywords, score, velocity,
			ST_X(location::geometry) as lng, ST_Y(location::geometry) as lat,
			location_radius, is_geo_local,
			first_detected, last_updated, related_trends,
			entity_types, sources,
			ST_Distance(geography(location), geography(ST_MakePoint($1, $2))) / 1000 as distance
		FROM trends
		WHERE location IS NOT NULL
		AND ST_DWithin(geography(location), geography(ST_MakePoint($1, $2)), $3 * 1000)
		ORDER BY score DESC, distance ASC
		LIMIT 50
	`

	// Execute query
	rows, err := s.db.Query(ctx, query, location.Longitude, location.Latitude, radiusKm)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	// Parse results
	var trends []trend.Trend
	for rows.Next() {
		var t trend.Trend
		var lng, lat *float64
		var distance float64
		var entityTypesJSON, sourcesJSON []byte

		err := rows.Scan(
			&t.ID,
			&t.Topic,
			&t.Description,
			&t.Keywords,
			&t.Score,
			&t.Velocity,
			&lng,
			&lat,
			&t.LocationRadius,
			&t.IsGeoLocal,
			&t.FirstDetected,
			&t.LastUpdated,
			&t.RelatedTrends,
			&entityTypesJSON,
			&sourcesJSON,
			&distance,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning trend: %w", err)
		}

		// Set location if coordinates are present
		if lng != nil && lat != nil {
			t.Location = &trend.Location{
				Longitude: *lng,
				Latitude:  *lat,
				Timestamp: t.LastUpdated,
			}
		}

		// Parse JSON fields
		if err := json.Unmarshal(entityTypesJSON, &t.EntityTypes); err != nil {
			return nil, fmt.Errorf("error unmarshaling entity types: %w", err)
		}

		if err := json.Unmarshal(sourcesJSON, &t.Sources); err != nil {
			return nil, fmt.Errorf("error unmarshaling sources: %w", err)
		}

		trends = append(trends, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trends: %w", err)
	}

	return trends, nil
}
