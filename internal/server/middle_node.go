package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/leakingtapan/craq/internal/store"
)

type MiddleNode struct {
	Id         int
	chainTable *ChainTable
	store      *store.Store
}

func NewMiddleNode(
	id int,
	chainTable *ChainTable,
	nodeFilePath string,
) (*MiddleNode, error) {
	store, err := store.New(filepath.Join(nodeFilePath, "wal"))
	if err != nil {
		return nil, err
	}

	return &MiddleNode{
		Id:         id,
		chainTable: chainTable,
		store:      store,
	}, nil
}

func (node *MiddleNode) HandleGet(w http.ResponseWriter, r *http.Request) {
	handlGet(node.store, w, r)
}

type PropagateWriteRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PropagateWriteResponse struct {
	Status string `json:"status"`
}

func (node *MiddleNode) HandlePropagateWrite(w http.ResponseWriter, r *http.Request) {
	var req PropagateWriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	object, err := node.store.Set(req.Key, req.Value)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to set %s=%s", req.Key, req.Value), http.StatusInternalServerError)
	}

	log.Printf("handle propagate write %s=%s", req.Key, req.Value)
	// propagate write
	// wait for the write to be committed
	err = node.propagateWrite(req.Key, req.Value)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	// mark the object as clean (commited) after the propagate is done
	object.Commit()

	resp := PropagateWriteResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (node *MiddleNode) propagateWrite(key, value string) error {
	nextId := node.Id + 1
	nextNode := node.chainTable.Nodes[nextId]
	return propagateWrite(nextNode.Addr, key, value)
}
