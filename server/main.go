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
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or error loading it. Using environment variables.")
	}

	// Initialize router
	r := mux.NewRouter()

	// Initialize services
	spaceService := services.NewSpaceService()
	messageService := services.NewMessageService()

	// Initialize handlers
	spaceHandler := handlers.NewSpaceHandler(spaceService)
	messageHandler := handlers.NewMessageHandler(messageService)
	socialHandler := handlers.NewSocialHandler()

	// Register routes
	spaceHandler.RegisterRoutes(r)
	messageHandler.RegisterRoutes(r)
	socialHandler.RegisterRoutes(r)

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

	// Configure server
	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Start server
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(srv.ListenAndServe())
}
