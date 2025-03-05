package services

import (
	"errors"
	"sync"
	"time"

	"essg/server/models"
)

// SpaceService handles business logic related to spaces
type SpaceService struct {
	spaces      map[string]*models.Space
	joinedUsers map[string][]string // map[spaceID][]userID
	mutex       sync.RWMutex
}

// NewSpaceService creates a new space service
func NewSpaceService() *SpaceService {
	return &SpaceService{
		spaces:      make(map[string]*models.Space),
		joinedUsers: make(map[string][]string),
	}
}

// CreateSpace creates a new space
func (s *SpaceService) CreateSpace(space *models.Space) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.spaces[space.ID]; exists {
		return errors.New("space with this ID already exists")
	}

	s.spaces[space.ID] = space
	return nil
}

// GetSpaceByID retrieves a space by its ID
func (s *SpaceService) GetSpaceByID(id string) (*models.Space, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	space, exists := s.spaces[id]
	if !exists {
		return nil, errors.New("space not found")
	}

	return space, nil
}

// GetTrendingSpaces retrieves trending spaces
func (s *SpaceService) GetTrendingSpaces() ([]*models.Space, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var trendingSpaces []*models.Space
	for _, space := range s.spaces {
		// Simple algorithm: spaces with more users and messages are trending
		// In a real implementation, you would use a more sophisticated algorithm
		if space.UserCount > 0 || space.MessageCount > 0 {
			trendingSpaces = append(trendingSpaces, space)
		}
	}

	return trendingSpaces, nil
}

// GetNearbySpaces retrieves spaces near a location
func (s *SpaceService) GetNearbySpaces(lat, lng float64, radius float64) ([]*models.Space, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var nearbySpaces []*models.Space
	for _, space := range s.spaces {
		if space.IsGeoLocal && space.Location != nil {
			// Simple algorithm: spaces within the radius are nearby
			// In a real implementation, you would use a more sophisticated algorithm
			// like Haversine formula to calculate distance
			if isNearby(lat, lng, space.Location.Latitude, space.Location.Longitude, radius) {
				nearbySpaces = append(nearbySpaces, space)
			}
		}
	}

	return nearbySpaces, nil
}

// GetJoinedSpaces retrieves spaces joined by a user
func (s *SpaceService) GetJoinedSpaces(userID string) ([]*models.Space, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var joinedSpaces []*models.Space
	for spaceID, users := range s.joinedUsers {
		for _, id := range users {
			if id == userID {
				if space, exists := s.spaces[spaceID]; exists {
					joinedSpaces = append(joinedSpaces, space)
				}
				break
			}
		}
	}

	return joinedSpaces, nil
}

// JoinSpace adds a user to a space
func (s *SpaceService) JoinSpace(spaceID, userID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	space, exists := s.spaces[spaceID]
	if !exists {
		return errors.New("space not found")
	}

	// Check if user is already joined
	users, exists := s.joinedUsers[spaceID]
	if exists {
		for _, id := range users {
			if id == userID {
				return nil // User already joined
			}
		}
	}

	// Add user to joined users
	s.joinedUsers[spaceID] = append(s.joinedUsers[spaceID], userID)

	// Update space user count
	space.UserCount++
	space.LastActive = time.Now()

	return nil
}

// LeaveSpace removes a user from a space
func (s *SpaceService) LeaveSpace(spaceID, userID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	space, exists := s.spaces[spaceID]
	if !exists {
		return errors.New("space not found")
	}

	// Check if user is joined
	users, exists := s.joinedUsers[spaceID]
	if !exists {
		return errors.New("user not joined to this space")
	}

	// Remove user from joined users
	var newUsers []string
	found := false
	for _, id := range users {
		if id != userID {
			newUsers = append(newUsers, id)
		} else {
			found = true
		}
	}

	if !found {
		return errors.New("user not joined to this space")
	}

	s.joinedUsers[spaceID] = newUsers

	// Update space user count
	space.UserCount--
	space.LastActive = time.Now()

	return nil
}

// isNearby checks if a location is within a radius of another location
// This is a simplified version that doesn't account for the curvature of the Earth
func isNearby(lat1, lng1, lat2, lng2, radius float64) bool {
	// Simple Euclidean distance for demonstration
	// In a real implementation, you would use the Haversine formula
	dx := lat1 - lat2
	dy := lng1 - lng2
	distance := (dx*dx + dy*dy)

	// Convert degrees to approximate kilometers (very rough approximation)
	// 1 degree of latitude is approximately 111 kilometers
	distanceKm := distance * 111.0

	return distanceKm <= radius
}
