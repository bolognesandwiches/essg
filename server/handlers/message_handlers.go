package handlers

import (
	"encoding/json"
	"net/http"

	"essg/server/services"

	"github.com/gorilla/mux"
)

// MessageHandler handles requests related to messages
type MessageHandler struct {
	messageService *services.MessageService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService *services.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// RegisterRoutes registers the message API routes
func (h *MessageHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/spaces/{id}/messages", h.GetMessages).Methods("GET")
	r.HandleFunc("/api/spaces/{id}/messages", h.CreateMessage).Methods("POST")
}

// GetMessages handles requests to get messages for a space
func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spaceID := vars["id"]

	messages, err := h.messageService.GetMessages(spaceID)
	if err != nil {
		http.Error(w, "Failed to get messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// CreateMessageRequest represents the request to create a message
type CreateMessageRequest struct {
	Content string `json:"content"`
}

// CreateMessage handles requests to create a new message in a space
func (h *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spaceID := vars["id"]

	// Get anonymous user info from headers
	userID := r.Header.Get("x-anonymous-user-id")
	userName := r.Header.Get("x-anonymous-user-name")
	userColor := r.Header.Get("x-anonymous-user-color")

	if userID == "" || userName == "" {
		http.Error(w, "Anonymous user ID and name required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Content == "" {
		http.Error(w, "Message content is required", http.StatusBadRequest)
		return
	}

	// Create the message
	message, err := h.messageService.CreateMessage(spaceID, userID, userName, userColor, req.Content)
	if err != nil {
		http.Error(w, "Failed to create message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the created message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}
