package server

import (
	"encoding/json"
	"log"
	"net/http"
)

type TailNode struct {
}

func NewTailNode() *TailNode {
	return &TailNode{}
}

func (node *TailNode) HandleGet(w http.ResponseWriter, r *http.Request) {

}

func (node *TailNode) HandlePropagateWrite(w http.ResponseWriter, r *http.Request) {
	var req PropagateWriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("commit write %s=%s", req.Key, req.Value)
	err := node.commitWrite()
	if err != nil {
		http.Error(w, "failed to commit write", http.StatusInternalServerError)
	}

	resp := PropagateWriteResponse{
		Status: "ok",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&resp)
}

func (node *TailNode) commitWrite() error {
	return nil
}

func (node *TailNode) HandleVersionQuery(w http.ResponseWriter, r *http.Request) {
}
