package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/leakingtapan/craq/internal/store"
)

type MiddleNode struct {
	Id         int
	chainTable *ChainTable
	store      *store.Store
}

func NewMiddleNode(id int, chainTable *ChainTable) *MiddleNode {
	return &MiddleNode{
		Id:         id,
		chainTable: chainTable,
		store:      store.New(),
	}
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

	node.store.Set(req.Key, req.Value)

	log.Printf("handle propagate write %s=%s", req.Key, req.Value)
	// propagate write
	// wait for the write to be committed
	err := node.propagateWrite(req.Key, req.Value)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

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
