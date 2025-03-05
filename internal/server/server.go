// internal/server/server.go

package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nats-io/nats.go"

	"essg/internal/config"
	"essg/internal/domain/geo"
	"essg/internal/domain/space"
	"essg/internal/domain/trend"
	"essg/internal/server/handlers"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	router *chi.Mux
}

// NewServer creates a new HTTP server
func NewServer(
	cfg config.ServerConfig,
	db *pgxpool.Pool,
	natsConn *nats.Conn,
	trendDetector trend.Detector,
	spaceManager space.Manager,
	geoService geo.Service,
) *Server {
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CorsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Create handler dependencies
	trendHandler := handlers.NewTrendHandler(trendDetector)
	spaceHandler := handlers.NewSpaceHandler(spaceManager)
	geoHandler := handlers.NewGeoHandler(geoService)

	// Routes
	router.Route("/api", func(r chi.Router) {
		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		// API version
		r.Route("/v1", func(r chi.Router) {
			// Trends API
			r.Route("/trends", func(r chi.Router) {
				r.Get("/", trendHandler.GetTrends)
				r.Get("/{id}", trendHandler.GetTrend)
				r.Get("/geo", trendHandler.GetGeoTrends)
			})

			// Spaces API
			r.Route("/spaces", func(r chi.Router) {
				r.Get("/", spaceHandler.ListSpaces)
				r.Post("/", spaceHandler.CreateSpace)
				r.Get("/{id}", spaceHandler.GetSpace)
				r.Get("/nearby", spaceHandler.GetNearbySpaces)

				// Space messages
				r.Route("/{id}/messages", func(r chi.Router) {
					r.Get("/", spaceHandler.GetMessages)
					r.Post("/", spaceHandler.SendMessage)
				})
			})

			// Geo API
			r.Route("/geo", func(r chi.Router) {
				r.Get("/context", geoHandler.GetLocationContext)
				r.Get("/trends", geoHandler.GetLocalTrends)
			})
		})
	})

	// WebSocket endpoint for real-time communications
	router.Get("/ws/spaces/{id}", handlers.SpaceWebSocketHandler(natsConn))

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &Server{
		server: httpServer,
		router: router,
	}
}

// ListenAndServe starts the HTTP server
func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
