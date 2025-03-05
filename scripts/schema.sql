-- schema.sql

-- Enable PostGIS extension for geospatial features
CREATE EXTENSION IF NOT EXISTS postgis;

-- Enable TimescaleDB for time-series data
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Create enum types
CREATE TYPE lifecycle_stage AS ENUM (
    'creating',
    'growing',
    'peak',
    'waning',
    'dissolving',
    'dissolved'
);

CREATE TYPE template_type AS ENUM (
    'general',
    'breaking_news',
    'event',
    'discussion',
    'local'
);

CREATE TYPE message_type AS ENUM (
    'text',
    'media',
    'system',
    'event',
    'location'
);

CREATE TYPE message_status AS ENUM (
    'sending',
    'delivered',
    'read',
    'failed',
    'removed'
);

CREATE TYPE location_sharing_level AS ENUM (
    'disabled',
    'approximate',
    'neighborhood',
    'precise'
);

-- Trends table
CREATE TABLE trends (
    id TEXT PRIMARY KEY,
    topic TEXT NOT NULL,
    description TEXT,
    keywords TEXT[],
    score FLOAT NOT NULL,
    velocity FLOAT,
    location GEOGRAPHY(POINT),
    location_radius FLOAT,
    is_geo_local BOOLEAN NOT NULL DEFAULT FALSE,
    first_detected TIMESTAMPTZ NOT NULL,
    last_updated TIMESTAMPTZ NOT NULL,
    related_trends TEXT[],
    entity_types JSONB,
    sources JSONB,
    raw_data JSONB
);

-- Create spatial index on trends location
CREATE INDEX trends_location_idx ON trends USING GIST (location);
CREATE INDEX trends_score_idx ON trends (score);
CREATE INDEX trends_first_detected_idx ON trends (first_detected);

-- Spaces table
CREATE TABLE spaces (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    trend_id TEXT REFERENCES trends(id),
    template_type template_type NOT NULL,
    lifecycle_stage lifecycle_stage NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    last_active TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    user_count INT NOT NULL DEFAULT 0,
    message_count INT NOT NULL DEFAULT 0,
    location GEOGRAPHY(POINT),
    location_radius FLOAT,
    is_geo_local BOOLEAN NOT NULL DEFAULT FALSE,
    topic_tags TEXT[],
    related_spaces TEXT[],
    engagement_metrics JSONB,
    features JSONB
);

-- Create spatial index on spaces location
CREATE INDEX spaces_location_idx ON spaces USING GIST (location);
CREATE INDEX spaces_lifecycle_idx ON spaces (lifecycle_stage);
CREATE INDEX spaces_created_at_idx ON spaces (created_at);
CREATE INDEX spaces_template_idx ON spaces (template_type);

-- Users table
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    external_ids JSONB,
    created_at TIMESTAMPTZ NOT NULL,
    last_seen TIMESTAMPTZ NOT NULL,
    location GEOGRAPHY(POINT),
    location_sharing_preference location_sharing_level NOT NULL DEFAULT 'neighborhood',
    default_anonymity BOOLEAN NOT NULL DEFAULT TRUE,
    notification_preferences JSONB
);

-- Create spatial index on users location
CREATE INDEX users_location_idx ON users USING GIST (location);

-- Ephemeral identities table
CREATE TABLE ephemeral_identities (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    space_id TEXT NOT NULL REFERENCES spaces(id),
    nickname TEXT NOT NULL,
    avatar TEXT,
    is_anonymous BOOLEAN NOT NULL DEFAULT TRUE,
    location GEOGRAPHY(POINT),
    location_share_level location_sharing_level NOT NULL DEFAULT 'neighborhood',
    created_at TIMESTAMPTZ NOT NULL,
    last_active TIMESTAMPTZ NOT NULL,
    reputation JSONB,
    UNIQUE (user_id, space_id)
);

-- Create index on ephemeral_identities for user and space lookups
CREATE INDEX ephemeral_identities_user_idx ON ephemeral_identities (user_id);
CREATE INDEX ephemeral_identities_space_idx ON ephemeral_identities (space_id);

-- Messages table
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    space_id TEXT NOT NULL REFERENCES spaces(id),
    user_id TEXT NOT NULL REFERENCES users(id),
    ephemeral_identity_id TEXT REFERENCES ephemeral_identities(id),
    type message_type NOT NULL,
    content TEXT,
    media_urls TEXT[],
    metadata JSONB,
    reply_to_id TEXT REFERENCES messages(id),
    status message_status NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    location GEOGRAPHY(POINT),
    distance_from_center FLOAT,
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    visible_to_roles TEXT[]
);

-- Make messages a hypertable for time-series optimization
SELECT create_hypertable('messages', 'created_at');

-- Create indexes on messages for common queries
CREATE INDEX messages_space_idx ON messages (space_id, created_at DESC);
CREATE INDEX messages_user_idx ON messages (user_id, created_at DESC);
CREATE INDEX messages_reply_idx ON messages (reply_to_id);

-- Reactions table
CREATE TABLE reactions (
    id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL REFERENCES messages(id),
    user_id TEXT NOT NULL REFERENCES users(id),
    reaction TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (message_id, user_id, reaction)
);

-- Create index on reactions for message lookup
CREATE INDEX reactions_message_idx ON reactions (message_id);

-- Tokens table for authentication
CREATE TABLE tokens (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

-- Create index on tokens for user lookup and expiration
CREATE INDEX tokens_user_idx ON tokens (user_id);
CREATE INDEX tokens_expiry_idx ON tokens (expires_at);

-- Rate limiting table
CREATE TABLE rate_limits (
    id SERIAL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    action_type TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    count INT NOT NULL DEFAULT 1,
    window_start TIMESTAMPTZ NOT NULL,
    UNIQUE (user_id, action_type, resource_id, window_start)
);

-- Create index on rate_limits for lookups
CREATE INDEX rate_limits_lookup_idx ON rate_limits (user_id, action_type, resource_id, window_start);

-- Create a hypertable for rate_limits for efficient time-series data
SELECT create_hypertable('rate_limits', 'window_start');

-- Local trends table for geo-specific trends
CREATE TABLE local_trends (
    id TEXT PRIMARY KEY,
    trend_id TEXT NOT NULL REFERENCES trends(id),
    location GEOGRAPHY(POINT) NOT NULL,
    location_context JSONB,
    score FLOAT NOT NULL,
    location_score FLOAT NOT NULL,
    local_relevance JSONB,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Create spatial index on local_trends location
CREATE INDEX local_trends_location_idx ON local_trends USING GIST (location);
CREATE INDEX local_trends_score_idx ON local_trends (score);

-- Space analytics table
CREATE TABLE space_analytics (
    id SERIAL PRIMARY KEY,
    space_id TEXT NOT NULL REFERENCES spaces(id),
    timestamp TIMESTAMPTZ NOT NULL,
    user_count INT NOT NULL,
    message_count INT NOT NULL,
    active_users INT NOT NULL,
    engagement_score FLOAT NOT NULL,
    lifecycle_stage lifecycle_stage NOT NULL,
    metrics JSONB
);

-- Create a hypertable for space_analytics
SELECT create_hypertable('space_analytics', 'timestamp');

-- Create index on space_analytics for space lookup
CREATE INDEX space_analytics_space_idx ON space_analytics (space_id, timestamp DESC);

-- Functions for space lifecycle management

-- Update space last_active timestamp
CREATE OR REPLACE FUNCTION update_space_last_active()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE spaces
    SET last_active = NOW()
    WHERE id = NEW.space_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update space last_active on message creation
CREATE TRIGGER update_space_last_active_trigger
AFTER INSERT ON messages
FOR EACH ROW
EXECUTE FUNCTION update_space_last_active();

-- Update space message_count
CREATE OR REPLACE FUNCTION update_space_message_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE spaces
    SET message_count = message_count + 1
    WHERE id = NEW.space_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update space message_count on message creation
CREATE TRIGGER update_space_message_count_trigger
AFTER INSERT ON messages
FOR EACH ROW
EXECUTE FUNCTION update_space_message_count();

-- Calculate distance from space center for geo messages
CREATE OR REPLACE FUNCTION calculate_distance_from_center()
RETURNS TRIGGER AS $
BEGIN
    IF NEW.location IS NOT NULL THEN
        WITH space_location AS (
            SELECT location FROM spaces WHERE id = NEW.space_id
        )
        UPDATE messages
        SET distance_from_center = ST_Distance(
            location::geography,
            (SELECT location FROM space_location)
        ) / 1000 -- Convert to kilometers
        WHERE id = NEW.id;
    END IF;
    RETURN NEW;
END;
$ LANGUAGE plpgsql;

-- Trigger to calculate distance from center on message creation
CREATE TRIGGER calculate_distance_from_center_trigger
AFTER INSERT ON messages
FOR EACH ROW
WHEN (NEW.location IS NOT NULL)
EXECUTE FUNCTION calculate_distance_from_center();

-- Clean up expired tokens
CREATE OR REPLACE FUNCTION cleanup_expired_tokens()
RETURNS void AS $
BEGIN
    DELETE FROM tokens WHERE expires_at < NOW();
END;
$ LANGUAGE plpgsql;

-- Automated cleanup of expired data

-- Function to clean up dissolved spaces and related data
CREATE OR REPLACE FUNCTION cleanup_dissolved_spaces()
RETURNS void AS $
DECLARE
    retention_period INTERVAL := '30 days'; -- Adjust based on requirements
BEGIN
    -- Find spaces to clean up
    WITH spaces_to_clean AS (
        SELECT id FROM spaces
        WHERE lifecycle_stage = 'dissolved'
        AND last_active < NOW() - retention_period
    )
    -- Delete related data
    DELETE FROM messages
    WHERE space_id IN (SELECT id FROM spaces_to_clean);
    
    DELETE FROM ephemeral_identities
    WHERE space_id IN (SELECT id FROM spaces_to_clean);
    
    DELETE FROM space_analytics
    WHERE space_id IN (SELECT id FROM spaces_to_clean);
    
    -- Finally, delete the spaces
    DELETE FROM spaces
    WHERE id IN (SELECT id FROM spaces_to_clean);
END;
$ LANGUAGE plpgsql;