package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	store := &Store{
		data: make(map[string]string),
	}

	http.HandleFunc("/set", store.handleSet)
	http.HandleFunc("/get", store.handleGet)

	fmt.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
