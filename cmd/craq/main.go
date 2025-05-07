package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/leakingtapan/craq/internal/server"
	"github.com/spf13/cobra"
)

var (
	port           int
	id             int
	chainTablePath string

	rootCmd = &cobra.Command{
		Use:   "craq",
		Short: "CRAQ - A key-value server the implement CRAQ",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startServer()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 8080, "port to run the server on")
	rootCmd.PersistentFlags().IntVar(&id, "id", 0, "port to run the server on")
	rootCmd.PersistentFlags().StringVar(&chainTablePath, "chain-table", "", "the path to chain table")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func startServer() error {
	fmt.Println("1")
	server, err := configureServer()
	if err != nil {
		return err
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Starting server with ID %d on %s...\n", id, server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting servr: %v", err)
	case sig := <-shutdown:
		log.Printf("Got signal: %v", sig)
		log.Printf("server is shutting down...")

		// Give outstanding requests 5 seconds to complete.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown did not complete in 5s: %v", err)
			if err := server.Close(); err != nil {
				log.Fatalf("Could not stop server: %v", err)
			}
		}
	}

	return nil
}

func configureServer() (*http.Server, error) {
	svr := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	chainTable, err := server.ParseChainTable(chainTablePath)
	if err != nil {
		return nil, err
	}
	log.Printf("parsed chain table: \n%+v", chainTable)

	currDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	nodeFilePath := filepath.Join(currDir, "states", fmt.Sprintf("%d", id))

	switch chainTable.Role(id) {
	case server.HEAD:
		log.Printf("creating head node")
		handler, err := server.NewHeadNode(id, chainTable, nodeFilePath)
		if err != nil {
			return nil, err
		}
		http.HandleFunc("/set", handler.HandleSet)
		http.HandleFunc("/get", handler.HandleGet)
	case server.MIDDLE:
		log.Printf("creating middle node")
		handler, err := server.NewMiddleNode(id, chainTable, nodeFilePath)
		if err != nil {
			return nil, err
		}
		http.HandleFunc("/get", handler.HandleGet)
		http.HandleFunc("/propagate", handler.HandlePropagateWrite)
	case server.TAIL:
		log.Printf("creating tail node")
		handler, err := server.NewTailNode(id, chainTable, nodeFilePath)
		if err != nil {
			return nil, err
		}

		http.HandleFunc("/get", handler.HandleGet)
		http.HandleFunc("/propagate", handler.HandlePropagateWrite)
		http.HandleFunc("/version", handler.HandleVersionQuery)
	case server.Unknown:
		log.Fatal("failed to create server, unknown node role")
	}

	return svr, nil
}
