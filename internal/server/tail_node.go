package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/leakingtapan/craq/internal/store"
)

type TailNode struct {
	Id         int
	chainTable *ChainTable
	store      *store.Store
}

func NewTailNode(id int, chainTable *ChainTable) *TailNode {
	return &TailNode{
		Id:         id,
		chainTable: chainTable,
		store:      store.New(),
	}
}

func (node *TailNode) HandleGet(w http.ResponseWriter, r *http.Request) {
	handlGet(node.store, w, r)
}

func (node *TailNode) HandlePropagateWrite(w http.ResponseWriter, r *http.Request) {
	var req PropagateWriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	object, err := node.store.Set(req.Key, req.Value)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to set %s=%s", req.Key, req.Value), http.StatusInternalServerError)
	}

	log.Printf("commit write %s=%s", req.Key, req.Value)
	err = node.commitWrite(object)
	if err != nil {
		http.Error(w, "failed to commit write", http.StatusInternalServerError)
	}

	resp := PropagateWriteResponse{
		Status: "ok",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&resp)
}

func (node *TailNode) commitWrite(object *store.Object) error {
	object.Commit()
	return nil
}

func (node *TailNode) HandleVersionQuery(w http.ResponseWriter, r *http.Request) {
}
