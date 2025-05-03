package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Store represents our in-memory key-value store
type Store struct {
	data map[string]string
	mu   sync.RWMutex
}

// WriteRequest represents the structure of our write request
type WriteRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// WriteResponse represents the response structure
type WriteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (s *Store) handleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
		return
	}

	if req.Value == "" {
		http.Error(w, "Value cannot be empty", http.StatusBadRequest)
	}

	s.mu.Lock()
	s.data[req.Key] = req.Value
	s.mu.Unlock()

	response := WriteResponse{
		Success: true,
		Message: fmt.Sprintf("wrote value for %s=%s", req.Key, req.Value),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type ReadRequest struct {
	Key string `json:"key"`
}

type ReadResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (s *Store) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	var req ReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	value := s.data[req.Key]
	s.mu.Unlock()

	resp := ReadResponse{
		Key:   req.Key,
		Value: value,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
