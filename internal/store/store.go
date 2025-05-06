package store

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Value struct {
	Value     string    `json:"value"`
	Version   int64     `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

type Object struct {
	Values []Value `json:"values"`
}

// the object is dirty if it has more than one value
func (o *Object) IsDirty() bool {
	return len(o.Values) > 1
}

func (o *Object) String() string {
	bytes, _ := json.Marshal(o)
	return string(bytes)
}

// Store represents our in-memory key-value store
type Store struct {
	data map[string]*Object
	mu   sync.RWMutex
}

// New creates a new Store instance
func New() *Store {
	return &Store{
		data: make(map[string]*Object),
	}
}

// Set stores a value for a given key
func (s *Store) Set(key, value string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; !exists {
		s.data[key] = &Object{
			Values: []Value{},
		}
	}

	object := s.data[key]
	if len(object.Values) == 0 {
		object.Values = append(object.Values, Value{
			Value:     value,
			Version:   0, // the version is zero indexed
			Timestamp: time.Now(),
		})
	} else {
		currVersion := object.Values[len(object.Values)-1].Version
		object.Values = append(object.Values, Value{
			Value:     value,
			Version:   currVersion + 1,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// Get retrieves a value for a given key
func (s *Store) Get(key string) (*Object, error) {
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	object, exists := s.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return object, nil
}

func (s *Store) GetByVersion(key string, version int64) (*Value, error) {
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	object, exists := s.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	if version < 0 || version >= int64(len(object.Values)) {
		return nil, fmt.Errorf("version out of range: %d", version)
	}

	return &object.Values[version], nil
}
