package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/leakingtapan/craq/internal/server"
	"github.com/spf13/cobra"
)

var (
	port int
	id   int
	//
	chainTablePath string
	rootCmd        = &cobra.Command{
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
	chainTable, err := server.ParseChainTable(chainTablePath)
	if err != nil {
		return err
	}
	svr := server.New(id, chainTable)

	http.HandleFunc("/set", svr.HandleSet)
	http.HandleFunc("/get", svr.HandleGet)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server starting on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		return err
	}
	return nil
}
