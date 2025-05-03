package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/leakingtapan/craq/internal/store"
)

type Server struct {
	// the server ID is the node ID within the chain
	Id int

	chainTable *ChainTable

	store *store.Store
}

// create a new server
func New(id int, chainTable *ChainTable) Server {
	return Server{
		Id:         id,
		chainTable: chainTable,
		store:      store.New(),
	}
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

func (svr *Server) HandleSet(w http.ResponseWriter, r *http.Request) {
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

	svr.store.Set(req.Key, req.Value)

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

func (svr *Server) HandleGet(w http.ResponseWriter, r *http.Request) {
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

	value, err := svr.store.Get(req.Key)
	if err != nil {
		log.Printf("failed to get key %s: %v", req.Key, err)
		http.Error(w, fmt.Sprintf("failed to get key %s from store", req.Key), http.StatusInternalServerError)
		return
	}

	resp := ReadResponse{
		Key:   req.Key,
		Value: value,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
