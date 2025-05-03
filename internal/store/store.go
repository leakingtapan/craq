package store

import (
	"fmt"
	"sync"
)

// Store represents our in-memory key-value store
type Store struct {
	data map[string]string
	mu   sync.RWMutex
}

// New creates a new Store instance
func New() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

// Set stores a value for a given key
func (s *Store) Set(key, value string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
	return nil
}

// Get retrieves a value for a given key
func (s *Store) Get(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}

	s.mu.RLock()
	value, exists := s.data[key]
	s.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}

	return value, nil
}
