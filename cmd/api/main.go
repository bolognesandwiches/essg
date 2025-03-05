// cmd/api/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nats-io/nats.go"

	"essg/internal/adapter/storage"
	"essg/internal/config"
	"essg/internal/domain/trend"
	"essg/internal/server"
	geoService "essg/internal/service/geo"
	"essg/internal/service/listening"
	spaceService "essg/internal/service/space"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Initialize dependencies
	db, err := initDatabase(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	natsConn, err := initNATS(cfg.NATS)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsConn.Close()

	// Initialize storage adapters
	trendStore := storage.NewTrendStore(db)
	spaceStore := storage.NewSpaceStore(db)

	// Initialize services
	trendAnalyzer := listening.NewAnalyzer()
	geoTagger := listening.NewGeoTagger()

	// Create local source registry
	localSourceRegistry := geoService.NewLocalSourceRegistry()

	// Create geocoder service
	geocoder := geoService.NewGeocoderService()

	// Initialize geospatial service
	geoSpatialService := geoService.NewGeoSpatialService(
		db,
		localSourceRegistry,
		geocoder,
		geoService.GeoSpatialConfig{
			DefaultRadius:               cfg.Geo.DefaultRadius,
			MinRadius:                   cfg.Geo.MinRadius,
			MaxRadius:                   cfg.Geo.MaxRadius,
			ClusterThreshold:            cfg.Geo.ClusterThreshold,
			PopulationDensityThresholds: cfg.Geo.PopulationDensityThresholds,
		},
	)

	// Initialize trend detector
	trendDetector := listening.NewTrendDetector(
		trendAnalyzer,
		geoTagger,
		trendStore,
		natsConn,
		listening.TrendDetectorConfig{
			TrendThreshold:         cfg.Trend.TrendThreshold,
			ScanInterval:           cfg.Trend.ScanInterval,
			GeoScanInterval:        cfg.Trend.GeoScanInterval,
			CorrelationThreshold:   cfg.Trend.CorrelationThreshold,
			MaxConcurrentPlatforms: cfg.Trend.MaxConcurrentPlatforms,
			EventsTopic:            cfg.Trend.EventsTopic,
		},
	)

	// Initialize engagement analyzer
	engagementAnalyzer := spaceService.NewEngagementAnalyzer(
		db,
		natsConn,
		geoSpatialService,
		spaceService.EngagementAnalyzerConfig{
			MonitoringInterval: cfg.Space.MonitoringInterval,
		},
	)

	// Initialize space manager
	spaceManager := spaceService.NewSpaceManager(
		spaceStore,
		engagementAnalyzer,
		natsConn,
		spaceService.SpaceManagerConfig{
			EventsTopic:         cfg.Space.EventsTopic,
			DefaultGracePeriod:  cfg.Space.DefaultGracePeriod,
			MonitoringInterval:  cfg.Space.MonitoringInterval,
			MaxConcurrentSpaces: cfg.Space.MaxConcurrentSpaces,
		},
	)

	// Register space templates
	registerSpaceTemplates(spaceManager)

	// Register trend handler to create spaces automatically
	trendDetector.RegisterTrendHandler(func(t trend.Trend) error {
		if t.Score >= cfg.Trend.TrendThreshold {
			_, err := spaceManager.CreateSpace(context.Background(), t)
			return err
		}
		return nil
	})

	// Start the trend detector
	if err := trendDetector.Start(ctx); err != nil {
		log.Fatalf("Failed to start trend detector: %v", err)
	}

	// Initialize HTTP server
	httpServer := server.NewServer(
		cfg.Server,
		db,
		natsConn,
		trendDetector,
		spaceManager,
		geoSpatialService,
	)

	// Start HTTP server
	go func() {
		log.Printf("Starting HTTP server on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	log.Println("Shutdown signal received")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	// Graceful shutdown
	log.Println("Shutting down services...")

	// Shutdown HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Stop trend detector
	if err := trendDetector.Stop(shutdownCtx); err != nil {
		log.Printf("Trend detector shutdown error: %v", err)
	}

	// Stop space manager
	if err := spaceManager.Stop(shutdownCtx); err != nil {
		log.Printf("Space manager shutdown error: %v", err)
	}

	log.Println("Shutdown complete")
}

// Initialize database connection
func initDatabase(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.MaxLifetime

	db, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Test connection
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return db, nil
}

// Initialize NATS connection
func initNATS(cfg config.NATSConfig) (*nats.Conn, error) {
	options := []nats.Option{
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
		nats.Timeout(cfg.ConnectTimeout),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %s", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Printf("NATS connection closed")
		}),
	}

	nc, err := nats.Connect(cfg.URL, options...)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to NATS: %w", err)
	}

	return nc, nil
}

// Register space templates
func registerSpaceTemplates(manager *spaceService.SpaceManager) {
	// General template
	manager.RegisterTemplate(spaceService.NewGeneralTemplate())

	// Breaking news template
	manager.RegisterTemplate(spaceService.NewBreakingNewsTemplate())

	// Event template
	manager.RegisterTemplate(spaceService.NewEventTemplate())

	// Discussion template
	manager.RegisterTemplate(spaceService.NewDiscussionTemplate())

	// Local template
	manager.RegisterTemplate(spaceService.NewLocalTemplate())
}
