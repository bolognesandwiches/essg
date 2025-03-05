// internal/config/config.go

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	NATS        NATSConfig
	Trend       TrendConfig
	Space       SpaceConfig
	Geo         GeoConfig
	Identity    IdentityConfig
	Messaging   MessagingConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	CorsOrigins     []string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
	SSLMode      string
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL            string
	MaxReconnects  int
	ReconnectWait  time.Duration
	ConnectTimeout time.Duration
}

// TrendConfig holds trend detection configuration
type TrendConfig struct {
	TrendThreshold         float64
	ScanInterval           time.Duration
	GeoScanInterval        time.Duration
	CorrelationThreshold   float64
	MaxConcurrentPlatforms int
	EventsTopic            string
}

// SpaceConfig holds space management configuration
type SpaceConfig struct {
	EventsTopic         string
	DefaultGracePeriod  time.Duration
	MonitoringInterval  time.Duration
	MaxConcurrentSpaces int
}

// GeoConfig holds geospatial service configuration
type GeoConfig struct {
	DefaultRadius               float64
	MinRadius                   float64
	MaxRadius                   float64
	ClusterThreshold            float64
	PopulationDensityThresholds map[string]float64
}

// IdentityConfig holds identity service configuration
type IdentityConfig struct {
	TokenSecret            string
	TokenExpiry            time.Duration
	DefaultLocationSharing string
	DefaultAnonymity       bool
}

// MessagingConfig holds messaging service configuration
type MessagingConfig struct {
	MessageLimit       int
	RateLimitWindow    time.Duration
	MaxMessageLength   int
	MessageRetention   time.Duration
	MonitoringInterval time.Duration
}

// Load loads configuration from environment variables
func Load() (Config, error) {
	config := Config{
		Environment: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
			CorsOrigins:     getEnvAsSlice("SERVER_CORS_ORIGINS", []string{"*"}),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvAsInt("DB_PORT", 5432),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "postgres"),
			Database:     getEnv("DB_NAME", "essg"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			MaxLifetime:  getEnvAsDuration("DB_MAX_LIFETIME", 5*time.Minute),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
		},
		NATS: NATSConfig{
			URL:            getEnv("NATS_URL", "nats://localhost:4222"),
			MaxReconnects:  getEnvAsInt("NATS_MAX_RECONNECTS", 10),
			ReconnectWait:  getEnvAsDuration("NATS_RECONNECT_WAIT", 1*time.Second),
			ConnectTimeout: getEnvAsDuration("NATS_CONNECT_TIMEOUT", 2*time.Second),
		},
		Trend: TrendConfig{
			TrendThreshold:         getEnvAsFloat("TREND_THRESHOLD", 50.0),
			ScanInterval:           getEnvAsDuration("TREND_SCAN_INTERVAL", 2*time.Minute),
			GeoScanInterval:        getEnvAsDuration("TREND_GEO_SCAN_INTERVAL", 5*time.Minute),
			CorrelationThreshold:   getEnvAsFloat("TREND_CORRELATION_THRESHOLD", 0.7),
			MaxConcurrentPlatforms: getEnvAsInt("TREND_MAX_CONCURRENT_PLATFORMS", 10),
			EventsTopic:            getEnv("TREND_EVENTS_TOPIC", "trend"),
		},
		Space: SpaceConfig{
			EventsTopic:         getEnv("SPACE_EVENTS_TOPIC", "space"),
			DefaultGracePeriod:  getEnvAsDuration("SPACE_DEFAULT_GRACE_PERIOD", 24*time.Hour),
			MonitoringInterval:  getEnvAsDuration("SPACE_MONITORING_INTERVAL", 1*time.Minute),
			MaxConcurrentSpaces: getEnvAsInt("SPACE_MAX_CONCURRENT_SPACES", 1000),
		},
		Geo: GeoConfig{
			DefaultRadius:    getEnvAsFloat("GEO_DEFAULT_RADIUS", 5.0),
			MinRadius:        getEnvAsFloat("GEO_MIN_RADIUS", 1.0),
			MaxRadius:        getEnvAsFloat("GEO_MAX_RADIUS", 50.0),
			ClusterThreshold: getEnvAsFloat("GEO_CLUSTER_THRESHOLD", 0.5),
			PopulationDensityThresholds: map[string]float64{
				"urban":    getEnvAsFloat("GEO_DENSITY_URBAN", 5000.0),
				"suburban": getEnvAsFloat("GEO_DENSITY_SUBURBAN", 1000.0),
				"rural":    getEnvAsFloat("GEO_DENSITY_RURAL", 100.0),
			},
		},
		Identity: IdentityConfig{
			TokenSecret:            getEnv("IDENTITY_TOKEN_SECRET", "your-secret-key"),
			TokenExpiry:            getEnvAsDuration("IDENTITY_TOKEN_EXPIRY", 24*time.Hour),
			DefaultLocationSharing: getEnv("IDENTITY_DEFAULT_LOCATION_SHARING", "neighborhood"),
			DefaultAnonymity:       getEnvAsBool("IDENTITY_DEFAULT_ANONYMITY", true),
		},
		Messaging: MessagingConfig{
			MessageLimit:       getEnvAsInt("MESSAGING_MESSAGE_LIMIT", 100),
			RateLimitWindow:    getEnvAsDuration("MESSAGING_RATE_LIMIT_WINDOW", 1*time.Minute),
			MaxMessageLength:   getEnvAsInt("MESSAGING_MAX_MESSAGE_LENGTH", 1000),
			MessageRetention:   getEnvAsDuration("MESSAGING_MESSAGE_RETENTION", 30*24*time.Hour),
			MonitoringInterval: getEnvAsDuration("MESSAGING_MONITORING_INTERVAL", 1*time.Minute),
		},
	}

	return config, validate(config)
}

// validate checks if config is valid
func validate(config Config) error {
	if config.Identity.TokenSecret == "your-secret-key" && config.Environment != "development" {
		return fmt.Errorf("token secret must be set in non-development environments")
	}

	return nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}
