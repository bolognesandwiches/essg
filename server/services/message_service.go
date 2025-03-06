package services

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Message represents a message in a space
type Message struct {
	ID              string    `json:"id"`
	SpaceID         string    `json:"spaceId"`
	UserID          string    `json:"userId"`
	UserName        string    `json:"userName"`
	UserColor       string    `json:"userColor,omitempty"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"createdAt"`
	ReplyToID       string    `json:"replyToId,omitempty"`
	ReplyToUserName string    `json:"replyToUserName,omitempty"`
}

// MessageService handles business logic related to messages
type MessageService struct {
	messages map[string][]Message // map[spaceID][]Message
	mutex    sync.RWMutex
}

// NewMessageService creates a new message service
func NewMessageService() *MessageService {
	return &MessageService{
		messages: make(map[string][]Message),
	}
}

// GetMessages retrieves messages for a space
func (s *MessageService) GetMessages(spaceID string) ([]Message, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	messages, exists := s.messages[spaceID]
	if !exists {
		return []Message{}, nil
	}

	return messages, nil
}

// CreateMessage creates a new message in a space
func (s *MessageService) CreateMessage(spaceID, userID, userName, userColor, content string, replyToID, replyToUserName string) (*Message, error) {
	if content == "" {
		return nil, errors.New("message content cannot be empty")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	message := Message{
		ID:              uuid.New().String(),
		SpaceID:         spaceID,
		UserID:          userID,
		UserName:        userName,
		UserColor:       userColor,
		Content:         content,
		CreatedAt:       time.Now(),
		ReplyToID:       replyToID,
		ReplyToUserName: replyToUserName,
	}

	s.messages[spaceID] = append(s.messages[spaceID], message)

	return &message, nil
}
