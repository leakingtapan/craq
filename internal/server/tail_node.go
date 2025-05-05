package server

import "net/http"

type TailNode struct {
}

func NewTailNode() *TailNode {
	return &TailNode{}
}

func (node *TailNode) HandleGet(w http.ResponseWriter, r *http.Request) {

}

func (node *TailNode) HandlePropagateWrite(w http.ResponseWriter, r *http.Request) {
}

func (node *TailNode) HandleVersionQuery(w http.ResponseWriter, r *http.Request) {
}
