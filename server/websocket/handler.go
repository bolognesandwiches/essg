package websocket

import (
	"log"
	"net/http"
	"sync"

	"essg/server/services"

	"github.com/gorilla/websocket"
)

// Handler manages WebSocket connections
type Handler struct {
	spaceService   *services.SpaceService
	messageService *services.MessageService
	clients        map[*Client]bool
	register       chan *Client
	unregister     chan *Client
	broadcast      chan Message
	mutex          sync.RWMutex
}

// Client represents a WebSocket client
type Client struct {
	conn      *websocket.Conn
	handler   *Handler
	send      chan Message
	userID    string
	userName  string
	userColor string
}

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Upgrader configures the WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// NewHandler creates a new WebSocket handler
func NewHandler(spaceService *services.SpaceService, messageService *services.MessageService) *Handler {
	return &Handler{
		spaceService:   spaceService,
		messageService: messageService,
		clients:        make(map[*Client]bool),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		broadcast:      make(chan Message),
	}
}

// HandleWebSocket handles WebSocket connections
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}

	// Get anonymous user info from headers
	userID := r.Header.Get("x-anonymous-user-id")
	userName := r.Header.Get("x-anonymous-user-name")
	userColor := r.Header.Get("x-anonymous-user-color")

	if userID == "" || userName == "" {
		log.Println("Anonymous user ID and name required")
		conn.Close()
		return
	}

	// Create a new client
	client := &Client{
		conn:      conn,
		handler:   h,
		send:      make(chan Message, 256),
		userID:    userID,
		userName:  userName,
		userColor: userColor,
	}

	// Register the client
	h.register <- client

	// Start goroutines for reading and writing
	go client.readPump()
	go client.writePump()
}

// Run starts the WebSocket handler
func (h *Handler) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.handler.unregister <- c
		c.conn.Close()
	}()

	for {
		var message Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.handler.broadcast <- message
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		message, ok := <-c.send
		if !ok {
			// The hub closed the channel
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.conn.WriteJSON(message)
		if err != nil {
			return
		}
	}
}
