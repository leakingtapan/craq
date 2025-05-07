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

func (o *Object) Commit() {
	o.Values = o.Values[len(o.Values)-1:]
}

func (o *Object) String() string {
	bytes, _ := json.Marshal(o)
	return string(bytes)
}

// Store represents our in-memory key-value store
type Store struct {
	data map[string]*Object
	mu   sync.RWMutex

	wal *WAL
}

// New creates a new Store instance
func New(walDir string) (*Store, error) {
	store := &Store{
		data: make(map[string]*Object),
	}

	wal, err := NewWAL(walDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create WAL: %w", err)
	}
	if err := wal.Recover(store); err != nil {
		return nil, fmt.Errorf("failed to recover from WAL: %w", err)
	}

	// only set the wal after recover is done to avoid
	// writing new WAL entry during recovery
	store.wal = wal

	return store, nil
}

// Set stores a value for a given key
func (s *Store) Set(key, value string) (*Object, error) {
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; !exists {
		s.data[key] = &Object{
			Values: []Value{},
		}
	}

	object := s.data[key]
	var version int64
	if len(object.Values) == 0 {
		version = 0
	} else {
		version = object.Values[len(object.Values)-1].Version + 1
	}

	// Create new value
	newValue := Value{
		Value:     value,
		Version:   version,
		Timestamp: time.Now(),
	}

	// wal will be nil during recover
	if s.wal != nil {
		// Write to WAL first
		if err := s.wal.Write(WALEntry{
			Ops:       "SET",
			Key:       key,
			Value:     value,
			Version:   version,
			Timestamp: newValue.Timestamp,
		}); err != nil {
			return nil, fmt.Errorf("failed to write to WAL: %w", err)
		}
	}

	// Then apply to memory
	object.Values = append(object.Values, newValue)

	return object, nil
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

	fmt.Println(object)

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
