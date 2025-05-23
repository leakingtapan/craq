package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/leakingtapan/craq/internal/store"
)

type HeadNode struct {
	// the server ID is the node ID within the chain
	Id int

	chainTable *ChainTable

	store *store.Store
}

// create a new server
func NewHeadNode(id int, chainTable *ChainTable, nodeFilePath string) (*HeadNode, error) {
	store, err := store.New(filepath.Join(nodeFilePath, "wal"))
	if err != nil {
		return nil, err
	}

	return &HeadNode{
		Id:         id,
		chainTable: chainTable,
		store:      store,
	}, nil
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

func (node *HeadNode) HandleSet(w http.ResponseWriter, r *http.Request) {
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
		return
	}

	var version int64
	object, err := node.store.Get(req.Key)
	if err != nil {
		if errors.Is(err, store.ErrObjectNonExists) {
			version = 0
		} else {
			http.Error(w, "failed to get object", http.StatusBadRequest)
			return
		}
	} else {
		version = object.NextVersion()
	}
	log.Printf("get latest version for %s (v=%d)", req.Key, version)

	log.Printf("handle set %s=%s (v=%d)", req.Key, req.Value, version)
	object, err = node.store.Set(req.Key, req.Value, version)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to set %s=%s", req.Key, req.Value), http.StatusInternalServerError)
		return
	}

	// propagate write
	// wait for the write to be committed
	err = node.propagateWrite(req.Key, object.LatestValue())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to write %s=%s", req.Key, req.Value), http.StatusInternalServerError)
		return
	}

	// mark the object as clean (commited) after the propagate is done
	object.Commit()

	response := WriteResponse{
		Success: true,
		Message: fmt.Sprintf("wrote value for %s=%s", req.Key, req.Value),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func propagateWrite(addr string, key string, value *store.Value) error {
	log.Printf("propagate write to %s for %s=%s", addr, key, value.Value)
	url := fmt.Sprintf("http://%s/propagate", addr)

	propagateReq := PropagateWriteRequest{
		Key:       key,
		Value:     value.Value,
		Version:   value.Version,
		Timestamp: value.Timestamp,
	}

	data, err := json.Marshal(propagateReq)
	if err != nil {
		return fmt.Errorf("failed to marchal propagate request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("propagation failed: %w", err)
	}

	var propagateResp PropagateWriteResponse

	if err := json.NewDecoder(resp.Body).Decode(&propagateResp); err != nil {
		return fmt.Errorf("failed to unmarchal the response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("failed to propagate write %v", propagateResp)
		return fmt.Errorf("propagation failed status: %v", resp.StatusCode)
	}

	if propagateResp.Status != "ok" {
		return fmt.Errorf("propagation failed response: %v", propagateResp.Status)
	}

	return nil
}

func (node *HeadNode) propagateWrite(key string, value *store.Value) error {
	nextId := node.Id + 1
	nextNode := node.chainTable.Nodes[nextId]
	return propagateWrite(nextNode.Addr, key, value)
}

type ReadRequest struct {
	Key string `json:"key"`
}

type ReadResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (node *HeadNode) HandleGet(w http.ResponseWriter, r *http.Request) {
	handlGet(node.store, w, r)
}

func handlGet(store *store.Store, w http.ResponseWriter, r *http.Request) {
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

	log.Printf("handle get key=%s", req.Key)
	object, err := store.Get(req.Key)
	if err != nil {
		log.Printf("failed to get key %s: %v", req.Key, err)
		http.Error(w, fmt.Sprintf("failed to get key %s from store", req.Key), http.StatusInternalServerError)
		return
	}

	if !object.IsDirty() {
		resp := ReadResponse{
			Key:   req.Key,
			Value: object.Values[0].Value,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// TODO:
	// if the object is dirty, do the version query against the tail
	// and look up the value for that version from local store
}
