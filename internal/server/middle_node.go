package server

import (
	"net/http"
)

type MiddleNode struct {
}

func NewMiddleNode() *MiddleNode {
	return &MiddleNode{}
}

func (node *MiddleNode) HandleGet(w http.ResponseWriter, r *http.Request) {

}

func (node *MiddleNode) HandlePropagateWrite(w http.ResponseWriter, r *http.Request) {

}
