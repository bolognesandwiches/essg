package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"essg/server/handlers"
	"essg/server/services"
	"essg/server/websocket"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or error loading it. Using environment variables.")
	}

	// Check if Twitter bearer token is set
	if os.Getenv("TWITTER_BEARER_TOKEN") == "" {
		log.Println("Warning: TWITTER_BEARER_TOKEN is not set. Twitter API features will not work.")
	} else {
		log.Println("TWITTER_BEARER_TOKEN is set. Twitter API features should work.")
	}

	// Initialize router
	r := mux.NewRouter()

	// Initialize services
	spaceService := services.NewSpaceService()
	messageService := services.NewMessageService()

	// Initialize handlers
	spaceHandler := handlers.NewSpaceHandler(spaceService)
	messageHandler := handlers.NewMessageHandler(messageService, spaceService)
	socialHandler := handlers.NewSocialHandler()

	// Register routes
	spaceHandler.RegisterRoutes(r)
	messageHandler.RegisterRoutes(r)
	socialHandler.RegisterRoutes(r)

	// Print all registered routes for debugging
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		fmt.Printf("Route: %s Methods: %v\n", path, methods)
		return nil
	})

	// Register WebSocket handler
	wsHandler := websocket.NewHandler(spaceService, messageService)
	r.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// Start WebSocket handler in a goroutine
	go wsHandler.Run()

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Allow requests from the frontend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Use CORS middleware
	handler := c.Handler(r)

	// Configure server
	srv := &http.Server{
		Handler:      handler,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Start server
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(srv.ListenAndServe())
}
