package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/leakingtapan/craq/internal/store"
	"github.com/spf13/cobra"
)

var (
	port    int
	rootCmd = &cobra.Command{
		Use:   "craq",
		Short: "CRAQ - A key-value server the implement CRAQ",
		Run: func(cmd *cobra.Command, args []string) {
			startServer()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 8080, "port to run the server on")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func startServer() {
	store := store.New()

	http.HandleFunc("/set", store.HandleSet)
	http.HandleFunc("/get", store.HandleGet)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server starting on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
