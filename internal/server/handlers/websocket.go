// internal/server/handlers/websocket.go

package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	conn              *websocket.Conn
	send              chan []byte
	spaceID           string
	userID            string
	natsConn          *nats.Conn
	subIDs            []string // Subscription IDs
	natsSubscriptions []*nats.Subscription
}

// WebSocketConfig contains configuration for WebSocket connections
type WebSocketConfig struct {
	// Time allowed to write a message to the peer
	WriteWait time.Duration

	// Time allowed to read the next pong message from the peer
	PongWait time.Duration

	// Send pings to peer with this period
	PingPeriod time.Duration

	// Maximum message size allowed from peer
	MaxMessageSize int64
}

// DefaultWebSocketConfig returns the default WebSocket configuration
func DefaultWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		WriteWait:      10 * time.Second,
		PongWait:       60 * time.Second,
		PingPeriod:     (60 * time.Second * 9) / 10,
		MaxMessageSize: 1024 * 1024, // 1MB
	}
}

// WebSocketUpgrader is used to upgrade HTTP connections to WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, this should be more restrictive
		return true
	},
}

// SpaceWebSocketHandler handles WebSocket connections for real-time space interaction
func SpaceWebSocketHandler(natsConn *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get space ID from URL
		spaceID := chi.URLParam(r, "id")
		if spaceID == "" {
			http.Error(w, "Missing space ID", http.StatusBadRequest)
			return
		}

		// Get user ID from query parameters or authentication
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			// In a real implementation, we would get this from the authentication token
			http.Error(w, "Missing user ID", http.StatusBadRequest)
			return
		}

		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade to WebSocket: %v", err)
			return
		}

		// Create new client
		client := &WebSocketClient{
			conn:     conn,
			send:     make(chan []byte, 256),
			spaceID:  spaceID,
			userID:   userID,
			natsConn: natsConn,
			subIDs:   []string{},
		}

		// Start client
		go client.writePump()
		go client.readPump()

		// Subscribe to space-related topics
		if err := client.subscribeToSpace(); err != nil {
			log.Printf("Failed to subscribe to space topics: %v", err)
			client.closeConnection()
			return
		}

		// Send welcome message
		welcomeMsg := map[string]interface{}{
			"type":     "welcome",
			"space_id": spaceID,
			"time":     time.Now(),
		}

		welcomeJSON, _ := json.Marshal(welcomeMsg)
		client.send <- welcomeJSON

		// Log connection
		log.Printf("New WebSocket connection for space %s from user %s", spaceID, userID)

		// Send recent messages
		client.sendRecentMessages()
	}
}

// readPump pumps messages from the WebSocket connection to NATS
func (c *WebSocketClient) readPump() {
	config := DefaultWebSocketConfig()

	defer func() {
		c.closeConnection()
	}()

	c.conn.SetReadLimit(config.MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(config.PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(config.PongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Process incoming message
		c.processIncomingMessage(message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *WebSocketClient) writePump() {
	config := DefaultWebSocketConfig()
	ticker := time.NewTicker(config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.closeConnection()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// processIncomingMessage processes an incoming WebSocket message
func (c *WebSocketClient) processIncomingMessage(message []byte) {
	// Parse message
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse WebSocket message: %v", err)
		return
	}

	// Get message type
	msgType, ok := msg["type"].(string)
	if !ok {
		log.Printf("Missing message type")
		return
	}

	// Add sender information
	msg["user_id"] = c.userID
	msg["time"] = time.Now()

	// Process based on message type
	switch msgType {
	case "message":
		// In a real implementation, we would validate and process the message
		c.handleChatMessage(msg)

	case "typing":
		// Handle typing indicator
		c.handleTypingIndicator(msg)

	case "reaction":
		// Handle reaction
		c.handleReaction(msg)

	default:
		log.Printf("Unknown message type: %s", msgType)
	}
}

// handleChatMessage handles a chat message
func (c *WebSocketClient) handleChatMessage(msg map[string]interface{}) {
	// Add message ID
	msg["id"] = fmt.Sprintf("msg_%d", time.Now().UnixNano())

	// In a real implementation, we would store the message in the database

	// Publish to NATS for all clients in this space
	msgJSON, _ := json.Marshal(msg)
	topic := fmt.Sprintf("space.%s.messages", c.spaceID)

	if err := c.natsConn.Publish(topic, msgJSON); err != nil {
		log.Printf("Failed to publish message to NATS: %v", err)
	}
}

// handleTypingIndicator handles a typing indicator
func (c *WebSocketClient) handleTypingIndicator(msg map[string]interface{}) {
	// Publish typing indicator to NATS
	msgJSON, _ := json.Marshal(msg)
	topic := fmt.Sprintf("space.%s.typing", c.spaceID)

	if err := c.natsConn.Publish(topic, msgJSON); err != nil {
		log.Printf("Failed to publish typing indicator to NATS: %v", err)
	}
}

// handleReaction handles a message reaction
func (c *WebSocketClient) handleReaction(msg map[string]interface{}) {
	// Get message ID
	messageID, ok := msg["message_id"].(string)
	if !ok {
		log.Printf("Missing message ID in reaction")
		return
	}

	// Get reaction
	reaction, ok := msg["reaction"].(string)
	if !ok {
		log.Printf("Missing reaction")
		return
	}

	// In a real implementation, we would update the reaction in the database

	// Publish reaction to NATS
	reactionMsg := map[string]interface{}{
		"type":       "reaction",
		"user_id":    c.userID,
		"message_id": messageID,
		"reaction":   reaction,
		"time":       time.Now(),
	}

	msgJSON, _ := json.Marshal(reactionMsg)
	topic := fmt.Sprintf("space.%s.reactions", c.spaceID)

	if err := c.natsConn.Publish(topic, msgJSON); err != nil {
		log.Printf("Failed to publish reaction to NATS: %v", err)
	}
}

// subscribeToSpace subscribes to space-related NATS topics
func (c *WebSocketClient) subscribeToSpace() error {
	// Subscribe to messages
	msgSub, err := c.natsConn.Subscribe(fmt.Sprintf("space.%s.messages", c.spaceID), func(msg *nats.Msg) {
		c.send <- msg.Data
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to messages: %w", err)
	}
	c.natsSubscriptions = append(c.natsSubscriptions, msgSub)

	// Subscribe to typing indicators
	typingSub, err := c.natsConn.Subscribe(fmt.Sprintf("space.%s.typing", c.spaceID), func(msg *nats.Msg) {
		c.send <- msg.Data
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to typing indicators: %w", err)
	}
	c.natsSubscriptions = append(c.natsSubscriptions, typingSub)

	// Subscribe to reactions
	reactionSub, err := c.natsConn.Subscribe(fmt.Sprintf("space.%s.reactions", c.spaceID), func(msg *nats.Msg) {
		c.send <- msg.Data
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to reactions: %w", err)
	}
	c.natsSubscriptions = append(c.natsSubscriptions, reactionSub)

	// Subscribe to lifecycle events
	lifecycleSub, err := c.natsConn.Subscribe(fmt.Sprintf("space.%s.lifecycle", c.spaceID), func(msg *nats.Msg) {
		c.send <- msg.Data
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to lifecycle events: %w", err)
	}
	c.natsSubscriptions = append(c.natsSubscriptions, lifecycleSub)

	return nil
}

// sendRecentMessages sends recent messages to the client
func (c *WebSocketClient) sendRecentMessages() {
	// In a real implementation, we would fetch recent messages from the database
	// This is a mock implementation

	// Create sample messages
	messages := []map[string]interface{}{
		{
			"type":      "message",
			"id":        "msg_1",
			"user_id":   "user_1",
			"content":   "Welcome to the space!",
			"time":      time.Now().Add(-5 * time.Minute),
			"reactions": map[string]int{"👍": 2},
		},
		{
			"type":    "message",
			"id":      "msg_2",
			"user_id": "user_2",
			"content": "Hello everyone!",
			"time":    time.Now().Add(-3 * time.Minute),
		},
	}

	// Send messages
	historyMsg := map[string]interface{}{
		"type":     "history",
		"messages": messages,
	}

	historyJSON, _ := json.Marshal(historyMsg)
	c.send <- historyJSON
}

// closeConnection closes the WebSocket connection and cleans up resources
func (c *WebSocketClient) closeConnection() {
	// Unsubscribe from all NATS topics
	for _, sub := range c.natsSubscriptions {
		sub.Unsubscribe()
	}

	// Close connection
	c.conn.Close()

	// Close send channel
	close(c.send)

	// Log disconnection
	log.Printf("WebSocket connection closed for space %s, user %s", c.spaceID, c.userID)
}
